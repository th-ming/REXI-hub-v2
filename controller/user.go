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
package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"rexi-hub-v2/common"
	"rexi-hub-v2/model"
	"rexi-hub-v2/setting"

	"rexi-hub-v2/constant"
	"rexi-hub-v2/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	if !common.PasswordLoginEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "The administrator has disabled password login",
			"success": false,
		})
		return
	}
	var loginRequest LoginRequest
	err := json.NewDecoder(c.Request.Body).Decode(&loginRequest)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Invalid parameters",
			"success": false,
		})
		return
	}
	username := loginRequest.Username
	password := loginRequest.Password
	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"message": "Invalid parameters",
			"success": false,
		})
		return
	}
	user := model.User{
		Username: username,
		Password: password,
	}
	err = user.ValidateAndFill()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	setupLogin(&user, c)
}

// setup session & cookies and then return user info
func setupLogin(user *model.User, c *gin.Context) {
	session := sessions.Default(c)

	// Clear any existing session data first
	session.Clear()
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Unable to clear old session, please retry",
			"success": false,
		})
		return
	}

	// Set new session data
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("status", user.Status)
	session.Set("group", user.Group)

	err = session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Unable to save session, please retry",
			"success": false,
		})
		return
	}

	cleanUser := model.User{
		Id:          user.Id,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Status:      user.Status,
		Group:       user.Group,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
		"data":    cleanUser,
	})
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
	})
}

func Register(c *gin.Context) {
	if !common.RegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "The administrator has disabled new user registration",
			"success": false,
		})
		return
	}
	if !common.PasswordRegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "The administrator has disabled password registration. Please register via third-party account verification",
			"success": false,
		})
		return
	}
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid parameters",
		})
		return
	}

	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid input: " + err.Error(),
		})
		return
	}
	if common.EmailVerificationEnabled {
		if user.Email == "" || user.VerificationCode == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Email verification is enabled. Please enter your email address and verification code",
			})
			return
		}
		if !common.VerifyCodeWithKey(user.Email, user.VerificationCode, common.EmailVerificationPurpose) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Verification code is incorrect or expired",
			})
			return
		}
	}
	exist, err := model.CheckUserExistOrDeleted(user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Database error, please try again later",
		})
		common.SysError(fmt.Sprintf("CheckUserExistOrDeleted error: %v", err))
		return
	}
	if exist {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Username already exists or has been deactivated",
		})
		return
	}
	affCode := user.AffCode // this code is the inviter's code, not the user's own code
	inviterId, err := model.GetUserIdByAffCode(affCode)
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.Username,
		InviterId:   inviterId,
	}
	if common.EmailVerificationEnabled {
		cleanUser.Email = user.Email
	}
	if err := cleanUser.Insert(inviterId); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Lấy ID người dùng sau khi chèn
	var insertedUser model.User
	if err := model.DB.Where("username = ?", cleanUser.Username).First(&insertedUser).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "User registration failed or failed to get the user ID",
		})
		return
	}
	// Tạo token mặc định
	if constant.GenerateDefaultToken {
		key, err := common.GenerateKey()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Failed to generate the default token",
			})
			common.SysError("failed to generate token key: " + err.Error())
			return
		}
		// Tạo token mặc định
		token := model.Token{
			UserId:             insertedUser.Id, // Dùng ID người dùng sau khi chèn
			Name:               cleanUser.Username + "'s initial token",
			Key:                key,
			CreatedTime:        common.GetTimestamp(),
			AccessedTime:       common.GetTimestamp(),
			ExpiredTime:        -1,     // Không bao giờ hết hạn
			RemainQuota:        500000, // Hạn mức mẫu
			UnlimitedQuota:     true,
			ModelLimitsEnabled: false,
		}
		if err := token.Insert(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Failed to create the default token",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func GetAllUsers(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if p < 1 {
		p = 1
	}
	if pageSize < 0 {
		pageSize = common.ItemsPerPage
	}
	users, total, err := model.GetAllUsers((p-1)*pageSize, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":     users,
			"total":     total,
			"page":      p,
			"page_size": pageSize,
		},
	})
	return
}

func SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	group := c.Query("group")
	p, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if p < 1 {
		p = 1
	}
	if pageSize < 0 {
		pageSize = common.ItemsPerPage
	}
	startIdx := (p - 1) * pageSize
	users, total, err := model.SearchUsers(keyword, group, startIdx, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":     users,
			"total":     total,
			"page":      p,
			"page_size": pageSize,
		},
	})
	return
}

func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= user.Role && myRole != common.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "No permission to get info of users with equal or higher role",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user,
	})
	return
}

func GenerateAccessToken(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	// get rand int 28-32
	randI := common.GetRandomInt(4)
	key, err := common.GenerateRandomKey(29 + randI)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Generation failed",
		})
		common.SysError("failed to generate key: " + err.Error())
		return
	}
	user.SetAccessToken(key)

	if model.DB.Where("access_token = ?", user.AccessToken).First(user).RowsAffected != 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Please retry, the system-generated UUID is a duplicate!",
		})
		return
	}

	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user.AccessToken,
	})
	return
}

type TransferAffQuotaRequest struct {
	Quota int `json:"quota" binding:"required"`
}

func TransferAffQuota(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	tran := TransferAffQuotaRequest{}
	if err := c.ShouldBindJSON(&tran); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = user.TransferAffQuotaToQuota(tran.Quota)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Transfer failed: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transfer successful",
	})
}

func GetAffCode(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if user.AffCode == "" {
		user.AffCode = common.GetRandomString(4)
		if err := user.Update(false); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user.AffCode,
	})
	return
}

func GetSelf(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user,
	})
	return
}

func GetUserModels(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		id = c.GetInt("id")
	}
	user, err := model.GetUserCache(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	groups := setting.GetUserUsableGroups(user.Group)
	var models []string
	addedModels := make(map[string]bool) // Track added models to avoid duplicates

	// Get prefix channels for the user's group
	prefixChannels := middleware.GetPrefixChannels(user.Group)
	modelPrefixMap := make(map[string][]string) // Map base model to all prefixed versions

	for prefix, channels := range prefixChannels {
		if prefix == "" {
			continue // Skip channels without prefixes
		}
		for _, channel := range channels {
			for _, modelName := range channel.GetModels() {
				prefixedModel := prefix + modelName
				// Check if this prefixed model already exists to avoid duplicates
				exists := false
				for _, existing := range modelPrefixMap[modelName] {
					if existing == prefixedModel {
						exists = true
						break
					}
				}
				if !exists {
					modelPrefixMap[modelName] = append(modelPrefixMap[modelName], prefixedModel)
				}
			}
		}
	}

	// First add all models with prefixes that are allowed for the user's group
	for group := range groups {
		groupModels := model.GetGroupModels(group)
		for _, baseModel := range groupModels {
			if prefixedModels, ok := modelPrefixMap[baseModel]; ok {
				for _, prefixedModel := range prefixedModels {
					if !addedModels[prefixedModel] {
						models = append(models, prefixedModel)
						addedModels[prefixedModel] = true
					}
				}
			}
		}
	}

	// Then add remaining models without prefixes that are allowed for the user's group
	// Only add base models that don't have any prefixed versions already added
	for group := range groups {
		groupModels := model.GetGroupModels(group)
		for _, baseModel := range groupModels {
			// Skip if this base model was already added
			if addedModels[baseModel] {
				continue
			}

			// Check if any prefixed version of this base model was already added
			hasAddedPrefixedVersion := false
			if prefixedModels, ok := modelPrefixMap[baseModel]; ok {
				for _, prefixedName := range prefixedModels {
					if addedModels[prefixedName] {
						hasAddedPrefixedVersion = true
						break
					}
				}
			}

			// Only add the base model if no prefixed version was added
			if !hasAddedPrefixedVersion {
				models = append(models, baseModel)
				addedModels[baseModel] = true
			}
		}
	}

	// Thêm mô hình ảo vào danh sách trả về
	virtualModels := model.GetAllVirtualModels()
	for _, virtualModel := range virtualModels {
		if !addedModels[virtualModel] {
			models = append(models, virtualModel)
			addedModels[virtualModel] = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    models,
	})
	return
}

