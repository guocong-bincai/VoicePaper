package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"voicepaper/config"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务
type AuthService struct {
	jwtSecret     string
	jwtExpiration time.Duration
	userRepo      *repository.UserRepository
	sessionRepo   *repository.UserSessionRepository
	emailService  *EmailService
	oauthService  *OAuthService
}

func NewAuthService(cfg *config.AuthConfig, userRepo *repository.UserRepository, sessionRepo *repository.UserSessionRepository, emailService *EmailService, oauthService *OAuthService) *AuthService {
	return &AuthService{
		jwtSecret:     cfg.JWT.Secret,
		jwtExpiration: time.Duration(cfg.JWT.Expiration) * time.Hour,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		emailService:  emailService,
		oauthService:  oauthService,
	}
}

// LoginWithEmail 邮箱验证码登录
func (s *AuthService) LoginWithEmail(email, code, ipAddress string) (*model.User, string, error) {
	// 验证验证码
	vc, err := s.emailService.VerifyCode(email, code)
	if err != nil {
		return nil, "", fmt.Errorf("验证码错误: %w", err)
	}
	fmt.Printf("✅ 验证码验证成功: email=%s, code=%s, vc_id=%d\n", email, code, vc.ID)

	// 查找或创建用户
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// 用户不存在，创建新用户
		fmt.Printf("📝 用户不存在，创建新用户: email=%s\n", email)
		now := time.Now()
		// 生成默认头像URL（使用邮箱首字母）
		defaultAvatar := generateDefaultAvatar(email)
		// 生成默认昵称（使用邮箱前缀）
		defaultNickname := generateDefaultNickname(email)

		user = &model.User{
			Email:           &email,
			EmailVerified:   true,
			EmailVerifiedAt: &now,
			Status:          "active",
			Avatar:          defaultAvatar,
			Nickname:        defaultNickname,
			// 对于uniqueIndex字段，不设置值（使用NULL），避免空字符串冲突
			// GORM会自动处理NULL值
		}

		if err := s.userRepo.Create(user); err != nil {
			fmt.Printf("❌ 创建用户失败: email=%s, error=%v\n", email, err)
			return nil, "", fmt.Errorf("创建用户失败: %w", err)
		}

		// 生成邀请码（基于用户ID）
		user.InviteCode = s.GenerateInviteCode(user.ID)
		if err := s.userRepo.Update(user); err != nil {
			fmt.Printf("⚠️  更新邀请码失败: user_id=%d, error=%v\n", user.ID, err)
			// 不影响登录流程
		}

		fmt.Printf("✅ 用户创建成功: id=%d, email=%s, invite_code=%s\n", user.ID, user.Email, user.InviteCode)
	} else {
		fmt.Printf("✅ 用户已存在: id=%d, email=%s\n", user.ID, user.Email)
		// 更新邮箱验证状态
		if !user.EmailVerified {
			now := time.Now()
			user.EmailVerified = true
			user.EmailVerifiedAt = &now
			s.userRepo.Update(user)
		}
	}

	// 更新最后登录信息
	if err := s.userRepo.UpdateLastLogin(user.ID, ipAddress); err != nil {
		fmt.Printf("⚠️  更新登录信息失败: user_id=%d, error=%v\n", user.ID, err)
		// 不影响登录流程
	}

	// 生成JWT token
	token, err := s.GenerateJWT(user.ID)
	if err != nil {
		fmt.Printf("❌ 生成token失败: user_id=%d, error=%v\n", user.ID, err)
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}
	fmt.Printf("✅ JWT token生成成功: user_id=%d\n", user.ID)

	// 创建会话（可选，失败不影响登录）
	_, _ = s.createSession(user.ID, ipAddress, "")

	return user, token, nil
}

// LoginWithPassword 密码登录（支持普通用户和管理员）
func (s *AuthService) LoginWithPassword(email, password, ipAddress string) (*model.User, string, error) {
	// 查找用户
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("用户不存在或密码错误")
	}

	// 检查密码哈希是否存在
	if user.PasswordHash == "" {
		return nil, "", fmt.Errorf("此账户未设置密码，请先注册或使用验证码登录")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		fmt.Printf("❌ 密码验证失败: email=%s, error=%v\n", email, err)
		return nil, "", fmt.Errorf("密码错误")
	}

	fmt.Printf("✅ 密码验证成功: email=%s, user_id=%d, role=%s\n", email, user.ID, user.Role)

	// 更新最后登录信息
	if err := s.userRepo.UpdateLastLogin(user.ID, ipAddress); err != nil {
		fmt.Printf("⚠️  更新登录信息失败: user_id=%d, error=%v\n", user.ID, err)
		// 不影响登录流程
	}

	// 生成JWT token
	token, err := s.GenerateJWT(user.ID)
	if err != nil {
		fmt.Printf("❌ 生成token失败: user_id=%d, error=%v\n", user.ID, err)
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}
	fmt.Printf("✅ JWT token生成成功: user_id=%d\n", user.ID)

	// 创建会话（可选，失败不影响登录）
	_, _ = s.createSession(user.ID, ipAddress, "")

	return user, token, nil
}

