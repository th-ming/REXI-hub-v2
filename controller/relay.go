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
package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"rexi-hub-v2/common"
	"rexi-hub-v2/dto"
	"rexi-hub-v2/middleware"
	"rexi-hub-v2/model"
	"rexi-hub-v2/relay"
	relaycommon "rexi-hub-v2/relay/common"
	"rexi-hub-v2/relay/constant"
	relayconstant "rexi-hub-v2/relay/constant"
	"rexi-hub-v2/relay/helper"
	"rexi-hub-v2/service"
	"rexi-hub-v2/setting/model_setting"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func relayHandler(c *gin.Context, relayMode int) *dto.OpenAIErrorWithStatusCode {
	var err *dto.OpenAIErrorWithStatusCode
	switch relayMode {
	case relayconstant.RelayModeImagesGenerations:
		err = relay.ImageHelper(c)
	case relayconstant.RelayModeAudioSpeech:
		fallthrough
	case relayconstant.RelayModeAudioTranslation:
		fallthrough
	case relayconstant.RelayModeAudioTranscription:
		err = relay.AudioHelper(c)
	case relayconstant.RelayModeRerank:
		err = relay.RerankHelper(c, relayMode)
	case relayconstant.RelayModeEmbeddings:
		err = relay.EmbeddingHelper(c)
	case relayconstant.RelayModeResponses:
		err = relay.ResponsesHelper(c)
	case relayconstant.RelayModeTokenCount:
		err = relay.TokenCountHelper(c)
	default:
		err = relay.TextHelper(c)
	}

	if err != nil && common.LogErrorEnabled { // If error log is enabled
		// Lưu log lỗi vào mysql
		userId := c.GetInt("id")
		tokenName := c.GetString("token_name")
		modelName := c.GetString("original_model")
		tokenId := c.GetInt("token_id")
		userGroup := c.GetString("group")
		channelId := c.GetInt("channel_id")
		other := make(map[string]interface{})
		other["error_type"] = err.Error.Type
		other["error_code"] = err.Error.Code
		other["status_code"] = err.StatusCode
		other["channel_id"] = channelId
		other["channel_name"] = c.GetString("channel_name")
		other["channel_type"] = c.GetInt("channel_type")

		model.RecordErrorLog(c, userId, channelId, modelName, tokenName, err.Error.Message, tokenId, 0, false, userGroup, other)
	}

	return err
}

