// Copyright (c) 2025 Tethys Plex
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
package model

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
	"rexi-hub-v2/common"

	"gorm.io/gorm"
)

type Redemption struct {
	Id           int            `json:"id"`
	UserId       int            `json:"user_id"`
	Key          string         `json:"key" gorm:"type:char(32);uniqueIndex"`
	Status       int            `json:"status" gorm:"default:1"`
	Name         string         `json:"name" gorm:"index"`
	Quota        int            `json:"quota" gorm:"default:100"`
	CreatedTime  int64          `json:"created_time" gorm:"bigint"`
	RedeemedTime int64          `json:"redeemed_time" gorm:"bigint"`
	ValidFrom    int64          `json:"valid_from" gorm:"bigint;default:0"`  // Effective time, 0 means immediate effect
	ValidUntil   int64          `json:"valid_until" gorm:"bigint;default:0"` // Expiry time, 0 means never expires
	Count        int            `json:"count" gorm:"-:all"`                  // only for api request
	UsedUserId   int            `json:"used_user_id"`
	IsGift       bool           `json:"is_gift" gorm:"default:false"`
	MaxUses      int            `json:"max_uses" gorm:"default:-1"` // -1 means unlimited
	UsedCount    int            `json:"used_count" gorm:"default:0"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// RedemptionLog ghi lại lịch sử dùng mã quà tặng
type RedemptionLog struct {
	Id           int   `json:"id"`
	RedemptionId int   `json:"redemption_id" gorm:"index"`
	UserId       int   `json:"user_id" gorm:"index"`
	UsedTime     int64 `json:"used_time" gorm:"bigint"`
}

func GetAllRedemptions(startIdx int, num int) (redemptions []*Redemption, total int64, err error) {
	// Bắt đầu giao dịch
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Lấy tổng số
	err = tx.Model(&Redemption{}).Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Lấy dữ liệu phân trang
	err = tx.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Xác nhận giao dịch
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return redemptions, total, nil
}

func SearchRedemptions(keyword string, startIdx int, num int) (redemptions []*Redemption, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Build query based on keyword type
	query := tx.Model(&Redemption{})

	// Only try to convert to ID if the string represents a valid integer
	if id, err := strconv.Atoi(keyword); err == nil {
		query = query.Where("id = ? OR name LIKE ?", id, keyword+"%")
	} else {
		query = query.Where("name LIKE ?", keyword+"%")
	}

	// Get total count
	err = query.Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Get paginated data
	err = query.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return redemptions, total, nil
}

func GetRedemptionById(id int) (*Redemption, error) {
	if id == 0 {
		return nil, errors.New("id is empty!")
	}
	redemption := Redemption{Id: id}
	var err error = nil
	err = DB.First(&redemption, "id = ?", id).Error
	return &redemption, err
}

// GetRedemptionByKey lấy mã đổi thưởng theo nội dung mã
func GetRedemptionByKey(key string) (*Redemption, error) {
	if key == "" {
		return nil, errors.New("Redemption code content is empty!")
	}
	redemption := Redemption{}
	var err error = nil
	err = DB.First(&redemption, "`key` = ?", key).Error
	return &redemption, err
}

func Redeem(key string, userId int) (quota int, isGift bool, err error) {
	if key == "" {
		return 0, false, errors.New("Redemption code not provided")
	}
	if userId == 0 {
		return 0, false, errors.New("Invalid user id")
	}
	redemption := &Redemption{}

	keyCol := "`key`"
	if common.UsingPostgreSQL {
		keyCol = `"key"`
	}

	if common.RedisEnabled && common.RDB != nil {
		lockKey := "rloc:" + key // Using a shorter prefix "rloc:" for brevity
		locked, redisErr := common.RDB.SetNX(context.Background(), lockKey, "1", 10*time.Second).Result()
		if redisErr != nil {
			common.SysError("Redis lock acquisition error: " + redisErr.Error())
			return 0, false, errors.New("The system is temporarily busy, please try again later") // System temporarily busy, please try again later
		}
		if !locked {
			return 0, false, errors.New("Too many operations, please try again later") // Operation too frequent, please try again later
		}
		defer common.RDB.Del(context.Background(), lockKey)
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where(keyCol+" = ?", key).First(redemption).Error
		if err != nil {
			return errors.New("Invalid redemption code")
		}

		// Kiểm tra khoảng thời gian của mã đổi thưởng
		currentTime := common.GetTimestamp()
		if redemption.ValidFrom > 0 && currentTime < redemption.ValidFrom {
			return errors.New("Redemption code is not yet effective")
		}
		if redemption.ValidUntil > 0 && currentTime > redemption.ValidUntil {
			return errors.New("Redemption code has expired")
		}

		if !redemption.IsGift {
			// Logic mã đổi thưởng thông thường
			if redemption.Status != common.RedemptionCodeStatusEnabled {
				return errors.New("This redemption code has already been used")
			}
			err = tx.Model(&User{}).Where("id = ?", userId).Update("quota", gorm.Expr("quota + ?", redemption.Quota)).Error
			if err != nil {
				return err
			}
			redemption.RedeemedTime = common.GetTimestamp()
			redemption.Status = common.RedemptionCodeStatusUsed
			redemption.UsedUserId = userId
		} else {
			// Logic mã quà tặng
			if redemption.MaxUses != -1 && redemption.UsedCount >= redemption.MaxUses {
				return errors.New("This gift code has reached its maximum usage count")
			}
			// Kiểm tra người dùng đã dùng mã quà tặng này chưa
			var usageCount int64
			err = tx.Model(&RedemptionLog{}).Where("redemption_id = ? AND user_id = ?", redemption.Id, userId).Count(&usageCount).Error
			if err != nil {
				return err
			}
			if usageCount > 0 {
				return errors.New("You have already used this gift code")
			}

			err = tx.Model(&User{}).Where("id = ?", userId).Update("quota", gorm.Expr("quota + ?", redemption.Quota)).Error
			if err != nil {
				return err
			}

			// Ghi log sử dụng
			log := RedemptionLog{
				RedemptionId: redemption.Id,
				UserId:       userId,
				UsedTime:     common.GetTimestamp(),
			}
			if err = tx.Create(&log).Error; err != nil {
				return err
			}

			redemption.UsedCount++
			if redemption.MaxUses != -1 && redemption.UsedCount >= redemption.MaxUses {
				redemption.Status = common.RedemptionCodeStatusUsed
			}
		}

		err = tx.Save(redemption).Error
		return err
	})
	if err != nil {
		return 0, false, errors.New("Redemption failed, " + err.Error())
	}
	RecordLog(userId, LogTypeTopup, fmt.Sprintf("Top-up %s via %s, redemption code ID %d",
		map[bool]string{true: "Gift code", false: "Redemption code"}[redemption.IsGift],
		common.LogQuota(redemption.Quota),
		redemption.Id))

	return redemption.Quota, redemption.IsGift, nil
}

func (redemption *Redemption) Insert() error {
	var err error
	err = DB.Create(redemption).Error
	return err
}

func (redemption *Redemption) SelectUpdate() error {
	// This can update zero values
	return DB.Model(redemption).Select("redeemed_time", "status").Updates(redemption).Error
}

// Update Make sure your token's fields is completed, because this will update non-zero values
func (redemption *Redemption) Update() error {
	var err error
	err = DB.Model(redemption).Select("name", "status", "quota", "redeemed_time", "valid_from", "valid_until").Updates(redemption).Error
	return err
}

func (redemption *Redemption) Delete() error {
	var err error
	err = DB.Delete(redemption).Error
	return err
}

func DeleteRedemptionById(id int) (err error) {
	if id == 0 {
		return errors.New("id is empty!")
	}
	redemption := Redemption{Id: id}
	err = DB.Where(redemption).First(&redemption).Error
	if err != nil {
		return err
	}
	return redemption.Delete()
}

// CountRedemptionsByName đếm số lượng mã đổi thưởng theo tên
func CountRedemptionsByName(name string) (count int64, err error) {
	if name == "" {
		return 0, errors.New("Name cannot be empty")
	}
	err = DB.Model(&Redemption{}).Where("name = ?", name).Count(&count).Error
	return count, err
}

// DeleteRedemptionsByName xóa hàng loạt mã đổi thưởng theo tên
func DeleteRedemptionsByName(name string) (count int64, err error) {
	if name == "" {
		return 0, errors.New("Name cannot be empty")
	}

	// Tính số lượng trước
	count, err = CountRedemptionsByName(name)
	if err != nil {
		return 0, err
	}

	// Không tìm thấy mã đổi thưởng khớp
	if count == 0 {
		return 0, nil
	}

	// Thực hiện xóa
	result := DB.Where("name = ?", name).Delete(&Redemption{})
	return result.RowsAffected, result.Error
}

// BatchDisableRedemptions vô hiệu hóa hàng loạt mã đổi thưởng theo ID chỉ định
func BatchDisableRedemptions(ids []int) (count int64, err error) {
	if len(ids) == 0 {
		return 0, errors.New("ID list cannot be empty")
	}

	result := DB.Model(&Redemption{}).Where("id IN ?", ids).Update("status", common.RedemptionCodeStatusDisabled)
	return result.RowsAffected, result.Error
}

// DeleteDisabledRedemptions xóa toàn bộ mã đổi thưởng đã bị vô hiệu hóa
func DeleteDisabledRedemptions() (count int64, err error) {
	// Delete both disabled and already used redemption codes
	result := DB.Where("status IN ?", []int{common.RedemptionCodeStatusDisabled, common.RedemptionCodeStatusUsed}).Delete(&Redemption{})
	return result.RowsAffected, result.Error
}
