package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"voicepaper/config"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// stringPtr 创建字符串指针
func stringPtr(s string) *string {
	return &s
}

// emailPtr 创建邮箱指针（空字符串返回nil）
func emailPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// emptyStringToNil 空字符串转nil（避免唯一索引冲突）
func emptyStringToNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// OAuthService OAuth登录服务
type OAuthService struct {
	githubConfig *oauth2.Config
	googleConfig *oauth2.Config
	wechatConfig config.WeChatOAuth
	userRepo     *repository.UserRepository
	bindingRepo  *repository.OAuthBindingRepository
}

func NewOAuthService(cfg *config.OAuthConfig, userRepo *repository.UserRepository, bindingRepo *repository.OAuthBindingRepository) *OAuthService {
	githubConfig := &oauth2.Config{
		ClientID:     cfg.GitHub.ClientID,
		ClientSecret: cfg.GitHub.ClientSecret,
		RedirectURL:  cfg.GitHub.RedirectURI,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	googleConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURI,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	return &OAuthService{
		githubConfig: githubConfig,
		googleConfig: googleConfig,
		wechatConfig: cfg.WeChat,
		userRepo:     userRepo,
		bindingRepo:  bindingRepo,
	}
}

// GetGitHubAuthURL 获取GitHub授权URL
func (s *OAuthService) GetGitHubAuthURL(state string) string {
	return s.githubConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// GetGoogleAuthURL 获取Google授权URL
func (s *OAuthService) GetGoogleAuthURL(state string) string {
	return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOnline, oauth2.ApprovalForce)
}

// HandleGitHubCallback 处理GitHub回调
func (s *OAuthService) HandleGitHubCallback(ctx context.Context, code string) (*model.User, error) {
	// 交换授权码获取token
	token, err := s.githubConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("交换token失败: %w", err)
	}

	// 获取用户信息
	userInfo, err := s.getGitHubUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("获取GitHub用户信息失败: %w", err)
	}

	// 查找或创建用户
	user, err := s.findOrCreateGitHubUser(userInfo, token)
	if err != nil {
		return nil, fmt.Errorf("处理GitHub用户失败: %w", err)
	}

	return user, nil
}

// WeChatSessionResponse 微信code2Session响应
type WeChatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// Code2Session 微信登录凭证校验
func (s *OAuthService) Code2Session(code string) (*WeChatSessionResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.wechatConfig.AppID, s.wechatConfig.AppSecret, code)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求微信接口失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	var session WeChatSessionResponse
	if err := json.Unmarshal(bodyBytes, &session); err != nil {
		return nil, fmt.Errorf("解析微信响应失败: %w", err)
	}

	if session.ErrCode != 0 {
		return nil, fmt.Errorf("微信接口错误: %d %s", session.ErrCode, session.ErrMsg)
	}

	return &session, nil
}

