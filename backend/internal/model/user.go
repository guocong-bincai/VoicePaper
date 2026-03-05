package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
// 对应数据库表 vp_users
func (User) TableName() string {
	return "vp_users"
}

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	// 用户基本信息
	Nickname string `gorm:"size:100" json:"nickname"` // 昵称
	Avatar   string `gorm:"size:512" json:"avatar"`   // 头像URL
	Bio      string `gorm:"size:500" json:"bio"`      // 个人简介

	// 手机号登录（可选）
	Phone           *string    `gorm:"size:20;uniqueIndex" json:"phone,omitempty"` // 手机号（加密存储，使用指针允许NULL）
	PhoneVerified   bool       `gorm:"default:false;column:phone_verified" json:"phone_verified"`
	PhoneVerifiedAt *time.Time `gorm:"column:phone_verified_at" json:"phone_verified_at"`

	// 邮箱登录（可选）
	Email           *string    `gorm:"size:255;uniqueIndex" json:"email,omitempty"` // 邮箱地址
	EmailVerified   bool       `gorm:"default:false;column:email_verified" json:"email_verified"`
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at" json:"email_verified_at"`

	// OAuth登录（GitHub/Google）
	GitHubID       *string `gorm:"size:100;uniqueIndex;column:github_id" json:"github_id,omitempty"` // GitHub用户ID（使用指针，允许NULL）
	GitHubUsername string  `gorm:"size:100;column:github_username" json:"github_username,omitempty"` // GitHub用户名
	GoogleID       *string `gorm:"size:100;uniqueIndex;column:google_id" json:"google_id,omitempty"` // Google用户ID（使用指针，允许NULL）
	GoogleEmail    string  `gorm:"size:255;column:google_email" json:"google_email,omitempty"`       // Google邮箱

	// WeChat Login
	WechatOpenID  *string `gorm:"size:100;uniqueIndex;column:wechat_openid" json:"wechat_openid,omitempty"`   // 微信OpenID
	WechatUnionID *string `gorm:"size:100;uniqueIndex;column:wechat_unionid" json:"wechat_unionid,omitempty"` // 微信UnionID

	// 用户状态
	Status      string     `gorm:"size:20;default:'active';index" json:"status"` // 状态：active/inactive/banned
	LastLoginAt *time.Time `gorm:"column:last_login_at" json:"last_login_at"`
	LastLoginIP string     `gorm:"size:45;column:last_login_ip" json:"last_login_ip"`

	// 角色和密码
	Role         string `gorm:"size:20;default:'user';index" json:"role"` // 角色：user(普通用户), admin(管理员)
	PasswordHash string `gorm:"size:255;column:password_hash" json:"-"`   // 密码哈希（bcrypt加密）

	// 邀请码系统
	InviteCode  string `gorm:"size:32;uniqueIndex;column:invite_code" json:"invite_code,omitempty"` // 用户自己的邀请码（用于邀请他人）
	InvitedByID *uint  `gorm:"index;column:invited_by_id" json:"invited_by_id,omitempty"`           // 邀请人ID（谁邀请的当前用户）

	// 学习统计
	TotalDurationMinutes int64 `gorm:"default:0;column:total_duration_minutes" json:"total_duration_minutes"` // 累积学习时长（分钟）

	// 关联
	OAuthBindings []OAuthBinding `gorm:"foreignKey:UserID" json:"oauth_bindings,omitempty"`
}

// VerificationCode 验证码表
// 对应数据库表 vp_verification_codes
func (VerificationCode) TableName() string {
	return "vp_verification_codes"
}

type VerificationCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`

	// 验证码类型和接收方
	CodeType string `gorm:"size:20;not null" json:"code_type"`                                       // 类型：phone/email
	Receiver string `gorm:"size:255;not null;index:idx_verification_codes_receiver" json:"receiver"` // 接收方：手机号或邮箱

	// 验证码信息
	Code      string     `gorm:"size:10;not null;index" json:"code"` // 验证码（6位数字）
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`   // 过期时间
	Used      bool       `gorm:"default:false;index" json:"used"`    // 是否已使用
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at"`      // 使用时间

	// 用途
	Purpose string `gorm:"size:20;default:'login'" json:"purpose"` // 用途：login/register/reset_password

	// IP和UA（防刷）
	IPAddress string `gorm:"size:45;column:ip_address" json:"ip_address"`  // 请求IP
	UserAgent string `gorm:"size:500;column:user_agent" json:"user_agent"` // User Agent
}

// OAuthBinding OAuth绑定表
// 对应数据库表 vp_oauth_bindings
func (OAuthBinding) TableName() string {
	return "vp_oauth_bindings"
}

type OAuthBinding struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	UserID uint `gorm:"not null;index" json:"user_id"` // 关联users.id

	Provider         string `gorm:"size:20;not null;uniqueIndex:idx_oauth_bindings_provider_user" json:"provider"`          // 提供商：github/google/wechat
	ProviderUserID   string `gorm:"size:100;not null;uniqueIndex:idx_oauth_bindings_provider_user" json:"provider_user_id"` // 第三方用户ID
	ProviderUsername string `gorm:"size:100" json:"provider_username"`                                                      // 第三方用户名
	ProviderEmail    string `gorm:"size:255" json:"provider_email"`                                                         // 第三方邮箱

	AccessToken    string     `gorm:"type:text" json:"-"`                              // 访问令牌（加密存储）
	RefreshToken   string     `gorm:"type:text" json:"-"`                              // 刷新令牌（加密存储）
	TokenExpiresAt *time.Time `gorm:"column:token_expires_at" json:"token_expires_at"` // 令牌过期时间

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// UserSession 用户会话表
// 对应数据库表 vp_user_sessions
func (UserSession) TableName() string {
	return "vp_user_sessions"
}

type UserSession struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`

	UserID       uint   `gorm:"not null;index" json:"user_id"`                      // 关联users.id
	SessionToken string `gorm:"size:128;not null;uniqueIndex" json:"session_token"` // 会话令牌
	IPAddress    string `gorm:"size:45;column:ip_address" json:"ip_address"`        // IP地址
	UserAgent    string `gorm:"size:500;column:user_agent" json:"user_agent"`       // User Agent
	DeviceInfo   string `gorm:"size:255;column:device_info" json:"device_info"`     // 设备信息

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}
