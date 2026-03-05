package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"
	"voicepaper/config"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("etc/config.yaml")
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	repository.DB = db
	log.Println("✅ 数据库连接成功")

	// 查找所有邀请码为空的用户
	var users []model.User
	result := db.Where("invite_code IS NULL OR invite_code = ''").Find(&users)
	if result.Error != nil {
		log.Fatalf("❌ 查询失败: %v", result.Error)
	}

	log.Printf("📊 找到 %d 个用户的邀请码为空", len(users))

	if len(users) == 0 {
		log.Println("✅ 没有需要修复的用户")
		return
	}

	// 为每个用户生成邀请码
	for i, user := range users {
		inviteCode := generateInviteCode(user.ID)
		user.InviteCode = inviteCode

		if err := db.Model(&user).Update("invite_code", inviteCode).Error; err != nil {
			log.Printf("❌ 更新用户 %d 失败: %v", user.ID, err)
			continue
		}

		log.Printf("✅ [%d/%d] 用户 ID:%d, Email:%s, 邀请码:%s",
			i+1, len(users), user.ID, user.Email, inviteCode)
	}

	log.Println("🎉 修复完成！")
}

// generateInviteCode 生成邀请码（与 AuthService 中的逻辑一致）
func generateInviteCode(userID uint) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("user_%d_%d", userID, time.Now().Unix())))
	inviteCode := hex.EncodeToString(hash.Sum(nil))[:16]
	return strings.ToUpper(inviteCode)
}
