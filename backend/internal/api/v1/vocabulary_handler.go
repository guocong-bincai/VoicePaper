package v1

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"
	"voicepaper/internal/service"

	"github.com/gin-gonic/gin"
)

var (
	vocabularyService     *service.VocabularyService
	vocabularyServiceOnce sync.Once
)

// getVocabularyService 惰性初始化生词本服务
func getVocabularyService() *service.VocabularyService {
	vocabularyServiceOnce.Do(func() {
		vocabularyService = service.NewVocabularyService()
	})
	return vocabularyService
}

// ==================== 生词 CRUD ====================

// AddVocabulary 添加生词
// POST /api/v1/vocabulary
func AddVocabulary(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req service.AddVocabularyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	vocab, err := getVocabularyService().AddVocabulary(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "添加成功",
		"data":    vocab,
	})
}

// GetVocabulary 获取单个生词
// GET /api/v1/vocabulary/:id
func GetVocabulary(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	vocab, err := getVocabularyService().GetVocabulary(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vocab)
}

// UpdateVocabulary 更新生词
// PUT /api/v1/vocabulary/:id
func UpdateVocabulary(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req service.UpdateVocabularyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	vocab, err := getVocabularyService().UpdateVocabulary(uint(id), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    vocab,
	})
}

// DeleteVocabulary 删除生词
// DELETE /api/v1/vocabulary/:id
func DeleteVocabulary(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := getVocabularyService().DeleteVocabulary(uint(id), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ListVocabulary 获取生词列表
// GET /api/v1/vocabulary
func ListVocabulary(c *gin.Context) {
	userID := c.GetUint("user_id")

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 解析查询参数
	params := &repository.VocabularyListParams{
		Type:        c.Query("type"),
		Keyword:     c.Query("keyword"),
		OrderBy:     c.Query("order_by"),
		WithArticle: c.Query("with_article") == "true",
	}

	if masteryStr := c.Query("mastery_level"); masteryStr != "" {
		if mastery, err := strconv.Atoi(masteryStr); err == nil {
			params.MasteryLevel = &mastery
		}
	}

	if starredStr := c.Query("is_starred"); starredStr != "" {
		starred := starredStr == "true"
		params.IsStarred = &starred
	}

	if articleIDStr := c.Query("article_id"); articleIDStr != "" {
		if articleID, err := strconv.ParseUint(articleIDStr, 10, 64); err == nil {
			aid := uint(articleID)
			params.ArticleID = &aid
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	} else {
		params.Limit = 20 // 默认每页20条
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params.Offset = offset
		}
	}

	vocabs, total, err := getVocabularyService().ListVocabulary(userID, params)
	if err != nil {
		log.Printf("[ListVocabulary] 错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

	// 确保返回空数组而不是 null
	if vocabs == nil {
		vocabs = make([]model.Vocabulary, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   vocabs,
		"total":  total,
		"limit":  params.Limit,
		"offset": params.Offset,
	})
}

// GetVocabularyStats 获取生词统计
// GET /api/v1/vocabulary/stats
func GetVocabularyStats(c *gin.Context) {
	userID := c.GetUint("user_id")

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	stats, err := getVocabularyService().GetStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ToggleVocabularyStar 切换标星状态
// POST /api/v1/vocabulary/:id/star
func ToggleVocabularyStar(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	newState, err := getVocabularyService().ToggleStar(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "操作成功",
		"is_starred": newState,
	})
}

// BatchDeleteVocabulary 批量删除生词
// DELETE /api/v1/vocabulary/batch
func BatchDeleteVocabulary(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := getVocabularyService().BatchDelete(userID, req.IDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ==================== 复习功能 ====================

// GetTodayReviewList 获取今日待复习列表
// GET /api/v1/vocabulary/review/today
func GetTodayReviewList(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	limit := 30 // 默认每次复习30个
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	vocabs, err := getVocabularyService().GetTodayReviewList(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  vocabs,
		"count": len(vocabs),
	})
}

// SubmitReview 提交复习结果
// POST /api/v1/vocabulary/:id/review
func SubmitReview(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req service.SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}
	req.VocabularyID = uint(id)

	result, err := getVocabularyService().SubmitReview(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "复习完成",
		"data":          result.Vocabulary,
		"points_earned": result.PointsEarned,
		"total_points":  result.TotalPoints,
	})
}

// ==================== 文件夹管理 ====================

// CreateFolder 创建文件夹
// POST /api/v1/vocabulary/folders
func CreateVocabularyFolder(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req service.CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	folder, err := getVocabularyService().CreateFolder(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"data":    folder,
	})
}

// ListFolders 获取文件夹列表
// GET /api/v1/vocabulary/folders
func ListVocabularyFolders(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	folders, err := getVocabularyService().ListFolders(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": folders})
}

// UpdateFolder 更新文件夹
// PUT /api/v1/vocabulary/folders/:id
func UpdateVocabularyFolder(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req service.CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	folder, err := getVocabularyService().UpdateFolder(uint(id), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    folder,
	})
}

// DeleteFolder 删除文件夹
// DELETE /api/v1/vocabulary/folders/:id
func DeleteVocabularyFolder(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := getVocabularyService().DeleteFolder(uint(id), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// AddVocabularyToFolder 添加生词到文件夹
// POST /api/v1/vocabulary/folders/:id/items
func AddVocabularyToFolder(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	folderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件夹ID"})
		return
	}

	var req struct {
		VocabularyID uint `json:"vocabulary_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := getVocabularyService().AddToFolder(uint(folderID), req.VocabularyID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "添加成功"})
}

// RemoveVocabularyFromFolder 从文件夹移除生词
// DELETE /api/v1/vocabulary/folders/:id/items/:vocab_id
func RemoveVocabularyFromFolder(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	folderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件夹ID"})
		return
	}

	vocabID, err := strconv.ParseUint(c.Param("vocab_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的生词ID"})
		return
	}

	if err := getVocabularyService().RemoveFromFolder(uint(folderID), uint(vocabID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "移除成功"})
}

// ListFolderVocabulary 获取文件夹中的生词
// GET /api/v1/vocabulary/folders/:id/items
func ListFolderVocabulary(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	folderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件夹ID"})
		return
	}

	limit := 20
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	vocabs, total, err := getVocabularyService().ListByFolder(uint(folderID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   vocabs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetDailyStats 获取每日学习统计
// GET /api/v1/vocabulary/stats/daily
func GetVocabularyDailyStats(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	days := 7 // 默认7天
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	stats, err := getVocabularyService().GetDailyStats(userID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
