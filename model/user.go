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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"rexi-hub-v2/common"

	"github.com/bytedance/gopkg/util/gopool"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// User if you add sensitive fields, don't forget to clean them in setupLogin function.
// Otherwise, the sensitive information will be saved on local storage in plain text!
type User struct {
	Id                int            `json:"id"`
	Username          string         `json:"username" gorm:"unique;index" validate:"max=12"`
	Password          string         `json:"password" gorm:"not null;" validate:"min=8,max=20"`
	DisplayName       string         `json:"display_name" gorm:"index" validate:"max=20"`
	Role              int            `json:"role" gorm:"type:int;default:1"`   // admin, common
	Status            int            `json:"status" gorm:"type:int;default:1"` // enabled, disabled
	Email             string         `json:"email" gorm:"index" validate:"max=50"`
	GitHubId          string         `json:"github_id" gorm:"column:github_id;index"`
	OidcId            string         `json:"oidc_id" gorm:"column:oidc_id;index"`
	WeChatId          string         `json:"wechat_id" gorm:"column:wechat_id;index"`
	TelegramId        string         `json:"telegram_id" gorm:"column:telegram_id;index"`
	VerificationCode  string         `json:"verification_code" gorm:"-:all"`                                    // this field is only for Email verification, don't save it to database!
	AccessToken       *string        `json:"access_token" gorm:"type:char(32);column:access_token;uniqueIndex"` // this token is for system management
	Quota             int            `json:"quota" gorm:"type:int;default:0"`
	UsedQuota         int            `json:"used_quota" gorm:"type:int;default:0;column:used_quota"` // used quota
	RequestCount      int            `json:"request_count" gorm:"type:int;default:0;"`               // request number
	Group             string         `json:"group" gorm:"type:varchar(64);default:'default'"`
	AffCode           string         `json:"aff_code" gorm:"type:varchar(32);column:aff_code;uniqueIndex"`
	AffCount          int            `json:"aff_count" gorm:"type:int;default:0;column:aff_count"`
	AffQuota          int            `json:"aff_quota" gorm:"type:int;default:0;column:aff_quota"`           // Remaining invitation quota
	AffHistoryQuota   int            `json:"aff_history_quota" gorm:"type:int;default:0;column:aff_history"` // Invitation history quota
	SubscriptionQuota int            `json:"subscription_quota" gorm:"type:int;default:0;column:subscription_quota"`
	InviterId         int            `json:"inviter_id" gorm:"type:int;column:inviter_id;index"`
	DeletedAt         gorm.DeletedAt `gorm:"index"`
	LinuxDOId         string         `json:"linux_do_id" gorm:"column:linux_do_id;index"`
	IDCFlareId        string         `json:"idc_flare_id" gorm:"column:idc_flare_id;index"`
	Setting           string         `json:"setting" gorm:"type:text;column:setting"`
	LastCheckInTime   *time.Time     `json:"last_check_in_time" gorm:"column:last_check_in_time"` // Thời gian điểm danh lần trước
}

func (user *User) ToBaseUser() *UserBase {
	cache := &UserBase{
		Id:                user.Id,
		Group:             user.Group,
		Quota:             user.Quota,
		SubscriptionQuota: user.SubscriptionQuota,
		Status:            user.Status,
		Username:          user.Username,
		Setting:           user.Setting,
		Email:             user.Email,
	}
	return cache
}

type UserQuotaBalance struct {
	Quota             int
	SubscriptionQuota int
}

func (b UserQuotaBalance) Total() int {
	return b.Quota + b.SubscriptionQuota
}

func (user *User) GetAccessToken() string {
	if user.AccessToken == nil {
		return ""
	}
	return *user.AccessToken
}

func (user *User) SetAccessToken(token string) {
	user.AccessToken = &token
}

func (user *User) GetSetting() map[string]interface{} {
	if user.Setting == "" {
		return nil
	}
	return common.StrToMap(user.Setting)
}

// GetShowIPInLogs returns whether the user has enabled IP logging in their settings
func (user *User) GetShowIPInLogs() bool {
	settings := user.GetSetting()
	if settings == nil {
		return false
	}

	if showIP, exists := settings["show_ip_in_logs"]; exists {
		if boolVal, ok := showIP.(bool); ok {
			return boolVal
		}
	}
	return false
}

func (user *User) SetSetting(setting map[string]interface{}) {
	settingBytes, err := json.Marshal(setting)
	if err != nil {
		common.SysError("failed to marshal setting: " + err.Error())
		return
	}
	user.Setting = string(settingBytes)
}

