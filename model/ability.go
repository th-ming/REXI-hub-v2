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
	"errors"
	"fmt"
	"strings"
	"rexi-hub-v2/common"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Ability struct {
	Group     string  `json:"group" gorm:"type:varchar(64);primaryKey;autoIncrement:false"`
	Model     string  `json:"model" gorm:"type:varchar(255);primaryKey;autoIncrement:false"`
	ChannelId int     `json:"channel_id" gorm:"primaryKey;autoIncrement:false;index"`
	Enabled   bool    `json:"enabled"`
	Priority  *int64  `json:"priority" gorm:"bigint;default:0;index"`
	Weight    uint    `json:"weight" gorm:"default:0;index"`
	Tag       *string `json:"tag" gorm:"index"`
}

func GetGroupModels(group string) []string {
	var models []string
	// Find distinct models
	DB.Table("abilities").Where(groupCol+" = ? and enabled = ?", group, true).Distinct("model").Pluck("model", &models)
	return models
}

func GetEnabledModels() []string {
	var models []string
	// Find distinct models
	DB.Table("abilities").Where("enabled = ?", true).Distinct("model").Pluck("model", &models)
	return models
}

func GetAllEnableAbilities() []Ability {
	var abilities []Ability
	DB.Find(&abilities, "enabled = ?", true)
	return abilities
}

// getModelCondition returns the appropriate model condition string for case-sensitive comparison
func getModelCondition() string {
	if common.UsingMySQL {
		// MySQL: Use BINARY for case-sensitive comparison
		return "BINARY model = ?"
	} else if common.UsingPostgreSQL {
		// PostgreSQL: Default is case-sensitive, but explicitly use COLLATE "C" for consistency
		return "model COLLATE \"C\" = ?"
	} else if common.UsingSQLite {
		// SQLite: Default BINARY collation is case-sensitive, but make it explicit
		return "model COLLATE BINARY = ?"
	}
	return "model = ?"
}

func getPriority(group string, model string, retry int) (int, error) {
	trueVal := "1"
	if common.UsingPostgreSQL {
		trueVal = "true"
	}

	modelCondition := getModelCondition()

	var priorities []int
	err := DB.Model(&Ability{}).
		Select("DISTINCT(priority)").
		Where(groupCol+" = ? and "+modelCondition+" and enabled = "+trueVal, group, model).
		Order("priority DESC").              // Sắp xếp giảm dần theo độ ưu tiên
		Pluck("priority", &priorities).Error // Pluck dùng để quét trực tiếp kết quả truy vấn vào một slice

	if err != nil {
		// Xử lý lỗi
		return 0, err
	}

	if len(priorities) == 0 {
		// Nếu không truy vấn được độ ưu tiên thì trả về lỗi
		return 0, errors.New("Database consistency broken")
	}

	// Xác định độ ưu tiên sẽ dùng
	var priorityToUse int
	if retry >= len(priorities) {
		// Nếu số lần thử lại lớn hơn số mức ưu tiên thì dùng mức ưu tiên nhỏ nhất
		priorityToUse = priorities[len(priorities)-1]
	} else {
		priorityToUse = priorities[retry]
	}
	return priorityToUse, nil
}

func getChannelQuery(group string, model string, retry int) *gorm.DB {
	trueVal := "1"
	if common.UsingPostgreSQL {
		trueVal = "true"
	}

	modelCondition := getModelCondition()

	maxPrioritySubQuery := DB.Model(&Ability{}).Select("MAX(priority)").Where(groupCol+" = ? and "+modelCondition+" and enabled = "+trueVal, group, model)
	channelQuery := DB.Where(groupCol+" = ? and "+modelCondition+" and enabled = "+trueVal+" and priority = (?)", group, model, maxPrioritySubQuery)
	if retry != 0 {
		priority, err := getPriority(group, model, retry)
		if err != nil {
			common.SysError(fmt.Sprintf("Get priority failed: %s", err.Error()))
		} else {
			channelQuery = DB.Where(groupCol+" = ? and "+modelCondition+" and enabled = "+trueVal+" and priority = ?", group, model, priority)
		}
	}

	return channelQuery
}

func GetRandomSatisfiedChannel(group string, model string, retry int) (*Channel, error) {
	// Gọi dịch vụ ánh xạ mô hình toàn cục để chuyển tên mô hình ảo thành tên mô hình thực
	actualModel, err := GetActualModel(model)
	if err != nil {
		common.SysError(fmt.Sprintf("Model mapping failed: Virtual Model=%s, Error=%s", model, err.Error()))
		return nil, fmt.Errorf("Model mapping failed: %w", err)
	}

	// Ghi log ánh xạ mô hình
	if actualModel != model {
		common.SysLog(fmt.Sprintf("Model Mapping: Virtual Model=%s -> Actual model=%s", model, actualModel))
	}

	var abilities []Ability

	var dbErr error = nil
	channelQuery := getChannelQuery(group, actualModel, retry)
	if common.UsingSQLite || common.UsingPostgreSQL {
		dbErr = channelQuery.Order("weight DESC").Find(&abilities).Error
	} else {
		dbErr = channelQuery.Order("weight DESC").Find(&abilities).Error
	}
	if dbErr != nil {
		common.SysError(fmt.Sprintf("Failed to query channel ability: group=%s, model=%s, error=%s", group, actualModel, dbErr.Error()))
		return nil, dbErr
	}

	channel := Channel{}
	if len(abilities) > 0 {
		// Randomly choose one
		weightSum := uint(0)
		for _, ability_ := range abilities {
			weightSum += ability_.Weight + 10
		}
		// Randomly choose one
		weight := common.GetRandomInt(int(weightSum))
		for _, ability_ := range abilities {
			weight -= int(ability_.Weight) + 10
			//log.Printf("weight: %d, ability weight: %d", weight, *ability_.Weight)
			if weight <= 0 {
				channel.Id = ability_.ChannelId
				break
			}
		}
	} else {
		common.SysError(fmt.Sprintf("No available channel found: group=%s, model=%s", group, actualModel))
		return nil, errors.New("no channels available")
	}

	dbErr = DB.First(&channel, "id = ?", channel.Id).Error
	if dbErr != nil {
		common.SysError(fmt.Sprintf("Failed to query channel info: channel_id=%d, error=%s", channel.Id, dbErr.Error()))
		return nil, dbErr
	}

	return &channel, dbErr
}