func Relay(c *gin.Context) {
	relayMode := constant.Path2RelayMode(c.Request.URL.Path)
	requestId := c.GetString(common.RequestIdKey)
	group := c.GetString("group")
	originalModel := c.GetString("original_model")
	var openaiErr *dto.OpenAIErrorWithStatusCode

	if model_setting.GetGlobalSettings().BlockBrowserExtensionEnabled &&
		relayMode == relayconstant.RelayModeChatCompletions {
		origin := c.Request.Header.Get("Origin")
		if strings.Contains(origin, "extension") && strings.Contains(origin, "://") {
			openaiErr = service.OpenAIErrorWrapperLocal(fmt.Errorf("browser extension blocked"), "browser_extension_blocked", http.StatusForbidden)
			c.JSON(openaiErr.StatusCode, gin.H{"error": openaiErr.Error})
			return
		}
	}

	autoRetryCount := model_setting.GetAutoRetryCount()
	maxRetries := autoRetryCount

	for i := 0; i <= maxRetries; i++ {
		// Nếu bật buộc chuyển kênh và không phải lần thử đầu thì tăng chỉ số thử lại để lấy kênh khác
		retryIndex := i
		if i > 0 && model_setting.ShouldForceChannelSwitch() {
			retryIndex = i + 1 // Buộc lấy kênh khác
		}

		channel, err := getChannel(c, group, originalModel, retryIndex)
		if err != nil {
			common.LogError(c, err.Error())
			openaiErr = service.OpenAIErrorWrapperLocal(err, "get_channel_failed", http.StatusInternalServerError)
			break
		}

		// Ghi lại trạng thái trước khi phản hồi, dùng để phát hiện phản hồi rỗng
		c.Set("response_written", false)
		openaiErr = relayRequest(c, relayMode, channel)

		// Phát hiện trường hợp phản hồi rỗng
		if openaiErr == nil {
			// Kiểm tra xem nội dung phản hồi có thực sự được ghi không
			responseWritten := c.GetBool("response_written")
			if !responseWritten && model_setting.GetGlobalSettings().AutoRetryEnabled {
				// Tạo một lỗi biểu thị phản hồi rỗng để kích hoạt thử lại
				openaiErr = service.OpenAIErrorWrapperLocal(
					fmt.Errorf("empty response from upstream"),
					"empty_response",
					http.StatusInternalServerError,
				)
				common.LogWarn(c, fmt.Sprintf("detected empty response from channel #%d, will retry", channel.Id))
			} else {
				return // Xử lý yêu cầu thành công, trả về luôn
			}
		}

		go processChannelError(c, channel.Id, channel.Type, channel.Name, channel.GetAutoBan(), openaiErr)

		if !shouldRetryWithAutoConfig(c, openaiErr, maxRetries-i) {
			break
		}
	}
	useChannel := c.GetStringSlice("use_channel")
	if len(useChannel) > 1 {
		retryLogStr := fmt.Sprintf("Retry: %s", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(useChannel)), "->"), "[]"))
		common.LogInfo(c, retryLogStr)
	}

	if openaiErr != nil {
		if openaiErr.StatusCode == http.StatusTooManyRequests {
			common.LogError(c, fmt.Sprintf("origin 429 error: %s", openaiErr.Error.Message))
			openaiErr.Error.Message = "Upstream load for this group is saturated, please try again later"
		}
		openaiErr.Error.Message = common.MessageWithRequestId(openaiErr.Error.Message, requestId)
		if c.GetString("relay_format") == relaycommon.RelayFormatGemini {
			geminiErr := service.OpenAIErrorToGeminiResponse(openaiErr)
			c.JSON(openaiErr.StatusCode, geminiErr)
		} else {
			c.JSON(openaiErr.StatusCode, gin.H{
				"error": openaiErr.Error,
			})
		}
	}
}

var upgrader = websocket.Upgrader{
	Subprotocols: []string{"realtime"}, // Giao thức mà WS handshake hỗ trợ; nếu dùng Sec-WebSocket-Protocol thì phải khai báo Protocol tương ứng ở đây. TODO add other protocol
	CheckOrigin: func(r *http.Request) bool {
		return true // Cho phép cross-origin
	},
}

func WssRelay(c *gin.Context) {
	// Nâng cấp kết nối HTTP thành kết nối WebSocket

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	defer ws.Close()

	if err != nil {
		openaiErr := service.OpenAIErrorWrapper(err, "get_channel_failed", http.StatusInternalServerError)
		helper.WssError(c, ws, openaiErr.Error)
		return
	}

	relayMode := constant.Path2RelayMode(c.Request.URL.Path)
	requestId := c.GetString(common.RequestIdKey)
	group := c.GetString("group")
	//wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview-2024-10-01
	originalModel := c.GetString("original_model")
	var openaiErr *dto.OpenAIErrorWithStatusCode

	for i := 0; i <= common.RetryTimes; i++ {
		channel, err := getChannel(c, group, originalModel, i)
		if err != nil {
			common.LogError(c, err.Error())
			openaiErr = service.OpenAIErrorWrapperLocal(err, "get_channel_failed", http.StatusInternalServerError)
			break
		}

		openaiErr = wssRequest(c, ws, relayMode, channel)

		if openaiErr == nil {
			return // Xử lý yêu cầu thành công, trả về luôn
		}

		go processChannelError(c, channel.Id, channel.Type, channel.Name, channel.GetAutoBan(), openaiErr)

		if !shouldRetry(c, openaiErr, common.RetryTimes-i) {
			break
		}
	}
	useChannel := c.GetStringSlice("use_channel")
	if len(useChannel) > 1 {
		retryLogStr := fmt.Sprintf("Retry: %s", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(useChannel)), "->"), "[]"))
		common.LogInfo(c, retryLogStr)
	}

	if openaiErr != nil {
		if openaiErr.StatusCode == http.StatusTooManyRequests {
			openaiErr.Error.Message = "Upstream load for this group is saturated, please try again later"
		}
		openaiErr.Error.Message = common.MessageWithRequestId(openaiErr.Error.Message, requestId)
		helper.WssError(c, ws, openaiErr.Error)
	}
}

