package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"voicepaper/config"
	"voicepaper/internal/repository"
	"voicepaper/internal/service"
	"voicepaper/internal/storage"

	"github.com/gin-gonic/gin"
)

type ArticleHandler struct {
	repo       *repository.ArticleRepository
	ttsService *service.TTSService
	storage    storage.Storage
	isOSS      bool
}

func NewArticleHandler() *ArticleHandler {
	cfg := config.GetConfig()
	var st storage.Storage
	var err error
	isOSS := false

	// 根据配置创建存储实例
	if cfg.Storage.Type == "oss" {
		st, err = storage.NewOSSStorage(cfg)
		if err != nil {
			log.Printf("❌ OSS存储初始化失败: %v", err)
			// 如果OSS初始化失败，使用本地存储作为降级方案
			st = storage.NewLocalStorage(cfg)
			log.Printf("⚠️  已降级使用本地存储")
		} else {
			log.Printf("✅ OSS存储初始化成功")
			isOSS = true
		}
	} else {
		st = storage.NewLocalStorage(cfg)
		log.Printf("✅ 使用本地存储")
	}

	return &ArticleHandler{
		repo:       repository.NewArticleRepository(),
		ttsService: service.NewTTSService(),
		storage:    st,
		isOSS:      isOSS,
	}
}

// GetArticles 获取文章列表
func (h *ArticleHandler) GetArticles(c *gin.Context) {
	// 判断是否是小程序请求
	userAgent := c.GetHeader("User-Agent")
	isMiniProgram := strings.Contains(userAgent, "MicroMessenger") || strings.Contains(userAgent, "miniProgram")
	if c.Query("from") == "miniprogram" {
		isMiniProgram = true
	}

	articles, err := h.repo.GetAll(isMiniProgram)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("⚡ GetArticles: 总文章数=%d", len(articles))
	if len(articles) > 0 {
		log.Printf("⚡ 第一篇文章 PicURL=%s", articles[0].PicURL)
	}

	c.JSON(http.StatusOK, articles)
}

// GetArticle 获取文章详情 (包含音频、时间轴、文章内容)
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	article, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 加载文章内容（从ArticleURL）
	var content string
	if article.ArticleURL != "" {
		content, err = h.loadArticleContent(article.ArticleURL)
		if err != nil {
			log.Printf("⚠️  Failed to load article content from %s: %v", article.ArticleURL, err)
		}
	}

	// 生成音频签名URL（如果audio_url是OSS URL且storage已配置）
	audioURL := article.AudioURL
	if article.AudioURL != "" && h.storage != nil && strings.Contains(article.AudioURL, "oss-cn-") {
		ctx := context.Background()
		// 提取OSS路径
		parts := strings.Split(article.AudioURL, ".aliyuncs.com/")
		if len(parts) > 1 {
			path := parts[1]
			// 生成签名URL，有效期24小时（86400秒）
			signedURL, err := h.storage.GetSignedURL(ctx, path, 86400)
			if err == nil {
				audioURL = signedURL
				log.Printf("✅ 生成音频签名URL成功: %s", signedURL[:50]+"...")
			} else {
				log.Printf("⚠️  生成音频签名URL失败: %v，使用原始URL", err)
			}
		} else {
			log.Printf("⚠️  无法解析OSS路径: %s", article.AudioURL)
		}
	} else {
		if article.AudioURL == "" {
			log.Printf("⚠️  文章没有audio_url")
		} else if h.storage == nil {
			log.Printf("⚠️  storage未初始化")
		} else if !strings.Contains(article.AudioURL, "oss-cn-") {
			log.Printf("⚠️  不是OSS URL: %s", article.AudioURL)
		}
	}

	// 🔧 原文URL不再由后端签名，交由前端统一处理
	// 原因：后端 SignURL 会对路径进行编码（/ -> %2F），导致小程序端 wx.request 请求时签名失败
	// 前端会在使用时通过 oss.ts 的 getSignedUrl 函数进行统一签名处理
	originalArticleURL := article.OriginalArticleURL
	if originalArticleURL != "" {
		log.Printf("📄 原文URL (交由前端签名): %s", originalArticleURL)
	}

	// 构建响应
	response := gin.H{
		"id":                   article.ID,
		"title":                article.Title,
		"pic_url":              article.PicURL,   // 主封面图
		"pic_1_1_url":          article.Pic11URL, // 1:1 朋友圈分享图
		"pic_5_4_url":          article.Pic54URL, // 5:4 聊天框分享图
		"online":               article.Online,
		"category_id":          article.CategoryID,
		"publish_date":         article.PublishDate,
		"is_daily":             article.IsDaily,
		"audio_url":            audioURL, // 使用签名URL（如果是OSS）
		"timeline_url":         article.TimelineURL,
		"article_url":          article.ArticleURL,
		"original_article_url": originalArticleURL, // 使用签名URL（如果是OSS）
		"content":              content,            // 从ArticleURL加载的内容
		"created_at":           article.CreatedAt,
		"updated_at":           article.UpdatedAt,
		"sentences":            article.Sentences,
		"words":                article.Words,
	}

	// 如果有分类信息，添加分类名称
	if article.Category != nil {
		response["category"] = gin.H{
			"id":   article.Category.ID,
			"name": article.Category.Name,
		}
	}

	c.JSON(http.StatusOK, response)
}