func (channel *Channel) AddAbilities() error {
	models_ := strings.Split(channel.Models, ",")
	groups_ := strings.Split(channel.Group, ",")
	abilities := make([]Ability, 0, len(models_))
	for _, model := range models_ {
		for _, group := range groups_ {
			ability := Ability{
				Group:     group,
				Model:     model,
				ChannelId: channel.Id,
				Enabled:   channel.Status == common.ChannelStatusEnabled,
				Priority:  channel.Priority,
				Weight:    uint(channel.GetWeight()),
				Tag:       channel.Tag,
			}
			abilities = append(abilities, ability)
		}
	}
	if len(abilities) == 0 {
		return nil
	}
	for _, chunk := range lo.Chunk(abilities, 50) {
		err := DB.Create(&chunk).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (channel *Channel) DeleteAbilities() error {
	return DB.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error
}

// UpdateAbilities updates abilities of this channel.
// Make sure the channel is completed before calling this function.
func (channel *Channel) UpdateAbilities(tx *gorm.DB) error {
	isNewTx := false
	// Nếu không truyền giao dịch vào thì tạo giao dịch mới
	if tx == nil {
		tx = DB.Begin()
		if tx.Error != nil {
			return tx.Error
		}
		isNewTx = true
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
	}

	// First delete all abilities of this channel
	err := tx.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error
	if err != nil {
		if isNewTx {
			tx.Rollback()
		}
		return err
	}

	// Then add new abilities
	models_ := strings.Split(channel.Models, ",")
	groups_ := strings.Split(channel.Group, ",")
	abilities := make([]Ability, 0, len(models_))
	for _, model := range models_ {
		for _, group := range groups_ {
			ability := Ability{
				Group:     group,
				Model:     model,
				ChannelId: channel.Id,
				Enabled:   channel.Status == common.ChannelStatusEnabled,
				Priority:  channel.Priority,
				Weight:    uint(channel.GetWeight()),
				Tag:       channel.Tag,
			}
			abilities = append(abilities, ability)
		}
	}

	if len(abilities) > 0 {
		for _, chunk := range lo.Chunk(abilities, 50) {
			err = tx.Create(&chunk).Error
			if err != nil {
				if isNewTx {
					tx.Rollback()
				}
				return err
			}
		}
	}

	// Nếu là giao dịch mới tạo thì cần xác nhận
	if isNewTx {
		return tx.Commit().Error
	}

	return nil
}

func UpdateAbilityStatus(channelId int, status bool) error {
	return DB.Model(&Ability{}).Where("channel_id = ?", channelId).Select("enabled").Update("enabled", status).Error
}

func UpdateAbilityStatusByTag(tag string, status bool) error {
	return DB.Model(&Ability{}).Where("tag = ?", tag).Select("enabled").Update("enabled", status).Error
}

func UpdateAbilityByTag(tag string, newTag *string, priority *int64, weight *uint) error {
	ability := Ability{}
	if newTag != nil {
		ability.Tag = newTag
	}
	if priority != nil {
		ability.Priority = priority
	}
	if weight != nil {
		ability.Weight = *weight
	}
	return DB.Model(&Ability{}).Where("tag = ?", tag).Updates(ability).Error
}

func FixAbility() (int, error) {
	var channelIds []int
	count := 0
	// Find all channel ids from channel table
	err := DB.Model(&Channel{}).Pluck("id", &channelIds).Error
	if err != nil {
		common.SysError(fmt.Sprintf("Get channel ids from channel table failed: %s", err.Error()))
		return 0, err
	}
	// Delete abilities of channels that are not in channel table
	err = DB.Where("channel_id NOT IN (?)", channelIds).Delete(&Ability{}).Error
	if err != nil {
		common.SysError(fmt.Sprintf("Delete abilities of channels that are not in channel table failed: %s", err.Error()))
		return 0, err
	}
	common.SysLog(fmt.Sprintf("Delete abilities of channels that are not in channel table successfully, ids: %v", channelIds))
	count += len(channelIds)

	// Use channelIds to find channel not in abilities table
	var abilityChannelIds []int
	err = DB.Table("abilities").Distinct("channel_id").Pluck("channel_id", &abilityChannelIds).Error
	if err != nil {
		common.SysError(fmt.Sprintf("Get channel ids from abilities table failed: %s", err.Error()))
		return 0, err
	}
	var channels []Channel
	if len(abilityChannelIds) == 0 {
		err = DB.Find(&channels).Error
	} else {
		err = DB.Where("id NOT IN (?)", abilityChannelIds).Find(&channels).Error
	}
	if err != nil {
		return 0, err
	}
	for _, channel := range channels {
		err := channel.UpdateAbilities(nil)
		if err != nil {
			common.SysError(fmt.Sprintf("Update abilities of channel %d failed: %s", channel.Id, err.Error()))
		} else {
			common.SysLog(fmt.Sprintf("Update abilities of channel %d successfully", channel.Id))
			count++
		}
	}
	InitChannelCache()
	return count, nil
}
