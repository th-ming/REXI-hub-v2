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
package common

import (
	"errors"
	"net/smtp"
	"strings"
)

type outlookAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &outlookAuth{username, password}
}

func (a *outlookAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *outlookAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unknown fromServer")
		}
	}
	return nil, nil
}

func isOutlookServer(server string) bool {
	// Tương thích email outlook và ofb ở nhiều khu vực
	// Thực ra nên thêm một Option để phân biệt có đăng nhập theo cách LOGIN hay không
	// Tạm thời tương thích một chút
	return strings.Contains(server, "outlook") || strings.Contains(server, "onmicrosoft")
}
