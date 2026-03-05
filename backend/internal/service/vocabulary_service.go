package service

import (
	"errors"
	"fmt"
	"math"
	"time"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"

	"gorm.io/gorm"
)

// VocabularyService 生词本服务
type VocabularyService struct {
	repo           *repository.VocabularyRepository
	db             *gorm.DB
	userPointsRepo *repository.UserPointsRepository
}

// NewVocabularyService 创建生词本服务实例
func NewVocabularyService() *VocabularyService {
	return &VocabularyService{
		repo:           repository.NewVocabularyRepository(),
		db:             repository.DB,
		userPointsRepo: repository.NewUserPointsRepository(repository.DB),
	}
}

// ==================== 生词 CRUD ====================

// AddVocabularyRequest 添加生词请求
type AddVocabularyRequest struct {
	ArticleID          *uint  `json:"article_id"`
	SentenceID         *uint  `json:"sentence_id"`
	Type               string `json:"type" binding:"required,oneof=word phrase sentence"`
	Content            string `json:"content" binding:"required,max=500"`
	Phonetic           string `json:"phonetic"`
	Meaning            string `json:"meaning"`
	Example            string `json:"example"`
	ExampleTranslation string `json:"example_translation"`
	Context            string `json:"context"`
	Note               string `json:"note"`
	Source             string `json:"source"` // 来源：article/manual（默认），由前端传递或后端自动判断
}