// loadArticleContent 从URL加载文章内容
func (h *ArticleHandler) loadArticleContent(url string) (string, error) {
	if h.storage == nil {
		return "", fmt.Errorf("storage not configured")
	}

	ctx := context.Background()

	// 如果是OSS URL，需要提取路径
	// 例如：https://voicepaper.oss-cn-chengdu.aliyuncs.com/article/article_1.md
	// 需要提取：article/article_1.md

	// 简单处理：如果URL包含OSS域名，提取路径部分
	// 否则直接使用URL作为路径
	path := url
	if strings.Contains(url, "oss-cn-") {
		// 提取OSS路径
		parts := strings.Split(url, ".aliyuncs.com/")
		if len(parts) > 1 {
			path = parts[1]
		}
	}

	// 从存储获取内容
	content, err := h.storage.Get(ctx, path)
	if err != nil {
		return "", fmt.Errorf("failed to load content: %w", err)
	}

	return string(content), nil
}

// GetArticleTimeline 获取文章时间轴数据
// GET /api/v1/articles/:id/timeline
func (h *ArticleHandler) GetArticleTimeline(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	article, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	if article.TimelineURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Timeline not found"})
		return
	}

	// 加载时间轴数据
	timelineData, err := h.loadTimelineData(article.TimelineURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load timeline data"})
		return
	}

	c.JSON(http.StatusOK, timelineData)
}

// loadTimelineData 从URL加载时间轴数据
func (h *ArticleHandler) loadTimelineData(url string) (interface{}, error) {
	if h.storage == nil {
		return nil, fmt.Errorf("storage not configured")
	}

	ctx := context.Background()

	// 提取路径
	path := url
	if strings.Contains(url, "oss-cn-") {
		parts := strings.Split(url, ".aliyuncs.com/")
		if len(parts) > 1 {
			path = parts[1]
		}
	}

	// 从存储获取JSON内容
	content, err := h.storage.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load timeline: %w", err)
	}

	// 解析JSON
	var timelineData interface{}
	if err := json.Unmarshal(content, &timelineData); err != nil {
		return nil, fmt.Errorf("failed to parse timeline JSON: %w", err)
	}

	return timelineData, nil
}

// CreateArticle 创建文章并触发 TTS (如果需要)
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 核心逻辑：检查是否已存在，不存在则生成
	article, err := h.ttsService.GetOrGenerateAudio(req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, article)
}

// GetWords 获取文章的重点单词
// GET /api/v1/articles/:id/words
func (h *ArticleHandler) GetWords(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	article, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 只返回重点单词
	var keyWords []gin.H
	for _, word := range article.Words {
		if word.IsKeyWord {
			keyWords = append(keyWords, gin.H{
				"id":                  word.ID,
				"article_id":          word.ArticleID,
				"text":                word.Text,
				"phonetic":            word.Phonetic,
				"meaning":             word.Meaning,
				"example":             word.Example,
				"example_translation": word.ExampleTranslation,
				"level":               word.Level,
				"frequency":           word.Frequency,
				"order":               word.Order,
				"is_key_word":         word.IsKeyWord,
			})
		}
	}

	c.JSON(http.StatusOK, keyWords)
}

