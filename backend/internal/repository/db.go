package repository

import (
	"context"
	"fmt"
	"log"
	"time"
	"voicepaper/config"
	"voicepaper/internal/model"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
)

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.Config) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		log.Printf("⚠️  Redis连接失败: %v (缓存功能将不可用)", err)
		RDB = nil
		return
	}
	log.Println("✅ Redis connected successfully.")
}

// InitDB 初始化数据库连接（MySQL）
func InitDB(cfg *config.Config) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%v&loc=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
		cfg.Database.Charset,
		cfg.Database.ParseTime,
		cfg.Database.Loc,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// 设置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("❌ Failed to get database instance:", err)
	}

	if cfg.Database.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	}
	if cfg.Database.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	}
	if cfg.Database.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)
	}

	log.Println("✅ Database connected successfully.")
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
	// 使用 Unscoped() 避免 deleted_at 过滤，因为可能没有软删除数据
	err := DB.Unscoped().Preload("Category").Preload("Sentences").Preload("Words").First(&article, id).Error
	if err != nil {
		return nil, err
	}

	// 如果ArticleURL存在，尝试加载文章内容
	if article.ArticleURL != "" {
		content, err := r.loadArticleContent(article.ArticleURL)
		if err == nil {
			article.Content = content
		}
	}

	return &article, nil
}

// loadArticleContent 从URL加载文章内容（支持OSS和HTTP）
// 这个方法现在由API层处理，repository不再负责加载内容
func (r *ArticleRepository) loadArticleContent(url string) (string, error) {
	return "", fmt.Errorf("use API layer to load content")
}

func (r *ArticleRepository) FindByHash(hash string) (*model.Article, error) {
	// 注意：新版本数据库可能没有content_hash字段
	// 如果需要，可以通过title或其他字段查找
	var article model.Article
	err := DB.Where("title = ?", hash).First(&article).Error // 临时实现，需要根据实际需求调整
	return &article, err
}

// UpdateAudioURL 更新音频URL
func (r *ArticleRepository) UpdateAudioURL(id uint, audioURL string) error {
	return DB.Model(&model.Article{}).Where("id = ?", id).Update("audio_url", audioURL).Error
}

// UpdateTimelineURL 更新时间轴URL
func (r *ArticleRepository) UpdateTimelineURL(id uint, timelineURL string) error {
	return DB.Model(&model.Article{}).Where("id = ?", id).Update("timeline_url", timelineURL).Error
}

// UpdateArticleURL 更新文章URL
func (r *ArticleRepository) UpdateArticleURL(id uint, articleURL string) error {
	return DB.Model(&model.Article{}).Where("id = ?", id).Update("article_url", articleURL).Error
}

// UpdateOnline 更新上线状态
func (r *ArticleRepository) UpdateOnline(id uint, online string) error {
	return DB.Model(&model.Article{}).Where("id = ?", id).Update("online", online).Error
}

// GetAll 获取所有文章列表（首页用）
func (r *ArticleRepository) GetAll(isMiniProgram bool) ([]model.Article, error) {
	var articles []model.Article

	// 获取今天的日期（格式：2025-12-10）
	today := time.Now().Format("2006-01-02")

	// 基础查询：每日精读分类、有音频
	query := DB.Unscoped().
		Select("id", "title", "pic_url", "pic_1_1_url", "pic_5_4_url", "online", "category_id", "publish_date", "is_daily", "audio_url", "timeline_url", "article_url", "original_article_url", "created_at", "updated_at").
		Where("category_id = ? AND audio_url != ''", 1)

	if isMiniProgram {
		// 小程序端逻辑：动态获取配置
		var appConfig model.AppConfig
		if err := DB.Where("config_key = ?", "miniprogram_audit").First(&appConfig).Error; err == nil {
			if appConfig.IsAuditMode {
				// 审核模式开启：严格按照配置的时间范围过滤
				log.Printf("📊 GetAll(MiniProgram): 审核模式开启")
				if appConfig.StartDate != "" {
					query = query.Where("publish_date >= ?", appConfig.StartDate)
					log.Printf("   -> 过滤 StartDate >= %s", appConfig.StartDate)
				}
				if appConfig.EndDate != "" {
					query = query.Where("publish_date <= ?", appConfig.EndDate)
					log.Printf("   -> 过滤 EndDate <= %s", appConfig.EndDate)
				}
			} else {
				log.Printf("📊 GetAll(MiniProgram): 审核模式关闭，按发布日期过滤")
				defaultStartDate := "2024-01-01"
				query = query.Where("publish_date <= ? AND publish_date >= ?", today, defaultStartDate)
			}
		} else {
			// 降级策略：如果获取配置失败，使用默认安全策略
			log.Printf("⚠️ GetAll(MiniProgram): 获取配置失败 %v, 使用默认策略", err)
			defaultStartDate := "2024-01-01"
			query = query.Where("publish_date <= ? AND publish_date >= ?", today, defaultStartDate)
		}
	} else {
		// 网站端/默认逻辑：
		// 只显示发布日期 <= 今天的文章
		// 保留基础过滤（例如过滤掉太早期的测试数据）
		defaultStartDate := "2024-01-01" // 改回 2024 年，避免加载过旧的脏数据导致前端崩溃
		query = query.Where("publish_date <= ? AND publish_date >= ?", today, defaultStartDate)
		log.Printf("📊 GetAll(Web): 正常模式, publish_date BETWEEN %s AND %s", defaultStartDate, today)
	}

	err := query.Order("publish_date DESC, created_at DESC").
		Find(&articles).Error

	// 调试日志
	log.Printf("📊 GetAll() 查询到 %d 篇文章", len(articles))
	for i, article := range articles {
		log.Printf("  [%d] ID=%d, Title=%s, PublishDate=%s, CreatedAt=%s", i+1, article.ID, article.Title, article.PublishDate, article.CreatedAt)
	}

	return articles, err
}

// GetDailyArticles 获取每日文章列表
func (r *ArticleRepository) GetDailyArticles() ([]model.Article, error) {
	var articles []model.Article
	err := DB.Unscoped().Where("is_daily = ?", true).
		Order("created_at ASC").
		Find(&articles).Error
	return articles, err
}

// GetByCategory 根据分类获取文章（返回该分类下的所有文章）
func (r *ArticleRepository) GetByCategory(categoryID uint) ([]model.Article, error) {
	var articles []model.Article
	// 返回该分类下的所有文章（文章类型由 category.type 区分）
	// 排序优化: 先按 publish_date 降序，再按 created_at 降序
	// 修复日期: 2025-12-09
	err := DB.Unscoped().Where("category_id = ?", categoryID).
		Order("publish_date DESC, created_at DESC"). // 先按发布日期降序，再按创建时间降序
		Find(&articles).Error
	return articles, err
}
