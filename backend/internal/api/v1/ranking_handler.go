package v1

import (
	"context"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"voicepaper/config"
	"voicepaper/internal/model"
	"voicepaper/internal/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RankingHandler struct {
	db      *gorm.DB
	storage storage.Storage
}

func NewRankingHandler(db *gorm.DB) *RankingHandler {
	cfg := config.GetConfig()
	var st storage.Storage
	var err error

	if cfg.Storage.Type == "oss" {
		st, err = storage.NewOSSStorage(cfg)
		if err != nil {
			st = storage.NewLocalStorage(cfg)
		}
	} else {
		st = storage.NewLocalStorage(cfg)
	}

	return &RankingHandler{
		db:      db,
		storage: st,
	}
}

// UserRankingResponse 用户排名响应结构
type UserRankingResponse struct {
	Rank          int64            `json:"rank"`
	TotalUsers    int64            `json:"total_users"`
	Score         float64          `json:"score"`
	TotalDuration int64            `json:"total_duration"`
	TotalPoints   int64            `json:"total_points"`
	NearbyUsers   []NearbyUserItem `json:"nearby_users"`
}

// NearbyUserItem 周围用户信息
type NearbyUserItem struct {
	Rank          int64   `json:"rank"`
	Nickname      string  `json:"nickname"`
	Avatar        string  `json:"avatar"`
	Score         float64 `json:"score"`
	IsCurrentUser bool    `json:"is_current_user,omitempty"`
}

// GetMyRanking 获取我的排名信息
// GET /api/v1/ranking/me
func (h *RankingHandler) GetMyRanking(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	ranking, err := h.getUserRanking(uid)
	if err != nil {
		log.Printf("❌ 获取排名失败 (user_id=%d): %v", uid, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排名信息失败", "details": err.Error()})
		return
	}

	log.Printf("✅ 获取排名成功 (user_id=%d, rank=%d)", uid, ranking.Rank)
	c.JSON(http.StatusOK, ranking)
}

// GetGlobalRanking 获取全球排名列表
// GET /api/v1/ranking/global?page=1&page_size=20&sort_by=score
func (h *RankingHandler) GetGlobalRanking(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	sortBy := c.DefaultQuery("sort_by", "score") // score, duration, points

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 构建排序字段 - 使用表别名避免歧义
	var orderBy string
	switch sortBy {
	case "duration":
		orderBy = "COALESCE(up.total_duration_minutes, 0) DESC"
	case "points":
		orderBy = "up.total_points DESC"
	default:
		orderBy = "(COALESCE(up.total_duration_minutes, 0) * 0.5 + up.total_points * 0.5) DESC"
	}

	// 获取总用户数
	var totalUsers int64
	if err := h.db.Model(&model.UserPoints{}).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排名信息失败"})
		return
	}

	// 获取排名列表
	var results []struct {
		UserID               uint
		Nickname             string
		Avatar               string
		TotalDurationMinutes int64
		TotalPoints          int64
		Score                float64
	}

	query := `
	SELECT
		up.user_id,
		COALESCE(u.nickname, '') as nickname,
		COALESCE(u.avatar, '') as avatar,
		COALESCE(up.total_duration_minutes, 0) as total_duration_minutes,
		COALESCE(up.total_points, 0) as total_points,
		(COALESCE(up.total_duration_minutes, 0) * 0.5 + COALESCE(up.total_points, 0) * 0.5) as score
	FROM vp_user_points up
	LEFT JOIN vp_users u ON up.user_id = u.id AND u.deleted_at IS NULL
	WHERE up.deleted_at IS NULL
	ORDER BY ` + orderBy + `
	LIMIT ? OFFSET ?
	`

	if err := h.db.Raw(query, pageSize, offset).Scan(&results).Error; err != nil {
		log.Printf("❌ 获取全球排名失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排名信息失败", "details": err.Error()})
		return
	}

	// 构建响应（排名 = offset + index + 1）
	items := make([]gin.H, len(results))
	for i, result := range results {
		rank := int64(offset + i + 1)

		// 处理头像签名
		avatarURL := result.Avatar
		if avatarURL != "" && h.storage != nil && strings.Contains(avatarURL, "oss-cn-") {
			ctx := context.Background()
			parts := strings.Split(avatarURL, ".aliyuncs.com/")
			if len(parts) > 1 {
				path := parts[1]
				signedURL, err := h.storage.GetSignedURL(ctx, path, 3153600000) // 100 years expiration
				if err == nil {
					avatarURL = signedURL
				}
			}
		}

		items[i] = gin.H{
			"rank":           rank,
			"user_id":        result.UserID,
			"nickname":       result.Nickname,
			"avatar":         avatarURL,
			"total_duration": result.TotalDurationMinutes,
			"total_points":   result.TotalPoints,
			"score":          result.Score,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"total":       totalUsers,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": int64(math.Ceil(float64(totalUsers) / float64(pageSize))),
	})
}

