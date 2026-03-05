package v1

import (
	"log"
	"net/http"
	"voicepaper/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReportHandler struct {
	db *gorm.DB
}

func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{
		db: db,
	}
}

type YearEndReportResponse struct {
	ArticlesRead    int    `json:"articles_read"`
	ConsecutiveDays int    `json:"consecutive_days"` // Using MaxContinuousCheckIns
	WordsLearned    int64  `json:"words_learned"`
	TotalDuration   int64  `json:"total_duration"` // In minutes
	TotalPoints     int    `json:"total_points"`
	GlobalRank      int64  `json:"global_rank"`
	Nickname        string `json:"nickname"`
	Avatar          string `json:"avatar"`
}

// GetYearEndReport 获取年终总结数据
// GET /api/v1/report/year-end
func (h *ReportHandler) GetYearEndReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	uid := userID.(uint)

	// 1. Get User Points Data (Articles, Duration, Points, Checkins)
	var userPoints model.UserPoints
	if err := h.db.Where("user_id = ?", uid).First(&userPoints).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return empty/zero data if not found
			c.JSON(http.StatusOK, YearEndReportResponse{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取数据失败"})
		return
	}

	// 2. Get User Info (Nickname, Avatar)
	var user model.User
	if err := h.db.Where("id = ?", uid).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 3. Count Words Learned (from Vocabulary table)
	var wordsLearned int64
	// Counting all words in vocabulary for now.
	// If "learned" implies mastery > X, we can adjust. Defaulting to all collected words.
	if err := h.db.Model(&model.Vocabulary{}).Where("user_id = ?", uid).Count(&wordsLearned).Error; err != nil {
		log.Printf("⚠️ Failed to count vocabulary: %v", err)
	}

	// 4. Calculate Global Rank
	// Reusing logic from RankingHandler to ensure consistency
	var userRank int64 = 1
	var rankResult struct {
		UserRank int64 `gorm:"column:user_rank"`
	}

	// Using the same weighted score logic as RankingHandler:
	// Score = (NormalizedDuration * 0.5 + NormalizedPoints * 0.5) * 10000
	rankQuery := `
	SELECT rank_position as user_rank
	FROM (
		SELECT
			user_id,
			ROW_NUMBER() OVER (
				ORDER BY (
					((LEAST(COALESCE(total_duration_minutes, 0) / 3000.0, 1.0) * 0.5) +
					 (LEAST(COALESCE(total_points, 0) / 10000.0, 1.0) * 0.5)) * 10000
				) DESC
			) as rank_position
		FROM vp_user_points
		WHERE deleted_at IS NULL
	) ranked_users
	WHERE user_id = ?
	`
	if err := h.db.Raw(rankQuery, uid).Scan(&rankResult).Error; err != nil {
		log.Printf("⚠️ Failed to calculate rank: %v", err)
		// Fallback simple rank
		userRank = 0
	} else {
		userRank = rankResult.UserRank
	}

	resp := YearEndReportResponse{
		ArticlesRead:    userPoints.TotalArticlesRead,
		ConsecutiveDays: userPoints.MaxContinuousCheckIns, // Using Max as it's a "Year Summary"
		WordsLearned:    wordsLearned,
		TotalDuration:   userPoints.TotalDurationMinutes,
		TotalPoints:     userPoints.TotalPoints,
		GlobalRank:      userRank,
		Nickname:        user.Nickname,
		Avatar:          user.Avatar,
	}

	// Fix Avatar URL if it's OSS (simplified, assume frontend handles or use full URL if possible)
	// For now sending as stored.

	c.JSON(http.StatusOK, resp)
}
