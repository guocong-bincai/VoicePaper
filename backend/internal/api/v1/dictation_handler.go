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

type DictationHandler struct {
	repo         *repository.DictationRecordRepository
	progressRepo *repository.DictationProgressRepository
	pointService *service.PointService
}

func NewDictationHandler(db *gorm.DB) *DictationHandler {
	return &DictationHandler{
		repo:         repository.NewDictationRecordRepository(),
		progressRepo: repository.NewDictationProgressRepository(db),
		pointService: service.NewPointService(db),
	}
}

// SaveRecord 保存默写记录
// POST /api/v1/dictation/record
func (h *DictationHandler) SaveRecord(c *gin.Context) {
	var req struct {
		ArticleID     uint   `json:"article_id" binding:"required"`
		DictationType string `json:"dictation_type" binding:"required"` // "word" | "sentence"
		WordID        *uint  `json:"word_id,omitempty"`
		SentenceID    *uint  `json:"sentence_id,omitempty"`
		UserAnswer    string `json:"user_answer" binding:"required"`
		IsCorrect     bool   `json:"is_correct"`
		Score         int    `json:"score"`
		AttemptCount  int    `json:"attempt_count"`
		TimeSpent     int    `json:"time_spent"` // 秒
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 验证类型
	dictationType := model.DictationType(req.DictationType)
	if dictationType != model.DictationTypeWord && dictationType != model.DictationTypeSentence {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dictation_type必须是'word'或'sentence'"})
		return
	}

	// 验证ID
	if dictationType == model.DictationTypeWord && req.WordID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "单词默写必须提供word_id"})
		return
	}
	if dictationType == model.DictationTypeSentence && req.SentenceID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "句子默写必须提供sentence_id"})
		return
	}

	// 获取用户ID（如果已登录）
	var userID *uint
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uint); ok {
			userID = &uid
		}
	}

	// 创建记录
	record := &model.DictationRecord{
		UserID:        userID,
		ArticleID:     req.ArticleID,
		DictationType: dictationType,
		WordID:        req.WordID,
		SentenceID:    req.SentenceID,
		UserAnswer:    req.UserAnswer,
		IsCorrect:     req.IsCorrect,
		Score:         req.Score,
		AttemptCount:  req.AttemptCount,
		TimeSpent:     req.TimeSpent,
		LastAttempt:   time.Now(),
	}

	// 保存或更新记录
	if err := h.repo.UpdateOrCreate(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存记录失败", "details": err.Error()})
		return
	}

	// 如果用户已登录且答案正确，奖励积分
	var pointsEarned int
	var pointRecord *model.PointRecord
	if userID != nil && req.IsCorrect {
		// 确定积分类型
		var pointType model.PointType
		var itemText string

		if dictationType == model.DictationTypeWord {
			pointType = model.PointTypeWordDictation
			// 获取单词文本（简化处理，可以从数据库查询）
			itemText = req.UserAnswer
		} else {
			pointType = model.PointTypeSentenceDictation
			itemText = req.UserAnswer
		}

		// 奖励积分
		_, pr, err := h.pointService.AwardDictationPoints(*userID, pointType, true, itemText, &record.ID)
		if err != nil {
			log.Printf("⚠️ 奖励默写积分失败: %v", err)
		} else {
			pointsEarned = pr.Points
			pointRecord = pr
		}
	}

	response := gin.H{
		"message": "记录保存成功",
		"record":  record,
	}

	// 如果有积分奖励，添加到响应中
	if pointsEarned > 0 {
		response["points_earned"] = pointsEarned
		response["point_record"] = pointRecord
	}

	c.JSON(http.StatusOK, response)
}

// GetStatistics 获取默写统计
// GET /api/v1/dictation/statistics/:article_id
func (h *DictationHandler) GetStatistics(c *gin.Context) {
	idStr := c.Param("article_id")
	articleID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	// 获取用户ID（如果已登录）
	var userID *uint
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uint); ok {
			userID = &uid
		}
	}

	stats, err := h.repo.GetStatistics(userID, uint(articleID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRecords 获取用户的默写记录
// GET /api/v1/dictation/records/:article_id
func (h *DictationHandler) GetRecords(c *gin.Context) {
	idStr := c.Param("article_id")
	articleID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	// 获取用户ID（如果已登录）
	var userID *uint
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uint); ok {
			userID = &uid
		}
	}

	records, err := h.repo.FindByUserIDAndArticleID(userID, uint(articleID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取记录失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

// GetProgress 获取默写进度
// GET /api/v1/dictation/progress/:article_id?type=word
func (h *DictationHandler) GetProgress(c *gin.Context) {
	idStr := c.Param("article_id")
	articleID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	dictationType := c.Query("type")
	if dictationType == "" {
		dictationType = "word"
	}

	// 获取用户ID（如果已登录）
	var userID uint
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uint); ok {
			userID = uid
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	progress, err := h.progressRepo.GetProgress(userID, uint(articleID), dictationType)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有进度记录，返回初始状态
			c.JSON(http.StatusOK, gin.H{
				"current_index": 0,
				"total_items":   0,
				"score":         0,
				"completed":     false,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取进度失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// SaveProgress 保存默写进度
// POST /api/v1/dictation/progress
func (h *DictationHandler) SaveProgress(c *gin.Context) {
	var req struct {
		ArticleID     uint   `json:"article_id" binding:"required"`
		DictationType string `json:"dictation_type" binding:"required"` // "word" | "sentence"
		CurrentIndex  int    `json:"current_index" binding:"required"`
		TotalItems    int    `json:"total_items" binding:"required"`
		Score         int    `json:"score"`
		Completed     bool   `json:"completed"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 获取用户ID（如果已登录）
	var userID uint
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uint); ok {
			userID = uid
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 保存进度
	err := h.progressRepo.CreateOrUpdateProgress(
		userID,
		req.ArticleID,
		req.DictationType,
		req.CurrentIndex,
		req.TotalItems,
		req.Score,
		req.Completed,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存进度失败", "details": err.Error()})
		return
	}

	response := gin.H{"message": "进度保存成功"}

	// 如果完成了整篇文章的默写，奖励额外积分
	if req.Completed {
		// 获取文章信息（需要查询数据库）
		articleRepo := repository.NewArticleRepository()
		article, err := articleRepo.FindByID(req.ArticleID)
		if err == nil {
			_, pr, err := h.pointService.AwardCompleteArticlePoints(userID, req.ArticleID, article.Title)
			if err != nil {
				log.Printf("⚠️ 奖励完成文章积分失败: %v", err)
			} else {
				response["bonus_points"] = pr.Points
				response["bonus_message"] = "恭喜完成整篇文章默写！"
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// ResetProgress 重置默写进度
// DELETE /api/v1/dictation/progress/:article_id?type=word
func (h *DictationHandler) ResetProgress(c *gin.Context) {
	idStr := c.Param("article_id")
	articleID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	dictationType := c.Query("type")
	if dictationType == "" {
		dictationType = "word"
	}

	// 获取用户ID（如果已登录）
	var userID uint
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uint); ok {
			userID = uid
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	err = h.progressRepo.ResetProgress(userID, uint(articleID), dictationType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置进度失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "进度重置成功"})
}
