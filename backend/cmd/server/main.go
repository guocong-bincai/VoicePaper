package main

import (
	"log"
	"strings"
	"voicepaper/config"
	v1 "voicepaper/internal/api/v1"
	"voicepaper/internal/repository"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}

	// 2. Initialize Database
	repository.InitDB(cfg)

	// 3. Initialize Redis
	repository.InitRedis(cfg)

	// 4. Initialize Gin
	r := gin.Default()

	// 5. Configure CORS
	r.Use(cors.Default())

	// 6. Static files (Fonts)
	r.Static("/static/fonts", "./assets/fonts")

	// 7. Register Routes
	v1.RegisterRoutes(r)

	// 8. Start Server
	serverAddr := cfg.Service.Port
	if serverAddr == "" {
		serverAddr = "8080"
	}
	if !strings.HasPrefix(serverAddr, ":") {
		serverAddr = ":" + serverAddr
	}

	log.Printf("🚀 Server starting on %s", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
