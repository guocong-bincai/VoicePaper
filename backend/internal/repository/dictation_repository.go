package repository

import (
	"voicepaper/internal/model"

	"gorm.io/gorm"
)

// DictationRecordRepository 默写记录仓库
type DictationRecordRepository struct{}

func NewDictationRecordRepository() *DictationRecordRepository {
	return &DictationRecordRepository{}
}

// Create 创建默写记录
func (r *DictationRecordRepository) Create(record *model.DictationRecord) error {
	return DB.Create(record).Error
}

// FindByUserIDAndArticleID 根据用户ID和文章ID查找记录
func (r *DictationRecordRepository) FindByUserIDAndArticleID(userID *uint, articleID uint) ([]model.DictationRecord, error) {
	var records []model.DictationRecord
	query := DB.Where("article_id = ?", articleID)
	
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}
	
	err := query.Order("created_at DESC").Find(&records).Error
	return records, err
}

// FindByUserIDAndItemID 根据用户ID和题目ID（单词或句子）查找记录
func (r *DictationRecordRepository) FindByUserIDAndItemID(
	userID *uint,
	dictationType model.DictationType,
	itemID uint,
) ([]model.DictationRecord, error) {
	var records []model.DictationRecord
	query := DB.Where("dictation_type = ?", dictationType)
	
	if dictationType == model.DictationTypeWord {
		query = query.Where("word_id = ?", itemID)
	} else {
		query = query.Where("sentence_id = ?", itemID)
	}
	
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}
	
	err := query.Order("created_at DESC").Find(&records).Error
	return records, err
}

// UpdateOrCreate 更新或创建记录（如果已存在相同用户、题目、类型的记录，则更新；否则创建）
func (r *DictationRecordRepository) UpdateOrCreate(record *model.DictationRecord) error {
	var existing model.DictationRecord
	query := DB.Where("article_id = ? AND dictation_type = ?", record.ArticleID, record.DictationType)
	
	if record.DictationType == model.DictationTypeWord {
		query = query.Where("word_id = ?", record.WordID)
	} else {
		query = query.Where("sentence_id = ?", record.SentenceID)
	}
	
	if record.UserID != nil {
		query = query.Where("user_id = ?", *record.UserID)
	} else {
		query = query.Where("user_id IS NULL")
	}
	
	err := query.First(&existing).Error
	
	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		return DB.Create(record).Error
	} else if err != nil {
		return err
	}
	
	// 存在，更新记录
	existing.UserAnswer = record.UserAnswer
	existing.IsCorrect = record.IsCorrect
	existing.Score = record.Score
	existing.AttemptCount = record.AttemptCount
	existing.TimeSpent = record.TimeSpent
	existing.LastAttempt = record.LastAttempt
	
	return DB.Save(&existing).Error
}

// GetStatistics 获取用户的默写统计
func (r *DictationRecordRepository) GetStatistics(userID *uint, articleID uint) (map[string]interface{}, error) {
	var totalCount int64
	var correctCount int64
	var totalScore int64
	
	query := DB.Model(&model.DictationRecord{}).Where("article_id = ?", articleID)
	
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("user_id IS NULL")
	}
	
	// 总数
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}
	
	// 正确数
	if err := query.Where("is_correct = ?", true).Count(&correctCount).Error; err != nil {
		return nil, err
	}
	
	// 平均分
	if err := query.Select("COALESCE(AVG(score), 0)").Scan(&totalScore).Error; err != nil {
		return nil, err
	}
	
	accuracy := 0.0
	if totalCount > 0 {
		accuracy = float64(correctCount) / float64(totalCount) * 100
	}
	
	return map[string]interface{}{
		"total_count":  totalCount,
		"correct_count": correctCount,
		"accuracy":     accuracy,
		"avg_score":    totalScore,
	}, nil
}

