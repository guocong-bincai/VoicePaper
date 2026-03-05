package main

import (
	"fmt"
	"voicepaper/config"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"
)

func main() {
	fmt.Println("🔍 排名功能验证脚本")

	// 加载配置
	cfg := config.GetConfig()

	// 初始化数据库
	repository.InitDB(cfg)
	db := repository.DB

	// 1. 验证字段是否存在
	fmt.Println("1️⃣  检查 vp_user_points 表结构...")
	hasField := db.Migrator().HasColumn(&model.UserPoints{}, "total_duration_minutes")
	if hasField {
		fmt.Println("   ✅ total_duration_minutes 字段存在")
	} else {
		fmt.Println("   ❌ total_duration_minutes 字段不存在！")
		return
	}

	// 2. 检查表中数据
	fmt.Println("\n2️⃣  检查表中数据...")
	var count int64
	if err := db.Model(&model.UserPoints{}).Count(&count).Error; err != nil {
		fmt.Printf("   ❌ 查询失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ UserPoints 表中有 %d 条记录\n", count)

	if count == 0 {
		fmt.Println("   ⚠️  表中没有数据，无法测试排名查询")
		return
	}

	// 3. 测试排名查询
	fmt.Println("\n3️⃣  执行排名查询...")
	type RankResult struct {
		Rank     int64
		UserID   uint
		Points   int64
		Duration int64
		Score    float64
	}

	var results []RankResult
	query := `
	SELECT
		ROW_NUMBER() OVER (ORDER BY (COALESCE(total_duration_minutes, 0) * 0.5 + total_points * 0.5) DESC) as rank,
		user_id,
		total_points as points,
		COALESCE(total_duration_minutes, 0) as duration,
		(COALESCE(total_duration_minutes, 0) * 0.5 + total_points * 0.5) as score
	FROM vp_user_points
	WHERE deleted_at IS NULL
	ORDER BY score DESC
	LIMIT 5
	`

	if err := db.Raw(query).Scan(&results).Error; err != nil {
		fmt.Printf("   ❌ 排名查询失败: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("   ⚠️  没有查询到排名数据")
		return
	}

	fmt.Println("   ✅ 排名查询成功！前5条结果：")
	for i, r := range results {
		fmt.Printf("   [%d] 排名#%d, UserID=%d, 积分=%d, 时长=%d分钟, 综合分=%v\n",
			i+1, r.Rank, r.UserID, r.Points, r.Duration, r.Score)
	}

	// 4. 测试单个用户的排名
	if len(results) > 0 {
		fmt.Println("\n4️⃣  获取第一个用户的详细排名...")
		userId := results[0].UserID

		var userInfo struct {
			Rank          int64
			TotalUsers    int64
			Score         float64
			TotalDuration int64
			TotalPoints   int64
		}

		countQuery := `SELECT COUNT(*) as total_users FROM vp_user_points WHERE deleted_at IS NULL`
		if err := db.Raw(countQuery).Scan(&userInfo).Error; err != nil {
			fmt.Printf("   ❌ 查询失败: %v\n", err)
		} else {
			userInfo.Rank = results[0].Rank
			userInfo.Score = results[0].Score
			userInfo.TotalDuration = results[0].Duration
			userInfo.TotalPoints = results[0].Points

			fmt.Printf("   ✅ UserID=%d 的排名信息:\n", userId)
			fmt.Printf("      - 全站排名: #%d/%d\n", userInfo.Rank, userInfo.TotalUsers)
			fmt.Printf("      - 综合分数: %.2f\n", userInfo.Score)
			fmt.Printf("      - 累积时长: %d 分钟\n", userInfo.TotalDuration)
			fmt.Printf("      - 累计积分: %d\n", userInfo.TotalPoints)
		}
	}

	fmt.Println("\n✅ 排名功能验证完成！前端应该可以正常显示排名了。")
}