// AddVocabulary 添加生词
func (s *VocabularyService) AddVocabulary(userID uint, req *AddVocabularyRequest) (*model.Vocabulary, error) {
	// 检查是否已存在
	existing, err := s.repo.GetByUserIDAndContent(userID, req.Content)
	if err == nil && existing != nil {
		return nil, errors.New("该生词已在生词本中")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 确定来源
	source := "manual" // 默认：手动添加
	if req.ArticleID != nil {
		source = "article" // 从文章中添加
	}
	if req.Source != "" {
		source = req.Source // 如果前端显式传递，则使用前端的值
	}

	// 新词默认下次复习时间为立即（表示新词待学习）
	now := time.Now()
	vocab := &model.Vocabulary{
		UserID:             userID,
		ArticleID:          req.ArticleID,
		SentenceID:         req.SentenceID,
		Type:               model.VocabularyType(req.Type),
		Content:            req.Content,
		Phonetic:           req.Phonetic,
		Meaning:            req.Meaning,
		Example:            req.Example,
		ExampleTranslation: req.ExampleTranslation,
		Context:            req.Context,
		Note:               req.Note,
		Source:             source, // 来源标识
		MasteryLevel:       0,
		EaseFactor:         2.5,
		IntervalDays:       0,
		Repetitions:        0,
		NextReviewAt:       &now, // 新词立即可复习
	}

	if err := s.repo.Create(vocab); err != nil {
		return nil, err
	}

	// 更新每日统计
	s.updateDailyStatsNewWord(userID)

	return vocab, nil
}

// GetVocabulary 获取单个生词
func (s *VocabularyService) GetVocabulary(id uint, userID uint) (*model.Vocabulary, error) {
	vocab, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if vocab.UserID != userID {
		return nil, errors.New("无权访问该生词")
	}
	return vocab, nil
}

// UpdateVocabularyRequest 更新生词请求
type UpdateVocabularyRequest struct {
	Phonetic           *string `json:"phonetic"`
	Meaning            *string `json:"meaning"`
	Example            *string `json:"example"`
	ExampleTranslation *string `json:"example_translation"`
	Context            *string `json:"context"`
	Note               *string `json:"note"`
	Tags               *string `json:"tags"`
	IsStarred          *bool   `json:"is_starred"`
}

// UpdateVocabulary 更新生词
func (s *VocabularyService) UpdateVocabulary(id uint, userID uint, req *UpdateVocabularyRequest) (*model.Vocabulary, error) {
	vocab, err := s.GetVocabulary(id, userID)
	if err != nil {
		return nil, err
	}

	if req.Phonetic != nil {
		vocab.Phonetic = *req.Phonetic
	}
	if req.Meaning != nil {
		vocab.Meaning = *req.Meaning
	}
	if req.Example != nil {
		vocab.Example = *req.Example
	}
	if req.ExampleTranslation != nil {
		vocab.ExampleTranslation = *req.ExampleTranslation
	}
	if req.Context != nil {
		vocab.Context = *req.Context
	}
	if req.Note != nil {
		vocab.Note = *req.Note
	}
	if req.Tags != nil {
		vocab.Tags = *req.Tags
	}
	if req.IsStarred != nil {
		vocab.IsStarred = *req.IsStarred
	}

	if err := s.repo.Update(vocab); err != nil {
		return nil, err
	}
	return vocab, nil
}

// DeleteVocabulary 删除生词
func (s *VocabularyService) DeleteVocabulary(id uint, userID uint) error {
	vocab, err := s.GetVocabulary(id, userID)
	if err != nil {
		return err
	}
	return s.repo.Delete(vocab.ID)
}

// ListVocabulary 获取生词列表
func (s *VocabularyService) ListVocabulary(userID uint, params *repository.VocabularyListParams) ([]model.Vocabulary, int64, error) {
	return s.repo.ListByUserID(userID, params)
}

// GetStats 获取统计信息
func (s *VocabularyService) GetStats(userID uint) (*repository.VocabularyStats, error) {
	return s.repo.GetStats(userID)
}

// ToggleStar 切换标星
func (s *VocabularyService) ToggleStar(id uint, userID uint) (bool, error) {
	vocab, err := s.GetVocabulary(id, userID)
	if err != nil {
		return false, err
	}
	newState := !vocab.IsStarred
	err = s.repo.ToggleStar(id, newState)
	return newState, err
}

// BatchDelete 批量删除
func (s *VocabularyService) BatchDelete(userID uint, ids []uint) error {
	return s.repo.BatchDelete(userID, ids)
}

// ==================== 复习功能 (SM-2算法) ====================

// GetTodayReviewList 获取今日待复习列表
func (s *VocabularyService) GetTodayReviewList(userID uint, limit int) ([]model.Vocabulary, error) {
	return s.repo.GetTodayReviewList(userID, limit)
}

// SubmitReviewRequest 提交复习结果请求
type SubmitReviewRequest struct {
	VocabularyID   uint   `json:"vocabulary_id"`                                                     // 从URL参数获取
	Quality        int    `json:"quality" binding:"min=0,max=5"`                                     // SM-2评分 0-5
	ReviewType     string `json:"review_type" binding:"omitempty,oneof=card spell choice dictation"` // 可选
	ResponseTimeMs *int   `json:"response_time_ms"`
	UserAnswer     string `json:"user_answer"`
}

// SubmitReviewResult 提交复习结果返回
type SubmitReviewResult struct {
	Vocabulary   *model.Vocabulary `json:"vocabulary"`
	PointsEarned int               `json:"points_earned"` // 本次获得的积分
	TotalPoints  int               `json:"total_points"`  // 当前总积分
}

// SubmitReview 提交复习结果（SM-2算法核心）
func (s *VocabularyService) SubmitReview(userID uint, req *SubmitReviewRequest) (*SubmitReviewResult, error) {
	vocab, err := s.GetVocabulary(req.VocabularyID, userID)
	if err != nil {
		return nil, err
	}

	// 保存复习前状态
	prevInterval := vocab.IntervalDays
	prevEaseFactor := vocab.EaseFactor

	// 应用SM-2算法
	s.applySM2Algorithm(vocab, req.Quality)

	// 更新统计
	vocab.ReviewCount++
	if req.Quality >= 3 {
		vocab.CorrectCount++
	} else {
		vocab.WrongCount++
	}

	now := time.Now()
	vocab.LastReviewAt = &now

	// 保存更新
	if err := s.repo.Update(vocab); err != nil {
		return nil, err
	}

	// 记录复习历史
	reviewType := model.ReviewTypeCard
	if req.ReviewType != "" {
		reviewType = model.ReviewType(req.ReviewType)
	}

	newInterval := vocab.IntervalDays
	newEaseFactor := vocab.EaseFactor
	review := &model.VocabularyReview{
		UserID:         userID,
		VocabularyID:   vocab.ID,
		ReviewType:     reviewType,
		Quality:        req.Quality,
		IsCorrect:      req.Quality >= 3,
		ResponseTimeMs: req.ResponseTimeMs,
		UserAnswer:     req.UserAnswer,
		PrevInterval:   &prevInterval,
		NewInterval:    &newInterval,
		PrevEaseFactor: &prevEaseFactor,
		NewEaseFactor:  &newEaseFactor,
	}
	s.repo.CreateReview(review)

	// 更新每日统计
	s.updateDailyStatsReview(userID, req.Quality >= 3, vocab.MasteryLevel >= 4)

	// 添加复习积分：忘记+1分，模糊+2分，认识+3分
	pointsEarned, totalPoints := s.addReviewPoints(userID, req.Quality, vocab.Content)

	return &SubmitReviewResult{
		Vocabulary:   vocab,
		PointsEarned: pointsEarned,
		TotalPoints:  totalPoints,
	}, nil
}

// addReviewPoints 添加复习积分
// quality: 1=忘记(+1分), 3=模糊(+2分), 5=认识(+3分)
// 返回：本次获得的积分，当前总积分
func (s *VocabularyService) addReviewPoints(userID uint, quality int, wordContent string) (int, int) {
	var points int
	var description string

	switch {
	case quality <= 1:
		points = 1
		description = fmt.Sprintf("复习单词「%s」- 忘记", wordContent)
	case quality <= 3:
		points = 2
		description = fmt.Sprintf("复习单词「%s」- 模糊", wordContent)
	default:
		points = 3
		description = fmt.Sprintf("复习单词「%s」- 认识", wordContent)
	}

	// 检查 db 是否为 nil
	if s.db == nil {
		fmt.Printf("❌ 复习积分添加失败: db 为 nil, userID=%d, points=%d\n", userID, points)
		return 0, 0
	}

	var totalPoints int

	// 使用事务添加积分
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 获取并锁定用户积分记录
		userPoints, err := s.userPointsRepo.AddPoints(tx, userID, points)
		if err != nil {
			return fmt.Errorf("AddPoints失败: %w", err)
		}

		totalPoints = userPoints.TotalPoints
		fmt.Printf("✅ 复习积分添加成功: userID=%d, points=+%d, total=%d, word=%s\n", userID, points, totalPoints, wordContent)

		// 创建积分记录
		pointRecord := &model.PointRecord{
			UserID:        userID,
			Points:        points,
			Type:          model.PointTypeVocabularyReview,
			Description:   description,
			BalanceBefore: userPoints.TotalPoints - points,
			BalanceAfter:  userPoints.TotalPoints,
		}

		if err := tx.Create(pointRecord).Error; err != nil {
			return fmt.Errorf("创建积分记录失败: %w", err)
		}

		return nil
	})

	if err != nil {
		// 积分添加失败不影响复习功能，只记录日志
		fmt.Printf("❌ 复习积分事务失败: userID=%d, error=%v\n", userID, err)
		return 0, 0
	}

	return points, totalPoints
}