// HashPassword 生成密码哈希（用于创建或更新超管账户密码）
func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("生成密码哈希失败: %w", err)
	}
	return string(hash), nil
}

// GenerateJWT 生成JWT token（公开方法）
func (s *AuthService) GenerateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.jwtExpiration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// VerifyJWT 验证JWT token
func (s *AuthService) VerifyJWT(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, fmt.Errorf("token解析失败: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, fmt.Errorf("无效的token claims")
		}
		return uint(userID), nil
	}

	return 0, fmt.Errorf("无效的token")
}

// createSession 创建会话
func (s *AuthService) createSession(userID uint, ipAddress, userAgent string) (string, error) {
	// 生成会话token
	sessionToken, err := s.generateSessionToken()
	if err != nil {
		return "", err
	}

	session := &model.UserSession{
		UserID:       userID,
		SessionToken: sessionToken,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    time.Now().Add(s.jwtExpiration),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return "", err
	}

	return sessionToken, nil
}

// generateSessionToken 生成会话token
func (s *AuthService) generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Logout 登出
func (s *AuthService) Logout(sessionToken string) error {
	if sessionToken != "" {
		return s.sessionRepo.DeleteByToken(sessionToken)
	}
	return nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(userID uint) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

// GetUserByEmail 根据邮箱获取用户
// BUG修复: 添加通过邮箱查询用户的方法，用于注册前检查邮箱是否已存在
// 修复策略: 封装 userRepo.FindByEmail 方法供外部调用
// 影响范围: backend/internal/service/auth_service.go
// 修复日期: 2025-12-10
func (s *AuthService) GetUserByEmail(email string) (*model.User, error) {
	return s.userRepo.FindByEmail(email)
}

// UpdateUserProfile 更新用户资料
// 如果某个字段为空字符串，则保留原有值（部分更新）
func (s *AuthService) UpdateUserProfile(userID uint, nickname, avatar, bio string) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 只更新非空字段，空字符串表示不更新该字段
	if nickname != "" {
		user.Nickname = nickname
	}
	if avatar != "" {
		user.Avatar = avatar
	}
	if bio != "" {
		user.Bio = bio
	}

	fmt.Printf("📝 更新用户资料: user_id=%d, nickname=%s (更新=%v), avatar=%s (更新=%v), bio=%s (更新=%v)\n",
		userID, user.Nickname, nickname != "", user.Avatar, avatar != "", user.Bio, bio != "")

	if err := s.userRepo.Update(user); err != nil {
		fmt.Printf("❌ 更新用户资料失败: user_id=%d, error=%v\n", userID, err)
		return nil, fmt.Errorf("更新用户资料失败: %w", err)
	}

	fmt.Printf("✅ 用户资料更新成功: user_id=%d\n", userID)
	return user, nil
}

// generateDefaultAvatar 生成默认头像URL（使用邮箱首字母）
func generateDefaultAvatar(email string) string {
	// 提取邮箱首字母（@前面的第一个字符）
	firstChar := ""
	if len(email) > 0 {
		firstChar = string(email[0])
	}
	// 转换为大写
	if firstChar >= "a" && firstChar <= "z" {
		firstChar = string(firstChar[0] - 32)
	}
	// 使用UI Avatars服务生成头像
	return fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=2563EB&color=fff&size=128&bold=true", firstChar)
}

// generateDefaultNickname 生成默认昵称（使用邮箱前缀）
func generateDefaultNickname(email string) string {
	// 提取@前面的部分作为昵称
	parts := strings.Split(email, "@")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "用户"
}

