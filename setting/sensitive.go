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
package setting

import (
	"regexp"
	"strings"
)

var CheckSensitiveEnabled = true
var CheckSensitiveOnPromptEnabled = true

var SafeCheckExemptEnabled = false
var SafeCheckExemptGroup = "nsfw-ok"

//var CheckSensitiveOnCompletionEnabled = true

// StopOnSensitiveEnabled nếu phát hiện từ nhạy cảm thì dừng sinh ngay, nếu không thì thay thế từ nhạy cảm
var StopOnSensitiveEnabled = true

// StreamCacheQueueLength độ dài hàng đợi đệm ở chế độ streaming, 0 nghĩa là không đệm
var StreamCacheQueueLength = 0

// SensitiveWords từ bị chặn thông thường
var SensitiveWords = []string{
	"test_sensitive",
}

// RegexSensitiveWords từ bị chặn dạng biểu thức chính quy
var RegexSensitiveWords = []string{}

// Kiểm tra chuỗi có phải biểu thức chính quy hợp lệ không
func isValidRegex(pattern string) bool {
	_, err := regexp.Compile(pattern)
	return err == nil
}

func SensitiveWordsToString() string {
	var builder strings.Builder

	// Add normal words
	for _, word := range SensitiveWords {
		builder.WriteString(word)
		builder.WriteString("\n")
	}

	// Add regex patterns with "regex:" prefix
	for _, pattern := range RegexSensitiveWords {
		builder.WriteString("regex:")
		builder.WriteString(pattern)
		builder.WriteString("\n")
	}

	return builder.String()
}

func SensitiveWordsFromString(s string) {
	SensitiveWords = []string{}
	RegexSensitiveWords = []string{}

	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line starts with "regex:"
		if strings.HasPrefix(line, "regex:") {
			pattern := strings.TrimPrefix(line, "regex:")
			if isValidRegex(pattern) {
				RegexSensitiveWords = append(RegexSensitiveWords, pattern)
			}
		} else {
			SensitiveWords = append(SensitiveWords, line)
		}
	}
}

func ShouldCheckPromptSensitive() bool {
	return CheckSensitiveEnabled && CheckSensitiveOnPromptEnabled
}

func ShouldCheckPromptSensitiveWithGroup(group string) bool {
	if SafeCheckExemptEnabled && group == SafeCheckExemptGroup {
		return false
	}
	return ShouldCheckPromptSensitive()
}

//func ShouldCheckCompletionSensitive() bool {
//	return CheckSensitiveEnabled && CheckSensitiveOnCompletionEnabled
//}