func RelayClaude(c *gin.Context) {
	//relayMode := constant.Path2RelayMode(c.Request.URL.Path)
	requestId := c.GetString(common.RequestIdKey)
	group := c.GetString("group")
	originalModel := c.GetString("original_model")
	var claudeErr *dto.ClaudeErrorWithStatusCode

	for i := 0; i <= common.RetryTimes; i++ {
		channel, err := getChannel(c, group, originalModel, i)
		if err != nil {
			common.LogError(c, err.Error())
			claudeErr = service.ClaudeErrorWrapperLocal(err, "get_channel_failed", http.StatusInternalServerError)
			break
		}

		claudeErr = claudeRequest(c, channel)

		if claudeErr == nil {
			return // Xử lý yêu cầu thành công, trả về luôn
		}

		openaiErr := service.ClaudeErrorToOpenAIError(claudeErr)

		go processChannelError(c, channel.Id, channel.Type, channel.Name, channel.GetAutoBan(), openaiErr)

		if !shouldRetry(c, openaiErr, common.RetryTimes-i) {
			break
		}
	}
	useChannel := c.GetStringSlice("use_channel")
	if len(useChannel) > 1 {
		retryLogStr := fmt.Sprintf("Retry: %s", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(useChannel)), "->"), "[]"))
		common.LogInfo(c, retryLogStr)
	}

	if claudeErr != nil {
		claudeErr.Error.Message = common.MessageWithRequestId(claudeErr.Error.Message, requestId)
		c.JSON(claudeErr.StatusCode, gin.H{
			"type":  "error",
			"error": claudeErr.Error,
		})
	}
}

func relayRequest(c *gin.Context, relayMode int, channel *model.Channel) *dto.OpenAIErrorWithStatusCode {
	addUsedChannel(c, channel.Id)
	requestBody, _ := common.GetRequestBody(c)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return relayHandler(c, relayMode)
}

func wssRequest(c *gin.Context, ws *websocket.Conn, relayMode int, channel *model.Channel) *dto.OpenAIErrorWithStatusCode {
	addUsedChannel(c, channel.Id)
	requestBody, _ := common.GetRequestBody(c)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return relay.WssHelper(c, ws)
}

func claudeRequest(c *gin.Context, channel *model.Channel) *dto.ClaudeErrorWithStatusCode {
	addUsedChannel(c, channel.Id)
	requestBody, _ := common.GetRequestBody(c)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return relay.ClaudeHelper(c)
}

func addUsedChannel(c *gin.Context, channelId int) {
	useChannel := c.GetStringSlice("use_channel")
	useChannel = append(useChannel, fmt.Sprintf("%d", channelId))
	c.Set("use_channel", useChannel)
}

func getChannel(c *gin.Context, group, originalModel string, retryCount int) (*model.Channel, error) {
	if retryCount == 0 {
		autoBan := c.GetBool("auto_ban")
		autoBanInt := 1
		if !autoBan {
			autoBanInt = 0
		}
		return &model.Channel{
			Id:      c.GetInt("channel_id"),
			Type:    c.GetInt("channel_type"),
			Name:    c.GetString("channel_name"),
			AutoBan: &autoBanInt,
		}, nil
	}
	channel, err := model.CacheGetRandomSatisfiedChannel(group, originalModel, retryCount)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get retry channel: %s", err.Error()))
	}
	middleware.SetupContextForSelectedChannel(c, channel, originalModel)
	return channel, nil
}

func shouldRetry(c *gin.Context, openaiErr *dto.OpenAIErrorWithStatusCode, retryTimes int) bool {
	if openaiErr == nil {
		return false
	}
	if openaiErr.LocalError {
		return false
	}
	if retryTimes <= 0 {
		return false
	}
	if _, ok := c.Get("specific_channel_id"); ok {
		return false
	}
	if openaiErr.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if openaiErr.StatusCode == 307 {
		return true
	}
	if openaiErr.StatusCode/100 == 5 {
		// Hết thời gian chờ thì không thử lại
		if openaiErr.StatusCode == 504 || openaiErr.StatusCode == 524 {
			return false
		}
		return true
	}
	if openaiErr.StatusCode == http.StatusBadRequest {
		channelType := c.GetInt("channel_type")
		if channelType == common.ChannelTypeAnthropic {
			return true
		}
		return false
	}
	if openaiErr.StatusCode == 408 {
		// azure xử lý timeout thì không thử lại
		return false
	}
	if openaiErr.StatusCode/100 == 2 {
		return false
	}
	return true
}

