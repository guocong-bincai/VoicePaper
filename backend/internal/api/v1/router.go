package v1

import (
	"net/http"
	"strconv"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"
	"voicepaper/internal/service"

	"github.com/gin-gonic/gin"
)

type ArticleHandler struct {
	repo       *repository.ArticleRepository
	ttsService *service.TTSService
}

func NewArticleHandler() *ArticleHandler {
	return &ArticleHandler{
		repo:       repository.NewArticleRepository(),
		ttsService: service.NewTTSService(),
	}
}

// GetArticles 获取文章列表
func (h *ArticleHandler) GetArticles(c *gin.Context) {
	articles, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, articles)
}

// GetArticle 获取文章详情 (包含音频信息)
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

	c.JSON(http.StatusOK, article)
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

// RegisterRoutes 注册路由
func RegisterRoutes(r *gin.Engine) {
	h := NewArticleHandler()
	v1 := r.Group("/api/v1")
	{
		v1.GET("/articles", h.GetArticles)
		v1.GET("/articles/:id", h.GetArticle)
		v1.POST("/articles", h.CreateArticle)
	}
}
