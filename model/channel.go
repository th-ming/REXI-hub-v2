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
package model

import (
	"encoding/json"
	"strings"
	"sync"
	"rexi-hub-v2/common"

	"gorm.io/gorm"
)

type Channel struct {
	Id                 int     `json:"id"`
	Type               int     `json:"type" gorm:"default:0"`
	Key                string  `json:"key" gorm:"not null"`
	OpenAIOrganization *string `json:"openai_organization"`
	TestModel          *string `json:"test_model"`
	Status             int     `json:"status" gorm:"default:1"`
	Name               string  `json:"name" gorm:"index"`
	Weight             *uint   `json:"weight" gorm:"default:0"`
	CreatedTime        int64   `json:"created_time" gorm:"bigint"`
	TestTime           int64   `json:"test_time" gorm:"bigint"`
	ResponseTime       int     `json:"response_time"` // in milliseconds
	BaseURL            *string `json:"base_url" gorm:"column:base_url;default:''"`
	Other              string  `json:"other"`
	Balance            float64 `json:"balance"` // in USD
	BalanceUpdatedTime int64   `json:"balance_updated_time" gorm:"bigint"`
	Models             string  `json:"models"`
	Group              string  `json:"group" gorm:"type:varchar(64);default:'default'"`
	UsedQuota          int64   `json:"used_quota" gorm:"bigint;default:0"`
	ModelMapping       *string `json:"model_mapping" gorm:"type:text"`
	//MaxInputTokens     *int    `json:"max_input_tokens" gorm:"default:0"`
	StatusCodeMapping *string `json:"status_code_mapping" gorm:"type:varchar(1024);default:''"`
	Priority          *int64  `json:"priority" gorm:"bigint;default:0"`
	AutoBan           *int    `json:"auto_ban" gorm:"default:1"`
	OtherInfo         string  `json:"other_info"`
	Tag               *string `json:"tag" gorm:"index"`
	Setting           *string `json:"setting" gorm:"type:text"`
	ParamOverride     *string `json:"param_override" gorm:"type:text"`
	ModelPrefix       *string `json:"model_prefix" gorm:"type:varchar(64);default:''"`
	SystemPrompt      *string `json:"system_prompt" gorm:"type:text"`
}

func (channel *Channel) GetModels() []string {
	if channel.Models == "" {
		return []string{}
	}
	return strings.Split(strings.Trim(channel.Models, ","), ",")
}

func (channel *Channel) GetOtherInfo() map[string]interface{} {
	otherInfo := make(map[string]interface{})
	if channel.OtherInfo != "" {
		err := json.Unmarshal([]byte(channel.OtherInfo), &otherInfo)
		if err != nil {
			common.SysError("failed to unmarshal other info: " + err.Error())
		}
	}
	return otherInfo
}

func (channel *Channel) SetOtherInfo(otherInfo map[string]interface{}) {
	otherInfoBytes, err := json.Marshal(otherInfo)
	if err != nil {
		common.SysError("failed to marshal other info: " + err.Error())
		return
	}
	channel.OtherInfo = string(otherInfoBytes)
}

func (channel *Channel) GetTag() string {
	if channel.Tag == nil {
		return ""
	}
	return *channel.Tag
}

func (channel *Channel) SetTag(tag string) {
	channel.Tag = &tag
}

func (channel *Channel) GetAutoBan() bool {
	if channel.AutoBan == nil {
		return false
	}
	return *channel.AutoBan == 1
}

func (channel *Channel) Save() error {
	return DB.Save(channel).Error
}

func GetAllChannels(startIdx int, num int, selectAll bool, idSort bool) ([]*Channel, error) {
	var channels []*Channel
	var err error
	order := "priority desc"
	if idSort {
		order = "id desc"
	}
	if selectAll {
		err = DB.Order(order).Find(&channels).Error
	} else {
		err = DB.Order(order).Limit(num).Offset(startIdx).Omit("key").Find(&channels).Error
	}
	return channels, err
}

