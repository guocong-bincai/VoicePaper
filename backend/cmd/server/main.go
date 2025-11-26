package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"voicepaper/internal/api/v1"
	"voicepaper/internal/repository"
	"voicepaper/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化数据库
	dbPath := filepath.Join("data", "voicepaper.db")
	repository.InitDB(dbPath)

	// 1.5 扫描旧数据 (Seed)
	seedLegacyData()

	// 2. 初始化 Gin
	r := gin.Default()

	// 3. 配置 CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 生产环境请改为具体域名
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 4. 静态文件服务 (音频文件)
	// 访问 /audio/xxx.mp3 -> data/audio/xxx.mp3
	r.Static("/audio", "./data/audio")

	// 5. 注册 API 路由
	v1.RegisterRoutes(r)

	// 6. 启动服务器
	port := ":8080"
	log.Printf("🚀 VoicePaper Backend running on http://localhost%s", port)
	if err := r.Run(port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func seedLegacyData() {
	manifestPath := "../data/manifest.json"
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		log.Printf("⚠️  No manifest found at %s, skipping seed.", manifestPath)
		return
	}

	var manifest struct {
		Articles []struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Markdown string `json:"markdown"`
			Audio    string `json:"audio"`
		} `json:"articles"`
	}

	if err := json.Unmarshal(content, &manifest); err != nil {
		log.Printf("❌ Failed to parse manifest: %v", err)
		return
	}

	ttsService := service.NewTTSService()
	repo := repository.NewArticleRepository()

	for _, item := range manifest.Articles {
		mdPath := filepath.Join("../data", item.Markdown)
		mdContent, err := os.ReadFile(mdPath)
		if err != nil {
			log.Printf("❌ Failed to read markdown %s: %v", mdPath, err)
			continue
		}

		log.Printf("📦 Seeding article: %s", item.Title)

		// 1. Create/Get Article
		article, err := ttsService.GetOrGenerateAudio(item.Title, string(mdContent))
		if err != nil {
			log.Printf("❌ Failed to seed article: %v", err)
			continue
		}

		// 2. Link existing audio if available
		// The manifest says "audio": "output.mp3" (relative to data/)
		// We need to check if that file exists and link it if the article doesn't have audio yet
		if article.Status != "completed" {
			audioPath := filepath.Join("../data", item.Audio)
			if _, err := os.Stat(audioPath); err == nil {
				// Copy or link the audio file to our managed directory?
				// For now, let's just point to it if it's in data/
				// But our system expects audio in backend/data/audio/
				// Let's copy it to be safe and consistent

				newFilename := "legacy_" + item.Audio
				newPath := filepath.Join("data", "audio", newFilename)

				// Ensure dir exists
				os.MkdirAll(filepath.Dir(newPath), 0755)

				// Copy file
				input, _ := os.ReadFile(audioPath)
				os.WriteFile(newPath, input, 0644)

				repo.UpdateStatus(article.ID, "completed", newPath)
				log.Printf("✅ Linked legacy audio for %s", item.Title)
			}
		}
	}
}
