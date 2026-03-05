import axios from 'axios';
import type { Article, User, Category, UserPoints, CheckInStatus, TitleConfig, UserTitle, TitleProgress } from '../types';

// 根据环境自动选择API地址
// 生产环境使用相对路径（Caddy会代理/api到后端）
// 开发环境通过Vite代理转发
const API_BASE_URL = '/api/v1';

export const api = axios.create({
    baseURL: API_BASE_URL,
});

// 请求拦截器：添加token
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// 响应拦截器：处理401错误
api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            // token过期或无效，清除token
            localStorage.removeItem('token');
            window.location.href = '/login';
        }
        return Promise.reject(error);
    }
);

export const getArticles = async () => {
    const response = await api.get<Article[]>('/articles');
    return response.data;
};

export const getArticle = async (id: number) => {
    const response = await api.get<Article>(`/articles/${id}`);
    return response.data;
};

export const getArticleTimeline = async (id: number) => {
    const response = await api.get(`/articles/${id}/timeline`);
    return response.data;
};

export const exportArticlePDF = async (articleId: number, title: string) => {
    const response = await api.get(`/articles/${articleId}/export/pdf`, {
        responseType: 'blob',
    });

    // Create blob link to download
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `${title}.pdf`);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
};

export const createArticle = async (title: string, content: string) => {
    const response = await api.post<Article>('/articles', { title, content });
    return response.data;
};

// ============================================
// 分类/合集相关API
// ============================================

// 获取所有分类/合集
export const getCategories = async () => {
    const response = await api.get<Category[]>('/categories');
    return response.data;
};

// 根据分类获取文章列表
export const getArticlesByCategory = async (categoryId: number) => {
    const response = await api.get<Article[]>(`/categories/${categoryId}/articles`);
    return response.data;
};

// ============================================
// 认证相关API
// ============================================

// 发送邮箱验证码
export const sendEmailCode = async (email: string, purpose: string = 'login') => {
    const response = await api.post('/auth/email/send', { email, purpose });
    return response.data;
};

// 邮箱验证码登录
export const loginWithEmail = async (email: string, code: string) => {
    const response = await api.post<{ token: string; user: User; message: string }>('/auth/email/login', {
        email,
        code,
    });
    return response.data;
};

// 验证邮箱验证码（注册流程用，只验证不创建用户）
// BUG修复: 注册流程验证邮箱时不应调用登录接口，应调用专门的验证接口
// 修复策略: 添加新的验证接口，只验证验证码和检查邮箱是否已注册
// 影响范围: frontend/src/services/api.ts
// 修复日期: 2025-12-10
export const verifyEmailCode = async (email: string, code: string) => {
    const response = await api.post<{ message: string; email: string }>('/auth/email/verify', {
        email,
        code,
    });
    return response.data;
};

// 密码登录（支持普通用户和管理员）
export const loginWithPassword = async (email: string, password: string) => {
    const response = await api.post<{ token: string; user: User; message: string }>('/auth/password/login', {
        email,
        password,
    });
    return response.data;
};

// 用户注册
export const register = async (email: string, password: string, code: string, inviteCode?: string) => {
    const response = await api.post<{ token: string; user: User; message: string }>('/auth/register', {
        email,
        password,
        code,
        invite_code: inviteCode || '',
    });
    return response.data;
};

// 重置密码
export const resetPassword = async (email: string, code: string, password: string) => {
    const response = await api.post<{ message: string }>('/auth/password/reset', {
        email,
        code,
        password,
    });
    return response.data;
};

// 验证邀请码
export const verifyInviteCode = async (inviteCode: string) => {
    const response = await api.post<{ valid: boolean; message: string; inviter?: { nickname: string; email: string } }>('/auth/invite/verify', {
        invite_code: inviteCode,
    });
    return response.data;
};

// GitHub OAuth登录（跳转）
export const loginWithGitHub = () => {
    window.location.href = `${API_BASE_URL}/auth/github`;
};

// Google OAuth登录（跳转）
export const loginWithGoogle = () => {
    window.location.href = `${API_BASE_URL}/auth/google`;
};

