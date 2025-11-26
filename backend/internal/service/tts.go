package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"
	"voicepaper/pkg/minimax"
)

type TTSService struct {
	repo *repository.ArticleRepository
}

func NewTTSService() *TTSService {
	return &TTSService{
		repo: repository.NewArticleRepository(),
	}
}

// GetOrGenerateAudio 核心逻辑：检查数据库 -> 检查文件 -> (如果不存在) 调用 API
func (s *TTSService) GetOrGenerateAudio(title, content string) (*model.Article, error) {
	// 1. 计算哈希
	hash := calculateHash(content)

	// 2. 检查数据库
	article, err := s.repo.FindByHash(hash)
	if err == nil {
		// 记录存在，检查本地文件是否真的存在
		if article.Status == "completed" && article.AudioPath != "" {
			if _, err := os.Stat(article.AudioPath); err == nil {
				log.Println("✅ Cache hit: Serving local audio for", title)
				return article, nil
			}
			log.Println("⚠️  Record exists but file missing, regenerating...", title)
		} else if article.Status == "processing" {
			return article, fmt.Errorf("audio is currently processing")
		}
	} else {
		// 记录不存在，创建新记录
		article = &model.Article{
			Title:       title,
			Content:     content,
			ContentHash: hash,
			Status:      "pending",
		}
		if err := s.repo.Create(article); err != nil {
			return nil, err
		}
	}

	// 3. 调用 MiniMax API 生成音频 (异步或同步)
	// 这里为了简单演示，先做同步调用，实际生产环境建议放进任务队列
	go s.processTTS(article)

	return article, nil
}

func (s *TTSService) processTTS(article *model.Article) {
	log.Println("🚀 Starting TTS generation for:", article.Title)
	s.repo.UpdateStatus(article.ID, "processing", "")

	// 调用 SDK (需要实现 pkg/minimax)
	audioData, err := minimax.GenerateSpeech(article.Content)
	if err != nil {
		log.Println("❌ TTS Generation failed:", err)
		s.repo.UpdateStatus(article.ID, "failed", "")
		return
	}

	// 保存文件
	filename := fmt.Sprintf("audio_%d_%s.mp3", article.ID, article.ContentHash[:8])
	savePath := filepath.Join("data", "audio", filename)

	// 确保目录存在
	os.MkdirAll(filepath.Dir(savePath), 0755)

	if err := os.WriteFile(savePath, audioData, 0644); err != nil {
		log.Println("❌ Failed to save audio file:", err)
		s.repo.UpdateStatus(article.ID, "failed", "")
		return
	}

	// 更新数据库
	s.repo.UpdateStatus(article.ID, "completed", savePath)
	log.Println("✅ TTS completed and saved to:", savePath)
}

func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
