package main

import (
	"fmt"
	"log"
	"voicepaper/config"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"
)

func main() {
	// 加载配置
	cfg := config.GetConfig()

	// 初始化数据库
	repository.InitDB(cfg)
	db := repository.DB

	// 1. 检查表中是否有数据
	var count int64
	if err := db.Model(&model.UserPoints{}).Count(&count).Error; err != nil {
		log.Fatalf("❌ 查询 UserPoints 表失败: %v", err)
	}
	fmt.Printf("📊 UserPoints 表中有 %d 条记录\n", count)

	// 2. 检查字段是否存在
	hasField := db.Migrator().HasColumn(&model.UserPoints{}, "total_duration_minutes")
	fmt.Printf("📋 total_duration_minutes 字段存在: %v\n", hasField)

	// 3. 列出前5条用户积分记录
	var records []model.UserPoints
	if err := db.Limit(5).Find(&records).Error; err != nil {
		log.Fatalf("❌ 查询用户积分记录失败: %v", err)
	}

	fmt.Println("\n📝 前5条用户积分记录:")
	for i, record := range records {
		fmt.Printf("  [%d] UserID=%d, TotalPoints=%d, TotalDurationMinutes=%d\n",
			i+1, record.UserID, record.TotalPoints, record.TotalDurationMinutes)
	}

	// 4. 尝试执行排名查询
	fmt.Println("\n🔍 尝试执行排名查询...")
	var results []map[string]interface{}
	query := `
	SELECT
		ROW_NUMBER() OVER (ORDER BY (COALESCE(total_duration_minutes, 0) * 0.5 + total_points * 0.5) DESC) as rank,
		user_id,
		total_points,
		COALESCE(total_duration_minutes, 0) as total_duration_minutes,
		(COALESCE(total_duration_minutes, 0) * 0.5 + total_points * 0.5) as score
	FROM vp_user_points
	WHERE deleted_at IS NULL
	LIMIT 5
	`

	if err := db.Raw(query).Scan(&results).Error; err != nil {
		log.Printf("❌ 排名查询失败: %v", err)
	} else {
		fmt.Println("✅ 排名查询成功:")
		for i, result := range results {
			fmt.Printf("  [%d] %v\n", i+1, result)
		}
	}
}
