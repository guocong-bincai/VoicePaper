package service

import (
	"fmt"
	"time"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"

	"gorm.io/gorm"
)

// TitleService 称号服务
type TitleService struct {
	db             *gorm.DB
	titleRepo      *repository.TitleRepository
	userPointsRepo *repository.UserPointsRepository
	vocabularyRepo *repository.VocabularyRepository
}

func NewTitleService(db *gorm.DB) *TitleService {
	return &TitleService{
		db:             db,
		titleRepo:      repository.NewTitleRepository(db),
		userPointsRepo: repository.NewUserPointsRepository(db),
		vocabularyRepo: repository.NewVocabularyRepository(),
	}
}

// GetAllTitles 获取所有称号配置
func (s *TitleService) GetAllTitles() ([]model.TitleConfig, error) {
	return s.titleRepo.GetAllTitleConfigs()
}

// GetUserTitles 获取用户已获得的称号
func (s *TitleService) GetUserTitles(userID uint) ([]model.UserTitle, error) {
	return s.titleRepo.GetUserTitles(userID)
}

// GetTitleProgress 获取用户的称号进度
func (s *TitleService) GetTitleProgress(userID uint) ([]map[string]interface{}, error) {
	// 1. 获取用户积分信息
	userPoints, err := s.userPointsRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 2. 获取所有启用的称号配置
	allTitles, err := s.titleRepo.GetAllTitleConfigs()
	if err != nil {
		return nil, err
	}

	// 3. 获取用户已拥有的称号
	userTitles, err := s.titleRepo.GetUserTitles(userID)
	if err != nil {
		return nil, err
	}

	// 构建已拥有称号的map 和 穿戴状态
	ownedTitles := make(map[uint]bool)
	equippedTitles := make(map[uint]bool)
	for _, ut := range userTitles {
		ownedTitles[ut.TitleConfigID] = true
		if ut.IsEquipped {
			equippedTitles[ut.TitleConfigID] = true
		}
	}

	// 自动为老用户颁发"新手上路"称号（如果还没有的话）
	s.ensureBeginnerTitle(userID, ownedTitles, allTitles)

	// 4. 计算每个称号的进度
	var progressList []map[string]interface{}
	for _, title := range allTitles {
		owned := ownedTitles[title.ID]

		var currentValue int
		var progress float64

		// 根据条件类型获取当前进度
		switch title.ConditionType {
		case model.TitleConditionArticlesRead:
			currentValue = userPoints.TotalArticlesRead
		case model.TitleConditionDictationsCompleted:
			currentValue = userPoints.TotalDictationsCompleted
		case model.TitleConditionTotalPoints:
			currentValue = userPoints.TotalPoints
		case model.TitleConditionContinuousCheckIns:
			currentValue = userPoints.ContinuousCheckIns
		case model.TitleConditionVocabularyCount:
			// 获取生词本收藏数
			if stats, err := s.vocabularyRepo.GetStats(userID); err == nil {
				currentValue = int(stats.Total)
			}
		case model.TitleConditionReviewCount:
			// 获取复习次数 - 从生词本统计复习总次数
			var reviewCount int64
			s.db.Model(&model.VocabularyReview{}).Where("user_id = ?", userID).Count(&reviewCount)
			currentValue = int(reviewCount)
		case model.TitleConditionTotalDuration:
			// 获取学习时长(分钟)
			currentValue = int(userPoints.TotalDurationMinutes)
		case model.TitleConditionPerfectStreak:
			// 连续全对次数 - 需要从默写记录中获取，暂设为0
			currentValue = 0 // TODO: 从默写记录中统计
		case model.TitleConditionCustom:
			currentValue = 0 // 自定义条件需要特殊处理
		}

		// 计算进度百分比
		if title.ConditionValue > 0 {
			progress = float64(currentValue) / float64(title.ConditionValue) * 100
			if progress > 100 {
				progress = 100
			}
		}

		// 自动解锁：如果满足条件但还未拥有，自动授予称号
		// 注意：beginner称号通过 ensureBeginnerTitle 单独处理
		if !owned && title.ConditionType != model.TitleConditionCustom && currentValue >= title.ConditionValue {
			userTitle := &model.UserTitle{
				UserID:        userID,
				TitleConfigID: title.ID,
				AwardedAt:     time.Now(),
				IsEquipped:    false,
			}
			if err := s.titleRepo.CreateUserTitle(s.db, userTitle); err == nil {
				owned = true // 更新状态为已拥有
			}
		}

		// 检查是否是刚颁发的beginner称号
		if title.TitleKey == "beginner" && ownedTitles[title.ID] {
			owned = true
		}

		progressList = append(progressList, map[string]interface{}{
			"title_id":              title.ID,
			"title_key":             title.TitleKey,
			"title_name":            title.TitleName,
			"title_icon":            title.TitleIcon,
			"description":           title.Description,
			"category":              title.Category,
			"condition_type":        title.ConditionType,
			"condition_value":       title.ConditionValue,
			"condition_description": title.ConditionDescription,
			"current_value":         currentValue,
			"progress":              progress,
			"owned":                 owned,
			"is_equipped":           equippedTitles[title.ID],
			"rarity":                title.Rarity,
		})
	}

	return progressList, nil
}

// EquipTitle 佩戴称号
func (s *TitleService) EquipTitle(userID, titleConfigID uint) error {
	// 1. 检查用户是否拥有该称号
	hasTitle, err := s.titleRepo.CheckUserHasTitle(userID, titleConfigID)
	if err != nil {
		return err
	}
	if !hasTitle {
		return fmt.Errorf("你还没有获得这个称号")
	}

	// 2. 佩戴称号
	return s.titleRepo.EquipTitle(userID, titleConfigID)
}

// UnequipTitle 取消佩戴称号
func (s *TitleService) UnequipTitle(userID uint) error {
	return s.titleRepo.UnequipTitle(userID)
}

// GetEquippedTitle 获取当前佩戴的称号
func (s *TitleService) GetEquippedTitle(userID uint) (*model.UserTitle, error) {
	return s.titleRepo.GetEquippedTitle(userID)
}

// GetTitlesByCategory 根据分类获取称号
func (s *TitleService) GetTitlesByCategory(category model.TitleCategory) ([]model.TitleConfig, error) {
	return s.titleRepo.GetTitleConfigsByCategory(category)
}

// ensureBeginnerTitle 确保用户拥有"新手上路"称号
// 这个方法用于处理老用户没有beginner称号的情况
func (s *TitleService) ensureBeginnerTitle(userID uint, ownedTitles map[uint]bool, allTitles []model.TitleConfig) {
	// 找到beginner称号的配置
	var beginnerTitle *model.TitleConfig
	for i := range allTitles {
		if allTitles[i].TitleKey == "beginner" {
			beginnerTitle = &allTitles[i]
			break
		}
	}

	if beginnerTitle == nil {
		return // 没有配置beginner称号
	}

	// 检查用户是否已拥有
	if ownedTitles[beginnerTitle.ID] {
		return // 已拥有，无需处理
	}

	// 颁发beginner称号
	userTitle := &model.UserTitle{
		UserID:        userID,
		TitleConfigID: beginnerTitle.ID,
		AwardedAt:     time.Now(),
		IsEquipped:    false,
	}
	if err := s.titleRepo.CreateUserTitle(s.db, userTitle); err == nil {
		ownedTitles[beginnerTitle.ID] = true // 更新map，让后续逻辑知道已拥有
	}
}
