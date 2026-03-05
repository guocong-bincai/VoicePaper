package service

import (
	"fmt"
	"time"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"

	"gorm.io/gorm"
)

// PointService 积分服务
type PointService struct {
	db              *gorm.DB
	userPointsRepo  *repository.UserPointsRepository
	pointRecordRepo *repository.PointRecordRepository
	titleRepo       *repository.TitleRepository
}

func NewPointService(db *gorm.DB) *PointService {
	return &PointService{
		db:              db,
		userPointsRepo:  repository.NewUserPointsRepository(db),
		pointRecordRepo: repository.NewPointRecordRepository(db),
		titleRepo:       repository.NewTitleRepository(db),
	}
}

// GetUserPoints 获取用户积分信息（如果不存在则创建）
func (s *PointService) GetUserPoints(userID uint) (*model.UserPoints, error) {
	userPoints, err := s.userPointsRepo.GetByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果不存在，创建初始积分记录
			userPoints = &model.UserPoints{
				UserID:        userID,
				TotalPoints:   0,
				CurrentPoints: 0,
				Level:         1,
				LevelName:     "初学者",
			}
			err = s.userPointsRepo.Create(userPoints)
			if err != nil {
				return nil, err
			}

			// 自动颁发"初学者"称号
			go s.AwardBeginnerTitle(userID)

			return userPoints, nil
		}
		return nil, err
	}
	return userPoints, nil
}