// Register 用户注册（邮箱+密码+邀请码）
func (s *AuthService) Register(email, password, code, inviteCode, ipAddress string) (*model.User, string, error) {
	// 检查邮箱是否已存在
	existingUser, err := s.userRepo.FindByEmail(email)
	if err == nil && existingUser != nil {
		return nil, "", fmt.Errorf("该邮箱已被注册")
	}

	// 验证邮箱验证码
	_, err = s.emailService.VerifyCode(email, code)
	if err != nil {
		return nil, "", fmt.Errorf("验证码错误: %w", err)
	}
	fmt.Printf("✅ 验证码验证成功: email=%s\n", email)

	// 验证邀请码（如果提供）
	// 邀请码是可选的，即使无效也不阻止注册，只是不建立邀请关系
	var invitedByID *uint
	if inviteCode != "" {
		inviter, err := s.userRepo.FindByInviteCode(inviteCode)
		if err != nil {
			// 邀请码无效，忽略它，继续注册
			fmt.Printf("⚠️  邀请码无效，忽略: invite_code=%s, error=%v\n", inviteCode, err)
		} else {
			invitedByID = &inviter.ID
			fmt.Printf("✅ 邀请码验证成功: invite_code=%s, inviter_id=%d\n", inviteCode, inviter.ID)
		}
	}

	// 生成密码哈希
	passwordHash, err := s.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	now := time.Now()
	defaultAvatar := generateDefaultAvatar(email)
	defaultNickname := generateDefaultNickname(email)

	user := &model.User{
		Email:           &email,
		EmailVerified:   true, // 注册时已验证（通过密码）
		EmailVerifiedAt: &now,
		PasswordHash:    passwordHash,
		Status:          "active",
		Role:            "user",
		Avatar:          defaultAvatar,
		Nickname:        defaultNickname,
		InvitedByID:     invitedByID,
	}

	// 先创建用户（获取ID）
	if err := s.userRepo.Create(user); err != nil {
		fmt.Printf("❌ 创建用户失败: email=%s, error=%v\n", email, err)
		return nil, "", fmt.Errorf("创建用户失败: %w", err)
	}

	// 生成邀请码（基于用户ID）
	user.InviteCode = s.GenerateInviteCode(user.ID)
	if err := s.userRepo.Update(user); err != nil {
		fmt.Printf("⚠️  更新邀请码失败: user_id=%d, error=%v\n", user.ID, err)
		// 不影响注册流程
	}

	fmt.Printf("✅ 用户注册成功: id=%d, email=%s, invite_code=%s\n", user.ID, user.Email, user.InviteCode)

	// 更新最后登录信息
	if err := s.userRepo.UpdateLastLogin(user.ID, ipAddress); err != nil {
		fmt.Printf("⚠️  更新登录信息失败: user_id=%d, error=%v\n", user.ID, err)
		// 不影响注册流程
	}

	// 生成JWT token
	token, err := s.GenerateJWT(user.ID)
	if err != nil {
		fmt.Printf("❌ 生成token失败: user_id=%d, error=%v\n", user.ID, err)
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}
	fmt.Printf("✅ JWT token生成成功: user_id=%d\n", user.ID)

	// 创建会话（可选，失败不影响注册）
	_, _ = s.createSession(user.ID, ipAddress, "")

	return user, token, nil
}

// GenerateInviteCode 生成邀请码（基于用户ID加密）
func (s *AuthService) GenerateInviteCode(userID uint) string {
	// 使用HMAC-SHA256基于用户ID和JWT密钥生成邀请码
	// 这样可以确保邀请码的唯一性和不可伪造性
	h := hmac.New(sha256.New, []byte(s.jwtSecret))
	h.Write([]byte(fmt.Sprintf("invite:%d", userID)))
	hash := h.Sum(nil)

	// 转换为16进制字符串，取前16位作为邀请码
	inviteCode := hex.EncodeToString(hash)[:16]

	// 转换为大写，便于输入和分享
	return strings.ToUpper(inviteCode)
}

// VerifyInviteCode 验证邀请码（可选，用于前端验证）
func (s *AuthService) VerifyInviteCode(inviteCode string) (*model.User, error) {
	return s.userRepo.FindByInviteCode(inviteCode)
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(email, code, newPassword string) error {
	// 1. 验证验证码
	_, err := s.emailService.VerifyCode(email, code)
	if err != nil {
		return fmt.Errorf("验证码错误: %w", err)
	}

	// 2. 查找用户
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 3. 生成新密码哈希
	passwordHash, err := s.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 4. 更新用户密码
	user.PasswordHash = passwordHash
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	fmt.Printf("✅ 用户密码重置成功: user_id=%d, email=%s\n", user.ID, user.Email)
	return nil
}

// UpdateLastLogin 更新最后登录信息
func (s *AuthService) UpdateLastLogin(userID uint, ipAddress string) error {
	return s.userRepo.UpdateLastLogin(userID, ipAddress)
}
