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
package service

import (
	"bytes"
	"fmt"
	goahocorasick "github.com/anknown/ahocorasick"
	"strings"
)

func SundaySearch(text string, pattern string) bool {
	// Tính bảng dịch
	offset := make(map[rune]int)
	for i, c := range pattern {
		offset[c] = len(pattern) - i
	}

	// Độ dài chuỗi văn bản và độ dài chuỗi mẫu
	n, m := len(text), len(pattern)

	// Vòng lặp chính, i là vị trí chuỗi văn bản đang căn hiện tại
	for i := 0; i <= n-m; {
		// Kiểm tra chuỗi con
		j := 0
		for j < m && text[i+j] == pattern[j] {
			j++
		}
		// Nếu khớp hoàn toàn thì trả về vị trí khớp
		if j == m {
			return true
		}

		// Nếu còn ký tự thì kiểm tra giá trị của ký tự tiếp theo trong bảng dịch
		if i+m < n {
			next := rune(text[i+m])
			if val, ok := offset[next]; ok {
				i += val // Có trong bảng dịch thì thực hiện nhảy
			} else {
				i += len(pattern) + 1 // Không có trong bảng dịch, bỏ qua toàn bộ độ dài chuỗi mẫu
			}
		} else {
			break
		}
	}
	return false // Nếu không tìm thấy kết quả khớp thì trả về -1
}

func RemoveDuplicate(s []string) []string {
	result := make([]string, 0, len(s))
	temp := map[string]struct{}{}
	for _, item := range s {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func InitAc(words []string) *goahocorasick.Machine {
	m := new(goahocorasick.Machine)
	dict := readRunes(words)
	if err := m.Build(dict); err != nil {
		fmt.Println(err)
		return nil
	}
	return m
}

func readRunes(words []string) [][]rune {
	var dict [][]rune

	for _, word := range words {
		word = strings.ToLower(word)
		l := bytes.TrimSpace([]byte(word))
		dict = append(dict, bytes.Runes(l))
	}

	return dict
}

func AcSearch(findText string, dict []string, stopImmediately bool) (bool, []string) {
	if len(dict) == 0 {
		return false, nil
	}
	if len(findText) == 0 {
		return false, nil
	}
	m := InitAc(dict)
	if m == nil {
		return false, nil
	}
	hits := m.MultiPatternSearch([]rune(findText), stopImmediately)
	if len(hits) > 0 {
		words := make([]string, 0)
		for _, hit := range hits {
			words = append(words, string(hit.Word))
		}
		return true, words
	}
	return false, nil
}
