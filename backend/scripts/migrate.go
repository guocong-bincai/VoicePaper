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
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// 初始化数据库
	repository.InitDB(cfg)

	// 执行迁移
	db := repository.DB

	// 1. 为 vp_users 表添加 total_duration_minutes 字段
	if !db.Migrator().HasColumn("vp_users", "total_duration_minutes") {
		if err := db.Migrator().AddColumn("vp_users", "total_duration_minutes BIGINT DEFAULT 0 COMMENT '累积学习时长（分钟）'"); err != nil {
			log.Fatalf("❌ 添加 vp_users.total_duration_minutes 字段失败: %v", err)
		}
		log.Println("✅ 成功添加 vp_users.total_duration_minutes 字段")
	} else {
		log.Println("⏭️  vp_users.total_duration_minutes 字段已存在")
	}

	// 2. 为 vp_user_points 表添加 total_duration_minutes 字段
	if !db.Migrator().HasColumn("vp_user_points", "total_duration_minutes") {
		if err := db.Migrator().AddColumn("vp_user_points", "total_duration_minutes BIGINT DEFAULT 0 COMMENT '累积学习时长（分钟）'"); err != nil {
			log.Fatalf("❌ 添加 vp_user_points.total_duration_minutes 字段失败: %v", err)
		}
		log.Println("✅ 成功添加 vp_user_points.total_duration_minutes 字段")
	} else {
		log.Println("⏭️  vp_user_points.total_duration_minutes 字段已存在")
	}

	// 3. 创建排名查询索引
	if !db.Migrator().HasIndex("vp_user_points", "idx_user_points_ranking") {
		if err := db.Migrator().CreateIndex("vp_user_points", "idx_user_points_ranking"); err != nil {
			log.Printf("⚠️  创建索引失败: %v（可能已存在）", err)
		} else {
			log.Println("✅ 成功创建排名查询索引")
		}
	} else {
		log.Println("⏭️  排名查询索引已存在")
	}

	// 4. 创建 vp_wordbook_progress 表
	if !db.Migrator().HasTable("vp_wordbook_progress") {
		// 这里直接使用 Exec 创建表比较简单
		if err := db.Exec(`
			CREATE TABLE IF NOT EXISTS vp_wordbook_progress (
				id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
				user_id BIGINT UNSIGNED NOT NULL,
				word_type VARCHAR(20) NOT NULL,
				current_index INT DEFAULT 0,
				total_words INT DEFAULT 0,
				last_word_id BIGINT UNSIGNED,
				created_at DATETIME(3),
				updated_at DATETIME(3),
				deleted_at DATETIME(3),
				INDEX idx_user_wordtype (user_id, word_type),
				INDEX idx_deleted_at (deleted_at)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
		`).Error; err != nil {
			log.Fatalf("❌ 创建 vp_wordbook_progress 表失败: %v", err)
		}
		log.Println("✅ 成功创建 vp_wordbook_progress 表")
	} else {
		log.Println("⏭️  vp_wordbook_progress 表已存在")
	}

	// 5. 创建 vp_app_configs 表
	if !db.Migrator().HasTable("vp_app_configs") {
		if err := db.Migrator().CreateTable(&model.AppConfig{}); err != nil {
			log.Fatalf("❌ 创建 vp_app_configs 表失败: %v", err)
		}
		log.Println("✅ 成功创建 vp_app_configs 表")
	} else {
		log.Println("⏭️  vp_app_configs 表已存在")
	}

	// 确保默认配置存在
	var count int64
	db.Model(&model.AppConfig{}).Where("config_key = ?", "miniprogram_audit").Count(&count)
	if count == 0 {
		defaultConfig := model.AppConfig{
			ConfigKey:   "miniprogram_audit",
			IsAuditMode: false,
			StartDate:   "2026-01-20",
			EndDate:     "",
			Description: "小程序审核模式配置",
		}
		if err := db.Create(&defaultConfig).Error; err != nil {
			log.Printf("⚠️ 插入默认配置失败: %v", err)
		} else {
			log.Println("✅ 插入默认配置成功")
		}
	} else {
		log.Println("⏭️  默认配置已存在")
	}

	if !db.Migrator().HasColumn("vp_articles", "original_article_url") {
		if err := db.Exec(`
			ALTER TABLE vp_articles
			ADD COLUMN original_article_url varchar(255) DEFAULT NULL COMMENT '英文原文 OSS 地址' AFTER article_url;
		`).Error; err != nil {
			log.Fatalf("❌ 添加 vp_articles.original_article_url 字段失败: %v", err)
		}
		log.Println("✅ 成功添加 vp_articles.original_article_url 字段")
	} else {
		log.Println("⏭️  vp_articles.original_article_url 字段已存在")
	}

	fmt.Println("\n✅ 所有迁移任务完成！")
}
