package main

import (
	"fmt"
	"log"
	"os"
	"voicepaper/config"
	"voicepaper/internal/repository"
)

func main() {
	// Set config path explicitly if needed, or rely on auto-detection
	// Since we are running from backend root usually, we can try to point to etc/config.yaml
	// But let's let LoadConfig handle it
    os.Setenv("CONFIG_PATH", "etc/config.yaml")

	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Initialize DB
	repository.InitDB(cfg)

	// Query articles
	var articles []struct {
		ID     uint   `gorm:"primaryKey"`
		Title  string
		PicURL string
	}
	
	result := repository.DB.Table("vp_articles").Select("id, title, pic_url").Limit(10).Order("id desc").Find(&articles)
	if result.Error != nil {
		log.Fatalf("Failed to query articles: %v", result.Error)
	}

	fmt.Printf("Found %d articles:\n", len(articles))
	for _, a := range articles {
		fmt.Printf("ID: %d, Title: %s, PicURL: '%s'\n", a.ID, a.Title, a.PicURL)
	}
}