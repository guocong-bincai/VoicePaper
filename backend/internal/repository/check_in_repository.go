package repository

import (
	"time"
	"voicepaper/internal/model"

	"gorm.io/gorm"
)

type CheckInRepository struct {
	db *gorm.DB
}

func NewCheckInRepository(db *gorm.DB) *CheckInRepository {
	return &CheckInRepository{db: db}
}

// Create 创建签到记录（在事务中执行）
func (r *CheckInRepository) Create(tx *gorm.DB, checkIn *model.UserCheckIn) error {
	return tx.Create(checkIn).Error
}

// CheckTodayCheckIn 检查今天是否已签到
func (r *CheckInRepository) CheckTodayCheckIn(userID uint) (bool, error) {
	today := time.Now().Format("2006-01-02")
	var count int64
	err := r.db.Model(&model.UserCheckIn{}).
		Where("user_id = ? AND check_in_date = ?", userID, today).
		Count(&count).Error
	return count > 0, err
}

// CheckDateCheckIn 检查指定日期是否已签到
func (r *CheckInRepository) CheckDateCheckIn(userID uint, date time.Time) (bool, error) {
	dateStr := date.Format("2006-01-02")
	var count int64
	err := r.db.Model(&model.UserCheckIn{}).
		Where("user_id = ? AND check_in_date = ?", userID, dateStr).
		Count(&count).Error
	return count > 0, err
}

// GetLastCheckIn 获取用户最后一次签到记录
func (r *CheckInRepository) GetLastCheckIn(userID uint) (*model.UserCheckIn, error) {
	var checkIn model.UserCheckIn
	err := r.db.Where("user_id = ?", userID).
		Order("check_in_date DESC").
		First(&checkIn).Error
	if err != nil {
		return nil, err
	}
	return &checkIn, nil
}

// GetCheckInHistory 获取用户签到历史（最近N天）
func (r *CheckInRepository) GetCheckInHistory(userID uint, days int) ([]model.UserCheckIn, error) {
	var checkIns []model.UserCheckIn
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	err := r.db.Where("user_id = ? AND check_in_date >= ?", userID, startDate).
		Order("check_in_date DESC").
		Find(&checkIns).Error
	return checkIns, err
}

// CalculateContinuousDays 计算连续签到天数
// 从今天开始向前遍历，计算连续签到的天数
func (r *CheckInRepository) CalculateContinuousDays(userID uint) (int, error) {
	// 获取用户所有签到记录，按日期倒序
	var checkIns []model.UserCheckIn
	err := r.db.Where("user_id = ?", userID).
		Order("check_in_date DESC").
		Find(&checkIns).Error
	if err != nil {
		return 0, err
	}

	if len(checkIns) == 0 {
		return 0, nil // 从未签到，连续天数为0
	}

	// 构建签到日期集合
	checkedDates := make(map[string]bool)
	for _, c := range checkIns {
		checkedDates[c.CheckInDate.Format("2006-01-02")] = true
	}

	// 从今天开始向前计算连续天数
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	continuousDays := 0
	for i := 0; i <= 365; i++ { // 最多检查365天
		checkDate := today.AddDate(0, 0, -i)
		dateStr := checkDate.Format("2006-01-02")

		if checkedDates[dateStr] {
			continuousDays++
		} else {
			// 如果是今天还没签到，检查昨天开始的连续天数
			if i == 0 {
				continue // 今天还没签到，继续检查昨天
			}
			break // 中断了，停止计算
		}
	}

	// 如果今天还没签到，返回连续天数（不预支）
	todayStr := today.Format("2006-01-02")
	if !checkedDates[todayStr] {
		return continuousDays, nil
	}

	return continuousDays, nil
}

// GetMonthCheckInCount 获取本月签到次数
func (r *CheckInRepository) GetMonthCheckInCount(userID uint) (int64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := startOfMonth.AddDate(0, 1, 0)

	var count int64
	err := r.db.Model(&model.UserCheckIn{}).
		Where("user_id = ? AND check_in_date >= ? AND check_in_date < ?",
			userID,
			startOfMonth.Format("2006-01-02"),
			nextMonth.Format("2006-01-02")).
		Count(&count).Error
	return count, err
}

// GetCheckInCalendar 获取签到日历（某个月的签到情况）
func (r *CheckInRepository) GetCheckInCalendar(userID uint, year, month int) ([]model.UserCheckIn, error) {
	var checkIns []model.UserCheckIn
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	nextMonth := startDate.AddDate(0, 1, 0)

	err := r.db.Where("user_id = ? AND check_in_date >= ? AND check_in_date < ?",
		userID,
		startDate.Format("2006-01-02"),
		nextMonth.Format("2006-01-02")).
		Order("check_in_date ASC").
		Find(&checkIns).Error
	return checkIns, err
}

// CheckInRankingItem 签到排行榜项
type CheckInRankingItem struct {
	UserID                uint   `gorm:"column:user_id"`
	Nickname              string `gorm:"column:nickname"`
	Avatar                string `gorm:"column:avatar"`
	TotalCheckIns         int64  `gorm:"column:total_check_ins"`
	MaxContinuousCheckIns int64  `gorm:"column:max_continuous_check_ins"`
	CurrentContinuousDays int64  `gorm:"column:current_continuous_days"`
}

// GetCheckInRanking 获取签到排行榜
// sortBy: "total" 按总签到次数排序, "continuous" 按最大连续签到天数排序
func (r *CheckInRepository) GetCheckInRanking(limit int, sortBy string) ([]CheckInRankingItem, error) {
	var results []CheckInRankingItem

	var orderBy string
	if sortBy == "continuous" {
		orderBy = "up.max_continuous_check_ins DESC, up.total_check_ins DESC"
	} else {
		orderBy = "up.total_check_ins DESC, up.max_continuous_check_ins DESC"
	}

	query := `
	SELECT
		up.user_id,
		COALESCE(u.nickname, '') as nickname,
		COALESCE(u.avatar, '') as avatar,
		COALESCE(up.total_check_ins, 0) as total_check_ins,
		COALESCE(up.max_continuous_check_ins, 0) as max_continuous_check_ins,
		COALESCE(up.continuous_check_ins, 0) as current_continuous_days
	FROM vp_user_points up
	LEFT JOIN vp_users u ON up.user_id = u.id AND u.deleted_at IS NULL
	WHERE up.deleted_at IS NULL
	ORDER BY ` + orderBy + `
	LIMIT ?
	`

	err := r.db.Raw(query, limit).Scan(&results).Error
	return results, err
}

// GetUserCheckInRank 获取用户的签到排名
func (r *CheckInRepository) GetUserCheckInRank(userID uint, sortBy string) (int64, error) {
	// 获取所有排名
	allRankings, err := r.GetCheckInRanking(10000, sortBy)
	if err != nil {
		return 0, err
	}

	// 查找用户排名
	for i, item := range allRankings {
		if item.UserID == userID {
			return int64(i + 1), nil
		}
	}

	return 0, nil
}