// LoginWithWechat 微信登录
func (s *OAuthService) LoginWithWechat(session *WeChatSessionResponse) (*model.User, error) {
	// 1. 查找是否已有绑定 (通过OpenID)
	// 注意：OAuthBinding表中ProviderUserID存储的是OpenID
	binding, err := s.bindingRepo.FindByProviderUserID("wechat", session.OpenID)
	if err == nil {
		// 检查关联的用户是否存在 (处理僵尸绑定)
		if binding.User.ID == 0 {
			fmt.Printf("⚠️ 发现僵尸绑定 (User ID=0), 正在清理: OpenID=%s\n", session.OpenID)
			// 删除僵尸绑定
			repository.DB.Delete(binding)
			// 继续执行后续逻辑，视为新用户
		} else {
			// 已存在绑定且用户有效，直接返回关联用户
			// 更新SessionKey等信息
			if binding.AccessToken != session.SessionKey {
				binding.AccessToken = session.SessionKey
				s.bindingRepo.Update(binding)
			}
			return &binding.User, nil
		}
	}

	// 2. 查找是否已有WechatOpenID的用户 (通过User表字段)
	// 这是为了兼容以前可能直接存储在User表的情况
	var existingUser *model.User
	if err := repository.DB.Where("wechat_openid = ?", session.OpenID).First(&existingUser).Error; err == nil {
		fmt.Printf("✅ 找到已存在的用户(通过User表): UserID=%d, OpenID=%s\n", existingUser.ID, session.OpenID)

		// 用户存在但没有绑定记录，重建绑定
		binding = &model.OAuthBinding{
			UserID:         existingUser.ID,
			Provider:       "wechat",
			ProviderUserID: session.OpenID,
			AccessToken:    session.SessionKey,
		}
		if err := s.bindingRepo.Create(binding); err != nil {
			fmt.Printf("⚠️ 重建绑定失败: %v\n", err)
		}

		return existingUser, nil
	}

	// 3. 创建新用户
	user := &model.User{
		WechatOpenID:  stringPtr(session.OpenID),
		WechatUnionID: emptyStringToNil(session.UnionID), // 空字符串转nil，避免唯一索引冲突
		Nickname:      "微信用户",                            // 默认昵称
		Status:        "active",
		Role:          "user",
		EmailVerified: false,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 3.5 生成邀请码
	user.InviteCode = s.generateInviteCode(user.ID)
	if err := s.userRepo.Update(user); err != nil {
		fmt.Printf("⚠️ 生成邀请码失败: %v\n", err)
		// 不阻断流程
	}

	// 4. 创建绑定
	binding = &model.OAuthBinding{
		UserID:         user.ID,
		Provider:       "wechat",
		ProviderUserID: session.OpenID,
		// AccessToken 暂时存储 session_key，虽然它不是access token
		AccessToken: session.SessionKey,
	}
	if err := s.bindingRepo.Create(binding); err != nil {
		return nil, fmt.Errorf("创建绑定失败: %w", err)
	}

	return user, nil
}

// BindWeChat 绑定微信
func (s *OAuthService) BindWeChat(userID uint, session *WeChatSessionResponse) error {
	// 1. 检查该微信号是否已被绑定
	_, err := s.bindingRepo.FindByProviderUserID("wechat", session.OpenID)
	if err == nil {
		return fmt.Errorf("该微信账号已被绑定")
	}

	// 2. 检查用户是否已绑定微信
	// 这里简化处理，允许用户绑定多个微信（通常不建议），或者可以在Handler层检查
	// 严格来说应该检查该用户是否已有wechat类型的绑定

	// 3. 创建绑定
	binding := &model.OAuthBinding{
		UserID:         userID,
		Provider:       "wechat",
		ProviderUserID: session.OpenID,
		AccessToken:    session.SessionKey,
	}

	if err := s.bindingRepo.Create(binding); err != nil {
		return err
	}

	// 4. 更新User表字段 (可选，为了冗余查询方便)
	user, err := s.userRepo.FindByID(userID)
	if err == nil {
		user.WechatOpenID = stringPtr(session.OpenID)
		user.WechatUnionID = emptyStringToNil(session.UnionID)
		s.userRepo.Update(user)
	}

	return nil
}

// BindEmail 微信用户绑定邮箱
// 逻辑：
// 1. 检查邮箱是否已存在
// 2. 如果邮箱账号存在且密码正确，则合并账号（将微信openid写入邮箱账号，删除微信账号）
// 3. 如果邮箱不存在，则为当前微信账号添加邮箱和密码
func (s *OAuthService) BindEmail(wechatUserID uint, email, password string) (uint, error) {
	// 1. 获取当前微信用户
	wechatUser, err := s.userRepo.FindByID(wechatUserID)
	if err != nil {
		return 0, fmt.Errorf("用户不存在")
	}

	// 2. 检查当前用户是否已有邮箱
	if wechatUser.Email != nil && *wechatUser.Email != "" {
		return 0, fmt.Errorf("该账号已绑定邮箱: %s", *wechatUser.Email)
	}

	// 3. 检查邮箱是否已被其他账号使用
	emailUser, err := s.userRepo.FindByEmail(email)

	if err == nil {
		// 邮箱账号已存在，需要验证密码并合并账号
		return s.mergeAccounts(wechatUser, emailUser, password)
	}

	// 4. 邮箱不存在，直接为当前微信账号添加邮箱
	return s.addEmailToWechatUser(wechatUser, email, password)
}

// mergeAccounts 合并微信账号和邮箱账号
// 策略：保留邮箱账号，将微信的openid迁移到邮箱账号，删除微信账号
func (s *OAuthService) mergeAccounts(wechatUser, emailUser *model.User, password string) (uint, error) {
	// 1. 验证邮箱账号密码
	if emailUser.PasswordHash == "" {
		return 0, fmt.Errorf("该邮箱账号未设置密码")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(emailUser.PasswordHash), []byte(password)); err != nil {
		return 0, fmt.Errorf("邮箱密码错误")
	}

	// 2. 将微信的openid/unionid迁移到邮箱账号
	// 为了避免Unique Index冲突，先清除当前微信账号的OpenID
	wechatOpenID := wechatUser.WechatOpenID
	wechatUnionID := wechatUser.WechatUnionID

	wechatUser.WechatOpenID = nil
	wechatUser.WechatUnionID = nil
	if err := s.userRepo.Update(wechatUser); err != nil {
		return 0, fmt.Errorf("清除微信账号OpenID失败: %w", err)
	}

	emailUser.WechatOpenID = wechatOpenID
	emailUser.WechatUnionID = wechatUnionID

	// 3. 合并数据（可选：合并学习时长、积分等）
	// 这里可以根据业务需求决定是否合并数据
	// 例如：emailUser.TotalDurationMinutes += wechatUser.TotalDurationMinutes

	// 3.1 迁移OAuthBinding (关键步骤：确保下次微信登录能找到邮箱账号)
	if wechatOpenID != nil {
		binding, err := s.bindingRepo.FindByProviderUserID("wechat", *wechatOpenID)
		if err == nil {
			binding.UserID = emailUser.ID
			if err := s.bindingRepo.Update(binding); err != nil {
				fmt.Printf("❌ 迁移OAuthBinding失败: %v\n", err)
				return 0, fmt.Errorf("账号绑定更新失败: %w", err)
			}
		} else {
			// 如果找不到绑定记录，这可能是个异常情况，但我们可以尝试创建一个新的绑定
			// 或者如果用户之前从未登录过微信（只是通过API创建了微信用户？不太可能），则忽略
			fmt.Printf("⚠️ 未找到原微信账号的OAuthBinding: OpenID=%s\n", *wechatOpenID)
		}
	}

	// 4. 更新邮箱账号
	if err := s.userRepo.Update(emailUser); err != nil {
		return 0, fmt.Errorf("账号合并失败: %w", err)
	}

	// 5. 迁移微信账号的学习数据到邮箱账号
	if err := s.migrateUserData(wechatUser.ID, emailUser.ID); err != nil {
		fmt.Printf("⚠️ 数据迁移失败: %v\n", err)
		// 不阻断绑定流程
	}

	// 6. 软删除微信账号
	if err := repository.DB.Delete(wechatUser).Error; err != nil {
		fmt.Printf("⚠️ 删除微信账号失败: %v\n", err)
		// 不阻断流程
	}

	return emailUser.ID, nil
}

// addEmailToWechatUser 为微信用户添加邮箱
func (s *OAuthService) addEmailToWechatUser(user *model.User, email, password string) (uint, error) {
	// 1. 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("密码加密失败: %w", err)
	}

	// 2. 更新用户信息
	user.Email = &email
	user.PasswordHash = string(hashedPassword)
	user.EmailVerified = true // 绑定即视为已验证

	if err := s.userRepo.Update(user); err != nil {
		return 0, fmt.Errorf("绑定邮箱失败: %w", err)
	}

	return user.ID, nil
}

