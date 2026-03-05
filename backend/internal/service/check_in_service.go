package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"voicepaper/config"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"
	"voicepaper/internal/storage"

	"gorm.io/gorm"
)

// CheckInService 签到服务
type CheckInService struct {
	db              *gorm.DB
	checkInRepo     *repository.CheckInRepository
	userPointsRepo  *repository.UserPointsRepository
	pointRecordRepo *repository.PointRecordRepository
	pointService    *PointService
	storage         storage.Storage
}

func NewCheckInService(db *gorm.DB) *CheckInService {
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

	return &CheckInService{
		db:              db,
		checkInRepo:     repository.NewCheckInRepository(db),
		userPointsRepo:  repository.NewUserPointsRepository(db),
		pointRecordRepo: repository.NewPointRecordRepository(db),
		pointService:    NewPointService(db),
		storage:         st,
	}
}

// CheckIn 用户签到
func (s *CheckInService) CheckIn(userID uint) (map[string]interface{}, error) {
	// 1. 检查今天是否已签到
	alreadyCheckedIn, err := s.checkInRepo.CheckTodayCheckIn(userID)
	if err != nil {
		return nil, err
	}
	if alreadyCheckedIn {
		return nil, fmt.Errorf("今天已签到，请明天再来")
	}

	// 2. 计算连续签到天数
	continuousDays, err := s.checkInRepo.CalculateContinuousDays(userID)
	if err != nil {
		return nil, err
	}

	// 3. 计算本次签到获得的积分
	basePoints := model.PointRewardConfig[model.PointTypeDailyCheckIn]
	bonusPoints := 0
	var bonusMessages []string

	// 检查连续签到奖励
	if reward, exists := model.ContinuousCheckInRewards[continuousDays]; exists {
		bonusPoints = reward
		bonusMessages = append(bonusMessages, fmt.Sprintf("连续签到%d天，额外奖励%d积分", continuousDays, reward))
	}

	totalPoints := basePoints + bonusPoints
	today := time.Now()

	var userPoints *model.UserPoints
	var checkInRecord *model.UserCheckIn
	var pointRecords []*model.PointRecord

	// 4. 使用事务创建签到记录和积分记录
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 4.1 创建签到记录
		checkInRecord = &model.UserCheckIn{
			UserID:         userID,
			CheckInDate:    today,
			PointsAwarded:  totalPoints,
			ContinuousDays: continuousDays,
			IsMakeup:       false,
		}
		err := s.checkInRepo.Create(tx, checkInRecord)
		if err != nil {
			return fmt.Errorf("创建签到记录失败: %w", err)
		}

		// 4.2 更新用户积分（基础签到积分）
		up, err := s.userPointsRepo.AddPoints(tx, userID, basePoints)
		if err != nil {
			return fmt.Errorf("添加签到积分失败: %w", err)
		}
		userPoints = up

		// 4.3 创建基础签到积分记录
		baseRecord := &model.PointRecord{
			UserID:        userID,
			Points:        basePoints,
			Type:          model.PointTypeDailyCheckIn,
			Description:   fmt.Sprintf("每日签到（连续%d天）", continuousDays),
			CheckInID:     &checkInRecord.ID,
			BalanceBefore: userPoints.CurrentPoints - basePoints,
			BalanceAfter:  userPoints.CurrentPoints,
		}
		err = s.pointRecordRepo.Create(tx, baseRecord)
		if err != nil {
			return fmt.Errorf("创建基础积分记录失败: %w", err)
		}
		pointRecords = append(pointRecords, baseRecord)

		// 4.4 如果有连续签到奖励，添加额外积分
		if bonusPoints > 0 {
			up, err = s.userPointsRepo.AddPoints(tx, userID, bonusPoints)
			if err != nil {
				return fmt.Errorf("添加奖励积分失败: %w", err)
			}
			userPoints = up

			bonusRecord := &model.PointRecord{
				UserID:        userID,
				Points:        bonusPoints,
				Type:          model.PointTypeContinuousCheckIn,
				Description:   bonusMessages[0],
				CheckInID:     &checkInRecord.ID,
				BalanceBefore: userPoints.CurrentPoints - bonusPoints,
				BalanceAfter:  userPoints.CurrentPoints,
			}
			err = s.pointRecordRepo.Create(tx, bonusRecord)
			if err != nil {
				return fmt.Errorf("创建奖励积分记录失败: %w", err)
			}
			pointRecords = append(pointRecords, bonusRecord)
		}

		// 4.5 更新用户签到统计
		err = s.userPointsRepo.UpdateCheckInStats(tx, userID, continuousDays)
		if err != nil {
			return fmt.Errorf("更新签到统计失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 5. 异步检查是否达成新称号
	go s.pointService.CheckAndAwardTitles(userID)

	// 6. 获取最新的用户积分统计（包含累计签到和最大连续签到）
	latestUserPoints, err := s.userPointsRepo.GetByUserID(userID)
	if err != nil {
		latestUserPoints = userPoints // 如果获取失败，使用之前的数据
	}

	// 7. 返回签到结果
	return map[string]interface{}{
		"success":             true,
		"message":             "签到成功",
		"continuous_days":     continuousDays,
		"base_points":         basePoints,
		"bonus_points":        bonusPoints,
		"total_points":        totalPoints,
		"bonus_messages":      bonusMessages,
		"current_points":      latestUserPoints.CurrentPoints,
		"level":               latestUserPoints.Level,
		"level_name":          latestUserPoints.LevelName,
		"check_in_date":       checkInRecord.CheckInDate.Format("2006-01-02"),
		"total_check_ins":     latestUserPoints.TotalCheckIns,         // 累计签到天数
		"max_continuous_days": latestUserPoints.MaxContinuousCheckIns, // 最大连续签到天数
	}, nil
}

// GetCheckInStatus 获取签到状态
func (s *CheckInService) GetCheckInStatus(userID uint) (map[string]interface{}, error) {
	// 1. 检查今天是否已签到
	todayCheckedIn, err := s.checkInRepo.CheckTodayCheckIn(userID)
	if err != nil {
		return nil, err
	}

	// 2. 获取用户积分信息（包含连续签到天数）
	userPoints, err := s.userPointsRepo.GetByUserID(userID)
	if err != nil {
		// 如果没有积分记录，实时计算连续签到天数
		continuousDays, _ := s.checkInRepo.CalculateContinuousDays(userID)
		userPoints = &model.UserPoints{
			ContinuousCheckIns: continuousDays,
		}
	}

	// 3. 获取本月签到次数
	monthCheckInCount, err := s.checkInRepo.GetMonthCheckInCount(userID)
	if err != nil {
		monthCheckInCount = 0
	}

	// 4. 获取最近7天的签到历史
	recentCheckIns, err := s.checkInRepo.GetCheckInHistory(userID, 7)
	if err != nil {
		recentCheckIns = []model.UserCheckIn{}
	}

	// 5. 计算下次奖励里程碑
	nextMilestone := 0
	nextMilestoneReward := 0
	for days, reward := range model.ContinuousCheckInRewards {
		if days > userPoints.ContinuousCheckIns {
			if nextMilestone == 0 || days < nextMilestone {
				nextMilestone = days
				nextMilestoneReward = reward
			}
		}
	}

	return map[string]interface{}{
		"today_checked_in":      todayCheckedIn,
		"continuous_days":       userPoints.ContinuousCheckIns,
		"month_check_in_count":  monthCheckInCount,
		"recent_check_ins":      recentCheckIns,
		"next_milestone":        nextMilestone,
		"next_milestone_reward": nextMilestoneReward,
		"check_in_rewards":      model.ContinuousCheckInRewards,
	}, nil
}

// GetCheckInRanking 获取签到排行榜
func (s *CheckInService) GetCheckInRanking(limit int, sortBy string, userID uint) (map[string]interface{}, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if sortBy != "total" && sortBy != "continuous" {
		sortBy = "total"
	}

	// 获取排行榜
	rankings, err := s.checkInRepo.GetCheckInRanking(limit, sortBy)
	if err != nil {
		return nil, err
	}

	// 获取用户排名
	userRank, err := s.checkInRepo.GetUserCheckInRank(userID, sortBy)
	if err != nil {
		userRank = 0
	}

	// 获取用户数据
	userPoints, err := s.userPointsRepo.GetByUserID(userID)
	if err != nil {
		userPoints = nil
	}

	var userNickname string
	if userPoints != nil {
		var user model.User
		if err := s.db.Where("id = ?", userID).First(&user).Error; err == nil {
			userNickname = user.Nickname
		}
	}

	// 构建排行榜项（添加排名）
	items := make([]map[string]interface{}, len(rankings))
	for i, item := range rankings {
		// 处理头像URL签名
		avatar := item.Avatar
		if avatar != "" && s.storage != nil && strings.Contains(avatar, "oss-cn-") {
			ctx := context.Background()
			parts := strings.Split(avatar, ".aliyuncs.com/")
			if len(parts) > 1 {
				path := parts[1]
				signedURL, err := s.storage.GetSignedURL(ctx, path, 3153600000) // 100 years expiration
				if err == nil {
					avatar = signedURL
				}
			}
		}

		if avatar == "" {
			// 如果没有头像，使用默认头像服务
			displayName := item.Nickname
			if displayName == "" {
				displayName = "U"
			}
			avatar = fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=2563EB&color=fff&size=128&bold=true", displayName)
		}

		items[i] = map[string]interface{}{
			"rank":                     i + 1,
			"user_id":                  item.UserID,
			"nickname":                 item.Nickname,
			"avatar":                   avatar,
			"total_check_ins":          item.TotalCheckIns,
			"max_continuous_check_ins": item.MaxContinuousCheckIns,
			"current_continuous_days":  item.CurrentContinuousDays,
			"is_me":                    item.UserID == userID,
		}
	}

	return map[string]interface{}{
		"items":     items,
		"user_rank": userRank,
		"user_data": map[string]interface{}{
			"user_id":                  userID,
			"nickname":                 userNickname,
			"total_check_ins":          userPoints.TotalCheckIns,
			"max_continuous_check_ins": userPoints.MaxContinuousCheckIns,
			"current_continuous_days":  userPoints.ContinuousCheckIns,
		},
		"sort_by": sortBy,
	}, nil
}

// GetCheckInCalendar 获取签到日历（某个月的签到情况）
func (s *CheckInService) GetCheckInCalendar(userID uint, year, month int) (map[string]interface{}, error) {
	checkIns, err := s.checkInRepo.GetCheckInCalendar(userID, year, month)
	if err != nil {
		return nil, err
	}

	// 转换为日期集合，方便前端渲染
	checkInDates := make([]string, 0, len(checkIns))
	for _, checkIn := range checkIns {
		checkInDates = append(checkInDates, checkIn.CheckInDate.Format("2006-01-02"))
	}

	return map[string]interface{}{
		"year":             year,
		"month":            month,
		"check_in_dates":   checkInDates,
		"check_in_records": checkIns,
		"total_days":       len(checkIns),
	}, nil
}

// MakeupCardPrice 补签卡价格（积分）
const MakeupCardPrice = 500

// MakeupCardMaxDays 可补签的最大天数范围
const MakeupCardMaxDays = 7

// GetMakeupCardInfo 获取补签卡信息
func (s *CheckInService) GetMakeupCardInfo(userID uint) (map[string]interface{}, error) {
	// 1. 获取用户积分信息
	userPoints, err := s.userPointsRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 2. 获取可补签的日期（最近7天内未签到的日期，不包括今天）
	makeupDates, err := s.getMakeupAvailableDates(userID)
	if err != nil {
		return nil, fmt.Errorf("获取可补签日期失败: %w", err)
	}

	return map[string]interface{}{
		"makeup_cards":    userPoints.MakeupCards,
		"current_points":  userPoints.CurrentPoints,
		"card_price":      MakeupCardPrice,
		"can_buy":         userPoints.CurrentPoints >= MakeupCardPrice,
		"makeup_dates":    makeupDates,
		"max_makeup_days": MakeupCardMaxDays,
	}, nil
}

// getMakeupAvailableDates 获取可补签的日期列表
func (s *CheckInService) getMakeupAvailableDates(userID uint) ([]string, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 获取最近7天的签到记录
	checkIns, err := s.checkInRepo.GetCheckInHistory(userID, MakeupCardMaxDays+1)
	if err != nil {
		return nil, err
	}

	// 转换为日期集合
	checkedDates := make(map[string]bool)
	for _, checkIn := range checkIns {
		checkedDates[checkIn.CheckInDate.Format("2006-01-02")] = true
	}

	// 找出未签到的日期（不包括今天）
	var makeupDates []string
	for i := 1; i <= MakeupCardMaxDays; i++ {
		date := today.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		if !checkedDates[dateStr] {
			makeupDates = append(makeupDates, dateStr)
		}
	}

	return makeupDates, nil
}

// BuyMakeupCard 购买补签卡
func (s *CheckInService) BuyMakeupCard(userID uint, count int) (map[string]interface{}, error) {
	if count <= 0 {
		count = 1
	}
	if count > 10 {
		return nil, fmt.Errorf("单次最多购买10张补签卡")
	}

	totalCost := MakeupCardPrice * count

	// 获取用户积分
	userPoints, err := s.userPointsRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	if userPoints.CurrentPoints < totalCost {
		return nil, fmt.Errorf("积分不足，需要%d积分，当前只有%d积分", totalCost, userPoints.CurrentPoints)
	}

	// 使用事务执行购买
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 扣除积分
		_, err := s.userPointsRepo.DeductPoints(tx, userID, totalCost)
		if err != nil {
			return fmt.Errorf("扣除积分失败: %w", err)
		}

		// 增加补签卡数量
		err = s.userPointsRepo.AddMakeupCards(tx, userID, count)
		if err != nil {
			return fmt.Errorf("增加补签卡失败: %w", err)
		}

		// 获取最新积分余额
		latestPoints, _ := s.userPointsRepo.GetByUserIDWithTx(tx, userID)
		balanceAfter := 0
		if latestPoints != nil {
			balanceAfter = latestPoints.CurrentPoints
		}

		// 创建积分消费记录
		record := &model.PointRecord{
			UserID:        userID,
			Points:        -totalCost,
			Type:          "makeup_card_purchase",
			Description:   fmt.Sprintf("购买%d张补签卡", count),
			BalanceBefore: userPoints.CurrentPoints,
			BalanceAfter:  balanceAfter,
		}
		return s.pointRecordRepo.Create(tx, record)
	})

	if err != nil {
		return nil, err
	}

	// 获取最新数据
	latestUserPoints, _ := s.userPointsRepo.GetByUserID(userID)

	return map[string]interface{}{
		"success":        true,
		"message":        fmt.Sprintf("成功购买%d张补签卡", count),
		"cards_bought":   count,
		"points_spent":   totalCost,
		"makeup_cards":   latestUserPoints.MakeupCards,
		"current_points": latestUserPoints.CurrentPoints,
	}, nil
}

