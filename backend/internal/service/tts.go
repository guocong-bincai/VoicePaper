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

// GetOrGenerateAudio 核心逻辑：检查数据库 -> (如果不存在) 调用 API 生成音频
// 注意：新版本中音频存储在OSS，通过audio_url访问
func (s *TTSService) GetOrGenerateAudio(title, content string) (*model.Article, error) {
	// 1. 计算哈希（用于去重，但新版本可能不需要）
	hash := calculateHash(content)

	// 2. 检查数据库（通过标题和内容查找）
	// 注意：新版本可能不需要ContentHash，可以根据实际需求调整
	article, err := s.repo.FindByHash(hash)
	if err == nil {
		// 记录存在，检查是否有音频URL
		if article.Online == "1" && article.AudioURL != "" {
			log.Println("✅ Cache hit: Article exists with audio URL for", title)
			return article, nil
		}
		log.Println("⚠️  Article exists but no audio URL, may need regeneration...", title)
	} else {
		// 记录不存在，创建新记录
		article = &model.Article{
			Title:  title,
			Online: "0", // 默认未上线
		}
		if err := s.repo.Create(article); err != nil {
			return nil, err
		}
	}

	// 3. 调用 MiniMax API 生成音频 (异步)
	// 注意：新版本中音频应该上传到OSS，并更新audio_url字段
	go s.processTTS(article, content)

	return article, nil
}

func (s *TTSService) processTTS(article *model.Article, content string) {
	log.Println("🚀 Starting TTS generation for:", article.Title)
	// 注意：新版本中不再使用Status字段，可以添加处理中标记到其他字段

	// 调用 SDK (需要实现 pkg/minimax)
	audioData, err := minimax.GenerateSpeech(content)
	if err != nil {
		log.Println("❌ TTS Generation failed:", err)
		// 可以更新Online为"0"表示失败
		return
	}

	// 保存文件到本地临时目录（后续应该上传到OSS）
	filename := fmt.Sprintf("audio_%d.mp3", article.ID)
	savePath := filepath.Join("data", "audio", filename)

	// 确保目录存在
	os.MkdirAll(filepath.Dir(savePath), 0755)

	if err := os.WriteFile(savePath, audioData, 0644); err != nil {
		log.Println("❌ Failed to save audio file:", err)
		return
	}

	// 注意：新版本中应该上传到OSS并更新audio_url字段
	// 这里暂时保存到本地，后续需要集成OSS上传功能
	log.Println("✅ TTS completed and saved to:", savePath)
	log.Println("⚠️  Note: Audio should be uploaded to OSS and audio_url should be updated")
}

func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
