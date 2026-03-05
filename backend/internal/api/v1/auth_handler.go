package v1

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"voicepaper/config"
	"voicepaper/internal/repository"
	"voicepaper/internal/service"
	"voicepaper/internal/storage"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService  *service.AuthService
	emailService *service.EmailService
	oauthService *service.OAuthService
	storage      storage.Storage
}

func NewAuthHandler() *AuthHandler {
	cfg := config.GetConfig()

	// 初始化存储（OSS或本地）
	var st storage.Storage
	var err error
	if cfg.Storage.Type == "oss" {
		st, err = storage.NewOSSStorage(cfg)
		if err != nil {
			fmt.Printf("⚠️  OSS存储初始化失败，将使用本地存储: %v\n", err)
			st = storage.NewLocalStorage(cfg)
		}
	} else {
		st = storage.NewLocalStorage(cfg)
	}

	userRepo := repository.NewUserRepository()
	sessionRepo := repository.NewUserSessionRepository()
	codeRepo := repository.NewVerificationCodeRepository()
	bindingRepo := repository.NewOAuthBindingRepository()

	emailService := service.NewEmailService(&cfg.Auth.Email, codeRepo)
	oauthService := service.NewOAuthService(&cfg.Auth.OAuth, userRepo, bindingRepo)
	authService := service.NewAuthService(&cfg.Auth, userRepo, sessionRepo, emailService, oauthService)

	return &AuthHandler{
		authService:  authService,
		emailService: emailService,
		oauthService: oauthService,
		storage:      st,
	}
}

// SendEmailCodeRequest 发送邮箱验证码请求
type SendEmailCodeRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Purpose string `json:"purpose" binding:"omitempty,oneof=login register reset_password"` // 默认login
}

// SendEmailCode 发送邮箱验证码
// POST /api/v1/auth/email/send
func (h *AuthHandler) SendEmailCode(c *gin.Context) {
	var req SendEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// Trim空格并转小写，避免苹果输入法自动大写导致验证失败
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if req.Purpose == "" {
		req.Purpose = "login"
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 发送验证码（即使邮件发送失败，验证码也已经保存到数据库）
	err := h.emailService.SendVerificationCode(email, ipAddress, userAgent, req.Purpose)
	if err != nil {
		// 检查是否是邮件发送失败（验证码已保存）
		// 如果是邮件发送失败，仍然返回成功，因为验证码已经保存到数据库
		// 用户可以从数据库或日志中获取验证码进行测试
		if err.Error() != "" && (strings.Contains(err.Error(), "发送邮件失败") || strings.Contains(err.Error(), "SMTP")) {
			// 邮件发送失败，但验证码已保存，返回成功
			c.JSON(http.StatusOK, gin.H{
				"message": "验证码已生成（邮件发送失败，请从数据库查看验证码）",
				"email":   email,
				"warning": "邮件发送失败，但验证码已保存到数据库",
			})
			return
		}
		// 其他错误（如防刷限制），返回错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送验证码失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "验证码已发送，请查收邮箱",
		"email":   email,
	})
}

// EmailLoginRequest 邮箱登录请求
type EmailLoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// PasswordLoginRequest 密码登录请求
type PasswordLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Code       string `json:"code" binding:"required"`         // 验证码（必填）
	InviteCode string `json:"invite_code" binding:"omitempty"` // 邀请码（可选）
}