// GetSentences 获取文章的句子
// GET /api/v1/articles/:id/sentences
func (h *ArticleHandler) GetSentences(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	article, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 返回所有句子
	var sentences []gin.H
	for _, sentence := range article.Sentences {
		sentences = append(sentences, gin.H{
			"id":          sentence.ID,
			"article_id":  sentence.ArticleID,
			"text":        sentence.Text,
			"translation": sentence.Translation,
			"order":       sentence.Order,
		})
	}

	c.JSON(http.StatusOK, sentences)
}

// GetCategories 获取所有分类/合集（包括文章和场景对话）
func (h *ArticleHandler) GetCategories(c *gin.Context) {
	var categories []struct {
		ID           uint   `json:"id"`
		Name         string `json:"name"`
		Description  string `json:"description"`
		Icon         string `json:"icon"`
		Sort         int    `json:"sort"`
		ArticleCount int64  `json:"article_count"`
	}

	// 查询所有启用的分类，并统计每个分类下的文章/句子数
	result := repository.DB.Table("vp_categories").
		Select("vp_categories.id, vp_categories.name, vp_categories.description, vp_categories.icon, vp_categories.sort, COUNT(vp_articles.id) as article_count").
		Joins("LEFT JOIN vp_articles ON vp_categories.id = vp_articles.category_id AND vp_articles.deleted_at IS NULL").
		Where("vp_categories.is_active = ? AND vp_categories.deleted_at IS NULL", true).
		Group("vp_categories.id").
		Order("vp_categories.sort ASC, vp_categories.id ASC").
		Scan(&categories)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 处理OSS图片URL签名
	if h.storage != nil {
		ctx := context.Background()
		for i := range categories {
			if categories[i].Icon != "" && strings.Contains(categories[i].Icon, "oss-cn-") {
				// 提取OSS路径
				parts := strings.Split(categories[i].Icon, ".aliyuncs.com/")
				if len(parts) > 1 {
					path := parts[1]
					// 生成签名URL，有效期24小时
					signedURL, err := h.storage.GetSignedURL(ctx, path, 3153600000) // 100 years expiration
					if err == nil {
						categories[i].Icon = signedURL
					} else {
						log.Printf("⚠️  生成分类图标签名URL失败: %v", err)
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, categories)
}

// GetArticlesByCategory 根据分类获取文章列表（支持所有类型）
func (h *ArticleHandler) GetArticlesByCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	// 验证分类是否存在
	var category struct {
		Type uint8 `json:"type"`
	}
	result := repository.DB.Table("vp_categories").
		Select("type").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&category)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	articles, err := h.repo.GetByCategory(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, articles)
}

// RegisterRoutes 注册路由
func RegisterRoutes(r *gin.Engine) {
	articleHandler := NewArticleHandler()
	authHandler := NewAuthHandler()
	dictationHandler := NewDictationHandler(repository.DB)
	feedbackHandler := NewFeedbackHandler(repository.DB)
	pointHandler := NewPointHandler()
	checkInHandler := NewCheckInHandler()
	titleHandler := NewTitleHandler()
	readingProgressHandler := NewReadingProgressHandler(repository.DB)
	rankingHandler := NewRankingHandler(repository.DB)
	durationHandler := NewDurationHandler(repository.DB)
	wordbookHandler := NewWordbookHandler(repository.DB)
	reportHandler := NewReportHandler(repository.DB)
	appConfigHandler := NewAppConfigHandler()
	bookHandler := NewBookHandler() // 添加书籍处理器

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		// 系统配置
		v1.GET("/config/audit", appConfigHandler.GetAuditConfig)

		// 书籍相关路由
		// 注意：更具体的路由要放在更通用的路由之前，避免路由冲突
		v1.GET("/books", bookHandler.GetBooks)                               // 获取书籍列表（支持搜索和类型筛选）
		v1.GET("/books/:book_id/points", bookHandler.GetBookPoints)          // 获取书籍章节列表
		v1.GET("/books/:book_id/points/:point_id", bookHandler.GetBookPoint) // 获取章节详情
		v1.GET("/books/:book_id", bookHandler.GetBookByBookID)               // 根据book_id获取书籍详情（统一使用book_id）

		// 分类/合集相关路由
		v1.GET("/categories", articleHandler.GetCategories)                      // 获取所有分类
		v1.GET("/categories/:id/articles", articleHandler.GetArticlesByCategory) // 获取分类下的文章

		// 单词书相关路由
		// 注意：更具体的路由要放在更通用的路由之前，避免路由冲突
		v1.GET("/wordbooks", wordbookHandler.GetWordbooks)
		v1.GET("/wordbooks/debug/books", wordbookHandler.DebugWordbookBooks) // 调试API：检查vp_wordbook_books表数据
		v1.POST("/wordbooks/progress", authHandler.AuthMiddleware(), wordbookHandler.SaveProgress)

		// 顺序/乱序学习相关路由（需要认证）- 必须放在 :wordbook_id 和 :type 之前
		v1.GET("/wordbooks/order/:type/words", authHandler.AuthMiddleware(), wordbookHandler.GetWordsWithOrder)

		// 单词导入和学习记录
		v1.POST("/wordbooks/word/:wordbook_id/import", authHandler.AuthMiddleware(), wordbookHandler.ImportWordToVocabulary)
		v1.POST("/wordbooks/word/:wordbook_id/study", authHandler.AuthMiddleware(), wordbookHandler.StudyWord)

		// 学习序列管理
		v1.GET("/wordbooks/seq/:type/order", authHandler.AuthMiddleware(), wordbookHandler.GetUserOrder)
		v1.POST("/wordbooks/seq/:type/order", authHandler.AuthMiddleware(), wordbookHandler.SwitchOrderMode)
		v1.PUT("/wordbooks/seq/:type/order/index", authHandler.AuthMiddleware(), wordbookHandler.UpdateOrderIndex)

		// 进度和单词列表
		v1.GET("/wordbooks/:type/progress", authHandler.AuthMiddleware(), wordbookHandler.GetProgress)
		v1.GET("/wordbooks/:type/words", wordbookHandler.GetWords)

		// 文章相关路由
		// 注意：更具体的路由要放在更通用的路由之前
		v1.GET("/articles", articleHandler.GetArticles)
		v1.GET("/articles/:id/timeline", articleHandler.GetArticleTimeline)
		v1.GET("/articles/:id/export/pdf", articleHandler.ExportArticlePDF) // 导出文章PDF
		v1.GET("/articles/:id/words", articleHandler.GetWords)              // 获取文章的重点单词
		v1.GET("/articles/:id/sentences", articleHandler.GetSentences)      // 获取文章的句子
		v1.GET("/articles/:id", articleHandler.GetArticle)                  // 这个要放在最后，因为它是通用路由
		v1.POST("/articles", articleHandler.CreateArticle)

		// 默写练习相关路由
		dictation := v1.Group("/dictation")
		{
			// BUG修复: 添加OptionalAuthMiddleware以支持已登录用户获取积分
			dictation.POST("/record", authHandler.OptionalAuthMiddleware(), dictationHandler.SaveRecord)                   // 保存默写记录（可选认证）
			dictation.GET("/statistics/:article_id", authHandler.OptionalAuthMiddleware(), dictationHandler.GetStatistics) // 获取统计（可选认证）
			dictation.GET("/records/:article_id", authHandler.OptionalAuthMiddleware(), dictationHandler.GetRecords)       // 获取记录（可选认证）

			// 进度相关路由（需要认证）
			dictation.GET("/progress/:article_id", authHandler.AuthMiddleware(), dictationHandler.GetProgress)      // 获取进度
			dictation.POST("/progress", authHandler.AuthMiddleware(), dictationHandler.SaveProgress)                // 保存进度
			dictation.DELETE("/progress/:article_id", authHandler.AuthMiddleware(), dictationHandler.ResetProgress) // 重置进度
		}

		// 认证相关路由
		auth := v1.Group("/auth")
		{
			// 邮箱验证码登录
			auth.POST("/email/send", authHandler.SendEmailCode)           // 发送验证码
			auth.POST("/email/verify", authHandler.VerifyEmailCode)       // 验证邮箱验证码（注册流程用）
			auth.POST("/email/login", authHandler.EmailLogin)             // 邮箱登录
			auth.POST("/email/clear", authHandler.ClearVerificationCodes) // 清理验证码（测试用）

			// 密码登录（支持普通用户和管理员）
			auth.POST("/password/login", authHandler.PasswordLogin) // 密码登录
			auth.POST("/password/reset", authHandler.ResetPassword) // 重置密码

			// 用户注册
			auth.POST("/register", authHandler.Register) // 用户注册

			// 邀请码相关
			auth.POST("/invite/verify", authHandler.VerifyInviteCode) // 验证邀请码

			// GitHub OAuth登录
			auth.GET("/github", authHandler.GitHubAuth)              // 跳转到GitHub授权
			auth.GET("/github/callback", authHandler.GitHubCallback) // GitHub回调

			// Google OAuth登录
			auth.GET("/google", authHandler.GoogleAuth)              // 跳转到Google授权
			auth.GET("/google/callback", authHandler.GoogleCallback) // Google回调

			// 微信小程序登录
			auth.POST("/wechat/login", authHandler.WeChatLogin) // 微信登录

			// 需要认证的路由
			auth.GET("/me", authHandler.AuthMiddleware(), authHandler.GetMe)                    // 获取当前用户信息
			auth.PUT("/profile", authHandler.AuthMiddleware(), authHandler.UpdateProfile)       // 更新用户资料
			auth.POST("/avatar/upload", authHandler.AuthMiddleware(), authHandler.UploadAvatar) // 上传头像
			auth.POST("/wechat/bind", authHandler.AuthMiddleware(), authHandler.BindWeChat)     // 绑定微信
			auth.POST("/email/bind", authHandler.AuthMiddleware(), authHandler.BindEmail)       // 微信用户绑定邮箱
			auth.POST("/logout", authHandler.AuthMiddleware(), authHandler.Logout)              // 登出
		}

		// 问题反馈相关路由
		// BUG修复: 添加可选认证中间件，已登录用户自动关联user_id
		v1.POST("/feedback", authHandler.OptionalAuthMiddleware(), feedbackHandler.SubmitFeedback) // 提交反馈（可选认证）

		// 积分系统相关路由（需要认证）
		points := v1.Group("/points")
		points.Use(authHandler.AuthMiddleware())
		{
			points.GET("/me", pointHandler.GetMyPoints)                // 获取我的积分信息
			points.GET("/records", pointHandler.GetPointRecords)       // 获取积分记录
			points.GET("/statistics", pointHandler.GetPointStatistics) // 获取积分统计
			points.POST("/award/read", pointHandler.AwardReadPoints)   // 奖励阅读积分
		}

		// 签到相关路由（需要认证）
		checkIn := v1.Group("/check-in")
		checkIn.Use(authHandler.AuthMiddleware())
		{
			checkIn.POST("", checkInHandler.CheckIn)                       // 每日签到
			checkIn.GET("/status", checkInHandler.GetCheckInStatus)        // 获取签到状态
			checkIn.GET("/calendar", checkInHandler.GetCheckInCalendar)    // 获取签到日历
			checkIn.GET("/ranking", checkInHandler.GetCheckInRanking)      // 获取签到排行榜
			checkIn.GET("/makeup-card", checkInHandler.GetMakeupCardInfo)  // 获取补签卡信息
			checkIn.POST("/makeup-card/buy", checkInHandler.BuyMakeupCard) // 购买补签卡
			checkIn.POST("/makeup-card/use", checkInHandler.UseMakeupCard) // 使用补签卡
		}

		// 称号相关路由
		titles := v1.Group("/titles")
		{
			titles.GET("/all", titleHandler.GetAllTitles) // 获取所有称号（无需认证）

			// 需要认证的称号路由
			titles.Use(authHandler.AuthMiddleware())
			titles.GET("/me", titleHandler.GetMyTitles)                         // 获取我的称号
			titles.GET("/progress", titleHandler.GetTitleProgress)              // 获取称号进度
			titles.GET("/equipped", titleHandler.GetEquippedTitle)              // 获取当前佩戴的称号
			titles.POST("/equip", titleHandler.EquipTitle)                      // 佩戴称号
			titles.POST("/unequip", titleHandler.UnequipTitle)                  // 取消佩戴称号
			titles.GET("/category/:category", titleHandler.GetTitlesByCategory) // 根据分类获取称号
		}

		// 阅读进度相关路由（需要认证）
		readingProgress := v1.Group("/reading-progress")
		readingProgress.Use(authHandler.AuthMiddleware())
		{
			readingProgress.GET("/stats", readingProgressHandler.GetStats)          // 获取阅读统计（放在前面避免被:article_id匹配）
			readingProgress.GET("/history", readingProgressHandler.GetHistory)      // 获取阅读历史
			readingProgress.GET("/last", readingProgressHandler.GetLastRead)        // 获取最后阅读的文章
			readingProgress.GET("/:article_id", readingProgressHandler.GetProgress) // 获取某篇文章的阅读进度
			readingProgress.POST("", readingProgressHandler.SaveProgress)           // 保存阅读进度
			readingProgress.POST("/quick", readingProgressHandler.QuickSave)        // 快速保存进度（高频更新用）
		}

		// 排名相关路由
		ranking := v1.Group("/ranking")
		{
			ranking.GET("/global", rankingHandler.GetGlobalRanking) // 获取全球排名列表（无需认证）
		}

		// 需要认证的排名路由
		rankingAuth := v1.Group("/ranking")
		rankingAuth.Use(authHandler.AuthMiddleware())
		{
			rankingAuth.GET("/me", rankingHandler.GetMyRanking) // 获取我的排名（需要认证）
		}

		// 学习时长相关路由（需要认证）
		duration := v1.Group("/duration")
		duration.Use(authHandler.AuthMiddleware())
		{
			duration.POST("/sync", durationHandler.SyncReadingDuration) // 同步阅读时长
			duration.GET("/me", durationHandler.GetUserDuration)        // 获取用户累积时长
		}

		// 每日统计相关路由（需要认证）- 小程序端使用
		dailyStats := v1.Group("/daily-stats")
		dailyStats.Use(authHandler.AuthMiddleware())
		{
			dailyStatsHandler := NewDailyStatsHandler(repository.DB)
			dailyStats.POST("/duration", dailyStatsHandler.SyncDuration)                  // 同步今日时长
			dailyStats.POST("/new-word", dailyStatsHandler.RecordNewWord)                 // 记录新学单词
			dailyStats.POST("/review", dailyStatsHandler.RecordReview)                    // 记录复习
			dailyStats.GET("/today", dailyStatsHandler.GetTodayStats)                     // 获取今日统计
			dailyStats.GET("/vocabulary-count", dailyStatsHandler.GetUserVocabularyCount) // 获取用户词汇量
		}

		// 年终报告相关路由（需要认证）
		report := v1.Group("/report")
		report.Use(authHandler.AuthMiddleware())
		{
			report.GET("/year-end", reportHandler.GetYearEndReport) // 获取年终总结数据
		}

		// 生词本相关路由（需要认证）
		vocabulary := v1.Group("/vocabulary")
		vocabulary.Use(authHandler.AuthMiddleware())
		{
			// 生词 CRUD
			vocabulary.POST("", AddVocabulary)                 // 添加生词
			vocabulary.GET("", ListVocabulary)                 // 获取生词列表
			vocabulary.GET("/stats", GetVocabularyStats)       // 获取统计（放在:id前面）
			vocabulary.GET("/:id", GetVocabulary)              // 获取单个生词
			vocabulary.PUT("/:id", UpdateVocabulary)           // 更新生词
			vocabulary.DELETE("/:id", DeleteVocabulary)        // 删除生词
			vocabulary.POST("/:id/star", ToggleVocabularyStar) // 切换标星
			vocabulary.DELETE("/batch", BatchDeleteVocabulary) // 批量删除

			// 复习功能
			vocabulary.GET("/review/today", GetTodayReviewList) // 获取今日待复习
			vocabulary.POST("/:id/review", SubmitReview)        // 提交复习结果

			// 每日统计
			vocabulary.GET("/stats/daily", GetVocabularyDailyStats) // 获取每日学习统计

			// 文件夹管理
			vocabulary.GET("/folders", ListVocabularyFolders)                             // 获取文件夹列表
			vocabulary.POST("/folders", CreateVocabularyFolder)                           // 创建文件夹
			vocabulary.PUT("/folders/:id", UpdateVocabularyFolder)                        // 更新文件夹
			vocabulary.DELETE("/folders/:id", DeleteVocabularyFolder)                     // 删除文件夹
			vocabulary.GET("/folders/:id/items", ListFolderVocabulary)                    // 获取文件夹中的生词
			vocabulary.POST("/folders/:id/items", AddVocabularyToFolder)                  // 添加生词到文件夹
			vocabulary.DELETE("/folders/:id/items/:vocab_id", RemoveVocabularyFromFolder) // 从文件夹移除生词
		}
	}
}