// migrateUserData 迁移用户数据（生词本、学习记录等）
func (s *OAuthService) migrateUserData(fromUserID, toUserID uint) error {
	// 这里需要迁移的数据包括：
	// - vp_vocabulary (生词本)
	// - vp_wordbook_progress (单词书进度)
	// - vp_user_daily_stats (每日统计)
	// - vp_vocabulary_reviews (复习记录)
	// - vp_point_records (积分记录)
	// - vp_user_points (用户积分) - 需要合并
	// - vp_wordbook_user_order (用户单词书学习序列)

	db := repository.DB

	// 1. 迁移生词本 (直接更新，假设没有唯一键冲突，或者允许重复)
	db.Model(&model.Vocabulary{}).Where("user_id = ?", fromUserID).Update("user_id", toUserID)

	// 2. 迁移单词书进度 (需要处理冲突)
	var fromProgress []model.WordbookProgress
	db.Where("user_id = ?", fromUserID).Find(&fromProgress)
	for _, fp := range fromProgress {
		var toProgress model.WordbookProgress
		// 检查目标用户是否有同类型的进度
		if err := db.Where("user_id = ? AND word_type = ?", toUserID, fp.WordType).First(&toProgress).Error; err == nil {
			// 目标已存在，保留进度较大的那个
			if fp.CurrentIndex > toProgress.CurrentIndex {
				toProgress.CurrentIndex = fp.CurrentIndex
				toProgress.TotalWords = fp.TotalWords
				toProgress.LastWordID = fp.LastWordID
				toProgress.UpdatedAt = time.Now()
				db.Save(&toProgress)
			}
			// 删除旧记录
			db.Delete(&fp)
		} else {
			// 目标不存在，直接迁移
			fp.UserID = toUserID
			db.Save(&fp)
		}
	}

	// 3. 迁移每日统计 (需要合并数据)
	var fromStats []model.UserDailyStats
	db.Where("user_id = ?", fromUserID).Find(&fromStats)
	for _, fs := range fromStats {
		var toStats model.UserDailyStats
		// 检查目标用户当天是否有统计
		if err := db.Where("user_id = ? AND stat_date = ?", toUserID, fs.StatDate).First(&toStats).Error; err == nil {
			// 目标已存在，累加数据
			toStats.TotalDurationSeconds += fs.TotalDurationSeconds
			toStats.NewWords += fs.NewWords
			toStats.ReviewedWords += fs.ReviewedWords
			toStats.CorrectCount += fs.CorrectCount
			toStats.TotalAttempts += fs.TotalAttempts
			toStats.UpdatedAt = time.Now()
			db.Save(&toStats)
			// 删除旧记录
			db.Delete(&fs)
		} else {
			// 目标不存在，直接迁移
			fs.UserID = toUserID
			db.Save(&fs)
		}
	}

	// 4. 迁移复习记录 (直接更新)
	db.Model(&model.VocabularyReview{}).Where("user_id = ?", fromUserID).Update("user_id", toUserID)

	// 5. 迁移积分记录 (直接更新)
	db.Model(&model.PointRecord{}).Where("user_id = ?", fromUserID).Update("user_id", toUserID)

	// 6. 迁移单词书学习序列 (需要处理冲突)
	var fromOrders []model.WordbookUserOrder
	db.Where("user_id = ?", fromUserID).Find(&fromOrders)
	for _, fo := range fromOrders {
		var toOrder model.WordbookUserOrder
		if err := db.Where("user_id = ? AND word_type = ?", toUserID, fo.WordType).First(&toOrder).Error; err == nil {
			// 目标已存在，保留目标用户的设置（通常保留已有账号的设置更合理）
			// 删除旧记录
			db.Delete(&fo)
		} else {
			// 目标不存在，直接迁移
			fo.UserID = toUserID
			db.Save(&fo)
		}
	}

	// 7. 合并用户积分
	var fromPoints, toPoints model.UserPoints
	db.Where("user_id = ?", fromUserID).First(&fromPoints)
	db.Where("user_id = ?", toUserID).First(&toPoints)

	if fromPoints.ID > 0 {
		if toPoints.ID > 0 {
			// 合并积分和时长
			toPoints.CurrentPoints += fromPoints.CurrentPoints
			toPoints.TotalDurationMinutes += fromPoints.TotalDurationMinutes
			db.Save(&toPoints)
			db.Delete(&fromPoints)
		} else {
			// 目标没有积分记录（罕见），直接迁移
			fromPoints.UserID = toUserID
			db.Save(&fromPoints)
		}
	}

	// 8. 迁移生词本文件夹
	db.Model(&model.VocabularyFolder{}).Where("user_id = ?", fromUserID).Update("user_id", toUserID)

	// 9. 迁移阅读进度 (需要处理冲突)
	var fromReading []model.ReadingProgress
	db.Where("user_id = ?", fromUserID).Find(&fromReading)
	for _, fr := range fromReading {
		var toReading model.ReadingProgress
		if err := db.Where("user_id = ? AND article_id = ?", toUserID, fr.ArticleID).First(&toReading).Error; err == nil {
			// 目标已存在，保留进度较高的那个
			if fr.Progress > toReading.Progress || (fr.IsCompleted && !toReading.IsCompleted) {
				toReading.CurrentTime = fr.CurrentTime
				toReading.Duration = fr.Duration
				toReading.Progress = fr.Progress
				toReading.IsCompleted = fr.IsCompleted
				toReading.ReadCount += fr.ReadCount // 累加阅读次数
				toReading.TotalTime += fr.TotalTime // 累加阅读时间
				toReading.LastReadAt = time.Now()
				db.Save(&toReading)
			} else {
				// 保留目标的进度，但累加阅读次数和时间
				toReading.ReadCount += fr.ReadCount
				toReading.TotalTime += fr.TotalTime
				db.Save(&toReading)
			}
			db.Delete(&fr)
		} else {
			fr.UserID = toUserID
			db.Save(&fr)
		}
	}

	// 10. 迁移默写进度 (需要处理冲突)
	var fromDictation []model.UserDictationProgress
	db.Where("user_id = ?", fromUserID).Find(&fromDictation)
	for _, fd := range fromDictation {
		var toDictation model.UserDictationProgress
		if err := db.Where("user_id = ? AND article_id = ? AND dictation_type = ?", toUserID, fd.ArticleID, fd.DictationType).First(&toDictation).Error; err == nil {
			// 目标已存在，保留得分较高或进度较远的
			if fd.Score > toDictation.Score || (fd.Score == toDictation.Score && fd.CurrentIndex > toDictation.CurrentIndex) {
				toDictation.CurrentIndex = fd.CurrentIndex
				toDictation.Score = fd.Score
				toDictation.Completed = fd.Completed
				toDictation.LastPracticeAt = time.Now()
				db.Save(&toDictation)
			}
			db.Delete(&fd)
		} else {
			fd.UserID = toUserID
			db.Save(&fd)
		}
	}

	return nil
}