// VerifyEmailCodeRequest 验证邮箱验证码请求（用于注册流程）
type VerifyEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// VerifyEmailCode 验证邮箱验证码（用于注册流程，只验证不创建用户）
// POST /api/v1/auth/email/verify
// BUG修复: 注册流程验证邮箱时不应创建用户，应该只验证验证码
// 修复策略: 创建新接口专门用于注册前验证邮箱，只验证验证码不创建用户
// 影响范围: backend/internal/api/v1/auth_handler.go:129-160
// 修复日期: 2025-12-10
func (h *AuthHandler) VerifyEmailCode(c *gin.Context) {
	var req VerifyEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ 验证邮箱验证码请求参数错误: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// Trim空格并转小写
	email := strings.ToLower(strings.TrimSpace(req.Email))
	code := strings.TrimSpace(req.Code)

	fmt.Printf("🔐 收到验证邮箱请求（注册流程）: email=%s, code=%s\n", email, code)

	// 检查验证码（不标记为已使用，注册时才标记）
	// BUG修复: 使用CheckCode而非VerifyCode，避免提前消耗验证码
	// 修复策略: 验证邮箱时只检查有效性，注册时才真正消耗验证码
	// 影响范围: backend/internal/api/v1/auth_handler.go:149-158
	// 修复日期: 2025-12-10
	_, err := h.emailService.CheckCode(email, code)
	if err != nil {
		fmt.Printf("❌ 验证码检查失败: email=%s, code=%s, error=%v\n", email, code, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期", "details": err.Error()})
		return
	}

	fmt.Printf("✅ 验证码检查成功（未消耗）: email=%s\n", email)

	// 检查邮箱是否已注册
	existingUser, err := h.authService.GetUserByEmail(email)
	if err == nil && existingUser != nil {
		fmt.Printf("⚠️  邮箱已被注册: email=%s, user_id=%d\n", req.Email, existingUser.ID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "该邮箱已被注册，请直接登录"})
		return
	}

	// 验证码正确且邮箱未注册，允许继续注册流程
	c.JSON(http.StatusOK, gin.H{
		"message": "验证成功，可以继续注册",
		"email":   email,
	})
}

// EmailLogin 邮箱验证码登录
// POST /api/v1/auth/email/login
func (h *AuthHandler) EmailLogin(c *gin.Context) {
	var req EmailLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ 登录请求参数错误: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// Trim空格并转小写
	email := strings.ToLower(strings.TrimSpace(req.Email))
	code := strings.TrimSpace(req.Code)

	fmt.Printf("🔐 收到登录请求: email=%s, code=%s\n", email, code)
	ipAddress := c.ClientIP()
	user, token, err := h.authService.LoginWithEmail(email, code, ipAddress)
	if err != nil {
		fmt.Printf("❌ 登录失败: email=%s, code=%s, error=%v\n", email, code, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "登录失败", "details": err.Error()})
		return
	}

	fmt.Printf("✅ 登录成功: user_id=%d, email=%s\n", user.ID, user.Email)

	// 头像 URL 处理: avatars 文件夹已设置为公共读,直接返回不带签名的 URL
	avatarURL := user.Avatar

	// 如果头像 URL 不为空且是 OSS 地址,确保使用 HTTPS
	if avatarURL != "" && strings.Contains(avatarURL, "oss-cn-") {
		// 强制使用 HTTPS
		if strings.HasPrefix(avatarURL, "http://") {
			avatarURL = strings.Replace(avatarURL, "http://", "https://", 1)
			fmt.Printf("🔒 Login头像URL转换为HTTPS: user_id=%d\n", user.ID)
		}
		fmt.Printf("✅ Login返回公开头像URL: user_id=%d, url=%s\n", user.ID, avatarURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"nickname": user.Nickname,
			"avatar":   avatarURL,
			"role":     user.Role,
		},
	})
}