// CheckUserExistOrDeleted check if user exist or deleted, if not exist, return false, nil, if deleted or exist, return true, nil
func CheckUserExistOrDeleted(username string, email string) (bool, error) {
	var user User

	// err := DB.Unscoped().First(&user, "username = ? or email = ?", username, email).Error
	// check email if empty
	var err error
	if email == "" {
		err = DB.Unscoped().First(&user, "username = ?", username).Error
	} else {
		err = DB.Unscoped().First(&user, "username = ? or email = ?", username, email).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// not exist, return false, nil
			return false, nil
		}
		// other error, return false, err
		return false, err
	}
	// exist, return true, nil
	return true, nil
}

func GetMaxUserId() int {
	var user User
	DB.Unscoped().Last(&user)
	return user.Id
}

func GetAllUsers(startIdx int, num int) (users []*User, total int64, err error) {
	// Start transaction
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get total count within transaction
	err = tx.Unscoped().Model(&User{}).Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Get paginated users within same transaction
	err = tx.Unscoped().Order("id desc").Limit(num).Offset(startIdx).Omit("password").Find(&users).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Commit transaction
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func SearchUsers(keyword string, group string, startIdx int, num int) ([]*User, int64, error) {
	var users []*User
	var total int64
	var err error

	// Bắt đầu giao dịch
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Dựng truy vấn cơ bản
	query := tx.Unscoped().Model(&User{})

	// Dựng điều kiện tìm kiếm
	likeCondition := "username LIKE ? OR email LIKE ? OR display_name LIKE ?"

	// Thử chuyển từ khóa thành ID số nguyên
	keywordInt, err := strconv.Atoi(keyword)
	if err == nil {
		// Nếu là số thì tìm cả ID và các trường khác
		likeCondition = "id = ? OR " + likeCondition
		if group != "" {
			query = query.Where("("+likeCondition+") AND "+groupCol+" = ?",
				keywordInt, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", group)
		} else {
			query = query.Where(likeCondition,
				keywordInt, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
	} else {
		// Từ khóa không phải số thì chỉ tìm trong trường chuỗi
		if group != "" {
			query = query.Where("("+likeCondition+") AND "+groupCol+" = ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", group)
		} else {
			query = query.Where(likeCondition,
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
	}

	// Lấy tổng số
	err = query.Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Lấy dữ liệu phân trang
	err = query.Omit("password").Order("id desc").Limit(num).Offset(startIdx).Find(&users).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Xác nhận giao dịch
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func GetUserById(id int, selectAll bool) (*User, error) {
	if id == 0 {
		return nil, errors.New("id is empty!")
	}
	user := User{Id: id}
	var err error = nil
	if selectAll {
		err = DB.First(&user, "id = ?", id).Error
	} else {
		err = DB.Omit("password").First(&user, "id = ?", id).Error
	}
	return &user, err
}

func GetUserIdByAffCode(affCode string) (int, error) {
	if affCode == "" {
		return 0, errors.New("affCode is empty!")
	}
	var user User
	err := DB.Select("id").First(&user, "aff_code = ?", affCode).Error
	return user.Id, err
}

func DeleteUserById(id int) (err error) {
	if id == 0 {
		return errors.New("id is empty!")
	}
	user := User{Id: id}
	return user.Delete()
}

func HardDeleteUserById(id int) error {
	if id == 0 {
		return errors.New("id is empty!")
	}
	err := DB.Unscoped().Delete(&User{}, "id = ?", id).Error
	return err
}

func inviteUser(inviterId int) (err error) {
	// Luôn cập nhật thống kê mời, bất kể phần thưởng có bằng 0 hay không
	updateFields := map[string]interface{}{
		"aff_count": gorm.Expr("aff_count + ?", 1),
	}

	// Chỉ cập nhật các trường liên quan đến thưởng khi phần thưởng lớn hơn 0
	if common.QuotaForInviter > 0 {
		updateFields["aff_quota"] = gorm.Expr("aff_quota + ?", common.QuotaForInviter)
		updateFields["aff_history"] = gorm.Expr("aff_history + ?", common.QuotaForInviter)
	}

	result := DB.Model(&User{}).Where("id = ?", inviterId).Updates(updateFields)

	if result.Error != nil {
		return result.Error
	}

	// Kiểm tra xem có tìm thấy người dùng không
	if result.RowsAffected == 0 {
		return errors.New("Inviter does not exist")
	}

	return nil
}

// ProcessRebate xử lý logic hoàn tiền
func ProcessRebate(userId int, amount int, rebateType string) error {
	// Kiểm tra chức năng mời và hoàn tiền có được bật không
	if !common.AffEnabled || !common.RebateEnabled || common.RebatePercentage <= 0 {
		return nil
	}

	// Lấy thông tin người dùng
	user, err := GetUserById(userId, true)
	if err != nil {
		return err
	}

	// Nếu người dùng không có người mời thì không hoàn tiền
	if user.InviterId == 0 {
		return nil
	}

	// Tính số tiền hoàn lại
	rebateAmount := int(float64(amount) * common.RebatePercentage / 100.0)
	if rebateAmount <= 0 {
		return nil
	}

	// Tăng hạn mức hoàn tiền cho người mời
	err = IncreaseUserQuota(user.InviterId, rebateAmount, false)
	if err != nil {
		return err
	}

	// Ghi log hoàn tiền
	RecordLog(user.InviterId, LogTypeSystem, fmt.Sprintf("Earned %s referral commission %s, rate: %.1f%%",
		rebateType, common.LogQuota(rebateAmount), common.RebatePercentage))
	RecordLog(userId, LogTypeSystem, fmt.Sprintf("Generated %s referral commission %s for the inviter",
		rebateType, common.LogQuota(rebateAmount)))

	return nil
}

func (user *User) TransferAffQuotaToQuota(quota int) error {
	// Kiểm tra quota có nhỏ hơn hạn mức tối thiểu không
	if float64(quota) < common.QuotaPerUnit {
		return fmt.Errorf("Minimum transfer amount is %s!", common.LogQuota(int(common.QuotaPerUnit)))
	}

	// Bắt đầu giao dịch cơ sở dữ liệu
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback() // Đảm bảo giao dịch có thể hoàn tác khi hàm thoát

	// Truy vấn người dùng có khóa để đảm bảo tính nhất quán dữ liệu
	err := tx.Set("gorm:query_option", "FOR UPDATE").First(&user, user.Id).Error
	if err != nil {
		return err
	}

	// Kiểm tra lại xem AffQuota của người dùng có đủ không
	if user.AffQuota < quota {
		return errors.New("Insufficient invitation quota!")
	}

	// Cập nhật hạn mức người dùng
	user.AffQuota -= quota
	user.Quota += quota

	// Lưu trạng thái người dùng
	if err := tx.Save(user).Error; err != nil {
		return err
	}

	// Xác nhận giao dịch
	return tx.Commit().Error
}

func (user *User) Insert(inviterId int) error {
	var err error
	if user.Password != "" {
		user.Password, err = common.Password2Hash(user.Password)
		if err != nil {
			return err
		}
	}
	user.Quota = common.QuotaForNewUser
	user.AffCode = common.GetRandomString(4)

	// Bắt đầu giao dịch cơ sở dữ liệu
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback() // Đảm bảo có thể hoàn tác giao dịch chưa xác nhận khi hàm thoát

	// Tạo người dùng
	result := tx.Create(user)
	if result.Error != nil {
		return result.Error
	}

	// Ghi log đăng ký người dùng mới
	if common.QuotaForNewUser > 0 {
		RecordLog(user.Id, LogTypeSystem, fmt.Sprintf("New user registration bonus %s", common.LogQuota(common.QuotaForNewUser)))
	}

	// Xử lý logic mời - kiểm tra chức năng mời có được bật không
	if inviterId != 0 && common.AffEnabled {
		// Tăng hạn mức cho người được mời (nếu đã cấu hình phần thưởng)
		if common.QuotaForInvitee > 0 {
			err = tx.Model(&User{}).Where("id = ?", user.Id).Update("quota", gorm.Expr("quota + ?", common.QuotaForInvitee)).Error
			if err != nil {
				return err
			}
			RecordLog(user.Id, LogTypeSystem, fmt.Sprintf("Invitation code bonus %s", common.LogQuota(common.QuotaForInvitee)))
		}

		// Luôn cập nhật thống kê người mời, bất kể phần thưởng có bằng 0 hay không
		updateFields := map[string]interface{}{
			"aff_count": gorm.Expr("aff_count + ?", 1),
		}

		// Chỉ cập nhật các trường liên quan đến thưởng khi phần thưởng lớn hơn 0
		if common.QuotaForInviter > 0 {
			updateFields["aff_quota"] = gorm.Expr("aff_quota + ?", common.QuotaForInviter)
			updateFields["aff_history"] = gorm.Expr("aff_history + ?", common.QuotaForInviter)
		}

		// Cập nhật nguyên tử thông tin người mời trong giao dịch
		result := tx.Model(&User{}).Where("id = ?", inviterId).Updates(updateFields)

		if result.Error != nil {
			return result.Error
		}

		// Kiểm tra xem có tìm thấy người dùng mời không
		if result.RowsAffected == 0 {
			return errors.New("Inviter does not exist")
		}

		// Ghi log (dù phần thưởng bằng 0 vẫn ghi sự kiện mời)
		if common.QuotaForInviter > 0 {
			RecordLog(inviterId, LogTypeSystem, fmt.Sprintf("Invitation bonus %s", common.LogQuota(common.QuotaForInviter)))
		} else {
			RecordLog(inviterId, LogTypeSystem, "Successfully invited a new user to register")
		}
	}

	// Xác nhận giao dịch
	return tx.Commit().Error
}

func (user *User) Update(updatePassword bool) error {
	var err error
	if updatePassword {
		user.Password, err = common.Password2Hash(user.Password)
		if err != nil {
			return err
		}
	}
	newUser := *user
	DB.First(&user, user.Id)
	if err = DB.Model(user).Updates(newUser).Error; err != nil {
		return err
	}

	// Update cache
	return updateUserCache(*user)
}

func (user *User) Edit(updatePassword bool) error {
	var err error
	if updatePassword {
		user.Password, err = common.Password2Hash(user.Password)
		if err != nil {
			return err
		}
	}

	newUser := *user
	updates := map[string]interface{}{
		"username":     newUser.Username,
		"display_name": newUser.DisplayName,
		"group":        newUser.Group,
		"quota":        newUser.Quota,
	}
	if updatePassword {
		updates["password"] = newUser.Password
	}

	DB.First(&user, user.Id)
	if err = DB.Model(user).Updates(updates).Error; err != nil {
		return err
	}

	// Update cache
	return updateUserCache(*user)
}

func (user *User) Delete() error {
	if user.Id == 0 {
		return errors.New("id is empty!")
	}
	if err := DB.Delete(user).Error; err != nil {
		return err
	}

	// Xóa bộ nhớ đệm
	return invalidateUserCache(user.Id)
}

func (user *User) HardDelete() error {
	if user.Id == 0 {
		return errors.New("id is empty!")
	}
	err := DB.Unscoped().Delete(user).Error
	return err
}

// ValidateAndFill check password & user status
func (user *User) ValidateAndFill() (err error) {
	// When querying with struct, GORM will only query with non-zero fields,
	// that means if your field's value is 0, '', false or other zero values,
	// it won't be used to build query conditions
	password := user.Password
	username := strings.TrimSpace(user.Username)
	if username == "" || password == "" {
		return errors.New("Username or password is empty")
	}
	// find buy username or email
	DB.Where("username = ? OR email = ?", username, username).First(user)
	okay := common.ValidatePasswordAndHash(password, user.Password)
	if !okay || user.Status != common.UserStatusEnabled {
		return errors.New("Incorrect username or password, or the user has been banned")
	}
	return nil
}

func (user *User) FillUserById() error {
	if user.Id == 0 {
		return errors.New("id is empty!")
	}
	DB.Where(User{Id: user.Id}).First(user)
	return nil
}

func (user *User) FillUserByEmail() error {
	if user.Email == "" {
		return errors.New("email is empty!")
	}
	DB.Where(User{Email: user.Email}).First(user)
	return nil
}

func (user *User) FillUserByGitHubId() error {
	if user.GitHubId == "" {
		return errors.New("GitHub id is empty!")
	}
	DB.Where(User{GitHubId: user.GitHubId}).First(user)
	return nil
}

func (user *User) FillUserByOidcId() error {
	if user.OidcId == "" {
		return errors.New("OIDC id is empty!")
	}
	DB.Where(User{OidcId: user.OidcId}).First(user)
	return nil
}

func (user *User) FillUserByWeChatId() error {
	if user.WeChatId == "" {
		return errors.New("WeChat id is empty!")
	}
	DB.Where(User{WeChatId: user.WeChatId}).First(user)
	return nil
}

func (user *User) FillUserByTelegramId() error {
	if user.TelegramId == "" {
		return errors.New("Telegram id is empty!")
	}
	err := DB.Where(User{TelegramId: user.TelegramId}).First(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("This Telegram account is not bound")
	}
	return nil
}

func IsEmailAlreadyTaken(email string) bool {
	return DB.Unscoped().Where("email = ?", email).Find(&User{}).RowsAffected == 1
}

func IsWeChatIdAlreadyTaken(wechatId string) bool {
	return DB.Unscoped().Where("wechat_id = ?", wechatId).Find(&User{}).RowsAffected == 1
}

func IsGitHubIdAlreadyTaken(githubId string) bool {
	return DB.Unscoped().Where("github_id = ?", githubId).Find(&User{}).RowsAffected == 1
}

func IsOidcIdAlreadyTaken(oidcId string) bool {
	return DB.Where("oidc_id = ?", oidcId).Find(&User{}).RowsAffected == 1
}

func IsTelegramIdAlreadyTaken(telegramId string) bool {
	return DB.Unscoped().Where("telegram_id = ?", telegramId).Find(&User{}).RowsAffected == 1
}

func ResetUserPasswordByEmail(email string, password string) error {
	if email == "" || password == "" {
		return errors.New("Email address or password is empty!")
	}
	hashedPassword, err := common.Password2Hash(password)
	if err != nil {
		return err
	}
	err = DB.Model(&User{}).Where("email = ?", email).Update("password", hashedPassword).Error
	return err
}

func IsAdmin(userId int) bool {
	if userId == 0 {
		return false
	}
	var user User
	err := DB.Where("id = ?", userId).Select("role").Find(&user).Error
	if err != nil {
		common.SysError("no such user " + err.Error())
		return false
	}
	return user.Role >= common.RoleAdminUser
}

//// IsUserEnabled checks user status from Redis first, falls back to DB if needed
//func IsUserEnabled(id int, fromDB bool) (status bool, err error) {
//	defer func() {
//		// Update Redis cache asynchronously on successful DB read
//		if shouldUpdateRedis(fromDB, err) {
//			gopool.Go(func() {
//				if err := updateUserStatusCache(id, status); err != nil {
//					common.SysError("failed to update user status cache: " + err.Error())
//				}
//			})
//		}
//	}()
//	if !fromDB && common.RedisEnabled {
//		// Try Redis first
//		status, err := getUserStatusCache(id)
//		if err == nil {
//			return status == common.UserStatusEnabled, nil
//		}
//		// Don't return error - fall through to DB
//	}
//	fromDB = true
//	var user User
//	err = DB.Where("id = ?", id).Select("status").Find(&user).Error
//	if err != nil {
//		return false, err
//	}
//
//	return user.Status == common.UserStatusEnabled, nil
//}

func ValidateAccessToken(token string) (user *User) {
	if token == "" {
		return nil
	}
	token = strings.Replace(token, "Bearer ", "", 1)
	user = &User{}
	if DB.Where("access_token = ?", token).First(user).RowsAffected == 1 {
		return user
	}
	return nil
}

// GetUserQuota gets quota from Redis first, falls back to DB if needed
func GetUserQuotaBalance(id int, fromDB bool) (balance UserQuotaBalance, err error) {
	defer func() {
		if shouldUpdateRedis(fromDB, err) {
			gopool.Go(func() {
				if cacheErr := updateUserQuotaCache(id, balance.Quota); cacheErr != nil {
					common.SysError("failed to update user quota cache: " + cacheErr.Error())
				}
				if cacheErr := updateUserSubscriptionQuotaCache(id, balance.SubscriptionQuota); cacheErr != nil {
					common.SysError("failed to update user subscription quota cache: " + cacheErr.Error())
				}
			})
		}
	}()
	if !fromDB && common.RedisEnabled {
		if cache, cacheErr := GetUserCache(id); cacheErr == nil {
			return UserQuotaBalance{Quota: cache.Quota, SubscriptionQuota: cache.SubscriptionQuota}, nil
		}
	}
	fromDB = true
	var result struct {
		Quota             int
		SubscriptionQuota int
	}
	err = DB.Model(&User{}).Where("id = ?", id).Select("quota", "subscription_quota").Take(&result).Error
	if err != nil {
		return UserQuotaBalance{}, err
	}
	return UserQuotaBalance{Quota: result.Quota, SubscriptionQuota: result.SubscriptionQuota}, nil
}

func GetUserQuota(id int, fromDB bool) (quota int, err error) {
	balance, err := GetUserQuotaBalance(id, fromDB)
	if err != nil {
		return 0, err
	}
	return balance.Total(), nil
}

func GetUserUsedQuota(id int) (quota int, err error) {
	err = DB.Model(&User{}).Where("id = ?", id).Select("used_quota").Find(&quota).Error
	return quota, err
}

func GetUserEmail(id int) (email string, err error) {
	err = DB.Model(&User{}).Where("id = ?", id).Select("email").Find(&email).Error
	return email, err
}

// GetUserGroup gets group from Redis first, falls back to DB if needed
func GetUserGroup(id int, fromDB bool) (group string, err error) {
	defer func() {
		// Update Redis cache asynchronously on successful DB read
		if shouldUpdateRedis(fromDB, err) {
			gopool.Go(func() {
				if err := updateUserGroupCache(id, group); err != nil {
					common.SysError("failed to update user group cache: " + err.Error())
				}
			})
		}
	}()
	if !fromDB && common.RedisEnabled {
		group, err := getUserGroupCache(id)
		if err == nil {
			return group, nil
		}
		// Don't return error - fall through to DB
	}
	fromDB = true
	err = DB.Model(&User{}).Where("id = ?", id).Select(groupCol).Find(&group).Error
	if err != nil {
		return "", err
	}

	return group, nil
}

// GetUserSetting gets setting from Redis first, falls back to DB if needed
func GetUserSetting(id int, fromDB bool) (settingMap map[string]interface{}, err error) {
	var setting string
	defer func() {
		// Update Redis cache asynchronously on successful DB read
		if shouldUpdateRedis(fromDB, err) {
			gopool.Go(func() {
				if err := updateUserSettingCache(id, setting); err != nil {
					common.SysError("failed to update user setting cache: " + err.Error())
				}
			})
		}
	}()
	if !fromDB && common.RedisEnabled {
		setting, err := getUserSettingCache(id)
		if err == nil {
			return setting, nil
		}
		// Don't return error - fall through to DB
	}
	fromDB = true
	err = DB.Model(&User{}).Where("id = ?", id).Select("setting").Find(&setting).Error
	if err != nil {
		return map[string]interface{}{}, err
	}

	return common.StrToMap(setting), nil
}

// GetUserShowIPInLogs gets the IP logging preference for a user by ID
func GetUserShowIPInLogs(id int, fromDB bool) (bool, error) {
	settings, err := GetUserSetting(id, fromDB)
	if err != nil {
		return false, err
	}

	if showIP, exists := settings["show_ip_in_logs"]; exists {
		if boolVal, ok := showIP.(bool); ok {
			return boolVal, nil
		}
	}
	return false, nil
}

func IncreaseUserQuota(id int, quota int, db bool) (err error) {
	if quota < 0 {
		return errors.New("quota cannot be negative!")
	}
	gopool.Go(func() {
		err := cacheIncrUserQuota(id, int64(quota))
		if err != nil {
			common.SysError("failed to increase user quota: " + err.Error())
		}
	})
	if !db && common.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeUserQuota, id, quota)
		return nil
	}
	return increaseUserQuota(id, quota)
}

func increaseUserQuota(id int, quota int) (err error) {
	err = DB.Model(&User{}).Where("id = ?", id).Update("quota", gorm.Expr("quota + ?", quota)).Error
	if err != nil {
		return err
	}
	return err
}

func DecreaseUserQuota(id int, quota int) (err error) {
	if quota < 0 {
		return errors.New("quota cannot be negative!")
	}
	gopool.Go(func() {
		err := cacheDecrUserQuota(id, int64(quota))
		if err != nil {
			common.SysError("failed to decrease user quota: " + err.Error())
		}
	})
	if common.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeUserQuota, id, -quota)
		return nil
	}
	return decreaseUserQuota(id, quota)
}

func decreaseUserQuota(id int, quota int) (err error) {
	err = DB.Model(&User{}).Where("id = ?", id).Update("quota", gorm.Expr("quota - ?", quota)).Error
	if err != nil {
		return err
	}
	return err
}

func DeltaUpdateUserQuota(id int, delta int) (err error) {
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		return IncreaseUserQuota(id, delta, false)
	} else {
		return DecreaseUserQuota(id, -delta)
	}
}

func ConsumeUserQuota(id int, amount int) (subscriptionUsed int, quotaUsed int, err error) {
	if amount < 0 {
		return 0, 0, errors.New("quota cannot be negative!")
	}
	if amount == 0 {
		return 0, 0, nil
	}
	tx := DB.Begin()
	if tx.Error != nil {
		return 0, 0, tx.Error
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	var snapshot struct {
		SubscriptionQuota int
		Quota             int
	}
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&User{}).
		Where("id = ?", id).
		Select("subscription_quota", "quota").
		Take(&snapshot).Error
	if err != nil {
		return 0, 0, err
	}
	total := snapshot.SubscriptionQuota + snapshot.Quota
	if total < amount {
		err = fmt.Errorf("user quota is not enough, user quota: %s, need quota: %s", common.FormatQuota(total), common.FormatQuota(amount))
		return 0, 0, err
	}
	subscriptionUsed = amount
	if subscriptionUsed > snapshot.SubscriptionQuota {
		subscriptionUsed = snapshot.SubscriptionQuota
	}
	quotaUsed = amount - subscriptionUsed
	if subscriptionUsed > 0 {
		err = tx.Model(&User{}).Where("id = ?", id).
			Update("subscription_quota", gorm.Expr("subscription_quota - ?", subscriptionUsed)).Error
		if err != nil {
			return 0, 0, err
		}
	}
	if quotaUsed > 0 {
		err = tx.Model(&User{}).Where("id = ?", id).
			Update("quota", gorm.Expr("quota - ?", quotaUsed)).Error
		if err != nil {
			return 0, 0, err
		}
	}
	err = tx.Commit().Error
	if err != nil {
		return 0, 0, err
	}
	if subscriptionUsed > 0 {
		gopool.Go(func() {
			if cacheErr := cacheDecrUserSubscriptionQuota(id, int64(subscriptionUsed)); cacheErr != nil {
				common.SysError("failed to decrease user subscription quota: " + cacheErr.Error())
			}
		})
	}
	if quotaUsed > 0 {
		gopool.Go(func() {
			if cacheErr := cacheDecrUserQuota(id, int64(quotaUsed)); cacheErr != nil {
				common.SysError("failed to decrease user quota: " + cacheErr.Error())
			}
		})
	}
	return subscriptionUsed, quotaUsed, nil
}

func RestoreUserQuota(id int, subscriptionAmount int, quotaAmount int) (err error) {
	if subscriptionAmount < 0 || quotaAmount < 0 {
		return errors.New("quota cannot be negative!")
	}
	if subscriptionAmount == 0 && quotaAmount == 0 {
		return nil
	}
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	if subscriptionAmount > 0 {
		err = tx.Model(&User{}).Where("id = ?", id).
			Update("subscription_quota", gorm.Expr("subscription_quota + ?", subscriptionAmount)).Error
		if err != nil {
			return err
		}
	}
	if quotaAmount > 0 {
		err = tx.Model(&User{}).Where("id = ?", id).
			Update("quota", gorm.Expr("quota + ?", quotaAmount)).Error
		if err != nil {
			return err
		}
	}
	err = tx.Commit().Error
	if err != nil {
		return err
	}
	if subscriptionAmount > 0 {
		gopool.Go(func() {
			if cacheErr := cacheIncrUserSubscriptionQuota(id, int64(subscriptionAmount)); cacheErr != nil {
				common.SysError("failed to increase user subscription quota: " + cacheErr.Error())
			}
		})
	}
	if quotaAmount > 0 {
		gopool.Go(func() {
			if cacheErr := cacheIncrUserQuota(id, int64(quotaAmount)); cacheErr != nil {
				common.SysError("failed to increase user quota: " + cacheErr.Error())
			}
		})
	}
	return nil
}

//func GetRootUserEmail() (email string) {
//	DB.Model(&User{}).Where("role = ?", common.RoleRootUser).Select("email").Find(&email)
//	return email
//}

func GetRootUser() (user *User) {
	DB.Where("role = ?", common.RoleRootUser).First(&user)
	return user
}

func UpdateUserUsedQuotaAndRequestCount(id int, quota int) {
	if common.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeUsedQuota, id, quota)
		addNewRecord(BatchUpdateTypeRequestCount, id, 1)
		return
	}
	updateUserUsedQuotaAndRequestCount(id, quota, 1)
}

