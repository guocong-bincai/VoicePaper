package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"voicepaper/config"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"

	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <user_id> <minutes>")
		fmt.Println("示例: go run main.go 1 120")
		fmt.Println("说明: 为用户ID=1添加120分钟的学习时长")
		return
	}

	userID, _ := strconv.Atoi(os.Args[1])
	minutes, _ := strconv.Atoi(os.Args[2])

	cfg := config.GetConfig()
	repository.InitDB(cfg)
	db := repository.DB

	// 更新 vp_user_points 表
	result := db.Model(&model.UserPoints{}).
		Where("user_id = ?", userID).
		Update("total_duration_minutes", gorm.Expr("COALESCE(total_duration_minutes, 0) + ?", minutes))

	if result.Error != nil {
		log.Fatalf("❌ 更新失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Fatalf("❌ 用户ID=%d不存在", userID)
	}

	// 更新 vp_users 表（备份字段）
	db.Model(&model.User{}).
		Where("id = ?", userID).
		Update("total_duration_minutes", gorm.Expr("COALESCE(total_duration_minutes, 0) + ?", minutes))

	fmt.Printf("✅ 成功为用户ID=%d添加了%d分钟的学习时长\n", userID, minutes)

	// 显示更新后的信息
	var userPoints model.UserPoints
	db.Where("user_id = ?", userID).First(&userPoints)
	fmt.Printf("📊 当前用户信息:\n")
	fmt.Printf("   - 累积时长: %d 分钟\n", userPoints.TotalDurationMinutes)
	fmt.Printf("   - 累计积分: %d\n", userPoints.TotalPoints)
	fmt.Printf("   - 综合分数: %.1f\n", float64(userPoints.TotalDurationMinutes)*0.5+float64(userPoints.TotalPoints)*0.5)
}