// PasswordLogin 密码登录（支持普通用户和管理员）
// POST /api/v1/auth/password/login
func (h *AuthHandler) PasswordLogin(c *gin.Context) {
	var req PasswordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ 密码登录请求参数错误: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// Trim空格，避免手机输入法自动加空格导致登录失败
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)

	fmt.Printf("🔐 收到密码登录请求: email=%s\n", email)
	ipAddress := c.ClientIP()
	user, token, err := h.authService.LoginWithPassword(email, password, ipAddress)
	if err != nil {
		fmt.Printf("❌ 密码登录失败: email=%s, error=%v\n", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "登录失败", "details": err.Error()})
		return
	}

	fmt.Printf("✅ 密码登录成功: user_id=%d, email=%s, role=%s\n", user.ID, user.Email, user.Role)

	// 头像 URL 处理: avatars 文件夹已设置为公共读,直接返回不带签名的 URL
	avatarURL := user.Avatar

	// 如果头像 URL 不为空且是 OSS 地址,确保使用 HTTPS
	if avatarURL != "" && strings.Contains(avatarURL, "oss-cn-") {
		// 强制使用 HTTPS
		if strings.HasPrefix(avatarURL, "http://") {
			avatarURL = strings.Replace(avatarURL, "http://", "https://", 1)
			fmt.Printf("🔒 PasswordLogin头像URL转换为HTTPS: user_id=%d\n", user.ID)
		}
		fmt.Printf("✅ PasswordLogin返回公开头像URL: user_id=%d, url=%s\n", user.ID, avatarURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"nickname": user.Nickname,
			"avatar":   avatarURL,
			"role":     user.Role,
		},
	})
}

// Register 用户注册
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ 注册请求参数错误: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// Trim空格并转小写
	email := strings.ToLower(strings.TrimSpace(req.Email))
	code := strings.TrimSpace(req.Code)
	inviteCode := strings.TrimSpace(req.InviteCode)

	fmt.Printf("📝 收到注册请求: email=%s, has_invite_code=%v\n", email, inviteCode != "")
	ipAddress := c.ClientIP()
	user, token, err := h.authService.Register(email, req.Password, code, inviteCode, ipAddress)
	if err != nil {
		fmt.Printf("❌ 注册失败: email=%s, error=%v\n", email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "注册失败", "details": err.Error()})
		return
	}

	fmt.Printf("✅ 注册成功: user_id=%d, email=%s, invite_code=%s\n", user.ID, user.Email, user.InviteCode)
	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
		"token":   token,
		"user": gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"nickname":    user.Nickname,
			"avatar":      user.Avatar,
			"role":        user.Role,
			"invite_code": user.InviteCode,
		},
	})
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required,min=6"`
}

// ResetPassword 重置密码
// POST /api/v1/auth/password/reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// Trim空格并转小写
	email := strings.ToLower(strings.TrimSpace(req.Email))
	code := strings.TrimSpace(req.Code)

	err := h.authService.ResetPassword(email, code, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "重置密码失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功，请使用新密码登录"})
}

// VerifyInviteCodeRequest 验证邀请码请求
type VerifyInviteCodeRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// VerifyInviteCode 验证邀请码（用于前端验证）
// POST /api/v1/auth/invite/verify
func (h *AuthHandler) VerifyInviteCode(c *gin.Context) {
	var req VerifyInviteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ 验证邀请码请求参数错误: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	inviter, err := h.authService.VerifyInviteCode(req.InviteCode)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": "邀请码无效",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "邀请码有效",
		"inviter": gin.H{
			"nickname": inviter.Nickname,
			"email":    inviter.Email,
		},
	})
}

// GitHubAuth 跳转到GitHub授权页面
// GET /api/v1/auth/github
func (h *AuthHandler) GitHubAuth(c *gin.Context) {
	// 生成state（用于防止CSRF攻击）
	state := c.Query("state")
	if state == "" {
		state = "random_state_" + generateRandomString(16)
	}

	authURL := h.oauthService.GetGitHubAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GitHubCallback 处理GitHub回调
// GET /api/v1/auth/github/callback
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	cfg := config.GetConfig()
	frontendURL := cfg.Service.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // 默认值
	}

	code := c.Query("code")
	if code == "" {
		// 重定向到前端首页，显示错误
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/?error=missing_code", frontendURL))
		return
	}

	ctx := context.Background()
	user, err := h.oauthService.HandleGitHubCallback(ctx, code)
	if err != nil {
		// 重定向到前端首页，显示错误
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/?error=oauth_failed", frontendURL))
		return
	}

	// 如果是新用户且没有邀请码，生成邀请码
	if user.InviteCode == "" {
		user.InviteCode = h.authService.GenerateInviteCode(user.ID)
		userRepo := repository.NewUserRepository()
		userRepo.Update(user)
		fmt.Printf("✅ 为新用户生成邀请码: user_id=%d, invite_code=%s\n", user.ID, user.InviteCode)
	}

	// 生成JWT token
	ipAddress := c.ClientIP()
	token, err := h.authService.GenerateJWT(user.ID)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/?error=token_failed", frontendURL))
		return
	}

	// 更新最后登录信息
	userRepo := repository.NewUserRepository()
	userRepo.UpdateLastLogin(user.ID, ipAddress)

	// 重定向到前端首页，携带token
	redirectURL := fmt.Sprintf("%s/?token=%s&provider=github", frontendURL, token)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GoogleAuth 跳转到Google授权页面
