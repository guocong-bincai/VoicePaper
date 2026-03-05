package main

import (
	"fmt"
	"log"
	"os"
	"voicepaper/config"
	"voicepaper/internal/repository"
	"voicepaper/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// 初始化数据库
	repository.InitDB(cfg)

	// 查找管理员账号
	// 优先从命令行参数或环境变量获取
	email := ""
	if len(os.Args) > 1 {
		email = os.Args[1]
	} else if envEmail := os.Getenv("USER_EMAIL"); envEmail != "" {
		email = envEmail
	} else {
		log.Fatalf("❌ 请提供邮箱地址。用法: go run main.go <email> 或设置 USER_EMAIL 环境变量")
	}

	userRepo := repository.NewUserRepository()
	user, err := userRepo.FindByEmail(email)
	if err != nil {
		log.Fatalf("❌ 找不到用户: email=%s, error=%v", email, err)
	}

	fmt.Printf("✅ 找到用户: id=%d, email=%s, nickname=%s\n", user.ID, user.Email, user.Nickname)
	fmt.Printf("📋 当前邀请码: %s\n", user.InviteCode)

	// 创建 AuthService 来生成邀请码
	codeRepo := repository.NewVerificationCodeRepository()
	bindingRepo := repository.NewOAuthBindingRepository()
	sessionRepo := repository.NewUserSessionRepository()
	emailService := service.NewEmailService(&cfg.Auth.Email, codeRepo)
	oauthService := service.NewOAuthService(&cfg.Auth.OAuth, userRepo, bindingRepo)
	authService := service.NewAuthService(&cfg.Auth, userRepo, sessionRepo, emailService, oauthService)

	// 生成新的邀请码
	newInviteCode := authService.GenerateInviteCode(user.ID)
	fmt.Printf("🔑 生成的邀请码: %s\n", newInviteCode)

	// 更新数据库
	user.InviteCode = newInviteCode
	if err := userRepo.Update(user); err != nil {
		log.Fatalf("❌ 更新邀请码失败: error=%v", err)
	}

	fmt.Printf("✅ 邀请码已更新到数据库\n")
	fmt.Printf("\n📝 邀请码信息:\n")
	fmt.Printf("   用户ID: %d\n", user.ID)
	fmt.Printf("   邮箱: %s\n", user.Email)
	fmt.Printf("   邀请码: %s\n", newInviteCode)
	fmt.Printf("\n💡 提示: 邀请码是基于用户ID和JWT密钥生成的，不能直接反向解析。\n")
	fmt.Printf("   但可以通过数据库查询找到对应的用户（因为邀请码存储在数据库中）。\n")
}