func GetChannelsByTag(tag string, idSort bool) ([]*Channel, error) {
	var channels []*Channel
	order := "priority desc"
	if idSort {
		order = "id desc"
	}
	err := DB.Where("tag = ?", tag).Order(order).Find(&channels).Error
	return channels, err
}

// buildWhereClause dựng mệnh đề WHERE và tham số dùng chung
func buildWhereClause(keyword string, group string, model string) (string, []interface{}) {
	keyCol := "`key`"
	groupCol := "`group`"
	modelsCol := "`models`"

	// Nếu là PostgreSQL thì dùng dấu ngoặc kép
	if common.UsingPostgreSQL {
		keyCol = `"key"`
		groupCol = `"group"`
		modelsCol = `"models"`
	}

	var whereClause string
	var args []interface{}
	if group != "" && group != "null" {
		var groupCondition string
		if common.UsingMySQL {
			groupCondition = `CONCAT(',', ` + groupCol + `, ',') LIKE ?`
		} else {
			// sqlite, PostgreSQL
			groupCondition = `(',' || ` + groupCol + ` || ',') LIKE ?`
		}
		whereClause = "(id = ? OR name LIKE ? OR " + keyCol + " = ?) AND " + modelsCol + ` LIKE ? AND ` + groupCondition
		args = append(args, common.String2Int(keyword), "%"+keyword+"%", keyword, "%"+model+"%", "%,"+group+",%")
	} else {
		whereClause = "(id = ? OR name LIKE ? OR " + keyCol + " = ?) AND " + modelsCol + " LIKE ?"
		args = append(args, common.String2Int(keyword), "%"+keyword+"%", keyword, "%"+model+"%")
	}
	return whereClause, args
}

// applyTagFilter áp dụng lọc thẻ, hỗ trợ hai chế độ khớp chính xác và khớp chuỗi con
func applyTagFilter(query *gorm.DB, tags []string, exactMatch bool) *gorm.DB {
	if len(tags) == 0 {
		return query
	}

	if exactMatch {
		// Chế độ khớp hoàn toàn
		return query.Where("tag IN ?", tags)
	} else {
		// Chế độ khớp chuỗi con
		var tagConditions []string
		var tagArgs []interface{}
		for _, tag := range tags {
			if common.UsingMySQL {
				tagConditions = append(tagConditions, "CONCAT(',', COALESCE(tag, ''), ',') LIKE ?")
			} else {
				// sqlite, PostgreSQL
				tagConditions = append(tagConditions, "(',' || COALESCE(tag, '') || ',') LIKE ?")
			}
			tagArgs = append(tagArgs, "%,"+tag+",%")
		}
		return query.Where("("+strings.Join(tagConditions, " OR ")+")", tagArgs...)
	}
}

// SearchChannels hỗ trợ lọc theo loại/trạng thái/thẻ (giữ tương thích ngược)
// exactTagMatch: true=khớp chính xác thẻ, false=khớp theo chuỗi con
func SearchChannels(keyword string, group string, model string, idSort bool, types []int, statuses []int, tags []string) ([]*Channel, error) {
	return SearchChannelsWithTagMode(keyword, group, model, idSort, types, statuses, tags, true)
}