// UseMakeupCard 使用补签卡补签
func (s *CheckInService) UseMakeupCard(userID uint, dateStr string) (map[string]interface{}, error) {
	// 1. 解析日期
	makeupDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("日期格式错误，请使用 YYYY-MM-DD 格式")
	}

	// 2. 检查日期是否在可补签范围内
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	makeupDateNormalized := time.Date(makeupDate.Year(), makeupDate.Month(), makeupDate.Day(), 0, 0, 0, 0, now.Location())

	if !makeupDateNormalized.Before(today) {
		return nil, fmt.Errorf("只能补签今天之前的日期")
	}

	daysDiff := int(today.Sub(makeupDateNormalized).Hours() / 24)
	if daysDiff > MakeupCardMaxDays {
		return nil, fmt.Errorf("只能补签最近%d天内的日期", MakeupCardMaxDays)
	}

	// 3. 检查该日期是否已签到
	alreadyCheckedIn, err := s.checkInRepo.CheckDateCheckIn(userID, makeupDate)
	if err != nil {
		return nil, fmt.Errorf("检查签到记录失败: %w", err)
	}
	if alreadyCheckedIn {
		return nil, fmt.Errorf("该日期已签到，无需补签")
	}

	// 4. 检查用户是否有补签卡
	userPoints, err := s.userPointsRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	if userPoints.MakeupCards <= 0 {
		return nil, fmt.Errorf("补签卡不足，请先购买")
	}

	// 5. 使用事务执行补签
	var checkInRecord *model.UserCheckIn
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 5.1 扣除补签卡
		err := s.userPointsRepo.DeductMakeupCards(tx, userID, 1)
		if err != nil {
			return fmt.Errorf("扣除补签卡失败: %w", err)
		}

		// 5.2 创建补签记录（补签不获得积分）
		checkInRecord = &model.UserCheckIn{
			UserID:         userID,
			CheckInDate:    makeupDate,
			PointsAwarded:  0, // 补签不给积分
			ContinuousDays: 0, // 补签时不计算连续天数
			IsMakeup:       true,
		}
		err = s.checkInRepo.Create(tx, checkInRecord)
		if err != nil {
			return fmt.Errorf("创建补签记录失败: %w", err)
		}

		// 5.3 更新累计签到天数
		return tx.Model(&model.UserPoints{}).Where("user_id = ?", userID).
			Update("total_check_ins", gorm.Expr("total_check_ins + 1")).Error
	})

	if err != nil {
		return nil, err
	}

	// 6. 重新计算连续签到天数（补签后可能会影响连续天数）
	newContinuousDays, err := s.checkInRepo.CalculateContinuousDays(userID)
	if err != nil {
		newContinuousDays = 0
	}

	// 7. 更新用户积分表中的连续签到天数
	s.db.Model(&model.UserPoints{}).Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"continuous_check_ins": newContinuousDays,
		})

	// 检查是否更新最大连续签到天数
	latestUserPoints, _ := s.userPointsRepo.GetByUserID(userID)
	if latestUserPoints != nil && newContinuousDays > latestUserPoints.MaxContinuousCheckIns {
		s.db.Model(&model.UserPoints{}).Where("user_id = ?", userID).
			Update("max_continuous_check_ins", newContinuousDays)
		latestUserPoints.MaxContinuousCheckIns = newContinuousDays
	}

	return map[string]interface{}{
		"success":         true,
		"message":         fmt.Sprintf("成功补签 %s", dateStr),
		"makeup_date":     dateStr,
		"makeup_cards":    latestUserPoints.MakeupCards,
		"total_check_ins": latestUserPoints.TotalCheckIns,
		"continuous_days": newContinuousDays,
	}, nil
}