func UpdateUser(c *gin.Context) {
	var updatedUser model.User
	err := json.NewDecoder(c.Request.Body).Decode(&updatedUser)
	if err != nil || updatedUser.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid parameters",
		})
		return
	}
	if updatedUser.Password == "" {
		updatedUser.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&updatedUser); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid input: " + err.Error(),
		})
		return
	}
	originUser, err := model.GetUserById(updatedUser.Id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= originUser.Role && myRole != common.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "No permission to update users with equal or higher role",
		})
		return
	}
	if myRole <= updatedUser.Role && myRole != common.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "No permission to promote a user to a role equal to or higher than your own",
		})
		return
	}
	if updatedUser.Password == "$I_LOVE_U" {
		updatedUser.Password = "" // rollback to what it should be
	}
	updatePassword := updatedUser.Password != ""
	if err := updatedUser.Edit(updatePassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if originUser.Quota != updatedUser.Quota {
		model.RecordLog(originUser.Id, model.LogTypeManage, fmt.Sprintf("Admin changed user quota from %s to %s", common.LogQuota(originUser.Quota), common.LogQuota(updatedUser.Quota)))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func UpdateSelf(c *gin.Context) {
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid parameters",
		})
		return
	}
	if user.Password == "" {
		user.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid input: " + err.Error(),
		})
		return
	}

	cleanUser := model.User{
		Id:          c.GetInt("id"),
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if user.Password == "$I_LOVE_U" {
		user.Password = "" // rollback to what it should be
		cleanUser.Password = ""
	}
	updatePassword := user.Password != ""
	if err := cleanUser.Update(updatePassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	originUser, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= originUser.Role {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "No permission to delete users with equal or higher role",
		})
		return
	}
	err = model.HardDeleteUserById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
		})
		return
	}
}

func DeleteSelf(c *gin.Context) {
	id := c.GetInt("id")
	user, _ := model.GetUserById(id, false)

	if user.Role == common.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Cannot delete the root administrator account",
		})
		return
	}

	err := model.DeleteUserById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func CreateUser(c *gin.Context) {
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	user.Username = strings.TrimSpace(user.Username)
	if err != nil || user.Username == "" || user.Password == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid parameters",
		})
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid input: " + err.Error(),
		})
		return
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	myRole := c.GetInt("role")
	if user.Role >= myRole {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Cannot create a user with a role equal to or higher than your own",
		})
		return
	}
	// Even for admin users, we cannot fully trust them!
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if err := cleanUser.Insert(0); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type ManageRequest struct {
	Id     int    `json:"id"`
	Action string `json:"action"`
}

// ManageUser Only admin user can do this
func ManageUser(c *gin.Context) {
	var req ManageRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid parameters",
		})
		return
	}
	user := model.User{
		Id: req.Id,
	}
	// Fill attributes
	model.DB.Unscoped().Where(&user).First(&user)
	if user.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "User does not exist",
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= user.Role && myRole != common.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "No permission to update users with equal or higher role",
		})
		return
	}
	switch req.Action {
	case "disable":
		user.Status = common.UserStatusDisabled
		if user.Role == common.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Cannot disable the root administrator",
			})
			return
		}
	case "enable":
		user.Status = common.UserStatusEnabled
	case "delete":
		if user.Role == common.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Cannot delete the root administrator",
			})
			return
		}
		if err := user.Delete(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "promote":
		if myRole != common.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "A regular admin cannot promote other users to admin",
			})
			return
		}
		if user.Role >= common.RoleAdminUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "This user is already an admin",
			})
			return
		}
		user.Role = common.RoleAdminUser
	case "demote":
		if user.Role == common.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Cannot demote the root administrator",
			})
			return
		}
		if user.Role == common.RoleCommonUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "This user is already a regular user",
			})
			return
		}
		user.Role = common.RoleCommonUser
	}

	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	clearUser := model.User{
		Role:   user.Role,
		Status: user.Status,
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    clearUser,
	})
	return
}