func updateUserUsedQuotaAndRequestCount(id int, quota int, count int) {
	err := DB.Model(&User{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"used_quota":    gorm.Expr("used_quota + ?", quota),
			"request_count": gorm.Expr("request_count + ?", count),
		},
	).Error
	if err != nil {
		common.SysError("failed to update user used quota and request count: " + err.Error())
		return
	}

	//// Cập nhật bộ nhớ đệm
	//if err := invalidateUserCache(id); err != nil {
	//	common.SysError("failed to invalidate user cache: " + err.Error())
	//}
}

func updateUserUsedQuota(id int, quota int) {
	err := DB.Model(&User{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"used_quota": gorm.Expr("used_quota + ?", quota),
		},
	).Error
	if err != nil {
		common.SysError("failed to update user used quota: " + err.Error())
	}
}

func updateUserRequestCount(id int, count int) {
	err := DB.Model(&User{}).Where("id = ?", id).Update("request_count", gorm.Expr("request_count + ?", count)).Error
	if err != nil {
		common.SysError("failed to update user request count: " + err.Error())
	}
}

// GetUsernameById gets username from Redis first, falls back to DB if needed
func GetUsernameById(id int, fromDB bool) (username string, err error) {
	defer func() {
		// Update Redis cache asynchronously on successful DB read
		if shouldUpdateRedis(fromDB, err) {
			gopool.Go(func() {
				if err := updateUserNameCache(id, username); err != nil {
					common.SysError("failed to update user name cache: " + err.Error())
				}
			})
		}
	}()
	if !fromDB && common.RedisEnabled {
		username, err := getUserNameCache(id)
		if err == nil {
			return username, nil
		}
		// Don't return error - fall through to DB
	}
	fromDB = true
	err = DB.Model(&User{}).Where("id = ?", id).Select("username").Find(&username).Error
	if err != nil {
		return "", err
	}

	return username, nil
}