// shouldRetryWithAutoConfig dùng cấu hình thử lại tự động để xác định có nên thử lại
func shouldRetryWithAutoConfig(c *gin.Context, openaiErr *dto.OpenAIErrorWithStatusCode, retryTimes int) bool {
	if openaiErr == nil {
		return false
	}
	if openaiErr.LocalError {
		return false
	}
	if retryTimes <= 0 {
		return false
	}
	if _, ok := c.Get("specific_channel_id"); ok {
		return false
	}

	// Kiểm tra xem có bật tự động thử lại không
	if !model_setting.GetGlobalSettings().AutoRetryEnabled {
		return shouldRetry(c, openaiErr, retryTimes) // Quay lại logic ban đầu
	}

	// Dùng mã trạng thái đã cấu hình để xác định có thử lại hay không
	if !model_setting.ShouldRetryForStatusCode(openaiErr.StatusCode) {
		return false
	}

	// Với trường hợp phản hồi rỗng thì cũng nên thử lại
	if openaiErr.Error.Message == "" || strings.Contains(strings.ToLower(openaiErr.Error.Message), "empty") {
		return true
	}

	// Hết thời gian chờ thì không thử lại
	if openaiErr.StatusCode == 504 || openaiErr.StatusCode == 524 || openaiErr.StatusCode == 408 {
		return false
	}

	// Mã trạng thái thành công thì không thử lại
	if openaiErr.StatusCode/100 == 2 {
		return false
	}

	return true
}

func processChannelError(c *gin.Context, channelId int, channelType int, channelName string, autoBan bool, err *dto.OpenAIErrorWithStatusCode) {
	// Không dùng context để lấy thông tin kênh, khi xử lý bất đồng bộ có thể xảy ra tình trạng thông tin kênh không nhất quán
	// do not use context to get channel info, there may be inconsistent channel info when processing asynchronously
	common.LogError(c, fmt.Sprintf("relay error (channel #%d, status code: %d): %s", channelId, err.StatusCode, err.Error.Message))
	if service.ShouldDisableChannel(channelType, err) && autoBan {
		service.DisableChannel(channelId, channelName, err.Error.Message)
	}
}

func RelayMidjourney(c *gin.Context) {
	relayMode := c.GetInt("relay_mode")
	var err *dto.MidjourneyResponse
	switch relayMode {
	case relayconstant.RelayModeMidjourneyNotify:
		err = relay.RelayMidjourneyNotify(c)
	case relayconstant.RelayModeMidjourneyTaskFetch, relayconstant.RelayModeMidjourneyTaskFetchByCondition:
		err = relay.RelayMidjourneyTask(c, relayMode)
	case relayconstant.RelayModeMidjourneyTaskImageSeed:
		err = relay.RelayMidjourneyTaskImageSeed(c)
	case relayconstant.RelayModeSwapFace:
		err = relay.RelaySwapFace(c)
	default:
		err = relay.RelayMidjourneySubmit(c, relayMode)
	}
	//err = relayMidjourneySubmit(c, relayMode)
	log.Println(err)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Code == 30 {
			err.Result = "Group load is saturated, please try again later, or upgrade your account for better service quality."
			statusCode = http.StatusTooManyRequests
		}
		c.JSON(statusCode, gin.H{
			"description": fmt.Sprintf("%s %s", err.Description, err.Result),
			"type":        "upstream_error",
			"code":        err.Code,
		})
		channelId := c.GetInt("channel_id")
		common.LogError(c, fmt.Sprintf("relay error (channel #%d, status code %d): %s", channelId, statusCode, fmt.Sprintf("%s %s", err.Description, err.Result)))
	}
}

func RelayNotImplemented(c *gin.Context) {
	err := dto.OpenAIError{
		Message: "API not implemented",
		Type:    "rexi_hub_v2_error",
		Param:   "",
		Code:    "api_not_implemented",
	}
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": err,
	})
}