func EmailBind(c *gin.Context) {
	email := c.Query("email")
	code := c.Query("code")
	if !common.VerifyCodeWithKey(email, code, common.EmailVerificationPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Verification code is incorrect or expired",
		})
		return
	}
	session := sessions.Default(c)
	id := session.Get("id")
	user := model.User{
		Id: id.(int),
	}
	err := user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.Email = email
	// no need to check if this email already taken, because we have used verification code to check it
	err = user.Update(false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type topUpRequest struct {
	Key string `json:"key"`
}

func TopUp(c *gin.Context) {
	req := topUpRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Validate turnstile if enabled
	if common.TurnstileCheckEnabled {
		response := c.Query("turnstile")
		if response == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Turnstile token is empty",
			})
			return
		}

		rawRes, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", url.Values{
			"secret":   {common.TurnstileSecretKey},
			"response": {response},
			"remoteip": {c.ClientIP()},
		})
		if err != nil {
			common.SysError(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		defer rawRes.Body.Close()

		var res struct {
			Success bool `json:"success"`
		}
		err = json.NewDecoder(rawRes.Body).Decode(&res)
		if err != nil {
			common.SysError(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		if !res.Success {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Turnstile verification failed, please refresh and retry!",
			})
			return
		}
	}

	id := c.GetInt("id")
	quota, isGift, err := model.Redeem(req.Key, id) // Sửa ở đây, nhận cả ba giá trị trả về
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"quota":   quota,
			"is_gift": isGift, // Có thể chọn trả về cho frontend để hiển thị thông báo khác nhau
		},
	})
	return
}

type UpdateUserSettingRequest struct {
	QuotaWarningType           string  `json:"notify_type"`
	QuotaWarningThreshold      float64 `json:"quota_warning_threshold"`
	WebhookUrl                 string  `json:"webhook_url,omitempty"`
	WebhookSecret              string  `json:"webhook_secret,omitempty"`
	NotificationEmail          string  `json:"notification_email,omitempty"`
	AcceptUnsetModelRatioModel bool    `json:"accept_unset_model_ratio_model"`
	ShowIPInLogs               bool    `json:"show_ip_in_logs"`
}

func UpdateUserSetting(c *gin.Context) {
	var req UpdateUserSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid parameters",
		})
		return
	}

	// Kiểm tra loại cảnh báo
	if req.QuotaWarningType != constant.NotifyTypeEmail && req.QuotaWarningType != constant.NotifyTypeWebhook {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid warning type",
		})
		return
	}

	// Kiểm tra ngưỡng cảnh báo
	if req.QuotaWarningThreshold <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Warning threshold must be greater than 0",
		})
		return
	}

	// Kiểm tra cài đặt log IP (giá trị boolean không cần kiểm tra thêm, nhưng có thể thêm kiểm tra logic nghiệp vụ)
	// ShowIPInLogs là kiểu boolean, Go tự động xử lý giá trị true/false

	// Nếu là loại webhook thì kiểm tra địa chỉ webhook
	if req.QuotaWarningType == constant.NotifyTypeWebhook {
		if req.WebhookUrl == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Webhook URL cannot be empty",
			})
			return
		}
		// Kiểm tra định dạng URL
		if _, err := url.ParseRequestURI(req.WebhookUrl); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Invalid webhook URL",
			})
			return
		}
	}

	// Nếu có cung cấp email thông báo thì thêm vào cài đặt
	if req.QuotaWarningType == constant.NotifyTypeEmail && req.NotificationEmail != "" {
		// Kiểm tra định dạng email
		if !strings.Contains(req.NotificationEmail, "@") {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Invalid email address",
			})
			return
		}
	}

	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Dựng cài đặt
	settings := map[string]interface{}{
		constant.UserSettingNotifyType:            req.QuotaWarningType,
		constant.UserSettingQuotaWarningThreshold: req.QuotaWarningThreshold,
		constant.UserAcceptUnsetRatioModel:        req.AcceptUnsetModelRatioModel,
		constant.UserSettingShowIPInLogs:          req.ShowIPInLogs,
	}

	// Nếu là loại webhook thì thêm các cài đặt liên quan đến webhook
	if req.QuotaWarningType == constant.NotifyTypeWebhook {
		settings[constant.UserSettingWebhookUrl] = req.WebhookUrl
		if req.WebhookSecret != "" {
			settings[constant.UserSettingWebhookSecret] = req.WebhookSecret
		}
	}

	// Nếu có cung cấp email thông báo thì thêm vào cài đặt
	if req.QuotaWarningType == constant.NotifyTypeEmail && req.NotificationEmail != "" {
		settings[constant.UserSettingNotificationEmail] = req.NotificationEmail
	}

	// Cập nhật cài đặt người dùng
	user.SetSetting(settings)
	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to update settings: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Settings updated",
	})
}

