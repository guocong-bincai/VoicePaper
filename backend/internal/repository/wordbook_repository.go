package repository

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"
	"voicepaper/internal/model"

	"gorm.io/gorm"
)

type WordbookRepository struct {
	db *gorm.DB
}

func NewWordbookRepository(db *gorm.DB) *WordbookRepository {
	return &WordbookRepository{db: db}
}

// DB 获取数据库连接
func (r *WordbookRepository) DB() *gorm.DB {
	return r.db
}

// GetWordsByType 获取指定类型的单词列表（带分页）
// 通过 vp_wordbook_books 关联表查询，支持单词多对多关联到不同单词书
func (r *WordbookRepository) GetWordsByType(wordType string, offset, limit int) ([]model.Wordbook, int64, error) {
	var words []model.Wordbook
	var total int64

	// 通过 JOIN vp_wordbook_books 关联表查询，只返回 vp_wordbook 中存在的记录
	// 使用 JOIN 而不是 IN 子查询，可以自动过滤掉关联表中无效的记录
	db := r.db.Model(&model.Wordbook{}).
		Joins("JOIN vp_wordbook_books ON vp_wordbook_books.word_id = vp_wordbook.id").
		Where("vp_wordbook_books.book_type = ?", wordType)

	if err := db.Count(&total).Error; err != nil {
		log.Printf("❌ GetWordsByType Count 失败 (book_type=%s): %v", wordType, err)
		return nil, 0, err
	}

	log.Printf("✅ GetWordsByType (book_type=%s): 总数 %d, offset %d, limit %d", wordType, total, offset, limit)

	err := db.Order("vp_wordbook.id ASC").Offset(offset).Limit(limit).Find(&words).Error
	if err != nil {
		log.Printf("❌ GetWordsByType Find 失败 (book_type=%s): %v", wordType, err)
	} else {
		log.Printf("✅ GetWordsByType Find 成功 (book_type=%s): 返回 %d 条记录", wordType, len(words))
	}

	return words, total, err
}

// GetWordByID 获取单个单词详情
func (r *WordbookRepository) GetWordByID(id uint) (*model.Wordbook, error) {
	var word model.Wordbook
	err := r.db.First(&word, id).Error
	return &word, err
}

// GetProgress 获取用户学习进度
func (r *WordbookRepository) GetProgress(userID uint, wordType string) (*model.WordbookProgress, error) {
	var progress model.WordbookProgress
	err := r.db.Where("user_id = ? AND word_type = ?", userID, wordType).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// SaveProgress 保存或更新学习进度
func (r *WordbookRepository) SaveProgress(progress *model.WordbookProgress) error {
	var existing model.WordbookProgress
	err := r.db.Where("user_id = ? AND word_type = ?", progress.UserID, progress.WordType).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.Create(progress).Error
	} else if err != nil {
		return err
	}

	existing.CurrentIndex = progress.CurrentIndex
	existing.LastWordID = progress.LastWordID
	if progress.TotalWords > 0 {
		existing.TotalWords = progress.TotalWords
	}

	return r.db.Save(&existing).Error
}

// GetWordbookList 获取所有单词书列表
func (r *WordbookRepository) GetWordbookList() ([]model.WordbookInfo, error) {
	var wordbooks []model.WordbookInfo
	err := r.db.Where("is_active = ?", true).Order("sort_order ASC").Find(&wordbooks).Error
	return wordbooks, err
}

