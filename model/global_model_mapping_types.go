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

package model

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// ModelMappingItem mục ánh xạ mô hình
type ModelMappingItem struct {
	Model      string `json:"model" binding:"required"`   // Actual model name
	Priorities int    `json:"priorities" binding:"min=0"` // Priority (non-negative integer)
}

// GlobalModelMapping cấu hình ánh xạ mô hình toàn cục
type GlobalModelMapping struct {
	Mapping map[string][]ModelMappingItem `json:"mapping"` // Tên mô hình ảo -> danh sách ánh xạ mô hình thực
}

// RoundRobinCounter bộ đếm luân phiên
type RoundRobinCounter struct {
	counters map[string]int
	mutex    sync.RWMutex
}

// NewRoundRobinCounter tạo bộ đếm luân phiên mới
func NewRoundRobinCounter() *RoundRobinCounter {
	return &RoundRobinCounter{
		counters: make(map[string]int),
	}
}

// GetNext lấy chỉ số luân phiên tiếp theo
func (rrc *RoundRobinCounter) GetNext(key string, maxIndex int) int {
	if maxIndex <= 0 {
		return 0
	}
	
	rrc.mutex.Lock()
	defer rrc.mutex.Unlock()
	
	current, exists := rrc.counters[key]
	if !exists {
		current = 0
	}
	
	next := (current + 1) % maxIndex
	rrc.counters[key] = next
	
	return current
}

// ClearAllCounters xóa toàn bộ bộ đếm
func (rrc *RoundRobinCounter) ClearAllCounters() {
	rrc.mutex.Lock()
	defer rrc.mutex.Unlock()
	
	rrc.counters = make(map[string]int)
}

// GetActualModel lấy tên mô hình thực theo tên mô hình ảo (có xét độ ưu tiên và luân phiên)
func GetActualModel(virtualModel string) (string, error) {
	ModelMappingMutex.RLock()
	defer ModelMappingMutex.RUnlock()
	
	if globalModelMapping == nil || len(globalModelMapping.Mapping) == 0 {
		// Nếu chưa cấu hình ánh xạ thì trả về tên mô hình gốc
		return virtualModel, nil
	}
	
	mappingItems, exists := globalModelMapping.Mapping[virtualModel]
	if !exists || len(mappingItems) == 0 {
		// Nếu mô hình ảo chưa có cấu hình ánh xạ thì trả về tên mô hình gốc
		return virtualModel, nil
	}
	
	// Nhóm theo độ ưu tiên
	priorityGroups := make(map[int][]ModelMappingItem)
	for _, item := range mappingItems {
		if item.Priorities < 0 {
			// Bỏ qua độ ưu tiên không hợp lệ (số âm)
			continue
		}
		priorityGroups[item.Priorities] = append(priorityGroups[item.Priorities], item)
	}
	
	if len(priorityGroups) == 0 {
		return virtualModel, nil
	}
	
	// Lấy độ ưu tiên cao nhất
	var maxPriority int = -1
	for priority := range priorityGroups {
		if priority > maxPriority {
			maxPriority = priority
		}
	}
	
	highestPriorityItems := priorityGroups[maxPriority]
	if len(highestPriorityItems) == 0 {
		return virtualModel, nil
	}
	
	// Nếu chỉ có một mục có độ ưu tiên cao nhất thì trả về luôn
	if len(highestPriorityItems) == 1 {
		return highestPriorityItems[0].Model, nil
	}
	
	// Dùng chiến lược luân phiên để chọn từ các mục có độ ưu tiên cao nhất
	if roundRobinCounter == nil {
		// Nếu bộ đếm luân phiên chưa khởi tạo thì chọn ngẫu nhiên
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(highestPriorityItems))
		return highestPriorityItems[index].Model, nil
	}
	
	// Dùng bộ đếm luân phiên
	index := roundRobinCounter.GetNext(virtualModel, len(highestPriorityItems))
	return highestPriorityItems[index].Model, nil
}

// ValidateModelMapping kiểm tra cấu hình ánh xạ mô hình
func ValidateModelMapping(mapping *GlobalModelMapping) error {
	if mapping == nil {
		return fmt.Errorf("Mapping config cannot be empty")
	}
	
	if mapping.Mapping == nil {
		return fmt.Errorf("Mapping dictionary cannot be empty")
	}
	
	for virtualModel, items := range mapping.Mapping {
		if strings.TrimSpace(virtualModel) == "" {
			return fmt.Errorf("Virtual model name cannot be empty")
		}
		
		if len(items) == 0 {
			return fmt.Errorf("Virtual model '%s' mapping items cannot be empty", virtualModel)
		}
		
		// Dùng để kiểm tra trùng tên mô hình thực trong cùng một mô hình ảo
		modelSet := make(map[string]bool)
		
		for i, item := range items {
			if strings.TrimSpace(item.Model) == "" {
				return fmt.Errorf("Actual model name of mapping item #%d of virtual model '%s' cannot be empty", virtualModel, i+1)
			}
			
			if item.Priorities < 0 {
				return fmt.Errorf("The priority of actual model '%s' of virtual model '%s' must be a non-negative integer", virtualModel, item.Model)
			}
			
			// Kiểm tra trùng tên mô hình thực trong cùng một mô hình ảo
			trimmedModel := strings.TrimSpace(item.Model)
			if modelSet[trimmedModel] {
				return fmt.Errorf("Virtual model '%s' contains duplicate actual model name '%s'", virtualModel, trimmedModel)
			}
			modelSet[trimmedModel] = true
		}
	}
	
	return nil
}