// HandleGoogleCallback 处理Google回调
func (s *OAuthService) HandleGoogleCallback(ctx context.Context, code string) (*model.User, error) {
	fmt.Printf("🔵 开始处理Google OAuth回调: code=%s\n", code)

	// 交换授权码获取token
	fmt.Printf("🔵 正在交换授权码获取token...\n")
	token, err := s.googleConfig.Exchange(ctx, code)
	if err != nil {
		fmt.Printf("❌ 交换token失败: %v\n", err)
		return nil, fmt.Errorf("交换token失败: %w", err)
	}
	fmt.Printf("✅ Token交换成功: access_token长度=%d\n", len(token.AccessToken))

	// 获取用户信息
	fmt.Printf("🔵 正在获取Google用户信息...\n")
	userInfo, err := s.getGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		fmt.Printf("❌ 获取Google用户信息失败: %v\n", err)
		return nil, fmt.Errorf("获取Google用户信息失败: %w", err)
	}
	fmt.Printf("✅ 获取Google用户信息成功: id=%s, email=%s, name=%s\n", userInfo.ID, userInfo.Email, userInfo.Name)

	// 查找或创建用户
	fmt.Printf("🔵 正在查找或创建用户...\n")
	user, err := s.findOrCreateGoogleUser(userInfo, token)
	if err != nil {
		fmt.Printf("❌ 处理Google用户失败: %v\n", err)
		return nil, fmt.Errorf("处理Google用户失败: %w", err)
	}
	fmt.Printf("✅ 用户处理成功: user_id=%d, email=%s\n", user.ID, user.Email)

	return user, nil
}