func RelayNotFound(c *gin.Context) {
	err := dto.OpenAIError{
		Message: fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path),
		Type:    "invalid_request_error",
		Param:   "",
		Code:    "",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}

func RelayTokenCount(c *gin.Context) {
	requestId := c.GetString(common.RequestIdKey)

	// Get request body and parse it
	requestBody, err := common.GetRequestBody(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "Failed to read request body",
			},
		})
		return
	}

	// Parse Claude request
	var claudeRequest dto.ClaudeRequest
	err = json.Unmarshal(requestBody, &claudeRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "Invalid request format",
			},
		})
		return
	}

	// Count tokens locally without needing channels
	tokenCount, err := service.CountTokenClaudeRequest(claudeRequest, claudeRequest.Model)
	if err != nil {
		common.LogError(c, fmt.Sprintf("Token counting failed: %s", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "internal_server_error",
				"message": common.MessageWithRequestId("Token counting failed", requestId),
			},
		})
		return
	}

	// Return token count in Claude API format
	c.JSON(http.StatusOK, gin.H{
		"input_tokens": tokenCount,
	})
}

func RelayTask(c *gin.Context) {
	retryTimes := model_setting.GetAutoRetryCount()
	channelId := c.GetInt("channel_id")
	relayMode := c.GetInt("relay_mode")
	group := c.GetString("group")
	originalModel := c.GetString("original_model")
	c.Set("use_channel", []string{fmt.Sprintf("%d", channelId)})
	taskErr := taskRelayHandler(c, relayMode)
	if taskErr == nil {
		retryTimes = 0
	}
	for i := 0; shouldRetryTaskRelay(c, channelId, taskErr, retryTimes) && i < retryTimes; i++ {
		channel, err := model.CacheGetRandomSatisfiedChannel(group, originalModel, i)
		if err != nil {
			common.LogError(c, fmt.Sprintf("CacheGetRandomSatisfiedChannel failed: %s", err.Error()))
			break
		}
		channelId = channel.Id
		useChannel := c.GetStringSlice("use_channel")
		useChannel = append(useChannel, fmt.Sprintf("%d", channelId))
		c.Set("use_channel", useChannel)
		common.LogInfo(c, fmt.Sprintf("using channel #%d to retry (remain times %d)", channel.Id, i))
		middleware.SetupContextForSelectedChannel(c, channel, originalModel)

		requestBody, err := common.GetRequestBody(c)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		taskErr = taskRelayHandler(c, relayMode)
	}
	useChannel := c.GetStringSlice("use_channel")
	if len(useChannel) > 1 {
		retryLogStr := fmt.Sprintf("Retry: %s", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(useChannel)), "->"), "[]"))
		common.LogInfo(c, retryLogStr)
	}
	if taskErr != nil {
		if taskErr.StatusCode == http.StatusTooManyRequests {
			taskErr.Message = "Upstream load for this group is saturated, please try again later"
		}
		c.JSON(taskErr.StatusCode, taskErr)
	}
}

func taskRelayHandler(c *gin.Context, relayMode int) *dto.TaskError {
	var err *dto.TaskError
	switch relayMode {
	case relayconstant.RelayModeSunoFetch, relayconstant.RelayModeSunoFetchByID:
		err = relay.RelayTaskFetch(c, relayMode)
	default:
		err = relay.RelayTaskSubmit(c, relayMode)
	}
	return err
}

func shouldRetryTaskRelay(c *gin.Context, channelId int, taskErr *dto.TaskError, retryTimes int) bool {
	if taskErr == nil {
		return false
	}
	if retryTimes <= 0 {
		return false
	}
	if _, ok := c.Get("specific_channel_id"); ok {
		return false
	}
	if taskErr.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if taskErr.StatusCode == 307 {
		return true
	}
	if taskErr.StatusCode/100 == 5 {
		// Hết thời gian chờ thì không thử lại
		if taskErr.StatusCode == 504 || taskErr.StatusCode == 524 {
			return false
		}
		return true
	}
	if taskErr.StatusCode == http.StatusBadRequest {
		return false
	}
	if taskErr.StatusCode == 408 {
		// azure xử lý timeout thì không thử lại
		return false
	}
	if taskErr.LocalError {
		return false
	}
	if taskErr.StatusCode/100 == 2 {
		return false
	}
	return true
}
