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
package constant

var (
	UserSettingNotifyType            = "notify_type"                    // QuotaWarningType loại cảnh báo hạn mức
	UserSettingQuotaWarningThreshold = "quota_warning_threshold"        // QuotaWarningThreshold ngưỡng cảnh báo hạn mức
	UserSettingWebhookUrl            = "webhook_url"                    // WebhookUrl địa chỉ webhook
	UserSettingWebhookSecret         = "webhook_secret"                 // WebhookSecret khóa webhook
	UserSettingNotificationEmail     = "notification_email"             // NotificationEmail địa chỉ email thông báo
	UserAcceptUnsetRatioModel        = "accept_unset_model_ratio_model" // AcceptUnsetRatioModel có chấp nhận mô hình chưa đặt Price hay không
	UserSettingShowIPInLogs          = "show_ip_in_logs"                // ShowIPInLogs có hiển thị IP trong log tiêu dùng hay không
)

var (
	NotifyTypeEmail   = "email"   // Email thư điện tử
	NotifyTypeWebhook = "webhook" // Webhook
)