// GET /api/v1/auth/google
func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		state = "random_state_" + generateRandomString(16)
	}

	authURL := h.oauthService.GetGoogleAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GoogleCallback 处理Google回调
// GET /api/v1/auth/google/callback
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	cfg := config.GetConfig()
	frontendURL := cfg.Service.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // 默认值
	}

	code := c.Query("code")
	if code == "" {
		// 重定向到前端首页，显示错误
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/?error=missing_code", frontendURL))
		return
	}

	ctx := context.Background()
	fmt.Printf("🔵 收到Google OAuth回调: code=%s\n", code)
	user, err := h.oauthService.HandleGoogleCallback(ctx, code)
	if err != nil {
		fmt.Printf("❌ Google OAuth回调处理失败: %v\n", err)
		// 重定向到前端首页，显示错误（包含详细错误信息用于调试）
		errorMsg := fmt.Sprintf("oauth_failed:%s", err.Error())
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/?error=%s", frontendURL, errorMsg))
		return
	}
	fmt.Printf("✅ Google OAuth回调处理成功: user_id=%d, email=%s\n", user.ID, user.Email)

	// 如果是新用户且没有邀请码，生成邀请码
	if user.InviteCode == "" {
		user.InviteCode = h.authService.GenerateInviteCode(user.ID)
		userRepo := repository.NewUserRepository()
		userRepo.Update(user)
		fmt.Printf("✅ 为新用户生成邀请码: user_id=%d, invite_code=%s\n", user.ID, user.InviteCode)
	}

	// 生成JWT token
	ipAddress := c.ClientIP()
	token, err := h.authService.GenerateJWT(user.ID)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/?error=token_failed", frontendURL))
		return
	}

	// 更新最后登录信息
	userRepo := repository.NewUserRepository()
	userRepo.UpdateLastLogin(user.ID, ipAddress)

	// 重定向到前端首页，携带token
	redirectURL := fmt.Sprintf("%s/?token=%s&provider=google", frontendURL, token)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// WeChatLoginRequest 微信登录请求
type WeChatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// WeChatLogin 微信登录
// POST /api/v1/auth/wechat/login
func (h *AuthHandler) WeChatLogin(c *gin.Context) {
	var req WeChatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 1. code2Session
	session, err := h.oauthService.Code2Session(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "微信登录失败", "details": err.Error()})
		return
	}

	// 2. 登录或注册
	user, err := h.oauthService.LoginWithWechat(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户处理失败", "details": err.Error()})
		return
	}

	// 3. 生成Token
	token, err := h.authService.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败", "details": err.Error()})
		return
	}

	// 4. 更新最后登录信息
	h.authService.UpdateLastLogin(user.ID, c.ClientIP())

	// 5. 返回结果
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user": gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"nickname":    user.Nickname,
			"avatar":      user.Avatar,
			"role":        user.Role,
			"invite_code": user.InviteCode,
		},
	})
}

// BindWeChatRequest 绑定微信请求
type BindWeChatRequest struct {
	Code string `json:"code" binding:"required"`
}

