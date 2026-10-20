// Copyright (c) 2026 REXI-hub team
//
// This file is part of REXI-hub v2.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"rexi-hub-v2/common"
	"rexi-hub-v2/setting"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	ModelRequestRateLimitCountMark        = "MRRL"
	ModelRequestRateLimitSuccessCountMark = "MRRLS"
)

// Kiểm tra giới hạn yêu cầu trong Redis
func checkRedisRateLimit(ctx context.Context, rdb *redis.Client, key string, maxCount int, duration int64) (bool, error) {
	// Nếu maxCount bằng 0 nghĩa là không giới hạn
	if maxCount == 0 {
		return true, nil
	}

	// Lấy số đếm hiện tại
	length, err := rdb.LLen(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Nếu chưa đạt giới hạn thì cho phép yêu cầu
	if length < int64(maxCount) {
		return true, nil
	}

	// Kiểm tra cửa sổ thời gian
	oldTimeStr, _ := rdb.LIndex(ctx, key, -1).Result()
	oldTime, err := time.Parse(timeFormat, oldTimeStr)
	if err != nil {
		return false, err
	}

	nowTimeStr := time.Now().Format(timeFormat)
	nowTime, err := time.Parse(timeFormat, nowTimeStr)
	if err != nil {
		return false, err
	}
	// Nếu đã đạt giới hạn trong cửa sổ thời gian thì từ chối yêu cầu
	subTime := nowTime.Sub(oldTime).Seconds()
	if int64(subTime) < duration {
		rdb.Expire(ctx, key, time.Duration(setting.ModelRequestRateLimitDurationMinutes)*time.Minute)
		return false, nil
	}

	return true, nil
}

// Ghi nhận yêu cầu Redis
func recordRedisRequest(ctx context.Context, rdb *redis.Client, key string, maxCount int) {
	// Nếu maxCount bằng 0 thì không ghi nhận yêu cầu
	if maxCount == 0 {
		return
	}

	now := time.Now().Format(timeFormat)
	rdb.LPush(ctx, key, now)
	rdb.LTrim(ctx, key, 0, int64(maxCount-1))
	rdb.Expire(ctx, key, time.Duration(setting.ModelRequestRateLimitDurationMinutes)*time.Minute)
}

// Bộ xử lý giới hạn tốc độ Redis
func redisRateLimitHandler(duration int64, totalMaxCount, successMaxCount int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := strconv.Itoa(c.GetInt("id"))
		ctx := context.Background()
		rdb := common.RDB

		// 1. Kiểm tra giới hạn tổng số yêu cầu (tự động bỏ qua khi totalMaxCount bằng 0)
		totalKey := fmt.Sprintf("rateLimit:%s:%s", ModelRequestRateLimitCountMark, userId)
		allowed, err := checkRedisRateLimit(ctx, rdb, totalKey, totalMaxCount, duration)
		if err != nil {
			fmt.Println("Failed to check total request limit:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
			return
		}
		if !allowed {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("You have reached the total request limit: max %d requests per %d minutes, including failures. Please check your request", setting.ModelRequestRateLimitDurationMinutes, totalMaxCount))
		}

		// 2. Kiểm tra giới hạn số yêu cầu thành công
		successKey := fmt.Sprintf("rateLimit:%s:%s", ModelRequestRateLimitSuccessCountMark, userId)
		allowed, err = checkRedisRateLimit(ctx, rdb, successKey, successMaxCount, duration)
		if err != nil {
			fmt.Println("Failed to check successful request limit:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
			return
		}
		if !allowed {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("You have reached the request limit: max %d requests per %d minutes", setting.ModelRequestRateLimitDurationMinutes, successMaxCount))
			return
		}

		// 3. Ghi nhận tổng yêu cầu (tự động bỏ qua khi totalMaxCount bằng 0)
		recordRedisRequest(ctx, rdb, totalKey, totalMaxCount)

		// 4. Xử lý yêu cầu
		c.Next()

		// 5. Nếu yêu cầu thành công thì ghi nhận yêu cầu thành công
		if c.Writer.Status() < 400 {
			recordRedisRequest(ctx, rdb, successKey, successMaxCount)
		}
	}
}

// Bộ xử lý giới hạn tốc độ trong bộ nhớ
func memoryRateLimitHandler(duration int64, totalMaxCount, successMaxCount int) gin.HandlerFunc {
	inMemoryRateLimiter.Init(time.Duration(setting.ModelRequestRateLimitDurationMinutes) * time.Minute)

	return func(c *gin.Context) {
		userId := strconv.Itoa(c.GetInt("id"))
		totalKey := ModelRequestRateLimitCountMark + userId
		successKey := ModelRequestRateLimitSuccessCountMark + userId

		// 1. Kiểm tra giới hạn tổng số yêu cầu (bỏ qua khi totalMaxCount bằng 0)
		if totalMaxCount > 0 && !inMemoryRateLimiter.Request(totalKey, totalMaxCount, duration) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}

		// 2. Kiểm tra giới hạn số yêu cầu thành công
		// Dùng một key tạm để kiểm tra giới hạn, nhờ đó tránh ghi nhận thật
		checkKey := successKey + "_check"
		if !inMemoryRateLimiter.Request(checkKey, successMaxCount, duration) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}

		// 3. Xử lý yêu cầu
		c.Next()

		// 4. Nếu yêu cầu thành công thì ghi vào bộ đếm số yêu cầu thành công thực tế
		if c.Writer.Status() < 400 {
			inMemoryRateLimiter.Request(successKey, successMaxCount, duration)
		}
	}
}

// ModelRequestRateLimit middleware giới hạn tốc độ yêu cầu mô hình
func ModelRequestRateLimit() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Kiểm tra xem có bật giới hạn tốc độ ở mỗi yêu cầu
		if !setting.ModelRequestRateLimitEnabled {
			c.Next()
			return
		}

		// Tính tham số giới hạn tốc độ
		duration := int64(setting.ModelRequestRateLimitDurationMinutes * 60)
		totalMaxCount := setting.ModelRequestRateLimitCount
		successMaxCount := setting.ModelRequestRateLimitSuccessCount

		// Chọn và thực thi bộ xử lý giới hạn tốc độ theo loại lưu trữ
		if common.RedisEnabled {
			redisRateLimitHandler(duration, totalMaxCount, successMaxCount)(c)
		} else {
			memoryRateLimitHandler(duration, totalMaxCount, successMaxCount)(c)
		}
	}
}
