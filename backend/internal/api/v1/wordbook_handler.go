package v1

import (
	"net/http"
	"strconv"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"
	"voicepaper/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WordbookHandler struct {
	repo         *repository.WordbookRepository
	pointService *service.PointService
}

func NewWordbookHandler(db *gorm.DB) *WordbookHandler {
	return &WordbookHandler{
		repo:         repository.NewWordbookRepository(db),
		pointService: service.NewPointService(db),
	}
}

// GetWordbooks 获取所有可用的单词书列表
func (h *WordbookHandler) GetWordbooks(c *gin.Context) {
	wordbooks, err := h.repo.GetWordbookList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取单词书列表失败"})
		return
	}

	// 调试日志
	for _, wb := range wordbooks {
		if wb.Type == "cet4" {
			// 使用 fmt.Printf 或 log.Printf
			println("🔍 Backend Debug CET4:", wb.Name, "CoverURL:", wb.CoverURL)
		}
	}

	c.JSON(http.StatusOK, wordbooks)
}

// DebugWordbookBooks 调试 API - 检查 vp_wordbook_books 表中的数据
// GET /api/v1/wordbooks/debug/books
func (h *WordbookHandler) DebugWordbookBooks(c *gin.Context) {
	bookType := c.Query("type")

	// 获取该类型的单词ID列表
	ids, err := h.repo.GetAllWordIDs(bookType)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"book_type": bookType,
			"error":     err.Error(),
			"count":     0,
			"ids":       []uint{},
		})
		return
	}

	// 检查这些ID在vp_wordbook表中是否存在
	var existingCount int64
	if len(ids) > 0 {
		h.repo.DB().Model(&model.Wordbook{}).Where("id IN ?", ids).Count(&existingCount)
	}

	// 获取vp_wordbook表的ID范围
	var minID, maxID uint
	h.repo.DB().Model(&model.Wordbook{}).Select("MIN(id)").Scan(&minID)
	h.repo.DB().Model(&model.Wordbook{}).Select("MAX(id)").Scan(&maxID)

	var totalWords int64
	h.repo.DB().Model(&model.Wordbook{}).Count(&totalWords)

	// 检查前20个ID是否存在
	var firstIDs []uint
	if len(ids) > 0 {
		firstCount := 20
		if len(ids) < 20 {
			firstCount = len(ids)
		}
		firstIDs = ids[:firstCount]
	}

	// 查询这些ID哪些不存在
	var existingIDs []uint
	h.repo.DB().Model(&model.Wordbook{}).Where("id IN ?", firstIDs).Pluck("id", &existingIDs)

	existingMap := make(map[uint]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	var missingIDs []uint
	for _, id := range firstIDs {
		if !existingMap[id] {
			missingIDs = append(missingIDs, id)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"book_type":            bookType,
		"count":                len(ids),
		"ids":                  ids[:min(10, len(ids))], // 只返回前10个
		"first_20_ids":         firstIDs,
		"existing_in_wordbook": existingCount,
		"missing_in_first_20":  missingIDs,
		"wordbook_id_range":    gin.H{"min": minID, "max": maxID},
		"wordbook_total":       totalWords,
	})
}

// GetWords 获取单词书中的单词
func (h *WordbookHandler) GetWords(c *gin.Context) {
	wordType := c.Param("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	words, total, err := h.repo.GetWordsByType(wordType, offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取单词失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"words": words,
		"total": total,
		"page":  page,
	})
}

// GetProgress 获取学习进度
func (h *WordbookHandler) GetProgress(c *gin.Context) {
	userID, _ := c.Get("user_id")
	wordType := c.Param("type")

	progress, err := h.repo.GetProgress(userID.(uint), wordType)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{"current_index": 0})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取进度失败"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// SaveProgress 保存学习进度
