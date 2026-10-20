// Copyright (c) 2025 REXI-hub team
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
package model_setting

import (
	"strconv"
	"strings"
	"rexi-hub-v2/setting/config"
)

type GlobalSettings struct {
	PassThroughRequestEnabled    bool   `json:"pass_through_request_enabled"`
	HideUpstreamErrorEnabled     bool   `json:"hide_upstream_error_enabled"`
	BlockBrowserExtensionEnabled bool   `json:"block_browser_extension_enabled"`
	RateLimitExemptEnabled       bool   `json:"rate_limit_exempt_enabled"`
	RateLimitExemptGroup         string `json:"rate_limit_exempt_group"`
	SafeCheckExemptEnabled       bool   `json:"safe_check_exempt_enabled"`
	SafeCheckExemptGroup         string `json:"safe_check_exempt_group"`
	AutoRetryEnabled             bool   `json:"auto_retry_enabled"`
	AutoRetryCount               int    `json:"auto_retry_count"`
	AutoRetryForceChannelSwitch  bool   `json:"auto_retry_force_channel_switch"`
	AutoRetryStatusCodes         string `json:"auto_retry_status_codes"`
}

// Cấu hình mặc định
var defaultOpenaiSettings = GlobalSettings{
	PassThroughRequestEnabled:    false,
	HideUpstreamErrorEnabled:     false,
	BlockBrowserExtensionEnabled: false,
	RateLimitExemptEnabled:       false,
	RateLimitExemptGroup:         "bulk-ok",
	SafeCheckExemptEnabled:       false,
	SafeCheckExemptGroup:         "nsfw-ok",
	AutoRetryEnabled:             false,
	AutoRetryCount:               3,
	AutoRetryForceChannelSwitch:  false,
	AutoRetryStatusCodes:         "5xx,4xx",
}

// Thực thể toàn cục
var globalSettings = defaultOpenaiSettings

func init() {
	// Đăng ký vào trình quản lý cấu hình toàn cục
	config.GlobalConfig.Register("global", &globalSettings)
}

func GetGlobalSettings() *GlobalSettings {
	return &globalSettings
}

func ShouldBypassRateLimit(group string) bool {
	return globalSettings.RateLimitExemptEnabled && group == globalSettings.RateLimitExemptGroup
}

func ShouldBypassSafeCheck(group string) bool {
	return globalSettings.SafeCheckExemptEnabled && group == globalSettings.SafeCheckExemptGroup
}

// GetAutoRetryCount lấy số lần tự động thử lại
func GetAutoRetryCount() int {
	if !globalSettings.AutoRetryEnabled {
		return 0
	}
	return globalSettings.AutoRetryCount
}

// ShouldForceChannelSwitch có buộc chuyển kênh hay không
func ShouldForceChannelSwitch() bool {
	return globalSettings.AutoRetryEnabled && globalSettings.AutoRetryForceChannelSwitch
}

// ShouldRetryForStatusCode kiểm tra mã trạng thái có nên thử lại hay không
func ShouldRetryForStatusCode(statusCode int) bool {
	if !globalSettings.AutoRetryEnabled {
		return false
	}

	if globalSettings.AutoRetryStatusCodes == "" {
		return true // Mặc định mọi mã trạng thái đều thử lại
	}

	codes := strings.Split(globalSettings.AutoRetryStatusCodes, ",")
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}

		// Xử lý các ký tự đại diện như 5xx, 4xx
		if strings.HasSuffix(code, "xx") {
			prefix := strings.TrimSuffix(code, "xx")
			if prefixNum, err := strconv.Atoi(prefix); err == nil {
				if statusCode/100 == prefixNum {
					return true
				}
			}
		} else {
			// Xử lý mã trạng thái cụ thể
			if codeNum, err := strconv.Atoi(code); err == nil {
				if statusCode == codeNum {
					return true
				}
			}
		}
	}

	return false
}