// AwardPoints 奖励积分（核心方法）
func (s *PointService) AwardPoints(
	userID uint,
	points int,
	pointType model.PointType,
	description string,
	articleID *uint,
	dictationRecordID *uint,
) (*model.UserPoints, *model.PointRecord, error) {
	var userPoints *model.UserPoints
	var pointRecord *model.PointRecord

	// 使用事务确保数据一致性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 获取并锁定用户积分记录
		up, err := s.userPointsRepo.AddPoints(tx, userID, points)
		if err != nil {
			return fmt.Errorf("添加积分失败: %w", err)
		}
		userPoints = up

		// 2. 创建积分记录
		pointRecord = &model.PointRecord{
			UserID:            userID,
			Points:            points,
			Type:              pointType,
			Description:       description,
			ArticleID:         articleID,
			DictationRecordID: dictationRecordID,
			BalanceBefore:     userPoints.CurrentPoints - points,
			BalanceAfter:      userPoints.CurrentPoints,
		}
		err = s.pointRecordRepo.Create(tx, pointRecord)
		if err != nil {
			return fmt.Errorf("创建积分记录失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// 异步检查是否达成新称号
	go s.CheckAndAwardTitles(userID)

	return userPoints, pointRecord, nil
}

// AwardReadArticlePoints 奖励阅读文章积分
func (s *PointService) AwardReadArticlePoints(userID, articleID uint, articleTitle string) (*model.UserPoints, *model.PointRecord, error) {
	// 检查是否已经获得过该文章的积分
	alreadyRead, err := s.pointRecordRepo.CheckArticleRead(userID, articleID)
	if err != nil {
		return nil, nil, err
	}
	if alreadyRead {
		return nil, nil, fmt.Errorf("该文章已阅读过，不重复奖励积分")
	}

	points := model.PointRewardConfig[model.PointTypeReadArticle]
	description := fmt.Sprintf("阅读文章《%s》", articleTitle)

	userPoints, record, err := s.AwardPoints(userID, points, model.PointTypeReadArticle, description, &articleID, nil)
	if err != nil {
		return nil, nil, err
	}

	// 更新阅读统计
	_ = s.userPointsRepo.IncrementArticlesRead(userID)

	return userPoints, record, nil
}

// AwardDictationPoints 奖励默写积分
func (s *PointService) AwardDictationPoints(
	userID uint,
	dictationType model.PointType, // PointTypeWordDictation 或 PointTypeSentenceDictation
	isCorrect bool,
	itemText string,
	dictationRecordID *uint,
) (*model.UserPoints, *model.PointRecord, error) {
	if !isCorrect {
		return nil, nil, fmt.Errorf("答案错误，不奖励积分")
	}

	points := model.PointRewardConfig[dictationType]
	var description string
	if dictationType == model.PointTypeWordDictation {
		description = fmt.Sprintf("单词默写正确: %s", itemText)
	} else {
		description = fmt.Sprintf("句子默写正确: %s", itemText)
	}

	userPoints, record, err := s.AwardPoints(userID, points, dictationType, description, nil, dictationRecordID)
	if err != nil {
		return nil, nil, err
	}

	// 更新默写统计
	_ = s.userPointsRepo.IncrementDictationsCompleted(userID)

	return userPoints, record, nil
}

// AwardCompleteArticlePoints 奖励完成整篇文章默写
func (s *PointService) AwardCompleteArticlePoints(userID, articleID uint, articleTitle string) (*model.UserPoints, *model.PointRecord, error) {
	points := model.PointRewardConfig[model.PointTypeCompleteArticle]
	description := fmt.Sprintf("完成文章《%s》全部默写", articleTitle)

	return s.AwardPoints(userID, points, model.PointTypeCompleteArticle, description, &articleID, nil)
}

// GetPointRecords 获取积分记录
func (s *PointService) GetPointRecords(userID uint, page, pageSize int) ([]model.PointRecord, int64, error) {
	offset := (page - 1) * pageSize
	records, err := s.pointRecordRepo.GetByUserID(userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.pointRecordRepo.CountByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetPointStatistics 获取积分统计
func (s *PointService) GetPointStatistics(userID uint) (map[string]interface{}, error) {
	userPoints, err := s.GetUserPoints(userID)
	if err != nil {
		return nil, err
	}

	// 获取各类型积分总数
	readPoints, _ := s.pointRecordRepo.GetTotalPointsByType(userID, model.PointTypeReadArticle)
	wordPoints, _ := s.pointRecordRepo.GetTotalPointsByType(userID, model.PointTypeWordDictation)
	sentencePoints, _ := s.pointRecordRepo.GetTotalPointsByType(userID, model.PointTypeSentenceDictation)
	checkInPoints, _ := s.pointRecordRepo.GetTotalPointsByType(userID, model.PointTypeDailyCheckIn)

	return map[string]interface{}{
		"total_points":               userPoints.TotalPoints,
		"current_points":             userPoints.CurrentPoints,
		"level":                      userPoints.Level,
		"level_name":                 userPoints.LevelName,
		"total_articles_read":        userPoints.TotalArticlesRead,
		"total_dictations_completed": userPoints.TotalDictationsCompleted,
		"total_check_ins":            userPoints.TotalCheckIns,
		"continuous_check_ins":       userPoints.ContinuousCheckIns,
		"max_continuous_check_ins":   userPoints.MaxContinuousCheckIns,
		"points_by_type": map[string]int{
			"read_article":       readPoints,
			"word_dictation":     wordPoints,
			"sentence_dictation": sentencePoints,
			"check_in":           checkInPoints,
		},
	}, nil
}

// AwardBeginnerTitle 自动颁发"初学者"称号
func (s *PointService) AwardBeginnerTitle(userID uint) error {
	titleConfig, err := s.titleRepo.GetTitleConfigByKey("beginner")
	if err != nil {
		return err
	}

	// 检查是否已拥有
	hasTitle, _ := s.titleRepo.CheckUserHasTitle(userID, titleConfig.ID)
	if hasTitle {
		return nil
	}

	// 颁发称号
	return s.db.Transaction(func(tx *gorm.DB) error {
		userTitle := &model.UserTitle{
			UserID:        userID,
			TitleConfigID: titleConfig.ID,
			AwardedAt:     time.Now(),
			IsEquipped:    true, // 默认佩戴
		}
		return s.titleRepo.CreateUserTitle(tx, userTitle)
	})
}

// CheckAndAwardTitles 检查并自动颁发称号
func (s *PointService) CheckAndAwardTitles(userID uint) error {
	userPoints, err := s.GetUserPoints(userID)
	if err != nil {
		return err
	}

	// 获取所有启用的称号配置
	allTitles, err := s.titleRepo.GetAllTitleConfigs()
	if err != nil {
		return err
	}

	// 检查每个称号的条件
	for _, titleConfig := range allTitles {
		// 跳过特殊称号
		if titleConfig.ConditionType == model.TitleConditionCustom {
			continue
		}

		// 检查是否已拥有
		hasTitle, _ := s.titleRepo.CheckUserHasTitle(userID, titleConfig.ID)
		if hasTitle {
			continue
		}

		// 检查是否满足条件
		shouldAward := false
		switch titleConfig.ConditionType {
		case model.TitleConditionArticlesRead:
			shouldAward = userPoints.TotalArticlesRead >= titleConfig.ConditionValue
		case model.TitleConditionDictationsCompleted:
			shouldAward = userPoints.TotalDictationsCompleted >= titleConfig.ConditionValue
		case model.TitleConditionTotalPoints:
			shouldAward = userPoints.TotalPoints >= titleConfig.ConditionValue
		case model.TitleConditionContinuousCheckIns:
			shouldAward = userPoints.ContinuousCheckIns >= titleConfig.ConditionValue
		case model.TitleConditionVocabularyCount:
			// 获取生词本收藏数
			var vocabCount int64
			s.db.Model(&model.Vocabulary{}).Where("user_id = ?", userID).Count(&vocabCount)
			shouldAward = int(vocabCount) >= titleConfig.ConditionValue
		case model.TitleConditionReviewCount:
			// 获取复习次数
			var reviewCount int64
			s.db.Model(&model.VocabularyReview{}).Where("user_id = ?", userID).Count(&reviewCount)
			shouldAward = int(reviewCount) >= titleConfig.ConditionValue
		case model.TitleConditionTotalDuration:
			// 获取学习时长(分钟)
			shouldAward = int(userPoints.TotalDurationMinutes) >= titleConfig.ConditionValue
		}

		// 颁发称号
		if shouldAward {
			_ = s.db.Transaction(func(tx *gorm.DB) error {
				userTitle := &model.UserTitle{
					UserID:        userID,
					TitleConfigID: titleConfig.ID,
					AwardedAt:     time.Now(),
					IsEquipped:    false,
				}
				return s.titleRepo.CreateUserTitle(tx, userTitle)
			})
		}
	}

	return nil
}

// GetEquippedTitle 获取用户佩戴的称号
func (s *PointService) GetEquippedTitle(userID uint) (*model.UserTitle, error) {
	return s.titleRepo.GetEquippedTitle(userID)
}

// AwardWordbookStudyPoints 奖励单词书学习积分
// quality: "forget"(不认识+1), "fuzzy"(模糊+2), "know"(认识+3)
func (s *PointService) AwardWordbookStudyPoints(
	userID uint,
	wordID uint,
	wordText string,
	quality string,
) (*model.UserPoints, *model.PointRecord, int, error) {
	// 根据质量获取积分
	points, exists := model.WordbookStudyPointsConfig[quality]
	if !exists {
		return nil, nil, 0, fmt.Errorf("无效的学习质量: %s", quality)
	}

	// 生成描述
	var qualityText string
	switch quality {
	case "forget":
		qualityText = "不认识"
	case "fuzzy":
		qualityText = "模糊"
	case "know":
		qualityText = "认识"
	default:
		qualityText = quality
	}

	description := fmt.Sprintf("单词书学习[%s]: %s", qualityText, wordText)

	userPoints, record, err := s.AwardPoints(userID, points, model.PointTypeWordbookStudy, description, nil, nil)
	if err != nil {
		return nil, nil, 0, err
	}

	return userPoints, record, points, nil
}