func (h *WordbookHandler) SaveProgress(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req struct {
		WordType     string `json:"word_type" binding:"required"`
		CurrentIndex int    `json:"current_index"`
		LastWordID   uint   `json:"last_word_id"`
		TotalWords   int    `json:"total_words"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	if err := h.repo.SaveProgress(&model.WordbookProgress{
		UserID:       userID.(uint),
		WordType:     req.WordType,
		CurrentIndex: req.CurrentIndex,
		LastWordID:   req.LastWordID,
		TotalWords:   req.TotalWords,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存进度失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "进度已保存"})
}

// ImportWordToVocabulary 将单词书中的单词导入到生词本
// POST /api/v1/wordbooks/:wordbook_id/import
func (h *WordbookHandler) ImportWordToVocabulary(c *gin.Context) {
	userID, _ := c.Get("user_id")
	wordbookID, err := strconv.ParseUint(c.Param("wordbook_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的单词书ID"})
		return
	}

	var req struct {
		Quality string `json:"quality" binding:"required,oneof=fuzzy forget"` // 模糊 或 不认识
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 调用 service 层方法导入单词
	vocab, err := h.repo.ImportWordToVocabulary(userID.(uint), uint(wordbookID), req.Quality)
	if err != nil {
		// 如果是 record not found，返回一个更有用的错误信息
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "单词不存在", "wordbook_id": wordbookID})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "wordbook_id": wordbookID})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "单词已添加到生词本",
		"data":    vocab,
	})
}

// ==================== 顺序/乱序学习相关接口 ====================

// GetUserOrder 获取用户的学习序列状态
// GET /api/v1/wordbooks/:type/order
func (h *WordbookHandler) GetUserOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	wordType := c.Param("type")

	order, err := h.repo.GetUserOrder(userID.(uint), wordType)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有序列，返回默认状态
			c.JSON(http.StatusOK, gin.H{
				"is_random":     false,
				"current_index": 0,
				"total_words":   0,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学习序列失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_random":     order.IsRandom,
		"current_index": order.CurrentIndex,
		"total_words":   order.TotalWords,
	})
}

// SwitchOrderMode 切换顺序/乱序模式
// POST /api/v1/wordbooks/:type/order
func (h *WordbookHandler) SwitchOrderMode(c *gin.Context) {
	userID, _ := c.Get("user_id")
	wordType := c.Param("type")

	var req struct {
		IsRandom bool `json:"is_random"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	order, err := h.repo.SwitchOrderMode(userID.(uint), wordType, req.IsRandom)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "单词书不存在或没有单词"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "切换模式失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "切换成功",
		"is_random":     order.IsRandom,
		"current_index": order.CurrentIndex,
		"total_words":   order.TotalWords,
	})
}

// GetWordsWithOrder 按用户序列获取单词（支持顺序/乱序）
// GET /api/v1/wordbooks/:type/words/ordered
func (h *WordbookHandler) GetWordsWithOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	wordType := c.Param("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}

	words, total, currentIndex, isRandom, err := h.repo.GetWordsByUserOrder(userID.(uint), wordType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取单词失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"words":         words,
		"total":         total,
		"page":          page,
		"current_index": currentIndex,
		"is_random":     isRandom,
	})
}

// UpdateOrderIndex 更新学习进度位置
// PUT /api/v1/wordbooks/:type/order/index
func (h *WordbookHandler) UpdateOrderIndex(c *gin.Context) {
	userID, _ := c.Get("user_id")
	wordType := c.Param("type")

	var req struct {
		CurrentIndex int `json:"current_index"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := h.repo.UpdateUserOrderIndex(userID.(uint), wordType, req.CurrentIndex); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新进度失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "进度已更新", "current_index": req.CurrentIndex})
}

// StudyWord 记录单词学习并奖励积分
// POST /api/v1/wordbooks/:wordbook_id/study
func (h *WordbookHandler) StudyWord(c *gin.Context) {
	userID, _ := c.Get("user_id")
	wordbookID, err := strconv.ParseUint(c.Param("wordbook_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的单词 ID"})
		return
	}

	var req struct {
		Quality  string `json:"quality" binding:"required,oneof=forget fuzzy know"` // 不认识/模糊/认识
		WordText string `json:"word_text"`                                          // 单词文本（用于积分记录描述）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	// 如果是模糊或不认识，导入到生词本
	if req.Quality == "fuzzy" || req.Quality == "forget" {
		_, _ = h.repo.ImportWordToVocabulary(userID.(uint), uint(wordbookID), req.Quality)
	}

	// 奖励积分
	userPoints, _, pointsEarned, err := h.pointService.AwardWordbookStudyPoints(
		userID.(uint),
		uint(wordbookID),
		req.WordText,
		req.Quality,
	)
	if err != nil {
		// 积分奖励失败不影响主流程，只记录日志
		println("奖励积分失败:", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"message":       "学习已记录",
			"points_earned": 0,
			"total_points":  0,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "学习已记录",
		"points_earned": pointsEarned,
		"total_points":  userPoints.CurrentPoints,
	})
}
