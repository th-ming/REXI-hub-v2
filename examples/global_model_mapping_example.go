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

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"rexi-hub-v2/model"
	"rexi-hub-v2/service"
)

func main() {
	// Khởi tạo dịch vụ ánh xạ mô hình
	if err := service.InitializeModelMappingService(); err != nil {
		log.Fatalf("Failed to initialize model mapping service: %v", err)
	}

	// Tạo cấu hình ánh xạ mô hình mẫu
	exampleMapping := &model.GlobalModelMapping{
		Mapping: map[string][]model.ModelMappingItem{
			"gpt-3.5-turbo": {
				{Model: "gpt-3.5-turbo-0613", Priorities: 10},
				{Model: "gpt-3.5-turbo-0301", Priorities: 10},
				{Model: "gpt-3.5-turbo-16k", Priorities: 5},
			},
			"gpt-4": {
				{Model: "gpt-4-0613", Priorities: 10},
				{Model: "gpt-4-0314", Priorities: 8},
			},
			"claude-instant": {
				{Model: "claude-instant-1.2", Priorities: 10},
				{Model: "claude-instant-1.1", Priorities: 5},
			},
		},
	}

	// Cập nhật cấu hình ánh xạ mô hình toàn cục
	if err := service.UpdateGlobalModelMapping(exampleMapping); err != nil {
		log.Fatalf("Failed to update global model mapping: %v", err)
	}

	// Lấy tên mô hình thực
	testModels := []string{"gpt-3.5-turbo", "gpt-4", "claude-instant", "non-existent-model"}

	fmt.Println("Model mapping test:")
	for _, virtualModel := range testModels {
		actualModel, err := service.GetActualModel(virtualModel)
		if err != nil {
			log.Printf("Error getting actual model for %s: %v", virtualModel, err)
			continue
		}
		fmt.Printf("Virtual model: %-20s -> actual model: %s\n", virtualModel, actualModel)
	}

	// Lấy cấu hình hiện tại
	currentConfig := service.GetGlobalModelMapping()
	fmt.Println("\nCurrent global model mapping config:")
	jsonData, _ := json.MarshalIndent(currentConfig, "", "  ")
	fmt.Println(string(jsonData))

	// Tải lại cấu hình
	if err := service.ReloadModelMapping(); err != nil {
		log.Printf("Failed to reload model mapping: %v", err)
	} else {
		fmt.Println("\nConfig reloaded successfully")
	}
}