// SearchChannelsWithTagMode tìm kiếm hỗ trợ chỉ định chế độ khớp thẻ
func SearchChannelsWithTagMode(keyword string, group string, model string, idSort bool, types []int, statuses []int, tags []string, exactTagMatch bool) ([]*Channel, error) {
	var channels []*Channel
	keyCol := "`key`"

	// Nếu là PostgreSQL thì dùng dấu ngoặc kép
	if common.UsingPostgreSQL {
		keyCol = `"key"`
	}

	order := "priority desc"
	if idSort {
		order = "id desc"
	}

	// Dựng truy vấn cơ bản
	baseQuery := DB.Model(&Channel{}).Omit(keyCol)

	// Dựng mệnh đề WHERE
	whereClause, args := buildWhereClause(keyword, group, model)

	// Áp dụng điều kiện lọc
	query := baseQuery.Where(whereClause, args...)
	if len(types) > 0 {
		query = query.Where("type IN ?", types)
	}
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	query = applyTagFilter(query, tags, exactTagMatch)

	// Thực thi truy vấn
	err := query.Order(order).Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

func GetChannelById(id int, selectAll bool) (*Channel, error) {
	channel := Channel{Id: id}
	var err error = nil
	if selectAll {
		err = DB.First(&channel, "id = ?", id).Error
	} else {
		err = DB.Omit("key").First(&channel, "id = ?", id).Error
	}
	return &channel, err
}

func BatchInsertChannels(channels []Channel) error {
	var err error
	err = DB.Create(&channels).Error
	if err != nil {
		return err
	}
	for _, channel_ := range channels {
		err = channel_.AddAbilities()
		if err != nil {
			return err
		}
	}
	return nil
}

func BatchDeleteChannels(ids []int) error {
	//Dùng giao dịch để xóa bảng channel và bảng channel_ability
	tx := DB.Begin()
	err := tx.Where("id in (?)", ids).Delete(&Channel{}).Error
	if err != nil {
		// Hoàn tác giao dịch
		tx.Rollback()
		return err
	}
	err = tx.Where("channel_id in (?)", ids).Delete(&Ability{}).Error
	if err != nil {
		// Hoàn tác giao dịch
		tx.Rollback()
		return err
	}
	// Xác nhận giao dịch
	tx.Commit()
	return err
}

func (channel *Channel) GetPriority() int64 {
	if channel.Priority == nil {
		return 0
	}
	return *channel.Priority
}

func (channel *Channel) GetWeight() int {
	if channel.Weight == nil {
		return 0
	}
	return int(*channel.Weight)
}

func (channel *Channel) GetBaseURL() string {
	if channel.BaseURL == nil {
		return ""
	}
	return *channel.BaseURL
}

func (channel *Channel) GetModelMapping() string {
	if channel.ModelMapping == nil {
		return ""
	}
	return *channel.ModelMapping
}

func (channel *Channel) GetStatusCodeMapping() string {
	if channel.StatusCodeMapping == nil {
		return ""
	}
	return *channel.StatusCodeMapping
}

func (channel *Channel) Insert() error {
	var err error
	err = DB.Create(channel).Error
	if err != nil {
		return err
	}
	err = channel.AddAbilities()
	return err
}

func (channel *Channel) Update() error {
	var err error
	err = DB.Model(channel).Updates(channel).Error
	if err != nil {
		return err
	}
	DB.Model(channel).First(channel, "id = ?", channel.Id)
	err = channel.UpdateAbilities(nil)
	return err
}

func (channel *Channel) UpdateResponseTime(responseTime int64) {
	err := DB.Model(channel).Select("response_time", "test_time").Updates(Channel{
		TestTime:     common.GetTimestamp(),
		ResponseTime: int(responseTime),
	}).Error
	if err != nil {
		common.SysError("failed to update response time: " + err.Error())
	}
}

func (channel *Channel) UpdateBalance(balance float64) {
	err := DB.Model(channel).Select("balance_updated_time", "balance").Updates(Channel{
		BalanceUpdatedTime: common.GetTimestamp(),
		Balance:            balance,
	}).Error
	if err != nil {
		common.SysError("failed to update balance: " + err.Error())
	}
}

func (channel *Channel) Delete() error {
	var err error
	err = DB.Delete(channel).Error
	if err != nil {
		return err
	}
	err = channel.DeleteAbilities()
	return err
}

var channelStatusLock sync.Mutex

func UpdateChannelStatusById(id int, status int, reason string) bool {
	if common.MemoryCacheEnabled {
		channelStatusLock.Lock()
		defer channelStatusLock.Unlock()

		channelCache, _ := CacheGetChannel(id)
		// Nếu kênh trong đệm tồn tại và trạng thái đã là trạng thái đích thì trả về luôn
		if channelCache != nil && channelCache.Status == status {
			return false
		}
		// Nếu kênh trong đệm không tồn tại (tức đã bị vô hiệu hóa) và trạng thái cần đặt không phải kích hoạt thì trả về luôn
		if channelCache == nil && status != common.ChannelStatusEnabled {
			return false
		}
		CacheUpdateChannelStatus(id, status)
	}
	err := UpdateAbilityStatus(id, status == common.ChannelStatusEnabled)
	if err != nil {
		common.SysError("failed to update ability status: " + err.Error())
		return false
	}
	channel, err := GetChannelById(id, true)
	if err != nil {
		// find channel by id error, directly update status
		result := DB.Model(&Channel{}).Where("id = ?", id).Update("status", status)
		if result.Error != nil {
			common.SysError("failed to update channel status: " + result.Error.Error())
			return false
		}
		if result.RowsAffected == 0 {
			return false
		}
	} else {
		if channel.Status == status {
			return false
		}
		// find channel by id success, update status and other info
		info := channel.GetOtherInfo()
		info["status_reason"] = reason
		info["status_time"] = common.GetTimestamp()
		channel.SetOtherInfo(info)
		channel.Status = status
		err = channel.Save()
		if err != nil {
			common.SysError("failed to update channel status: " + err.Error())
			return false
		}
	}
	return true
}

func EnableChannelByTag(tag string) error {
	err := DB.Model(&Channel{}).Where("tag = ?", tag).Update("status", common.ChannelStatusEnabled).Error
	if err != nil {
		return err
	}
	err = UpdateAbilityStatusByTag(tag, true)
	return err
}

func DisableChannelByTag(tag string) error {
	err := DB.Model(&Channel{}).Where("tag = ?", tag).Update("status", common.ChannelStatusManuallyDisabled).Error
	if err != nil {
		return err
	}
	err = UpdateAbilityStatusByTag(tag, false)
	return err
}

func EditChannelByTag(tag string, newTag *string, modelMapping *string, models *string, group *string, priority *int64, weight *uint) error {
	updateData := Channel{}
	shouldReCreateAbilities := false
	updatedTag := tag
	// Nếu newTag không rỗng và khác tag thì cập nhật tag
	if newTag != nil && *newTag != tag {
		updateData.Tag = newTag
		updatedTag = *newTag
	}
	if modelMapping != nil && *modelMapping != "" {
		updateData.ModelMapping = modelMapping
	}
	if models != nil && *models != "" {
		shouldReCreateAbilities = true
		updateData.Models = *models
	}
	if group != nil && *group != "" {
		shouldReCreateAbilities = true
		updateData.Group = *group
	}
	if priority != nil {
		updateData.Priority = priority
	}
	if weight != nil {
		updateData.Weight = weight
	}

	err := DB.Model(&Channel{}).Where("tag = ?", tag).Updates(updateData).Error
	if err != nil {
		return err
	}
	if shouldReCreateAbilities {
		channels, err := GetChannelsByTag(updatedTag, false)
		if err == nil {
			for _, channel := range channels {
				err = channel.UpdateAbilities(nil)
				if err != nil {
					common.SysError("failed to update abilities: " + err.Error())
				}
			}
		}
	} else {
		err := UpdateAbilityByTag(tag, newTag, priority, weight)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateChannelUsedQuota(id int, quota int) {
	if common.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeChannelUsedQuota, id, quota)
		return
	}
	updateChannelUsedQuota(id, quota)
}

func updateChannelUsedQuota(id int, quota int) {
	err := DB.Model(&Channel{}).Where("id = ?", id).Update("used_quota", gorm.Expr("used_quota + ?", quota)).Error
	if err != nil {
		common.SysError("failed to update channel used quota: " + err.Error())
	}
}

func DeleteChannelByStatus(status int64) (int64, error) {
	result := DB.Where("status = ?", status).Delete(&Channel{})
	return result.RowsAffected, result.Error
}

func DeleteDisabledChannel() (int64, error) {
	result := DB.Where("status = ? or status = ?", common.ChannelStatusAutoDisabled, common.ChannelStatusManuallyDisabled).Delete(&Channel{})
	return result.RowsAffected, result.Error
}

func GetPaginatedTags(offset int, limit int) ([]*string, error) {
	var tags []*string
	err := DB.Model(&Channel{}).Select("DISTINCT tag").Where("tag != ''").Offset(offset).Limit(limit).Find(&tags).Error
	return tags, err
}

// SearchTags hỗ trợ lọc theo loại/trạng thái/thẻ (dùng để lọc tập thẻ trước khi gộp khi tag_mode=true)
// exactTagMatch: true=khớp chính xác thẻ, false=khớp theo chuỗi con
func SearchTags(keyword string, group string, model string, idSort bool, types []int, statuses []int, tagsFilter []string) ([]*string, error) {
	return SearchTagsWithTagMode(keyword, group, model, idSort, types, statuses, tagsFilter, true)
}

// SearchTagsWithTagMode tìm kiếm thẻ hỗ trợ chỉ định chế độ khớp thẻ
func SearchTagsWithTagMode(keyword string, group string, model string, idSort bool, types []int, statuses []int, tagsFilter []string, exactTagMatch bool) ([]*string, error) {
	var tags []*string
	keyCol := "`key`"

	// Nếu là PostgreSQL thì dùng dấu ngoặc kép
	if common.UsingPostgreSQL {
		keyCol = `"key"`
	}

	order := "priority desc"
	if idSort {
		order = "id desc"
	}

	// Dựng truy vấn cơ bản
	baseQuery := DB.Model(&Channel{}).Omit(keyCol)

	// Dựng mệnh đề WHERE
	whereClause, args := buildWhereClause(keyword, group, model)

	// Áp dụng điều kiện lọc
	query := baseQuery.Where(whereClause, args...)
	if len(types) > 0 {
		query = query.Where("type IN ?", types)
	}
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	query = applyTagFilter(query, tagsFilter, exactTagMatch)

	// Dựng truy vấn con và lấy các thẻ duy nhất
	subQuery := query.Select("tag").
		Where("tag != ''").
		Order(order)

	err := DB.Table("(?) as sub", subQuery).
		Select("DISTINCT tag").
		Find(&tags).Error

	if err != nil {
		return nil, err
	}

	return tags, nil
}

// Get model prefix from channel
func (channel *Channel) GetModelPrefix() string {
	if channel.ModelPrefix == nil {
		return ""
	}
	return *channel.ModelPrefix
}

func (channel *Channel) GetSystemPrompt() string {
	if channel.SystemPrompt == nil {
		return ""
	}
	return *channel.SystemPrompt
}

func (channel *Channel) GetSetting() map[string]interface{} {
	setting := make(map[string]interface{})
	if channel.Setting != nil && *channel.Setting != "" {
		err := json.Unmarshal([]byte(*channel.Setting), &setting)
		if err != nil {
			common.SysError("failed to unmarshal setting: " + err.Error())
		}
	}
	return setting
}

func (channel *Channel) SetSetting(setting map[string]interface{}) {
	settingBytes, err := json.Marshal(setting)
	if err != nil {
		common.SysError("failed to marshal setting: " + err.Error())
		return
	}
	channel.Setting = common.GetPointer[string](string(settingBytes))
}

func (channel *Channel) GetParamOverride() map[string]interface{} {
	paramOverride := make(map[string]interface{})
	if channel.ParamOverride != nil && *channel.ParamOverride != "" {
		err := json.Unmarshal([]byte(*channel.ParamOverride), &paramOverride)
		if err != nil {
			common.SysError("failed to unmarshal param override: " + err.Error())
		}
	}
	return paramOverride
}

func GetChannelsByIds(ids []int) ([]*Channel, error) {
	var channels []*Channel
	err := DB.Where("id in (?)", ids).Find(&channels).Error
	return channels, err
}

func BatchSetChannelTag(ids []int, tag *string) error {
	// Mở giao dịch
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Cập nhật thẻ
	err := tx.Model(&Channel{}).Where("id in (?)", ids).Update("tag", tag).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// update ability status
	channels, err := GetChannelsByIds(ids)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, channel := range channels {
		err = channel.UpdateAbilities(tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Xác nhận giao dịch
	return tx.Commit().Error
}