// GitHubUserInfo GitHub用户信息
type GitHubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
}

// getGitHubUserInfo 获取GitHub用户信息
func (s *OAuthService) getGitHubUserInfo(ctx context.Context, accessToken string) (*GitHubUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API错误: %s", string(body))
	}

	var userInfo GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	// 如果email为空，尝试获取邮箱
	if userInfo.Email == "" {
		email, err := s.getGitHubEmail(ctx, accessToken)
		if err == nil {
			userInfo.Email = email
		}
	}

	return &userInfo, nil
}

// getGitHubEmail 获取GitHub邮箱（如果用户信息中没有）
func (s *OAuthService) getGitHubEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取邮箱失败")
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", fmt.Errorf("未找到邮箱")
}

// GoogleUserInfo Google用户信息
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// getGoogleUserInfo 获取Google用户信息
func (s *OAuthService) getGoogleUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google API错误: %s", string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// findOrCreateGitHubUser 查找或创建GitHub用户
func (s *OAuthService) findOrCreateGitHubUser(userInfo *GitHubUserInfo, token *oauth2.Token) (*model.User, error) {
	// 先查找是否已有绑定
	binding, err := s.bindingRepo.FindByProviderUserID("github", fmt.Sprintf("%d", userInfo.ID))
	if err == nil {
		// 已存在，更新token
		binding.AccessToken = token.AccessToken
		if token.RefreshToken != "" {
			binding.RefreshToken = token.RefreshToken
		}
		if !token.Expiry.IsZero() {
			binding.TokenExpiresAt = &token.Expiry
		}
		s.bindingRepo.Update(binding)
		return &binding.User, nil
	}

	// 查找是否已有GitHub ID的用户
	user, err := s.userRepo.FindByGitHubID(fmt.Sprintf("%d", userInfo.ID))
	if err == nil {
		// 创建绑定
		binding := &model.OAuthBinding{
			UserID:           user.ID,
			Provider:         "github",
			ProviderUserID:   fmt.Sprintf("%d", userInfo.ID),
			ProviderUsername: userInfo.Login,
			ProviderEmail:    userInfo.Email,
			AccessToken:      token.AccessToken,
			RefreshToken:     token.RefreshToken,
		}
		if !token.Expiry.IsZero() {
			binding.TokenExpiresAt = &token.Expiry
		}
		s.bindingRepo.Create(binding)
		return user, nil
	}

	// 创建新用户
	now := time.Now()
	user = &model.User{
		GitHubID:        stringPtr(fmt.Sprintf("%d", userInfo.ID)),
		GitHubUsername:  userInfo.Login,
		Email:           emailPtr(userInfo.Email), // 保存邮箱
		EmailVerified:   true,                     // OAuth登录的邮箱视为已验证
		EmailVerifiedAt: &now,
		Nickname:        userInfo.Name,
		Avatar:          userInfo.AvatarURL,
		Bio:             userInfo.Bio,
		Status:          "active",
		Role:            "user",
	}
	if user.Nickname == "" {
		user.Nickname = userInfo.Login
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// 创建绑定
	binding = &model.OAuthBinding{
		UserID:           user.ID,
		Provider:         "github",
		ProviderUserID:   fmt.Sprintf("%d", userInfo.ID),
		ProviderUsername: userInfo.Login,
		ProviderEmail:    userInfo.Email,
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
	}
	if !token.Expiry.IsZero() {
		binding.TokenExpiresAt = &token.Expiry
	}
	s.bindingRepo.Create(binding)

	return user, nil
}

// findOrCreateGoogleUser 查找或创建Google用户
func (s *OAuthService) findOrCreateGoogleUser(userInfo *GoogleUserInfo, token *oauth2.Token) (*model.User, error) {
	// 先查找是否已有绑定
	binding, err := s.bindingRepo.FindByProviderUserID("google", userInfo.ID)
	if err == nil {
		// 已存在，更新token
		binding.AccessToken = token.AccessToken
		if token.RefreshToken != "" {
			binding.RefreshToken = token.RefreshToken
		}
		if !token.Expiry.IsZero() {
			binding.TokenExpiresAt = &token.Expiry
		}
		s.bindingRepo.Update(binding)
		return &binding.User, nil
	}

	// 查找是否已有Google ID的用户
	user, err := s.userRepo.FindByGoogleID(userInfo.ID)
	if err == nil {
		// 创建绑定
		binding := &model.OAuthBinding{
			UserID:         user.ID,
			Provider:       "google",
			ProviderUserID: userInfo.ID,
			ProviderEmail:  userInfo.Email,
			AccessToken:    token.AccessToken,
			RefreshToken:   token.RefreshToken,
		}
		if !token.Expiry.IsZero() {
			binding.TokenExpiresAt = &token.Expiry
		}
		s.bindingRepo.Create(binding)
		return user, nil
	}

	// 创建新用户
	now := time.Now()
	user = &model.User{
		GoogleID:        stringPtr(userInfo.ID),
		GoogleEmail:     userInfo.Email,
		Email:           emailPtr(userInfo.Email), // 保存邮箱
		EmailVerified:   true,                     // OAuth登录的邮箱视为已验证
		EmailVerifiedAt: &now,
		Nickname:        userInfo.Name,
		Avatar:          userInfo.Picture,
		Status:          "active",
		Role:            "user",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// 创建绑定
	binding = &model.OAuthBinding{
		UserID:         user.ID,
		Provider:       "google",
		ProviderUserID: userInfo.ID,
		ProviderEmail:  userInfo.Email,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
	}
	if !token.Expiry.IsZero() {
		binding.TokenExpiresAt = &token.Expiry
	}
	s.bindingRepo.Create(binding)

	return user, nil
}

// generateInviteCode 生成邀请码（基于用户ID加密）
func (s *OAuthService) generateInviteCode(userID uint) string {
	// 使用SHA256生成邀请码
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("user_%d_%d", userID, time.Now().Unix())))

	// 取前16位作为邀请码
	inviteCode := hex.EncodeToString(hash.Sum(nil))[:16]

	// 转大写
	return strings.ToUpper(inviteCode)
}