// BindWeChat 绑定微信
// POST /api/v1/auth/wechat/bind
func (h *AuthHandler) BindWeChat(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req BindWeChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 1. code2Session
	session, err := h.oauthService.Code2Session(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "微信授权失败", "details": err.Error()})
		return
	}

	// 2. 绑定
	if err := h.oauthService.BindWeChat(userID.(uint), session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "绑定失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "绑定成功"})
}

// BindEmailRequest 绑定邮箱请求
type BindEmailRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// BindEmail 微信用户绑定邮箱
// POST /api/v1/auth/email/bind
// 场景：微信登录的用户想要绑定已有的邮箱账号，或创建新的邮箱账号
func (h *AuthHandler) BindEmail(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req BindEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 调用服务层处理绑定逻辑
	finalUserID, err := h.oauthService.BindEmail(userID.(uint), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "绑定失败", "details": err.Error()})
		return
	}

	// 如果用户ID发生变化（发生了账号合并），需要生成新的Token
	var token string
	if finalUserID != userID.(uint) {
		token, err = h.authService.GenerateJWT(finalUserID)
		if err != nil {
			// 这种情况很少见，但如果发生了，用户需要重新登录
			c.JSON(http.StatusOK, gin.H{"message": "绑定成功，请重新登录"})
			return
		}
	}

	if token != "" {
		c.JSON(http.StatusOK, gin.H{"message": "绑定成功", "token": token})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "绑定成功"})
	}
}

// GetMe 获取当前用户信息
// GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	fmt.Printf("📋 GetMe获取用户信息: user_id=%d, avatar=%s\n", userID.(uint), user.Avatar)

	// 头像 URL 处理: avatars 文件夹已设置为公共读,直接返回不带签名的 URL
	avatarURL := user.Avatar

	// 如果头像 URL 不为空且是 OSS 地址,确保使用 HTTPS
	if avatarURL != "" && strings.Contains(avatarURL, "oss-cn-") {
		// 强制使用 HTTPS
		if strings.HasPrefix(avatarURL, "http://") {
			avatarURL = strings.Replace(avatarURL, "http://", "https://", 1)
			fmt.Printf("🔒 GetMe头像URL转换为HTTPS: user_id=%d\n", userID.(uint))
		}
		fmt.Printf("✅ GetMe返回公开头像URL: user_id=%d, url=%s\n", userID.(uint), avatarURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":              user.ID,
			"email":           user.Email,
			"phone":           getStringValue(user.Phone),
			"github_id":       getStringValue(user.GitHubID),
			"github_username": user.GitHubUsername,
			"google_id":       getStringValue(user.GoogleID),
			"google_email":    user.GoogleEmail,
			"nickname":        user.Nickname,
			"avatar":          avatarURL,
			"bio":             user.Bio,
			"role":            user.Role,
			"status":          user.Status,
			"invite_code":     user.InviteCode,
			"created_at":      user.CreatedAt,
		},
	})
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"`   // 头像URL
	Bio      string `json:"bio"`      // 个人简介
}

