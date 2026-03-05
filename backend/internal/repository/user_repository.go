package repository

import (
	"time"
	"voicepaper/internal/model"
)

// UserRepository 用户相关的数据库操作
type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return DB.Create(user).Error
}

// FindByID 根据ID查找用户
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := DB.Unscoped().First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByPhone 根据手机号查找用户
func (r *UserRepository) FindByPhone(phone string) (*model.User, error) {
	var user model.User
	err := DB.Unscoped().Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := DB.Unscoped().Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByGitHubID 根据GitHub ID查找用户
func (r *UserRepository) FindByGitHubID(githubID string) (*model.User, error) {
	var user model.User
	err := DB.Unscoped().Where("github_id = ?", githubID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByGoogleID 根据Google ID查找用户
func (r *UserRepository) FindByGoogleID(googleID string) (*model.User, error) {
	var user model.User
	err := DB.Unscoped().Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByInviteCode 根据邀请码查找用户
func (r *UserRepository) FindByInviteCode(inviteCode string) (*model.User, error) {
	var user model.User
	err := DB.Unscoped().Where("invite_code = ?", inviteCode).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *UserRepository) Update(user *model.User) error {
	return DB.Save(user).Error
}

// UpdateLastLogin 更新最后登录时间和IP
func (r *UserRepository) UpdateLastLogin(userID uint, ip string) error {
	now := time.Now()
	return DB.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": ip,
	}).Error
}

// VerificationCodeRepository 验证码相关的数据库操作
type VerificationCodeRepository struct{}

func NewVerificationCodeRepository() *VerificationCodeRepository {
	return &VerificationCodeRepository{}
}

// Create 创建验证码
func (r *VerificationCodeRepository) Create(code *model.VerificationCode) error {
	return DB.Create(code).Error
}

// FindValidCode 查找有效的验证码
func (r *VerificationCodeRepository) FindValidCode(receiver, codeType, code string) (*model.VerificationCode, error) {
	var vc model.VerificationCode
	err := DB.Where("receiver = ? AND code_type = ? AND code = ? AND used = ? AND expires_at > ?",
		receiver, codeType, code, false, time.Now()).
		Order("created_at DESC").
		First(&vc).Error
	if err != nil {
		return nil, err
	}
	return &vc, nil
}

// MarkAsUsed 标记验证码为已使用
func (r *VerificationCodeRepository) MarkAsUsed(id uint) error {
	now := time.Now()
	return DB.Model(&model.VerificationCode{}).Where("id = ?", id).Updates(map[string]interface{}{
		"used":    true,
		"used_at": now,
	}).Error
}

// CountRecentCodes 统计最近发送的验证码数量（防刷）
func (r *VerificationCodeRepository) CountRecentCodes(receiver, codeType string, minutes int) (int64, error) {
	var count int64
	cutoffTime := time.Now().Add(-time.Duration(minutes) * time.Minute)
	err := DB.Model(&model.VerificationCode{}).
		Where("receiver = ? AND code_type = ? AND created_at > ?", receiver, codeType, cutoffTime).
		Count(&count).Error
	return count, err
}

// CleanExpiredCodes 清理过期验证码（定期任务）
func (r *VerificationCodeRepository) CleanExpiredCodes() error {
	return DB.Where("expires_at < ? OR (used = ? AND used_at < ?)",
		time.Now(), true, time.Now().Add(-24*time.Hour)).
		Delete(&model.VerificationCode{}).Error
}

// DeleteByReceiver 删除指定接收方的所有验证码（用于测试）
func (r *VerificationCodeRepository) DeleteByReceiver(receiver, codeType string) error {
	return DB.Where("receiver = ? AND code_type = ?", receiver, codeType).
		Delete(&model.VerificationCode{}).Error
}

// OAuthBindingRepository OAuth绑定相关的数据库操作
type OAuthBindingRepository struct{}

func NewOAuthBindingRepository() *OAuthBindingRepository {
	return &OAuthBindingRepository{}
}

// Create 创建OAuth绑定
func (r *OAuthBindingRepository) Create(binding *model.OAuthBinding) error {
	return DB.Create(binding).Error
}

// FindByProviderUserID 根据提供商和用户ID查找绑定
func (r *OAuthBindingRepository) FindByProviderUserID(provider, providerUserID string) (*model.OAuthBinding, error) {
	var binding model.OAuthBinding
	err := DB.Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		Preload("User").
		First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// FindByUserID 根据用户ID查找所有绑定
func (r *OAuthBindingRepository) FindByUserID(userID uint) ([]model.OAuthBinding, error) {
	var bindings []model.OAuthBinding
	err := DB.Where("user_id = ?", userID).Find(&bindings).Error
	return bindings, err
}

// Update 更新OAuth绑定
func (r *OAuthBindingRepository) Update(binding *model.OAuthBinding) error {
	return DB.Save(binding).Error
}

// UserSessionRepository 用户会话相关的数据库操作
type UserSessionRepository struct{}

func NewUserSessionRepository() *UserSessionRepository {
	return &UserSessionRepository{}
}

// Create 创建会话
func (r *UserSessionRepository) Create(session *model.UserSession) error {
	return DB.Create(session).Error
}

// FindByToken 根据令牌查找会话
func (r *UserSessionRepository) FindByToken(token string) (*model.UserSession, error) {
	var session model.UserSession
	err := DB.Where("session_token = ? AND expires_at > ?", token, time.Now()).
		Preload("User").
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteByToken 根据令牌删除会话
func (r *UserSessionRepository) DeleteByToken(token string) error {
	return DB.Where("session_token = ?", token).Delete(&model.UserSession{}).Error
}

// DeleteByUserID 删除用户的所有会话
func (r *UserSessionRepository) DeleteByUserID(userID uint) error {
	return DB.Where("user_id = ?", userID).Delete(&model.UserSession{}).Error
}

// CleanExpiredSessions 清理过期会话
func (r *UserSessionRepository) CleanExpiredSessions() error {
	return DB.Where("expires_at < ?", time.Now()).Delete(&model.UserSession{}).Error
}