// ImportWordToVocabulary 将单词书中的单词导入到生词本
func (r *WordbookRepository) ImportWordToVocabulary(userID, wordbookID uint, quality string) (*model.Vocabulary, error) {
	// 1. 先从单词书表获取单词信息
	wordbookWord := &model.Wordbook{}
	log.Printf("DEBUG: 尝试查询单词 ID=%d", wordbookID)
	if err := r.db.First(wordbookWord, wordbookID).Error; err != nil {
		log.Printf("❌ 查询单词失败 (ID=%d): %v", wordbookID, err)
		return nil, err
	}
	log.Printf("✅ 查询到单词: %s (ID=%d)", wordbookWord.Word, wordbookWord.ID)

	// 2. 检查该单词是否已在用户生词本中（先检查是否来自同一本单词书）
	var existingVocab model.Vocabulary
	err := r.db.Where("user_id = ? AND content = ? AND wordbook_id = ?", userID, wordbookWord.Word, wordbookID).First(&existingVocab).Error

	if err == nil {
		// 同一本单词书的单词已存在，直接返回
		return &existingVocab, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 3. 创建新的生词本记录
	newVocab := &model.Vocabulary{
		UserID:             userID,
		Type:               model.VocabularyTypeWord,
		Content:            wordbookWord.Word,
		Phonetic:           wordbookWord.Phonetic,
		Meaning:            wordbookWord.Meaning,
		Example:            wordbookWord.Example,
		ExampleTranslation: wordbookWord.ExampleTranslation,
		Tags:               "[\"wordbook\"]",      // 标记为来自单词书
		MasteryLevel:       1,                     // 初始掌握度：1（不认识）或 2（模糊）
		WordbookID:         &wordbookID,           // 记录来自哪个单词书ID
		WordbookType:       wordbookWord.WordType, // 记录单词书类型（cet4/cet6等）
		Source:             "wordbook",            // 来源标识：来自单词书
	}

	// 根据质量设置掌握度
	if quality == "fuzzy" {
		newVocab.MasteryLevel = 2 // 模糊
	} else {
		newVocab.MasteryLevel = 1 // 不认识
	}

	// 4. 保存到生词本
	if err := r.db.Create(newVocab).Error; err != nil {
		return nil, err
	}

	return newVocab, nil
}

// ==================== 顺序/乱序学习相关方法 ====================

// GetUserOrder 获取用户的学习序列
func (r *WordbookRepository) GetUserOrder(userID uint, wordType string) (*model.WordbookUserOrder, error) {
	var order model.WordbookUserOrder
	err := r.db.Where("user_id = ? AND word_type = ?", userID, wordType).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetAllWordIDs 获取指定类型的所有单词ID（按ID顺序）
// 通过 vp_wordbook_books 关联表查询，只返回 vp_wordbook 中存在的记录
func (r *WordbookRepository) GetAllWordIDs(wordType string) ([]uint, error) {
	var ids []uint
	// 通过 JOIN vp_wordbook 表，只查询存在的单词ID
	err := r.db.Model(&model.WordbookBook{}).
		Select("vp_wordbook_books.word_id").
		Joins("JOIN vp_wordbook ON vp_wordbook.id = vp_wordbook_books.word_id").
		Where("vp_wordbook_books.book_type = ?", wordType).
		Order("vp_wordbook_books.word_id ASC").
		Pluck("vp_wordbook_books.word_id", &ids).Error

	if err != nil {
		log.Printf("❌ GetAllWordIDs 查询失败 (book_type=%s): %v", wordType, err)
	} else {
		log.Printf("✅ GetAllWordIDs 查询成功 (book_type=%s): 找到 %d 个单词ID", wordType, len(ids))
	}

	return ids, err
}

// CreateOrUpdateUserOrder 创建或更新用户学习序列
func (r *WordbookRepository) CreateOrUpdateUserOrder(order *model.WordbookUserOrder) error {
	var existing model.WordbookUserOrder
	err := r.db.Where("user_id = ? AND word_type = ?", order.UserID, order.WordType).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.Create(order).Error
	} else if err != nil {
		return err
	}

	// 更新现有记录
	existing.IsRandom = order.IsRandom
	existing.WordSequence = order.WordSequence
	existing.CurrentIndex = order.CurrentIndex
	existing.TotalWords = order.TotalWords
	return r.db.Save(&existing).Error
}

// UpdateUserOrderIndex 只更新学习位置
func (r *WordbookRepository) UpdateUserOrderIndex(userID uint, wordType string, currentIndex int) error {
	return r.db.Model(&model.WordbookUserOrder{}).Where("user_id = ? AND word_type = ?", userID, wordType).
		Update("current_index", currentIndex).Error
}

// SwitchOrderMode 切换顺序/乱序模式
func (r *WordbookRepository) SwitchOrderMode(userID uint, wordType string, isRandom bool) (*model.WordbookUserOrder, error) {
	// 1. 获取所有单词ID
	allIDs, err := r.GetAllWordIDs(wordType)
	if err != nil {
		return nil, err
	}
	if len(allIDs) == 0 {
		log.Printf("⚠️ 单词书 %s 没有单词", wordType)
		return nil, gorm.ErrRecordNotFound
	}

	// 2. 获取现有序列（如果有）
	existingOrder, err := r.GetUserOrder(userID, wordType)
	currentIndex := 0
	if err == nil && existingOrder != nil {
		currentIndex = existingOrder.CurrentIndex
	}

	// 3. 生成新序列
	var newSequence []uint
	if isRandom {
		// 乱序模式：打乱剩余未学单词
		if currentIndex > 0 && currentIndex < len(allIDs) {
			// 保留已学部分，打乱剩余部分
			learned := allIDs[:currentIndex]
			remaining := make([]uint, len(allIDs)-currentIndex)
			copy(remaining, allIDs[currentIndex:])
			shuffleIDs(remaining)
			newSequence = append(learned, remaining...)
		} else {
			// 全部打乱
			newSequence = make([]uint, len(allIDs))
			copy(newSequence, allIDs)
			shuffleIDs(newSequence)
		}
	} else {
		// 顺序模式：恢复原始顺序
		newSequence = allIDs
	}

	// 4. 序列化为JSON
	sequenceJSON, err := json.Marshal(newSequence)
	if err != nil {
		return nil, err
	}

	// 5. 保存
	order := &model.WordbookUserOrder{
		UserID:       userID,
		WordType:     wordType,
		IsRandom:     isRandom,
		WordSequence: string(sequenceJSON),
		CurrentIndex: currentIndex,
		TotalWords:   len(newSequence),
	}

	if err := r.CreateOrUpdateUserOrder(order); err != nil {
		return nil, err
	}

	return order, nil
}

// GetWordsByUserOrder 按用户序列获取单词（带分页）
func (r *WordbookRepository) GetWordsByUserOrder(userID uint, wordType string, page, pageSize int) ([]model.Wordbook, int64, int, bool, error) {
	// 1. 获取用户序列
	order, err := r.GetUserOrder(userID, wordType)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有序列，创建默认顺序序列
			order, err = r.SwitchOrderMode(userID, wordType, false)
			if err != nil {
				return nil, 0, 0, false, err
			}
		} else {
			return nil, 0, 0, false, err
		}
	}

	// 2. 解析序列
	var sequence []uint
	if err := json.Unmarshal([]byte(order.WordSequence), &sequence); err != nil {
		return nil, 0, 0, false, err
	}

	// 3. 检查序列中的单词是否都存在（数据整合后可能有些ID失效了）
	// 获取当前有效的单词ID列表
	validIDs, err := r.GetAllWordIDs(wordType)
	if err != nil {
		return nil, 0, 0, false, err
	}

	// 创建有效ID的map
	validIDMap := make(map[uint]bool)
	for _, id := range validIDs {
		validIDMap[id] = true
	}

	// 过滤序列，只保留有效的ID
	var filteredSequence []uint
	for _, id := range sequence {
		if validIDMap[id] {
			filteredSequence = append(filteredSequence, id)
		}
	}

	// 如果过滤后的序列和原序列长度不一致，说明有些ID失效了，需要更新序列
	if len(filteredSequence) != len(sequence) {
		log.Printf("⚠️ 用户 %d 的单词书 %s 序列中有 %d 个无效ID，重新生成序列", userID, wordType, len(sequence)-len(filteredSequence))

		// 重新生成序列（保持原来的乱序/顺序模式）
		order, err = r.SwitchOrderMode(userID, wordType, order.IsRandom)
		if err != nil {
			return nil, 0, 0, false, err
		}

		// 重新解析序列
		if err := json.Unmarshal([]byte(order.WordSequence), &sequence); err != nil {
			return nil, 0, 0, false, err
		}
	} else {
		sequence = filteredSequence
	}

	total := int64(len(sequence))
	offset := (page - 1) * pageSize

	// 4. 获取当前页的单词ID
	if offset >= len(sequence) {
		return []model.Wordbook{}, total, order.CurrentIndex, order.IsRandom, nil
	}
	end := offset + pageSize
	if end > len(sequence) {
		end = len(sequence)
	}
	pageIDs := sequence[offset:end]

	// 5. 按ID查询单词（保持序列顺序）
	var words []model.Wordbook
	if err := r.db.Where("id IN ?", pageIDs).Find(&words).Error; err != nil {
		return nil, 0, 0, false, err
	}

	// 6. 按序列顺序排序
	wordMap := make(map[uint]model.Wordbook)
	for _, w := range words {
		wordMap[w.ID] = w
	}
	sortedWords := make([]model.Wordbook, 0, len(pageIDs))
	for _, id := range pageIDs {
		if w, ok := wordMap[id]; ok {
			sortedWords = append(sortedWords, w)
		}
	}

	return sortedWords, total, order.CurrentIndex, order.IsRandom, nil
}

// shuffleIDs 打乱切片
func shuffleIDs(ids []uint) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(ids) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		ids[i], ids[j] = ids[j], ids[i]
	}
}