// applySM2Algorithm 应用SM-2间隔重复算法
// quality: 0-5 的评分
// 0 - 完全忘记
// 1 - 错误，看到答案后想起来
// 2 - 错误，答案感觉熟悉
// 3 - 正确，很费力
// 4 - 正确，有些犹豫
// 5 - 正确，非常轻松
func (s *VocabularyService) applySM2Algorithm(vocab *model.Vocabulary, quality int) {
	// 更新易度因子 (EF)
	// EF' = EF + (0.1 - (5 - q) * (0.08 + (5 - q) * 0.02))
	ef := vocab.EaseFactor + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
	if ef < 1.3 {
		ef = 1.3 // 最小值
	}
	vocab.EaseFactor = math.Round(ef*100) / 100

	if quality < 3 {
		// 回答错误，重置
		vocab.Repetitions = 0
		vocab.IntervalDays = 0
		vocab.MasteryLevel = max(0, vocab.MasteryLevel-1) // 降级
	} else {
		// 回答正确
		vocab.Repetitions++

		// 计算新的间隔
		switch vocab.Repetitions {
		case 1:
			vocab.IntervalDays = 1
		case 2:
			vocab.IntervalDays = 6
		default:
			vocab.IntervalDays = int(math.Round(float64(vocab.IntervalDays) * vocab.EaseFactor))
		}

		// 更新掌握等级
		if vocab.Repetitions >= 5 && quality >= 4 {
			vocab.MasteryLevel = 5
		} else if vocab.Repetitions >= 3 {
			vocab.MasteryLevel = min(4, vocab.MasteryLevel+1)
		} else if vocab.Repetitions >= 1 {
			vocab.MasteryLevel = min(2, vocab.MasteryLevel+1)
		}
	}

	// 计算下次复习时间
	if vocab.IntervalDays == 0 {
		// 立即复习（1分钟后），方便用户快速重新学习
		next := time.Now().Add(1 * time.Minute)
		vocab.NextReviewAt = &next
	} else {
		next := time.Now().AddDate(0, 0, vocab.IntervalDays)
		vocab.NextReviewAt = &next
	}
}

