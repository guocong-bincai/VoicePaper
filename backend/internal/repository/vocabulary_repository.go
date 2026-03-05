package repository

import (
	"time"
	"voicepaper/internal/model"

	"gorm.io/gorm"
)

// VocabularyRepository 生词本数据访问层
type VocabularyRepository struct {
	db *gorm.DB
}

// NewVocabularyRepository 创建生词本仓库实例
func NewVocabularyRepository() *VocabularyRepository {
	return &VocabularyRepository{db: DB}
}

// ==================== 生词 CRUD ====================

// Create 创建生词
func (r *VocabularyRepository) Create(vocab *model.Vocabulary) error {
	return r.db.Create(vocab).Error
}

// GetByID 根据ID获取生词
func (r *VocabularyRepository) GetByID(id uint) (*model.Vocabulary, error) {
	var vocab model.Vocabulary
	err := r.db.First(&vocab, id).Error
	if err != nil {
		return nil, err
	}
	return &vocab, nil
}

// GetByUserIDAndContent 检查用户是否已添加过该生词
func (r *VocabularyRepository) GetByUserIDAndContent(userID uint, content string) (*model.Vocabulary, error) {
	var vocab model.Vocabulary
	err := r.db.Where("user_id = ? AND content = ?", userID, content).First(&vocab).Error
	if err != nil {
		return nil, err
	}
	return &vocab, nil
}

// Update 更新生词
func (r *VocabularyRepository) Update(vocab *model.Vocabulary) error {
	return r.db.Save(vocab).Error
}

// Delete 软删除生词
func (r *VocabularyRepository) Delete(id uint) error {
	return r.db.Delete(&model.Vocabulary{}, id).Error
}

