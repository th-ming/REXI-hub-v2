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
	"encoding/json"
	"net/http"
	"rexi-hub-v2/model"
	"rexi-hub-v2/service"

	"github.com/gin-gonic/gin"
)

// GetGlobalModelMapping lấy cấu hình ánh xạ mô hình toàn cục hiện tại
func GetGlobalModelMapping(c *gin.Context) {
	mapping := service.GetGlobalModelMapping()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    mapping,
	})
}

// UpdateGlobalModelMapping cập nhật cấu hình ánh xạ mô hình toàn cục
func UpdateGlobalModelMapping(c *gin.Context) {
	var mapping model.GlobalModelMapping
	
	if err := c.ShouldBindJSON(&mapping); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to parse request parameters: " + err.Error(),
		})
		return
	}
	
	if err := service.UpdateGlobalModelMapping(&mapping); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update global model mapping config: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Global model mapping config updated successfully",
	})
}



// GetModelMappingConfig lấy cấu hình ánh xạ mô hình dưới dạng chuỗi JSON
func GetModelMappingConfig(c *gin.Context) {
	configStr := service.GlobalModelMappingToJSONString()
	
	// Thử phân tích chuỗi JSON để kiểm tra định dạng
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Config format error: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    config,
	})
}

// UpdateModelMappingConfig cập nhật cấu hình ánh xạ mô hình qua chuỗi JSON
func UpdateModelMappingConfig(c *gin.Context) {
	// Đọc phần thân yêu cầu
	jsonStr, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to read request body: " + err.Error(),
		})
		return
	}
	
	// Nếu phần thân yêu cầu rỗng thì đặt thành chuỗi rỗng
	if len(jsonStr) == 0 {
		jsonStr = []byte("{}")
	}
	
	// Cập nhật cấu hình
	if err := service.UpdateGlobalModelMappingFromJSONString(string(jsonStr)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update global model mapping config: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Global model mapping config updated successfully",
	})
}


// ReloadModelMapping tải lại cấu hình ánh xạ mô hình
func ReloadModelMapping(c *gin.Context) {
	if err := service.ReloadModelMapping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to reload model mapping config: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Model mapping config reloaded successfully",
	})
}