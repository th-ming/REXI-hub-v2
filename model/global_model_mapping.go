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

package model

import (
	"encoding/json"
	"fmt"
	"sync"

	"rexi-hub-v2/common"
)

// Biến toàn cục
var (
	globalModelMapping *GlobalModelMapping
	roundRobinCounter  *RoundRobinCounter
	ModelMappingMutex  sync.RWMutex
)

// Hàm khởi tạo package init
func init() {
	// Khởi tạo biến toàn cục, cấu hình thực tế sẽ được nạp trong InitializeModelMapping
	globalModelMapping = &GlobalModelMapping{
		Mapping: make(map[string][]ModelMappingItem),
	}
	roundRobinCounter = NewRoundRobinCounter()
}

// LoadModelMappingConfig tải cấu hình từ cơ sở dữ liệu
func LoadModelMappingConfig() error {
	ModelMappingMutex.Lock()
	defer ModelMappingMutex.Unlock()

	return loadModelMappingConfigUnsafe()
}

// loadModelMappingConfigUnsafe tải cấu hình từ cơ sở dữ liệu (gọi nội bộ, không khóa)
func loadModelMappingConfigUnsafe() error {
	// Tải cấu hình từ cơ sở dữ liệu
	common.OptionMapRWMutex.RLock()
	modelMappingStr, ok := common.OptionMap["global_model_mapping"]
	common.OptionMapRWMutex.RUnlock()

	if !ok || modelMappingStr == "" || modelMappingStr == "{}" {
		// Nếu chưa cấu hình hoặc cấu hình rỗng thì dùng cấu hình mặc định
		globalModelMapping = &GlobalModelMapping{
			Mapping: make(map[string][]ModelMappingItem),
		}
		return nil
	}

	// Phân tích cấu hình JSON
	var mapping GlobalModelMapping
	if err := json.Unmarshal([]byte(modelMappingStr), &mapping); err != nil {
		common.SysError(fmt.Sprintf("Failed to parse global model mapping configuration: %v", err))
		return err
	}

	globalModelMapping = &mapping
	return nil
}

// GetGlobalModelMapping lấy bản sao cấu hình ánh xạ mô hình toàn cục
func GetGlobalModelMapping() *GlobalModelMapping {
	ModelMappingMutex.RLock()
	defer ModelMappingMutex.RUnlock()

	if globalModelMapping == nil {
		return &GlobalModelMapping{
			Mapping: make(map[string][]ModelMappingItem),
		}
	}

	// Trả về bản sao cấu hình để tránh bị sửa từ bên ngoài
	configCopy := &GlobalModelMapping{
		Mapping: make(map[string][]ModelMappingItem),
	}

	for k, v := range globalModelMapping.Mapping {
		itemsCopy := make([]ModelMappingItem, len(v))
		copy(itemsCopy, v)
		configCopy.Mapping[k] = itemsCopy
	}

	return configCopy
}

// GlobalModelMappingToJSONString chuyển cấu hình ánh xạ mô hình toàn cục thành chuỗi JSON
func GlobalModelMappingToJSONString() string {
	ModelMappingMutex.RLock()
	defer ModelMappingMutex.RUnlock()

	if globalModelMapping == nil {
		return "{}"
	}

	jsonBytes, err := json.Marshal(globalModelMapping)
	if err != nil {
		common.SysError("Failed to serialize global model mapping configuration: " + err.Error())
		return "{}"
	}
	return string(jsonBytes)
}

// UpdateGlobalModelMappingFromJSONString cập nhật cấu hình ánh xạ mô hình toàn cục từ chuỗi JSON
func UpdateGlobalModelMappingFromJSONString(jsonStr string) error {
	if jsonStr == "" || jsonStr == "{}" {
		// Nếu cấu hình rỗng thì dùng cấu hình mặc định
		defaultMapping := &GlobalModelMapping{
			Mapping: make(map[string][]ModelMappingItem),
		}
		return UpdateGlobalModelMapping(defaultMapping)
	}

	// Phân tích cấu hình JSON
	var mapping GlobalModelMapping
	if err := json.Unmarshal([]byte(jsonStr), &mapping); err != nil {
		common.SysError(fmt.Sprintf("Failed to parse global model mapping configuration: %v", err))
		return err
	}

	return UpdateGlobalModelMapping(&mapping)
}

// UpdateGlobalModelMapping cập nhật cấu hình ánh xạ mô hình toàn cục
func UpdateGlobalModelMapping(mapping *GlobalModelMapping) error {
	ModelMappingMutex.Lock()
	defer ModelMappingMutex.Unlock()

	// Kiểm tra định dạng cấu hình
	if err := ValidateModelMapping(mapping); err != nil {
		return err
	}

	// Tuần tự hóa cấu hình thành JSON
	jsonData, err := json.Marshal(mapping)
	if err != nil {
		common.SysError(fmt.Sprintf("Failed to serialize global model mapping configuration: %v", err))
		return err
	}

	// Lưu cấu hình vào cơ sở dữ liệu
	if err := UpdateOption("global_model_mapping", string(jsonData)); err != nil {
		common.SysError(fmt.Sprintf("Failed to save global model mapping configuration to the database.: %v", err))
		return err
	}

	globalModelMapping = mapping

	// Đặt lại bộ đếm luân phiên
	roundRobinCounter.ClearAllCounters()

	common.SysLog("Global model mapping configuration updated successfully")
	return nil
}

// InitializeModelMapping khởi tạo ánh xạ mô hình
func InitializeModelMapping() error {
	ModelMappingMutex.Lock()
	defer ModelMappingMutex.Unlock()

	// Khởi tạo bộ đếm luân phiên
	roundRobinCounter = NewRoundRobinCounter()

	// Gọi trực tiếp logic nạp nội bộ để tránh khóa trùng
	if err := loadModelMappingConfigUnsafe(); err != nil {
		common.SysError(fmt.Sprintf("Failed to load model mapping configuration: %v", err))
		return err
	}

	common.SysLog("Model mapping initialized successfully")
	return nil
}

// ReloadModelMapping tải lại cấu hình ánh xạ mô hình
func ReloadModelMapping() error {
	ModelMappingMutex.Lock()
	defer ModelMappingMutex.Unlock()

	// Tải lại cấu hình
	if err := loadModelMappingConfigUnsafe(); err != nil {
		common.SysError(fmt.Sprintf("Failed to reload model mapping configuration: %v", err))
		return err
	}

	// Đặt lại bộ đếm luân phiên
	roundRobinCounter.ClearAllCounters()

	common.SysLog("Model mapping configuration reload successful")
	return nil
}

// GetAllVirtualModels lấy tên toàn bộ mô hình ảo đã cấu hình
func GetAllVirtualModels() []string {
	ModelMappingMutex.RLock()
	defer ModelMappingMutex.RUnlock()

	if globalModelMapping == nil || len(globalModelMapping.Mapping) == 0 {
		return []string{}
	}

	virtualModels := make([]string, 0, len(globalModelMapping.Mapping))
	for virtualModel := range globalModelMapping.Mapping {
		virtualModels = append(virtualModels, virtualModel)
	}

	return virtualModels
}