// UpdateProfile 更新用户资料
// PUT /api/v1/auth/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ 更新用户资料请求参数错误: user_id=%v, error=%v\n", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	fmt.Printf("📥 收到更新用户资料请求: user_id=%v, nickname=%s, avatar=%s, bio=%s\n",
		userID, req.Nickname, req.Avatar, req.Bio)

	// 先获取当前用户信息，如果某个字段未提供，使用原有值
	currentUser, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		fmt.Printf("❌ 获取用户信息失败: user_id=%v, error=%v\n", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败", "details": err.Error()})
		return
	}

	// 如果前端没有提供某个字段（空字符串），使用原有值
	nickname := req.Nickname
	if nickname == "" {
		nickname = currentUser.Nickname
	}
	avatar := req.Avatar
	if avatar == "" {
		avatar = currentUser.Avatar
	}
	bio := req.Bio
	if bio == "" {
		bio = currentUser.Bio
	}

	user, err := h.authService.UpdateUserProfile(userID.(uint), nickname, avatar, bio)
	if err != nil {
		fmt.Printf("❌ 更新用户资料失败: user_id=%v, error=%v\n", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败", "details": err.Error()})
		return
	}

	fmt.Printf("✅ 更新用户资料成功: user_id=%v\n", userID)

	// 头像 URL 处理: avatars 文件夹已设置为公共读,直接返回不带签名的 URL
	avatarURL := user.Avatar

	// 如果头像 URL 不为空且是 OSS 地址,确保使用 HTTPS
	if avatarURL != "" && strings.Contains(avatarURL, "oss-cn-") {
		// 强制使用 HTTPS
		if strings.HasPrefix(avatarURL, "http://") {
			avatarURL = strings.Replace(avatarURL, "http://", "https://", 1)
			fmt.Printf("🔒 UpdateProfile头像URL转换为HTTPS: user_id=%d\n", userID.(uint))
		}
		fmt.Printf("✅ UpdateProfile返回公开头像URL: user_id=%d, url=%s\n", userID.(uint), avatarURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"nickname":   user.Nickname,
			"avatar":     avatarURL,
			"bio":        user.Bio,
			"created_at": user.CreatedAt,
		},
	})
}

// UploadAvatar 上传头像
// POST /api/v1/auth/avatar/upload
func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("avatar")
	if err != nil {
		fmt.Printf("❌ 获取上传文件失败: user_id=%v, error=%v\n", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件", "details": err.Error()})
		return
	}

	// 验证文件类型（只允许图片）
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	allowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			allowed = true
			break
		}
	}
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型，仅支持：jpg, jpeg, png, gif, webp"})
		return
	}

	// 验证文件大小（最大5MB）
	maxSize := int64(5 * 1024 * 1024) // 5MB
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过5MB"})
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		fmt.Printf("❌ 打开文件失败: user_id=%v, error=%v\n", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败", "details": err.Error()})
		return
	}
	defer src.Close()

	// 生成唯一文件名：avatars/{user_id}/{timestamp}_{random}.{ext}
	timestamp := time.Now().Format("20060102150405")
	randomStr := generateRandomString(8)
	filename := fmt.Sprintf("%s_%s%s", timestamp, randomStr, ext)
	ossPath := fmt.Sprintf("avatars/%d/%s", userID.(uint), filename)

	fmt.Printf("📤 开始上传头像: user_id=%v, filename=%s, size=%d\n", userID, filename, file.Size)

	// 上传到OSS (设置为公开可读，防止签名过期)
	ctx := context.Background()
	avatarURL, err := h.storage.SaveFromReaderWithPublicRead(ctx, ossPath, src, file.Size)
	if err != nil {
		fmt.Printf("❌ 上传头像到OSS失败: user_id=%v, error=%v\n", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传失败", "details": err.Error()})
		return
	}

	fmt.Printf("✅ 头像上传成功: user_id=%v, url=%s\n", userID, avatarURL)

	// 生成签名URL（OSS Bucket是私有的）
	signedAvatarURL := avatarURL
	if h.storage != nil && strings.Contains(avatarURL, "oss-cn-") {
		parts := strings.Split(avatarURL, ".aliyuncs.com/")
		if len(parts) > 1 {
			path := parts[1]
			signedURL, err := h.storage.GetSignedURL(ctx, path, 3153600000) // 100 years expiration
			if err == nil {
				signedAvatarURL = signedURL
				fmt.Printf("✅ 生成头像签名URL成功\n")
			} else {
				fmt.Printf("⚠️  生成头像签名URL失败: %v\n", err)
			}
		}
	}

	// 先获取当前用户信息，保留昵称和简介
	currentUser, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		fmt.Printf("❌ 获取用户信息失败: user_id=%v, error=%v\n", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败", "details": err.Error()})
		return
	}

	// 更新用户头像字段（保留原有的昵称和简介）
	user, err := h.authService.UpdateUserProfile(userID.(uint), currentUser.Nickname, avatarURL, currentUser.Bio)
	if err != nil {
		fmt.Printf("❌ 更新用户头像失败: user_id=%v, error=%v\n", userID, err)
		c.JSON(http.StatusOK, gin.H{
			"message":    "上传成功，但更新用户资料失败",
			"avatar_url": signedAvatarURL,
			"error":      err.Error(),
		})
		return
	}

	fmt.Printf("✅ 用户头像更新成功: user_id=%v\n", userID)
	c.JSON(http.StatusOK, gin.H{
		"message":    "上传成功",
		"avatar_url": signedAvatarURL,
		"user": gin.H{
			"id":              user.ID,
			"email":           user.Email,
			"phone":           getStringValue(user.Phone),
			"github_id":       getStringValue(user.GitHubID),
			"github_username": user.GitHubUsername,
			"google_id":       getStringValue(user.GoogleID),
			"google_email":    user.GoogleEmail,
			"nickname":        user.Nickname,
			"avatar":          signedAvatarURL,
			"bio":             user.Bio,
			"role":            user.Role,
			"status":          user.Status,
			"created_at":      user.CreatedAt,
		},
	})
}