// 获取当前用户信息
export const getCurrentUser = async () => {
    const response = await api.get<{ user: User }>('/auth/me');
    return response.data.user;
};

// 更新用户资料
export const updateUserProfile = async (data: { nickname?: string; avatar?: string; bio?: string }) => {
    const response = await api.put<{ user: User; message: string }>('/auth/profile', data);
    return response.data.user;
};

// 上传头像
export const uploadAvatar = async (file: File): Promise<{ avatar_url: string; user: User }> => {
    const formData = new FormData();
    formData.append('avatar', file);

    const response = await api.post<{ avatar_url: string; user: User; message: string }>('/auth/avatar/upload', formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
    return response.data;
};

// 登出
export const logout = async () => {
    await api.post('/auth/logout');
};

// ============================================
// 默写练习相关API
// ============================================

// 获取文章的重点单词
export const getWords = async (articleId: number) => {
    const response = await api.get<import('../types').Word[]>(`/articles/${articleId}/words`);
    return response.data;
};

// 获取文章的句子
export const getSentences = async (articleId: number) => {
    const response = await api.get<import('../types').Sentence[]>(`/articles/${articleId}/sentences`);
    return response.data;
};

// 保存默写记录
export const saveDictationRecord = async (data: {
    article_id: number;
    dictation_type: 'word' | 'sentence';
    word_id?: number;
    sentence_id?: number;
    user_answer: string;
    is_correct: boolean;
    score: number;
    attempt_count: number;
    time_spent: number;
}) => {
    const response = await api.post('/dictation/record', data);
    return response.data;
};

// 获取默写统计
export const getDictationStatistics = async (articleId: number) => {
    const response = await api.get(`/dictation/statistics/${articleId}`);
    return response.data;
};

// 获取默写进度
export const getDictationProgress = async (articleId: number, dictationType: 'word' | 'sentence' = 'word') => {
    const response = await api.get(`/dictation/progress/${articleId}?type=${dictationType}`);
    return response.data;
};

// 保存默写进度
export const saveDictationProgress = async (data: {
    article_id: number;
    dictation_type: 'word' | 'sentence';
    current_index: number;
    total_items: number;
    score: number;
    completed: boolean;
}) => {
    const response = await api.post('/dictation/progress', data);
    return response.data;
};

// 重置默写进度
export const resetDictationProgress = async (articleId: number, dictationType: 'word' | 'sentence' = 'word') => {
    const response = await api.delete(`/dictation/progress/${articleId}?type=${dictationType}`);
    return response.data;
};

// ============================================
// 问题反馈相关API
// ============================================

// 提交反馈
export const submitFeedback = async (data: {
    type: 'feature' | 'bug' | 'ui' | 'other';
    description: string;
    contact?: string;
}) => {
    const response = await api.post('/feedback', data);
    return response.data;
};

// ============================================
// 积分系统相关API
// ============================================

// 获取我的积分信息
export const getMyPoints = async (): Promise<UserPoints> => {
    const response = await api.get('/points/me');
    return response.data;
};

// 获取积分记录
export const getPointRecords = async (page: number = 1, pageSize: number = 20) => {
    const response = await api.get('/points/records', {
        params: { page, page_size: pageSize }
    });
    return response.data;
};

// 获取积分统计
export const getPointStatistics = async () => {
    const response = await api.get('/points/statistics');
    return response.data;
};

// 奖励阅读积分
export const awardReadPoints = async (data: {
    article_id: number;
    article_title: string;
    read_duration?: number;
}) => {
    const response = await api.post('/points/award/read', data);
    return response.data;
};

// ============================================
// 排名相关API
// ============================================

// 获取我的排名
export const getMyRanking = async () => {
    const response = await api.get('/ranking/me');
    return response.data;
};

// 获取全球排名
export const getGlobalRanking = async (page: number = 1, pageSize: number = 20, sortBy: string = 'score') => {
    const response = await api.get('/ranking/global', {
        params: { page, page_size: pageSize, sort_by: sortBy }
    });
    return response.data;
};

// ============================================
// 年终总结相关API
// ============================================

export interface YearEndReport {
    articles_read: number;
    consecutive_days: number;
    words_learned: number;
    total_duration: number; // minutes
    total_points: number;
    global_rank: number;
    nickname: string;
    avatar: string;
}

// 获取年终总结数据
export const getYearEndReport = async (): Promise<YearEndReport> => {
    const response = await api.get('/report/year-end');
    return response.data;
};

// ============================================
// 学习时长相关API
// ============================================

// 同步阅读时长到排名系统
export const syncReadingDuration = async (totalReadingSeconds: number) => {
    const response = await api.post('/duration/sync', {
        total_reading_seconds: totalReadingSeconds
    });
    return response.data;
};

// 获取用户累积时长
export const getUserDuration = async () => {
    const response = await api.get('/duration/me');
    return response.data;
};

// ============================================
// 签到相关API
// ============================================

// 每日签到
export const checkIn = async () => {
    const response = await api.post('/check-in');
    return response.data;
};

// 获取签到状态
export const getCheckInStatus = async (): Promise<CheckInStatus> => {
    const response = await api.get('/check-in/status');
    return response.data;
};

// 获取签到日历
export const getCheckInCalendar = async (year?: number, month?: number) => {
    const response = await api.get('/check-in/calendar', {
        params: { year, month }
    });
    return response.data;
};

// 获取签到排行榜
export const getCheckInRanking = async (limit: number = 20, sortBy: 'total' | 'continuous' = 'total') => {
    const response = await api.get('/check-in/ranking', {
        params: { limit, sort_by: sortBy }
    });
    return response.data;
};

// 补签卡信息类型
export interface MakeupCardInfo {
    makeup_cards: number;      // 拥有的补签卡数量
    current_points: number;    // 当前积分
    card_price: number;        // 卡价格
    can_buy: boolean;          // 是否可以购买
    makeup_dates: string[];    // 可补签的日期列表
    max_makeup_days: number;   // 最大可补签天数
}

// 获取补签卡信息
export const getMakeupCardInfo = async (): Promise<MakeupCardInfo> => {
    const response = await api.get('/check-in/makeup-card');
    return response.data;
};

// 购买补签卡
export const buyMakeupCard = async (count: number = 1): Promise<{
    success: boolean;
    message: string;
    cards_bought: number;
    points_spent: number;
    makeup_cards: number;
    current_points: number;
}> => {
    const response = await api.post('/check-in/makeup-card/buy', { count });
    return response.data;
};

// 使用补签卡补签
export const useMakeupCard = async (date: string): Promise<{
    success: boolean;
    message: string;
    makeup_date: string;
    makeup_cards: number;
    total_check_ins: number;
    continuous_days: number;
}> => {
    const response = await api.post('/check-in/makeup-card/use', { date });
    return response.data;
};

// ============================================
// 称号相关API
// ============================================

// 获取所有称号
export const getAllTitles = async (): Promise<TitleConfig[]> => {
    const response = await api.get('/titles/all');
    return response.data;
};

// 获取我的称号
export const getMyTitles = async (): Promise<UserTitle[]> => {
    const response = await api.get('/titles/me');
    return response.data;
};

// 获取称号进度
export const getTitleProgress = async (): Promise<TitleProgress[]> => {
    const response = await api.get('/titles/progress');
    return response.data;
};

// 获取当前佩戴的称号
export const getEquippedTitle = async (): Promise<UserTitle> => {
    const response = await api.get('/titles/equipped');
    return response.data;
};

// 佩戴称号
export const equipTitle = async (titleConfigId: number) => {
    const response = await api.post('/titles/equip', { title_config_id: titleConfigId });
    return response.data;
};

// 取消佩戴称号
export const unequipTitle = async () => {
    const response = await api.post('/titles/unequip');
    return response.data;
};

// 根据分类获取称号
export const getTitlesByCategory = async (category: string): Promise<TitleConfig[]> => {
    const response = await api.get(`/titles/category/${category}`);
    return response.data;
};

// ============================================
// 阅读进度相关API
// ============================================

// 阅读进度类型
export interface ReadingProgress {
    id?: number;
    user_id?: number;
    article_id: number;
    current_time: number;
    duration: number;
    progress: number;
    is_completed: boolean;
    read_count: number;
    total_time?: number;
    last_read_at?: string;
    article?: import('../types').Article;
}

// 获取某篇文章的阅读进度
export const getReadingProgress = async (articleId: number): Promise<ReadingProgress> => {
    const response = await api.get(`/reading-progress/${articleId}`);
    return response.data;
};

// 保存阅读进度
export const saveReadingProgress = async (data: {
    article_id: number;
    current_time: number;
    duration: number;
}): Promise<{ message: string; progress: ReadingProgress; points_earned?: number; points_message?: string }> => {
    const response = await api.post('/reading-progress', data);
    return response.data;
};

// 快速保存阅读进度（用于高频更新）
export const quickSaveReadingProgress = async (data: {
    article_id: number;
    current_time: number;
    duration: number;
}): Promise<{ ok: boolean }> => {
    const response = await api.post('/reading-progress/quick', data);
    return response.data;
};

// 获取阅读历史
export const getReadingHistory = async (page: number = 1, pageSize: number = 20): Promise<{
    history: ReadingProgress[];
    total: number;
    page: number;
    page_size: number;
}> => {
    const response = await api.get('/reading-progress/history', {
        params: { page, page_size: pageSize }
    });
    return response.data;
};

// 获取最后阅读的文章
export const getLastReadArticle = async (): Promise<ReadingProgress | { message: string }> => {
    const response = await api.get('/reading-progress/last');
    return response.data;
};

// 获取阅读统计
export const getReadingStats = async (): Promise<{
    total_reading_time: number;  // 累计阅读时长（秒）
    total_articles: number;      // 阅读文章数
    completed_count: number;     // 完成阅读的文章数
}> => {
    const response = await api.get('/reading-progress/stats');
    return response.data;
};

// ============================================
// 生词本相关API
// ============================================

// 生词类型
export interface Vocabulary {
    id: number;
    user_id: number;
    article_id?: number;
    sentence_id?: number;
    type: 'word' | 'phrase' | 'sentence';
    content: string;
    phonetic?: string;
    meaning?: string;
    example?: string;
    example_translation?: string;
    context?: string;
    note?: string;
    morphemes?: string;  // 词根词缀分析结果，JSON格式
    mastery_level: number;  // 0-5
    ease_factor: number;
    interval_days: number;
    repetitions: number;
    next_review_at?: string;
    last_review_at?: string;
    review_count: number;
    correct_count: number;
    wrong_count: number;
    tags?: string;
    is_starred: boolean;
    created_at: string;
    updated_at: string;
}

// 生词统计
export interface VocabularyStats {
    total: number;
    new: number;
    learning: number;
    reviewing: number;
    mastered: number;
    today_review: number;
    starred: number;
}

// 生词文件夹
export interface VocabularyFolder {
    id: number;
    user_id: number;
    name: string;
    description?: string;
    color: string;
    icon?: string;
    sort_order: number;
    created_at: string;
}

// 每日统计
export interface VocabularyDailyStats {
    id: number;
    user_id: number;
    stat_date: string;
    new_words: number;
    reviewed_words: number;
    mastered_words: number;
    review_time_seconds: number;
    correct_rate?: number;
}

// 添加生词
export const addVocabulary = async (data: {
    article_id?: number;
    sentence_id?: number;
    type: 'word' | 'phrase' | 'sentence';
    content: string;
    phonetic?: string;
    meaning?: string;
    example?: string;
    example_translation?: string;
    context?: string;
    note?: string;
    morphemes?: string; // 词根词缀分析结果，JSON格式
    source?: 'article' | 'manual'; // 来源：article=从文章添加，manual=手动添加（默认）
}): Promise<{ message: string; data: Vocabulary }> => {
    const response = await api.post('/vocabulary', data);
    return response.data;
};

// 获取生词列表
export const getVocabularyList = async (params?: {
    type?: string;
    mastery_level?: number;
    is_starred?: boolean;
    keyword?: string;
    article_id?: number;
    order_by?: string;
    limit?: number;
    offset?: number;
    with_article?: boolean;
}): Promise<{ data: Vocabulary[]; total: number; limit: number; offset: number }> => {
    const response = await api.get('/vocabulary', { params });
    return response.data;
};

// 获取单个生词
export const getVocabulary = async (id: number): Promise<Vocabulary> => {
    const response = await api.get(`/vocabulary/${id}`);
    return response.data;
};

// 更新生词
export const updateVocabulary = async (id: number, data: {
    phonetic?: string;
    meaning?: string;
    example?: string;
    example_translation?: string;
    context?: string;
    note?: string;
    tags?: string;
    is_starred?: boolean;
}): Promise<{ message: string; data: Vocabulary }> => {
    const response = await api.put(`/vocabulary/${id}`, data);
    return response.data;
};

// 删除生词
export const deleteVocabulary = async (id: number): Promise<{ message: string }> => {
    const response = await api.delete(`/vocabulary/${id}`);
    return response.data;
};

// 切换标星
export const toggleVocabularyStar = async (id: number): Promise<{ message: string; is_starred: boolean }> => {
    const response = await api.post(`/vocabulary/${id}/star`);
    return response.data;
};

// 批量删除
export const batchDeleteVocabulary = async (ids: number[]): Promise<{ message: string }> => {
    const response = await api.delete('/vocabulary/batch', { data: { ids } });
    return response.data;
};

// 获取生词统计
export const getVocabularyStats = async (): Promise<VocabularyStats> => {
    const response = await api.get('/vocabulary/stats');
    return response.data;
};

// 获取今日待复习列表
export const getTodayReviewList = async (limit?: number): Promise<{ data: Vocabulary[]; count: number }> => {
    const response = await api.get('/vocabulary/review/today', { params: { limit } });
    return response.data;
};

// 提交复习结果
export const submitVocabularyReview = async (id: number, data: {
    quality: number;  // 0-5
    review_type?: 'card' | 'spell' | 'choice' | 'dictation';
    response_time_ms?: number;
    user_answer?: string;
}): Promise<{ message: string; data: Vocabulary; points_earned: number; total_points: number }> => {
    const response = await api.post(`/vocabulary/${id}/review`, data);
    return response.data;
};

// 获取每日学习统计
export const getVocabularyDailyStats = async (days?: number): Promise<{ data: VocabularyDailyStats[] }> => {
    const response = await api.get('/vocabulary/stats/daily', { params: { days } });
    return response.data;
};

// 获取文件夹列表
export const getVocabularyFolders = async (): Promise<{ data: VocabularyFolder[] }> => {
    const response = await api.get('/vocabulary/folders');
    return response.data;
};

// 创建文件夹
export const createVocabularyFolder = async (data: {
    name: string;
    description?: string;
    color?: string;
    icon?: string;
}): Promise<{ message: string; data: VocabularyFolder }> => {
    const response = await api.post('/vocabulary/folders', data);
    return response.data;
};

// 更新文件夹
export const updateVocabularyFolder = async (id: number, data: {
    name: string;
    description?: string;
    color?: string;
    icon?: string;
}): Promise<{ message: string; data: VocabularyFolder }> => {
    const response = await api.put(`/vocabulary/folders/${id}`, data);
    return response.data;
};

// 删除文件夹
export const deleteVocabularyFolder = async (id: number): Promise<{ message: string }> => {
    const response = await api.delete(`/vocabulary/folders/${id}`);
    return response.data;
};

// 获取文件夹中的生词
export const getFolderVocabulary = async (folderId: number, params?: {
    limit?: number;
    offset?: number;
}): Promise<{ data: Vocabulary[]; total: number; limit: number; offset: number }> => {
    const response = await api.get(`/vocabulary/folders/${folderId}/items`, { params });
    return response.data;
};

// 添加生词到文件夹
export const addVocabularyToFolder = async (folderId: number, vocabularyId: number): Promise<{ message: string }> => {
    const response = await api.post(`/vocabulary/folders/${folderId}/items`, { vocabulary_id: vocabularyId });
    return response.data;
};

// 从文件夹移除生词
export const removeVocabularyFromFolder = async (folderId: number, vocabularyId: number): Promise<{ message: string }> => {
    const response = await api.delete(`/vocabulary/folders/${folderId}/items/${vocabularyId}`);
    return response.data;
};

// ============================================
// 单词书相关API
// ============================================

export interface WordbookInfo {
    type: string;
    name: string;
    description: string;
    icon: string;
    cover_url?: string;
    word_count: number;
    hotness: number;
    study_count: number;
    category: string;
}

export interface WordbookWord {
    id: number;
    word: string;
    phonetic: string;
    word_type: string;
    meaning: string;
    meaning_analysis: string;
    example: string;
    example_translation: string;
    examples_json: string;
    phrases: string;
    root: string;
    root_analysis: string;
    affix: string;
    affix_analysis: string;
    etymology: string;
    cultural_background: string;
    word_forms: string;
    memory_tips: string;
    story_en: string;
    story_cn: string;
    draw_explain: string;
    draw_prompt: string;
    image_url: string;
}

export interface WordbookProgress {
    current_index: number;
    total_words: number;
    last_word_id: number;
}

// 获取所有单词书
export const getWordbooks = async (): Promise<WordbookInfo[]> => {
    const response = await api.get('/wordbooks');
    return response.data;
};

// 获取单词书中的单词
export const getWordbookWords = async (type: string, page: number = 1, pageSize: number = 20): Promise<{ words: WordbookWord[], total: number, page: number }> => {
    const response = await api.get(`/wordbooks/${type}/words`, {
        params: { page, page_size: pageSize }
    });
    return response.data;
};

// 获取学习进度
export const getWordbookProgress = async (type: string): Promise<WordbookProgress> => {
    const response = await api.get(`/wordbooks/${type}/progress`);
    return response.data;
};

// 保存学习进度
export const saveWordbookProgress = async (data: {
    word_type: string;
    current_index: number;
    last_word_id: number;
    total_words: number;
}) => {
    const response = await api.post('/wordbooks/progress', data);
    return response.data;
};

// 将单词书中的单词导入到生词本
export const importWordToVocabulary = async (wordbookId: number, quality: 'fuzzy' | 'forget'): Promise<{ message: string; data: Vocabulary }> => {
    const response = await api.post(`/wordbooks/word/${wordbookId}/import`, { quality });
    return response.data;
};

// 单词书学习并获取积分
// quality: 'forget'(不认识+1分), 'fuzzy'(模糊+2分), 'know'(认识+3分)
export const studyWordbookWord = async (
    wordbookId: number,
    quality: 'forget' | 'fuzzy' | 'know',
    wordText: string
): Promise<{ message: string; points_earned: number; total_points: number }> => {
    const response = await api.post(`/wordbooks/word/${wordbookId}/study`, {
        quality,
        word_text: wordText
    });
    return response.data;
};

// ==================== 单词书顺序/乱序学习 API ====================

export interface WordbookOrderStatus {
    is_random: boolean;
    current_index: number;
    total_words: number;
}

export interface WordbookOrderedWordsResponse {
    words: WordbookWord[];
    total: number;
    page: number;
    current_index: number;
    is_random: boolean;
}

// 获取用户学习序列状态
export const getWordbookOrderStatus = async (type: string): Promise<WordbookOrderStatus> => {
    const response = await api.get(`/wordbooks/seq/${type}/order`);
    return response.data;
};

// 切换顺序/乱序模式
export const switchWordbookOrderMode = async (type: string, isRandom: boolean): Promise<WordbookOrderStatus & { message: string }> => {
    const response = await api.post(`/wordbooks/seq/${type}/order`, { is_random: isRandom });
    return response.data;
};

// 更新学习进度位置
export const updateWordbookOrderIndex = async (type: string, currentIndex: number): Promise<{ message: string; current_index: number }> => {
    const response = await api.put(`/wordbooks/seq/${type}/order/index`, { current_index: currentIndex });
    return response.data;
};

// 按用户序列获取单词（支持顺序/乱序）
export const getWordbookWordsOrdered = async (type: string, page: number = 1, pageSize: number = 20): Promise<WordbookOrderedWordsResponse> => {
    const response = await api.get(`/wordbooks/order/${type}/words`, {
        params: { page, page_size: pageSize }
    });
    return response.data;
};

// ==================== 翻译/词典 API ====================

// 词典API响应类型
export interface DictionaryPhonetic {
    text?: string;
    audio?: string;
}

export interface DictionaryDefinition {
    definition: string;
    example?: string;
    synonyms?: string[];
    antonyms?: string[];
}

export interface DictionaryMeaning {
    partOfSpeech: string;
    definitions: DictionaryDefinition[];
}

export interface DictionaryResult {
    word: string;
    phonetic?: string;
    phonetics?: DictionaryPhonetic[];
    meanings?: DictionaryMeaning[];
    origin?: string;
}

// 查询单词释义（使用免费的 Free Dictionary API）
export const lookupWord = async (word: string): Promise<DictionaryResult | null> => {
    try {
        const response = await axios.get<DictionaryResult[]>(
            `https://api.dictionaryapi.dev/api/v2/entries/en/${encodeURIComponent(word.trim().toLowerCase())}`
        );
        if (response.data && response.data.length > 0) {
            return response.data[0];
        }
        return null;
    } catch (error) {
        console.warn('词典API查询失败:', error);
        return null;
    }
};

// 获取单词的音标
export const getPhonetic = (result: DictionaryResult): string => {
    if (result.phonetic) {
        return result.phonetic;
    }
    if (result.phonetics) {
        const phoneticWithText = result.phonetics.find(p => p.text);
        return phoneticWithText?.text || '';
    }
    return '';
};

// 获取单词的发音音频URL
export const getPronunciationAudio = (result: DictionaryResult): string => {
    if (result.phonetics) {
        const phoneticWithAudio = result.phonetics.find(p => p.audio && p.audio.length > 0);
        return phoneticWithAudio?.audio || '';
    }
    return '';
};

// 获取单词的主要释义
export const getPrimaryMeaning = (result: DictionaryResult): string => {
    if (result.meanings && result.meanings.length > 0) {
        const firstMeaning = result.meanings[0];
        if (firstMeaning.definitions && firstMeaning.definitions.length > 0) {
            const partOfSpeech = firstMeaning.partOfSpeech;
            const definition = firstMeaning.definitions[0].definition;
            return `(${partOfSpeech}) ${definition}`;
        }
    }
    return '';
};

// 播放单词发音
export const playPronunciation = (audioUrl: string): void => {
    if (audioUrl) {
        const audio = new Audio(audioUrl);
        audio.play().catch(err => console.warn('播放发音失败:', err));
    }
};

// ==================== 翻译 API（中文） ====================

// 清理翻译结果中的HTML标签（如 <g id="...">...</g>）
const cleanTranslationTags = (text: string): string => {
    if (!text) return '';
    // 移除所有 <g id="..."> 和 </g> 标签
    return text.replace(/<g[^>]*>/gi, '').replace(/<\/g>/gi, '').trim();
};

// 使用 MyMemory 免费翻译API（每天5000字符免费额度）
export const translateToChineseAPI = async (text: string): Promise<string> => {
    try {
        const response = await axios.get(
            `https://api.mymemory.translated.net/get?q=${encodeURIComponent(text)}&langpair=en|zh-CN`
        );
        if (response.data && response.data.responseData) {
            const rawText = response.data.responseData.translatedText || '';
            // 清理HTML标签后返回
            return cleanTranslationTags(rawText);
        }
        return '';
    } catch (error) {
        console.warn('翻译API调用失败:', error);
        return '';
    }
};