func IsIDCFlareIdAlreadyTaken(idcFlareId string) bool {
	var user User
	err := DB.Unscoped().Where("idc_flare_id = ?", idcFlareId).First(&user).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (user *User) FillUserByIDCFlareId() error {
	if user.IDCFlareId == "" {
		return errors.New("idc flare id is empty")
	}
	err := DB.Where("idc_flare_id = ?", user.IDCFlareId).First(user).Error
	return err
}

func IsLinuxDOIdAlreadyTaken(linuxDOId string) bool {
	var user User
	err := DB.Unscoped().Where("linux_do_id = ?", linuxDOId).First(&user).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (user *User) FillUserByLinuxDOId() error {
	if user.LinuxDOId == "" {
		return errors.New("linux do id is empty")
	}
	err := DB.Where("linux_do_id = ?", user.LinuxDOId).First(user).Error
	return err
}

func RootUserExists() bool {
	var user User
	err := DB.Where("role = ?", common.RoleRootUser).First(&user).Error
	if err != nil {
		return false
	}
	return true
}

// CheckIn performs a check-in for the user and rewards them with tokens
func (user *User) CheckIn(minQuota int, maxQuota int) (rewarded int, err error) {
	// Get current transaction and perform initial checks
	tx, err := beginCheckInTransaction(user.Id)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Check eligibility and calculate reward
	reward, canCheckIn, err := calculateCheckInReward(user, minQuota, maxQuota)
	if err != nil {
		return 0, err
	}

	if !canCheckIn {
		return 0, errors.New("You have already checked in today")
	}

	// Apply check-in reward and save to database
	if err = applyAndSaveCheckIn(tx, user, reward); err != nil {
		return 0, err
	}

	// Commit transaction and update cache
	if err = finalizeCheckIn(tx, user.Id, user.Quota); err != nil {
		return 0, err
	}

	return reward, nil
}

// beginCheckInTransaction starts a transaction and locks the user row
func beginCheckInTransaction(userId int) (*gorm.DB, error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Lock the user row for update
	var user User
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, err
	}

	return tx, nil
}

// calculateCheckInReward checks if user can check in and calculates the reward
func calculateCheckInReward(user *User, minQuota int, maxQuota int) (reward int, canCheckIn bool, err error) {
	// Check if user can check in today
	canCheckIn = true
	nowUTC := time.Now().UTC() // Get current time in UTC

	if user.LastCheckInTime != nil {
		lastCheckIn := *user.LastCheckInTime
		lastCheckInUTC := lastCheckIn.UTC() // Convert stored time to UTC for comparison
		// Compare Year, Month, Day in UTC
		if lastCheckInUTC.Year() == nowUTC.Year() && lastCheckInUTC.Month() == nowUTC.Month() && lastCheckInUTC.Day() == nowUTC.Day() {
			canCheckIn = false
			return 0, false, nil // Already checked in today (UTC)
		}
	}

	// Calculate reward amount
	reward = minQuota
	if maxQuota > minQuota {
		// Generate random reward between min and max
		reward = minQuota + common.RandomInt(maxQuota-minQuota+1)
	}

	return reward, true, nil
}

// applyAndSaveCheckIn applies the check-in reward to the user and saves changes
func applyAndSaveCheckIn(tx *gorm.DB, user *User, reward int) error {
	// Update user data
	now := time.Now()
	user.LastCheckInTime = &now
	user.Quota += reward

	// Save changes
	if err := tx.Save(user).Error; err != nil {
		return err
	}

	// Record this activity in log
	RecordLog(user.Id, LogTypeCheckIn, fmt.Sprintf("Check-in reward %s", common.LogQuota(reward)))

	return nil
}

// finalizeCheckIn commits the transaction and updates the user cache
func finalizeCheckIn(tx *gorm.DB, userId int, quota int) error {
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Update cache
	if err := updateUserQuotaCache(userId, quota); err != nil {
		common.SysError("failed to update user quota cache after check-in: " + err.Error())
	}

	return nil
}

// CanCheckInToday checks if the user can check in today based on UTC date
func (user *User) CanCheckInToday() bool {
	if user.LastCheckInTime == nil {
		return true
	}

	nowUTC := time.Now().UTC()
	lastCheckIn := *user.LastCheckInTime
	lastCheckInUTC := lastCheckIn.UTC() // Convert stored time to UTC

	// Compare Year, Month, Day in UTC
	alreadyCheckedInToday := lastCheckInUTC.Year() == nowUTC.Year() &&
		lastCheckInUTC.Month() == nowUTC.Month() &&
		lastCheckInUTC.Day() == nowUTC.Day()

	return !alreadyCheckedInToday
}