// ListByUserID 获取用户的生词列表
func (r *VocabularyRepository) ListByUserID(userID uint, params *VocabularyListParams) ([]model.Vocabulary, int64, error) {
	var vocabs []model.Vocabulary
	var total int64

	query := r.db.Model(&model.Vocabulary{}).Where("user_id = ?", userID)

	// 筛选条件
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.MasteryLevel != nil {
		query = query.Where("mastery_level = ?", *params.MasteryLevel)
	}
	if params.IsStarred != nil {
		query = query.Where("is_starred = ?", *params.IsStarred)
	}
	if params.Keyword != "" {
		query = query.Where("content LIKE ? OR meaning LIKE ?", "%"+params.Keyword+"%", "%"+params.Keyword+"%")
	}
	if params.ArticleID != nil {
		query = query.Where("article_id = ?", *params.ArticleID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "created_at DESC"
	if params.OrderBy != "" {
		orderBy = params.OrderBy
	}
	query = query.Order(orderBy)

	// 分页
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// 预加载文章信息
	if params.WithArticle {
		query = query.Preload("Article")
	}

	err := query.Find(&vocabs).Error

	// 确保返回空数组而不是 nil
	if len(vocabs) == 0 {
		vocabs = make([]model.Vocabulary, 0)
	}

	return vocabs, total, err
}

// VocabularyListParams 列表查询参数
type VocabularyListParams struct {
	Type         string
	MasteryLevel *int
	IsStarred    *bool
	Keyword      string
	ArticleID    *uint
	OrderBy      string
	Limit        int
	Offset       int
	WithArticle  bool
}

// GetTodayReviewList 获取今日待复习列表
func (r *VocabularyRepository) GetTodayReviewList(userID uint, limit int) ([]model.Vocabulary, error) {
	var vocabs []model.Vocabulary
	now := time.Now()

	query := r.db.Where("user_id = ? AND (next_review_at IS NULL OR next_review_at <= ?)", userID, now).
		Order("CASE WHEN next_review_at IS NULL THEN 0 ELSE 1 END, next_review_at ASC, mastery_level ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&vocabs).Error
	return vocabs, err
}

// CountTodayReview 统计今日待复习数量
func (r *VocabularyRepository) CountTodayReview(userID uint) (int64, error) {
	var count int64
	now := time.Now()
	err := r.db.Model(&model.Vocabulary{}).
		Where("user_id = ? AND (next_review_at IS NULL OR next_review_at <= ?)", userID, now).
		Count(&count).Error
	return count, err
}

// GetStats 获取用户生词统计
func (r *VocabularyRepository) GetStats(userID uint) (*VocabularyStats, error) {
	var stats VocabularyStats

	// 总数
	r.db.Model(&model.Vocabulary{}).Where("user_id = ?", userID).Count(&stats.Total)

	// 各掌握等级数量
	r.db.Model(&model.Vocabulary{}).Where("user_id = ? AND mastery_level = 0", userID).Count(&stats.New)
	r.db.Model(&model.Vocabulary{}).Where("user_id = ? AND mastery_level BETWEEN 1 AND 2", userID).Count(&stats.Learning)
	r.db.Model(&model.Vocabulary{}).Where("user_id = ? AND mastery_level BETWEEN 3 AND 4", userID).Count(&stats.Reviewing)
	r.db.Model(&model.Vocabulary{}).Where("user_id = ? AND mastery_level >= 5", userID).Count(&stats.Mastered)

	// 今日待复习
	now := time.Now()
	r.db.Model(&model.Vocabulary{}).
		Where("user_id = ? AND (next_review_at IS NULL OR next_review_at <= ?)", userID, now).
		Count(&stats.TodayReview)

	// 标星数量
	r.db.Model(&model.Vocabulary{}).Where("user_id = ? AND is_starred = ?", userID, true).Count(&stats.Starred)

	return &stats, nil
}

// VocabularyStats 生词统计
type VocabularyStats struct {
	Total       int64 `json:"total"`
	New         int64 `json:"new"`          // 未学习
	Learning    int64 `json:"learning"`     // 学习中 (1-2)
	Reviewing   int64 `json:"reviewing"`    // 复习中 (3-4)
	Mastered    int64 `json:"mastered"`     // 已掌握 (5)
	TodayReview int64 `json:"today_review"` // 今日待复习
	Starred     int64 `json:"starred"`      // 标星
}

// ToggleStar 切换标星状态
func (r *VocabularyRepository) ToggleStar(id uint, starred bool) error {
	return r.db.Model(&model.Vocabulary{}).Where("id = ?", id).Update("is_starred", starred).Error
}

// BatchDelete 批量删除
func (r *VocabularyRepository) BatchDelete(userID uint, ids []uint) error {
	return r.db.Where("user_id = ? AND id IN ?", userID, ids).Delete(&model.Vocabulary{}).Error
}

// ==================== 复习记录 ====================

// CreateReview 创建复习记录
func (r *VocabularyRepository) CreateReview(review *model.VocabularyReview) error {
	return r.db.Create(review).Error
}

// GetReviewHistory 获取复习历史
func (r *VocabularyRepository) GetReviewHistory(userID uint, vocabID uint, limit int) ([]model.VocabularyReview, error) {
	var reviews []model.VocabularyReview
	query := r.db.Where("user_id = ? AND vocabulary_id = ?", userID, vocabID).
		Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&reviews).Error
	return reviews, err
}

// ==================== 文件夹 ====================

// CreateFolder 创建文件夹
func (r *VocabularyRepository) CreateFolder(folder *model.VocabularyFolder) error {
	return r.db.Create(folder).Error
}

// ListFolders 获取用户的文件夹列表
func (r *VocabularyRepository) ListFolders(userID uint) ([]model.VocabularyFolder, error) {
	var folders []model.VocabularyFolder
	err := r.db.Where("user_id = ?", userID).Order("sort_order ASC, created_at ASC").Find(&folders).Error
	return folders, err
}

// UpdateFolder 更新文件夹
func (r *VocabularyRepository) UpdateFolder(folder *model.VocabularyFolder) error {
	return r.db.Save(folder).Error
}

// DeleteFolder 删除文件夹
func (r *VocabularyRepository) DeleteFolder(id uint) error {
	// 先删除关联
	r.db.Where("folder_id = ?", id).Delete(&model.VocabularyFolderItem{})
	return r.db.Delete(&model.VocabularyFolder{}, id).Error
}

// AddToFolder 添加生词到文件夹
func (r *VocabularyRepository) AddToFolder(folderID, vocabID uint) error {
	item := model.VocabularyFolderItem{
		FolderID:     folderID,
		VocabularyID: vocabID,
	}
	return r.db.Create(&item).Error
}

// RemoveFromFolder 从文件夹移除生词
func (r *VocabularyRepository) RemoveFromFolder(folderID, vocabID uint) error {
	return r.db.Where("folder_id = ? AND vocabulary_id = ?", folderID, vocabID).
		Delete(&model.VocabularyFolderItem{}).Error
}

// ListByFolder 获取文件夹中的生词
func (r *VocabularyRepository) ListByFolder(folderID uint, limit, offset int) ([]model.Vocabulary, int64, error) {
	var vocabs []model.Vocabulary
	var total int64

	// 子查询获取 vocabulary_ids
	subQuery := r.db.Model(&model.VocabularyFolderItem{}).
		Select("vocabulary_id").
		Where("folder_id = ?", folderID)

	query := r.db.Model(&model.Vocabulary{}).Where("id IN (?)", subQuery)

	query.Count(&total)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("created_at DESC").Find(&vocabs).Error
	return vocabs, total, err
}

// ==================== 每日统计 ====================

// GetOrCreateDailyStats 获取或创建当日统计
func (r *VocabularyRepository) GetOrCreateDailyStats(userID uint, date time.Time) (*model.VocabularyDailyStats, error) {
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	var stats model.VocabularyDailyStats
	err := r.db.Where("user_id = ? AND stat_date = ?", userID, dateOnly).First(&stats).Error
	if err == gorm.ErrRecordNotFound {
		stats = model.VocabularyDailyStats{
			UserID:   userID,
			StatDate: dateOnly,
		}
		if err := r.db.Create(&stats).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &stats, nil
}

// UpdateDailyStats 更新每日统计
// 注意：使用 Updates 只更新有变化的字段，避免覆盖其他字段
func (r *VocabularyRepository) UpdateDailyStats(stats *model.VocabularyDailyStats) error {
	updates := map[string]interface{}{
		"new_words":           stats.NewWords,
		"reviewed_words":      stats.ReviewedWords,
		"mastered_words":      stats.MasteredWords,
		"review_time_seconds": stats.ReviewTimeSeconds,
	}
	if stats.CorrectRate != nil {
		updates["correct_rate"] = *stats.CorrectRate
	}
	return r.db.Model(&model.VocabularyDailyStats{}).
		Where("id = ?", stats.ID).
		Updates(updates).Error
}

// IncrementNewWords 增加新学单词数（仅更新这一个字段）
func (r *VocabularyRepository) IncrementNewWords(statsID uint) error {
	return r.db.Model(&model.VocabularyDailyStats{}).
		Where("id = ?", statsID).
		Updates(map[string]interface{}{
			"new_words": gorm.Expr("COALESCE(new_words, 0) + 1"),
		}).Error
}

// IncrementReviewedWords 增加复习单词数（仅更新这一个字段）
func (r *VocabularyRepository) IncrementReviewedWords(statsID uint) error {
	return r.db.Model(&model.VocabularyDailyStats{}).
		Where("id = ?", statsID).
		Updates(map[string]interface{}{
			"reviewed_words": gorm.Expr("COALESCE(reviewed_words, 0) + 1"),
		}).Error
}

// IncrementMasteredWords 增加掌握单词数（仅更新这一个字段）
func (r *VocabularyRepository) IncrementMasteredWords(statsID uint) error {
	return r.db.Model(&model.VocabularyDailyStats{}).
		Where("id = ?", statsID).
		Updates(map[string]interface{}{
			"mastered_words": gorm.Expr("COALESCE(mastered_words, 0) + 1"),
		}).Error
}

// GetDailyStatsRange 获取日期范围内的统计
func (r *VocabularyRepository) GetDailyStatsRange(userID uint, startDate, endDate time.Time) ([]model.VocabularyDailyStats, error) {
	var stats []model.VocabularyDailyStats
	err := r.db.Where("user_id = ? AND stat_date BETWEEN ? AND ?", userID, startDate, endDate).
		Order("stat_date ASC").Find(&stats).Error
	return stats, err
}