// ==================== 文件夹管理 ====================

// CreateFolderRequest 创建文件夹请求
type CreateFolderRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

// CreateFolder 创建文件夹
func (s *VocabularyService) CreateFolder(userID uint, req *CreateFolderRequest) (*model.VocabularyFolder, error) {
	folder := &model.VocabularyFolder{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	}
	if folder.Color == "" {
		folder.Color = "#3b82f6"
	}
	if err := s.repo.CreateFolder(folder); err != nil {
		return nil, err
	}
	return folder, nil
}

// ListFolders 获取文件夹列表
func (s *VocabularyService) ListFolders(userID uint) ([]model.VocabularyFolder, error) {
	return s.repo.ListFolders(userID)
}

// UpdateFolder 更新文件夹
func (s *VocabularyService) UpdateFolder(id uint, userID uint, req *CreateFolderRequest) (*model.VocabularyFolder, error) {
	folders, err := s.repo.ListFolders(userID)
	if err != nil {
		return nil, err
	}

	var folder *model.VocabularyFolder
	for i := range folders {
		if folders[i].ID == id {
			folder = &folders[i]
			break
		}
	}
	if folder == nil {
		return nil, errors.New("文件夹不存在")
	}

	folder.Name = req.Name
	folder.Description = req.Description
	if req.Color != "" {
		folder.Color = req.Color
	}
	folder.Icon = req.Icon

	if err := s.repo.UpdateFolder(folder); err != nil {
		return nil, err
	}
	return folder, nil
}

// DeleteFolder 删除文件夹
func (s *VocabularyService) DeleteFolder(id uint, userID uint) error {
	folders, err := s.repo.ListFolders(userID)
	if err != nil {
		return err
	}

	found := false
	for _, f := range folders {
		if f.ID == id {
			found = true
			break
		}
	}
	if !found {
		return errors.New("文件夹不存在")
	}

	return s.repo.DeleteFolder(id)
}

// AddToFolder 添加生词到文件夹
func (s *VocabularyService) AddToFolder(folderID, vocabID, userID uint) error {
	// 验证权限
	_, err := s.GetVocabulary(vocabID, userID)
	if err != nil {
		return err
	}
	return s.repo.AddToFolder(folderID, vocabID)
}

// RemoveFromFolder 从文件夹移除生词
func (s *VocabularyService) RemoveFromFolder(folderID, vocabID uint) error {
	return s.repo.RemoveFromFolder(folderID, vocabID)
}

// ListByFolder 获取文件夹中的生词
func (s *VocabularyService) ListByFolder(folderID uint, limit, offset int) ([]model.Vocabulary, int64, error) {
	return s.repo.ListByFolder(folderID, limit, offset)
}

// ==================== 辅助方法 ====================

func (s *VocabularyService) updateDailyStatsNewWord(userID uint) {
	stats, err := s.repo.GetOrCreateDailyStats(userID, time.Now())
	if err != nil {
		return
	}
	// 使用增量更新，避免覆盖其他字段
	s.repo.IncrementNewWords(stats.ID)
}

func (s *VocabularyService) updateDailyStatsReview(userID uint, isCorrect bool, isMastered bool) {
	stats, err := s.repo.GetOrCreateDailyStats(userID, time.Now())
	if err != nil {
		return
	}

	// 使用增量更新，避免覆盖其他字段
	s.repo.IncrementReviewedWords(stats.ID)

	if isMastered {
		s.repo.IncrementMasteredWords(stats.ID)
	}

	// 正确率计算暂时跳过，避免覆盖问题
	// TODO: 如果需要计算正确率，应该使用原子更新
}

// GetDailyStats 获取每日统计
func (s *VocabularyService) GetDailyStats(userID uint, days int) ([]model.VocabularyDailyStats, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)
	return s.repo.GetDailyStatsRange(userID, startDate, endDate)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
