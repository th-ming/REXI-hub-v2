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
	"github.com/gin-gonic/gin"
	"time"
	"rexi-hub-v2/common"
	"rexi-hub-v2/constant"
	"rexi-hub-v2/model"
	"rexi-hub-v2/setting/operation_setting"
)

type Setup struct {
	Status       bool   `json:"status"`
	RootInit     bool   `json:"root_init"`
	DatabaseType string `json:"database_type"`
}

type SetupRequest struct {
	Username           string `json:"username"`
	Password           string `json:"password"`
	ConfirmPassword    string `json:"confirmPassword"`
	SelfUseModeEnabled bool   `json:"SelfUseModeEnabled"`
	DemoSiteEnabled    bool   `json:"DemoSiteEnabled"`
}

func GetSetup(c *gin.Context) {
	setup := Setup{
		Status: constant.Setup,
	}
	if constant.Setup {
		c.JSON(200, gin.H{
			"success": true,
			"data":    setup,
		})
		return
	}
	setup.RootInit = model.RootUserExists()
	if common.UsingMySQL {
		setup.DatabaseType = "mysql"
	}
	if common.UsingPostgreSQL {
		setup.DatabaseType = "postgres"
	}
	if common.UsingSQLite {
		setup.DatabaseType = "sqlite"
	}
	c.JSON(200, gin.H{
		"success": true,
		"data":    setup,
	})
}

func PostSetup(c *gin.Context) {
	// Check if setup is already completed
	if constant.Setup {
		c.JSON(400, gin.H{
			"success": false,
			"message": "The system has already been initialized",
		})
		return
	}

	// Check if root user already exists
	rootExists := model.RootUserExists()

	var req SetupRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "Invalid request parameters",
		})
		return
	}

	// If root doesn't exist, validate and create admin account
	if !rootExists {
		// Validate password
		if req.Password != req.ConfirmPassword {
			c.JSON(400, gin.H{
				"success": false,
				"message": "The two passwords do not match",
			})
			return
		}

		if len(req.Password) < 8 {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Password must be at least 8 characters",
			})
			return
		}

		// Create root user
		hashedPassword, err := common.Password2Hash(req.Password)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "System error: " + err.Error(),
			})
			return
		}
		rootUser := model.User{
			Username:    req.Username,
			Password:    hashedPassword,
			Role:        common.RoleRootUser,
			Status:      common.UserStatusEnabled,
			DisplayName: "Root User",
			AccessToken: nil,
			Quota:       100000000,
		}
		err = model.DB.Create(&rootUser).Error
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to create admin account: " + err.Error(),
			})
			return
		}
	}

	// Set operation modes
	operation_setting.SelfUseModeEnabled = req.SelfUseModeEnabled
	operation_setting.DemoSiteEnabled = req.DemoSiteEnabled

	// Save operation modes to database for persistence
	err = model.UpdateOption("SelfUseModeEnabled", boolToString(req.SelfUseModeEnabled))
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"message": "Failed to save self-use mode settings: " + err.Error(),
		})
		return
	}

	err = model.UpdateOption("DemoSiteEnabled", boolToString(req.DemoSiteEnabled))
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"message": "Failed to save demo site mode settings: " + err.Error(),
		})
		return
	}

	// Auto-detect reverse proxy configuration based on request headers
	err = initializeProxySettings(c)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"message": "Auto-detect reverse proxy config failed: " + err.Error(),
		})
		return
	}

	// Update setup status
	constant.Setup = true

	setup := model.Setup{
		Version:       common.Version,
		InitializedAt: time.Now().Unix(),
	}
	err = model.DB.Create(&setup).Error
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"message": "System initialization failed: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "System initialized successfully",
	})
}

// initializeProxySettings automatically detects and configures reverse proxy settings
// based on the headers present in the initialization request
func initializeProxySettings(c *gin.Context) error {
	// Use the DetectProxyHeaders function from common/utils.go
	provider, detectedIP := common.DetectProxyHeaders(c)
	
	if provider != "" {
		// Proxy headers detected, enable reverse proxy with detected provider
		err := model.UpdateOption("ReverseProxyEnabled", "true")
		if err != nil {
			common.SysError("Failed to enable reverse proxy during initialization: " + err.Error())
			return err
		}
		
		err = model.UpdateOption("ReverseProxyProvider", provider)
		if err != nil {
			common.SysError("Failed to set reverse proxy provider during initialization: " + err.Error())
			return err
		}
		
		// Log the auto-detection result
		common.SysLog("Auto-detected reverse proxy configuration: provider=" + provider + ", detected_ip=" + detectedIP)
	} else {
		// No proxy headers detected, ensure reverse proxy is disabled
		err := model.UpdateOption("ReverseProxyEnabled", "false")
		if err != nil {
			common.SysError("Failed to disable reverse proxy during initialization: " + err.Error())
			return err
		}
		
		// Set default provider (nginx) for future use
		err = model.UpdateOption("ReverseProxyProvider", "nginx")
		if err != nil {
			common.SysError("Failed to set default reverse proxy provider during initialization: " + err.Error())
			return err
		}
		
		common.SysLog("No reverse proxy headers detected during initialization, reverse proxy disabled")
	}
	
	return nil
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