// getUserRanking 获取用户排名信息
func (h *RankingHandler) getUserRanking(userID uint) (*UserRankingResponse, error) {
	// 获取用户的积分信息
	var userPoints model.UserPoints
	if err := h.db.Where("user_id = ?", userID).First(&userPoints).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果用户没有积分记录，创建一个
			userPoints = model.UserPoints{
				UserID:               userID,
				TotalPoints:          0,
				CurrentPoints:        0,
				Level:                1,
				LevelName:            "初学者",
				TotalDurationMinutes: 0,
			}
			if err := h.db.Create(&userPoints).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// 获取用户信息
	var user model.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	// 获取总用户数
	var totalUsers int64
	if err := h.db.Model(&model.UserPoints{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	// 计算用户排名 - 需要用和显示分数一致的归一化计算方法
	// 因为显示分数用的是 (duration/3000 * 0.5 + points/10000 * 0.5) * 10000
	var userRank int64 = 1
	var rankResult struct {
		UserRank int64 `gorm:"column:user_rank"`
	}

	// 先获取当前用户的分数（用于比较）
	userScore := h.calculateScore(userPoints.TotalDurationMinutes, int64(userPoints.TotalPoints))

	log.Printf("📊 用户 %d 的计算数据: duration=%d, points=%d, score=%.1f", userID, userPoints.TotalDurationMinutes, userPoints.TotalPoints, userScore)

	// BUG修复: 修复排名计算逻辑，使用ROW_NUMBER窗口函数确保排名计算准确
	// 修复策略: 改用窗口函数计算排名，避免浮点数精度问题，确保与nearby_users列表中的排名一致
	// 注意: rank是MySQL保留字，必须使用user_rank作为别名
	// 影响范围: backend/internal/api/v1/ranking_handler.go:194-238
	// 修复日期: 2025-01-27
	// 使用窗口函数计算排名，确保准确性（MySQL 8.0+支持）
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
	if err := h.db.Raw(rankQuery, userID).Scan(&rankResult).Error; err != nil {
		log.Printf("⚠️  计算排名失败，使用默认排名: %v", err)
		// 如果窗口函数失败，回退到原来的COUNT方法
		fallbackQuery := `
		SELECT COUNT(*) + 1 as user_rank
		FROM vp_user_points
		WHERE deleted_at IS NULL
		AND user_id != ?
		AND (
			((LEAST(COALESCE(total_duration_minutes, 0) / 3000.0, 1.0) * 0.5) +
			 (LEAST(COALESCE(total_points, 0) / 10000.0, 1.0) * 0.5)) * 10000 > ?
		)
		`
		if err2 := h.db.Raw(fallbackQuery, userID, userScore).Scan(&rankResult).Error; err2 != nil {
			log.Printf("⚠️  回退排名计算也失败: %v", err2)
			userRank = 1
		} else {
			userRank = rankResult.UserRank
			log.Printf("✅ 使用回退方法计算排名成功: user_id=%d, rank=%d, score=%.1f", userID, userRank, userScore)
		}
	} else {
		userRank = rankResult.UserRank
		log.Printf("✅ 使用窗口函数计算排名成功: user_id=%d, rank=%d, score=%.1f", userID, userRank, userScore)
	}

	// 计算综合分数
	score := h.calculateScore(userPoints.TotalDurationMinutes, int64(userPoints.TotalPoints))

	// 获取周围排名用户（前1名、用户自己、后1名）
	nearbyUsers := h.getNearbyUsers(userID, user.Nickname, score)

	return &UserRankingResponse{
		Rank:          userRank,
		TotalUsers:    totalUsers,
		Score:         score,
		TotalDuration: userPoints.TotalDurationMinutes,
		TotalPoints:   int64(userPoints.TotalPoints),
		NearbyUsers:   nearbyUsers,
	}, nil
}

// getNearbyUsers 获取周围排名用户
func (h *RankingHandler) getNearbyUsers(userID uint, nickname string, userScore float64) []NearbyUserItem {
	var results []struct {
		UserID               uint
		Nickname             string
		Avatar               string
		TotalDurationMinutes int64
		TotalPoints          int64
	}

	// BUG修复: 修复排名排序逻辑，使用与calculateScore一致的归一化计算方法
	// 修复策略: 使用归一化后的分数进行排序，确保与排名计算逻辑一致
	// 影响范围: backend/internal/api/v1/ranking_handler.go:236-247
	// 修复日期: 2025-01-27
	query := `
	SELECT
		up.user_id,
		COALESCE(u.nickname, '用户') as nickname,
		COALESCE(u.avatar, '') as avatar,
		COALESCE(up.total_duration_minutes, 0) as total_duration_minutes,
		COALESCE(up.total_points, 0) as total_points
	FROM vp_user_points up
	LEFT JOIN vp_users u ON up.user_id = u.id AND u.deleted_at IS NULL
	WHERE up.deleted_at IS NULL
	ORDER BY (
		((LEAST(COALESCE(up.total_duration_minutes, 0) / 3000.0, 1.0) * 0.5) +
		 (LEAST(COALESCE(up.total_points, 0) / 10000.0, 1.0) * 0.5)) * 10000
	) DESC
	LIMIT 5
	`

	if err := h.db.Raw(query).Scan(&results).Error; err != nil {
		log.Printf("⚠️  获取周围排名用户失败: %v", err)
	}

	nearbyUsers := make([]NearbyUserItem, 0)
	for i, result := range results {
		score := h.calculateScore(result.TotalDurationMinutes, result.TotalPoints)

		// 处理头像签名
		avatarURL := result.Avatar
		if avatarURL != "" && h.storage != nil && strings.Contains(avatarURL, "oss-cn-") {
			ctx := context.Background()
			parts := strings.Split(avatarURL, ".aliyuncs.com/")
			if len(parts) > 1 {
				path := parts[1]
				signedURL, err := h.storage.GetSignedURL(ctx, path, 3153600000) // 100 years expiration
				if err == nil {
					avatarURL = signedURL
				}
			}
		}

		item := NearbyUserItem{
			Rank:     int64(i + 1), // 前5名的排名就是 1-5
			Nickname: result.Nickname,
			Avatar:   avatarURL,
			Score:    score,
		}

		if result.UserID == userID {
			item.IsCurrentUser = true
		}

		nearbyUsers = append(nearbyUsers, item)
	}

	return nearbyUsers
}

// calculateScore 计算综合分数 = 累积时长 * 0.5 + 累计积分 * 0.5
func (h *RankingHandler) calculateScore(duration int64, points int64) float64 {
	// 为了数值平衡，需要进行归一化
	// 假设：3000分钟为满分（约50小时），10000积分为满分
	maxDuration := 3000.0
	maxPoints := 10000.0

	normalizedDuration := float64(duration) / maxDuration
	if normalizedDuration > 1 {
		normalizedDuration = 1
	}

	normalizedPoints := float64(points) / maxPoints
	if normalizedPoints > 1 {
		normalizedPoints = 1
	}

	// 综合分数 = (时长比重 * 0.5 + 积分比重 * 0.5) * 10000
	score := (normalizedDuration*0.5 + normalizedPoints*0.5) * 10000
	return math.Round(score*100) / 100 // 保留两位小数
}