// Logout 登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if exists {
		// 清除用户会话
		sessionToken := c.GetHeader("X-Session-Token")
		_ = h.authService.Logout(sessionToken)
		fmt.Printf("✅ 用户登出: user_id=%d\n", userID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// getStringValue 获取字符串指针的值（如果为nil返回空字符串）
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ClearVerificationCodes 清理验证码（临时接口，用于测试）
// POST /api/v1/auth/email/clear
func (h *AuthHandler) ClearVerificationCodes(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	codeRepo := repository.NewVerificationCodeRepository()
	// 清理该邮箱的所有验证码
	err := codeRepo.DeleteByReceiver(req.Email, "email")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清理失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码已清理，可以重新发送"})
}

// AuthMiddleware JWT认证中间件
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少Authorization header"})
			c.Abort()
			return
		}

		// 提取token（格式：Bearer <token>）
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization格式错误，应为: Bearer <token>"})
			c.Abort()
			return
		}

		// 验证token
		userID, err := h.authService.VerifyJWT(tokenString)
		log.Printf("[AUTH] VerifyJWT result: userID=%d, error=%v", userID, err)

		if err != nil {
			log.Printf("[AUTH] Token verification failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token无效或已过期", "details": err.Error()})
			c.Abort()
			return
		}

		// 将userID存储到context
		log.Printf("[AUTH] Setting userID=%d in context", userID)
		c.Set("user_id", userID)
		c.Next()
	}
}

// OptionalAuthMiddleware 可选的JWT认证中间件（如果有token则验证并设置user_id，没有token则继续）
// BUG修复: 用于支持可选认证的路由（如反馈功能），已登录用户自动关联user_id，未登录用户也可以使用
// 修复策略: 创建新的中间件，尝试验证token但不强制要求
// 影响范围: backend/internal/api/v1/auth_handler.go, backend/internal/api/v1/router.go
// 修复日期: 2025-12-10
func (h *AuthHandler) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有token，继续处理请求（不设置user_id）
			c.Next()
			return
		}

		// 提取token（格式：Bearer <token>）
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			// token格式错误，继续处理请求（不设置user_id）
			c.Next()
			return
		}

		// 验证token
		userID, err := h.authService.VerifyJWT(tokenString)
		if err != nil {
			// token无效，继续处理请求（不设置user_id）
			log.Printf("⚠️  可选认证中token验证失败（允许继续）: %v", err)
			c.Next()
			return
		}

		// token有效，将userID存储到context
		c.Set("user_id", userID)
		log.Printf("✅ 可选认证成功，用户ID: %d", userID)
		c.Next()
	}
}
