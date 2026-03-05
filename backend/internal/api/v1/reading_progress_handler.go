package v1

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"
	"voicepaper/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReadingProgressHandler struct {
	repo         *repository.ReadingProgressRepository
	pointService *service.PointService
	articleRepo  *repository.ArticleRepository
}

func NewReadingProgressHandler(db *gorm.DB) *ReadingProgressHandler {
	return &ReadingProgressHandler{
		repo:         repository.NewReadingProgressRepository(db),
		pointService: service.NewPointService(db),
		articleRepo:  repository.NewArticleRepository(),
	}
}

// GetProgress 获取某篇文章的阅读进度
// GET /api/v1/reading-progress/:article_id
func (h *ReadingProgressHandler) GetProgress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	articleIDStr := c.Param("article_id")
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文章ID"})
		return
	}

	uid := userID.(uint)
	progress, err := h.repo.GetByUserAndArticle(uid, uint(articleID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有阅读记录，返回初始状态
			c.JSON(http.StatusOK, gin.H{
				"current_time": 0,
				"duration":     0,
				"progress":     0,
				"is_completed": false,
				"read_count":   0,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取阅读进度失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// SaveProgress 保存阅读进度
// POST /api/v1/reading-progress
func (h *ReadingProgressHandler) SaveProgress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		ArticleID   uint    `json:"article_id" binding:"required"`
		CurrentTime float64 `json:"current_time" binding:"required"`
		Duration    float64 `json:"duration" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 计算进度百分比
	var progressPercent float64
	if req.Duration > 0 {
		progressPercent = (req.CurrentTime / req.Duration) * 100
		if progressPercent > 100 {
			progressPercent = 100
		}
	}

	uid := userID.(uint)

	// 先获取之前的进度状态
	oldProgress, _ := h.repo.GetByUserAndArticle(uid, req.ArticleID)
	wasCompleted := oldProgress != nil && oldProgress.IsCompleted

	// 保存进度
	progress, err := h.repo.SaveProgress(uid, req.ArticleID, req.CurrentTime, req.Duration, progressPercent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存进度失败", "details": err.Error()})
		return
	}

	response := gin.H{
		"message":  "进度保存成功",
		"progress": progress,
	}

	// 如果首次达到80%完成度，奖励阅读积分
	if progressPercent >= 80 && !wasCompleted {
		// 获取文章信息
		article, err := h.articleRepo.FindByID(req.ArticleID)
		if err == nil {
			_, record, err := h.pointService.AwardReadArticlePoints(uid, req.ArticleID, article.Title)
			if err != nil {
				log.Printf("⚠️ 奖励阅读积分失败: %v", err)
			} else if record != nil {
				response["points_earned"] = record.Points
				response["points_message"] = "恭喜完成阅读！"
				log.Printf("✅ 用户 %d 完成文章《%s》阅读，获得 %d 积分", uid, article.Title, record.Points)
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetHistory 获取阅读历史
// GET /api/v1/reading-progress/history?page=1&page_size=20
func (h *ReadingProgressHandler) GetHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	uid := userID.(uint)
	progresses, total, err := h.repo.GetUserReadingHistory(uid, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取阅读历史失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history":   progresses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetLastRead 获取最后阅读的文章
// GET /api/v1/reading-progress/last
func (h *ReadingProgressHandler) GetLastRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	progress, err := h.repo.GetLastReadArticle(uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{"message": "暂无阅读记录"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取最后阅读记录失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetStats 获取阅读统计
// GET /api/v1/reading-progress/stats
func (h *ReadingProgressHandler) GetStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	stats, err := h.repo.GetUserReadingStats(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取阅读统计失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// QuickSave 快速保存进度（用于高频更新，如每5秒自动保存）
// POST /api/v1/reading-progress/quick
func (h *ReadingProgressHandler) QuickSave(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		ArticleID   uint    `json:"article_id" binding:"required"`
		CurrentTime float64 `json:"current_time" binding:"required"`
		Duration    float64 `json:"duration" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 计算进度百分比
	var progressPercent float64
	if req.Duration > 0 {
		progressPercent = (req.CurrentTime / req.Duration) * 100
		if progressPercent > 100 {
			progressPercent = 100
		}
	}

	uid := userID.(uint)

	// 使用UPSERT快速保存
	progress := &model.ReadingProgress{
		UserID:      uid,
		ArticleID:   req.ArticleID,
		CurrentTime: req.CurrentTime,
		Duration:    req.Duration,
		Progress:    progressPercent,
		IsCompleted: progressPercent >= 80,
		TotalTime:   int(req.CurrentTime), // 设置初始值
		LastReadAt:  time.Now(),
	}

	err := h.repo.CreateOrUpdate(progress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}

	// 更新total_time（取最大值）
	_ = h.repo.UpdateTotalTime(uid, req.ArticleID, req.CurrentTime)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
