package repository

import (
	"log"
	"voicepaper/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// 自动迁移 Schema
	err = DB.AutoMigrate(&model.Article{}, &model.Sentence{})
	if err != nil {
		log.Fatal("❌ Failed to migrate database:", err)
	}

	log.Println("✅ Database connected and migrated successfully.")
}

// ArticleRepository 封装文章相关的数据库操作
type ArticleRepository struct{}

func NewArticleRepository() *ArticleRepository {
	return &ArticleRepository{}
}

func (r *ArticleRepository) Create(article *model.Article) error {
	return DB.Create(article).Error
}

func (r *ArticleRepository) FindByID(id uint) (*model.Article, error) {
	var article model.Article
	err := DB.Preload("Sentences").First(&article, id).Error
	return &article, err
}

func (r *ArticleRepository) FindByHash(hash string) (*model.Article, error) {
	var article model.Article
	err := DB.Where("content_hash = ?", hash).First(&article).Error
	return &article, err
}

func (r *ArticleRepository) UpdateStatus(id uint, status string, audioPath string) error {
	return DB.Model(&model.Article{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"audio_path": audioPath,
	}).Error
}

func (r *ArticleRepository) GetAll() ([]model.Article, error) {
	var articles []model.Article
	err := DB.Select("id", "title", "created_at", "status").Find(&articles).Error
	return articles, err
}