// CheckInStatus returns whether a user can check in today
func CheckInStatus(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"can_check_in": user.CanCheckInToday(),
		},
	})
}

// CheckIn performs a daily check-in for the user
func CheckIn(c *gin.Context) {
	// Get check-in settings and validate them
	checkInQuota, checkInMaxQuota, err := getAndValidateCheckInSettings()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Validate turnstile if enabled
	if common.TurnstileCheckEnabled {
		session := sessions.Default(c)
		turnstileChecked := session.Get("turnstile")
		if turnstileChecked == nil {
			response := c.Query("turnstile")
			if response == "" {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "Turnstile token is empty",
				})
				return
			}

			rawRes, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", url.Values{
				"secret":   {common.TurnstileSecretKey},
				"response": {response},
				"remoteip": {c.ClientIP()},
			})
			if err != nil {
				common.SysError(err.Error())
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
			defer rawRes.Body.Close()

			var res struct {
				Success bool `json:"success"`
			}
			err = json.NewDecoder(rawRes.Body).Decode(&res)
			if err != nil {
				common.SysError(err.Error())
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}

			if !res.Success {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "Turnstile verification failed, please refresh and retry!",
				})
				return
			}

			session.Set("turnstile", true)
			err = session.Save()
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"message": "Unable to save session, please retry",
					"success": false,
				})
				return
			}
		}
	}

	// Get user and perform check-in
	id := c.GetInt("id")
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	reward, err := user.CheckIn(checkInQuota, checkInMaxQuota)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Check-in successful",
		"data": gin.H{
			"quota": reward,
		},
	})
}

// getAndValidateCheckInSettings retrieves and validates check-in settings
func getAndValidateCheckInSettings() (checkInQuota, checkInMaxQuota int, err error) {
	common.OptionMapRWMutex.RLock()
	checkInEnabled := common.OptionMap["CheckInEnabled"] == "true"
	checkInQuotaStr := common.OptionMap["CheckInQuota"]
	checkInMaxQuotaStr := common.OptionMap["CheckInMaxQuota"]
	common.OptionMapRWMutex.RUnlock()

	if !checkInEnabled {
		return 0, 0, errors.New("Check-in feature is not enabled")
	}

	checkInQuota, err = strconv.Atoi(checkInQuotaStr)
	if err != nil || checkInQuota <= 0 {
		return 0, 0, errors.New("Check-in quota configuration error")
	}

	checkInMaxQuota = checkInQuota
	if checkInMaxQuotaStr != "" {
		maxQuota, err := strconv.Atoi(checkInMaxQuotaStr)
		if err == nil && maxQuota > checkInQuota {
			checkInMaxQuota = maxQuota
		}
	}

	return checkInQuota, checkInMaxQuota, nil
}
