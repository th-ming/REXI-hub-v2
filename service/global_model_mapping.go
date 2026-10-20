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
	"rexi-hub-v2/model"
)

// GetActualModel lấy tên mô hình thực theo tên mô hình ảo (có xét độ ưu tiên và luân phiên)
// Đây là hàm bọc cho GetActualModel trong package model
func GetActualModel(virtualModel string) (string, error) {
	return model.GetActualModel(virtualModel)
}

// InitializeModelMappingService khởi tạo dịch vụ ánh xạ mô hình
// Đây là hàm bọc cho InitializeModelMapping trong package model
func InitializeModelMappingService() error {
	return model.InitializeModelMapping()
}

// ReloadModelMapping tải lại cấu hình ánh xạ mô hình
// Đây là hàm bọc cho ReloadModelMapping trong package model
func ReloadModelMapping() error {
	return model.ReloadModelMapping()
}

// GetGlobalModelMapping lấy bản sao cấu hình ánh xạ mô hình toàn cục
// Đây là hàm bọc cho GetGlobalModelMapping trong package model
func GetGlobalModelMapping() *model.GlobalModelMapping {
	return model.GetGlobalModelMapping()
}

// GlobalModelMappingToJSONString chuyển cấu hình ánh xạ mô hình toàn cục thành chuỗi JSON
// Đây là hàm bọc cho GlobalModelMappingToJSONString trong package model
func GlobalModelMappingToJSONString() string {
	return model.GlobalModelMappingToJSONString()
}

// UpdateGlobalModelMappingFromJSONString cập nhật cấu hình ánh xạ mô hình toàn cục từ chuỗi JSON
// Đây là hàm bọc cho UpdateGlobalModelMappingFromJSONString trong package model
func UpdateGlobalModelMappingFromJSONString(jsonStr string) error {
	return model.UpdateGlobalModelMappingFromJSONString(jsonStr)
}

// UpdateGlobalModelMapping cập nhật cấu hình ánh xạ mô hình toàn cục
// Đây là hàm bọc cho UpdateGlobalModelMapping trong package model
func UpdateGlobalModelMapping(mapping *model.GlobalModelMapping) error {
	return model.UpdateGlobalModelMapping(mapping)
}