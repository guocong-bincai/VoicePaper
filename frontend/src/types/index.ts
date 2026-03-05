export interface User {
  id: number;
  email?: string;
  phone?: string;
  github_id?: string;
  github_username?: string;
  google_id?: string;
  google_email?: string;
  nickname?: string;
  avatar?: string;
  bio?: string;
  role?: string; // 角色：user(普通用户), admin(管理员)
  status: string;
  created_at: string;
  invite_code?: string; // 邀请码
}

export interface Article {
    id: number;
    title: string;
    content?: string; // 从article_url加载的内容
    pic_url?: string; // 文章封面图完整访问URL
    online: string; // '0' | '1' - 是否上线
    category_id?: number;
    publish_date?: string;
    is_daily: boolean;
    audio_url?: string; // 音频完整访问URL
    timeline_url?: string; // 时间轴完整访问URL
    article_url?: string; // 文章URL
    created_at: string;
    updated_at: string;
    sentences?: Sentence[];
    words?: Word[];
    category?: {
        id: number;
        name: string;
    };
}

export interface Sentence {
    id: number;
    article_id: number;
    text: string;
    translation?: string; // 句子中文翻译
    order: number;
    created_at: string;
    updated_at: string;
}

export interface Word {
    id: number;
    article_id: number;
    text: string;
    phonetic?: string;
    meaning?: string;
    example?: string;
    example_translation?: string; // 例句中文翻译
    level: number;
    frequency: number;
    order: number;
    is_key_word: boolean;
}

export interface Category {
    id: number;
    name: string;
    description: string;
    icon?: string;
    sort: number;
    article_count: number;
}

export interface Feedback {
    id?: number;
    type: 'feature' | 'bug' | 'ui' | 'other';
    description: string;
    contact?: string;
    user_id?: number;
    created_at?: string;
}

// ========== 积分系统相关类型 ==========

export interface UserPoints {
    id: number;
    user_id: number;
    total_points: number;           // 累计总积分
    current_points: number;          // 当前可用积分
    level: number;                   // 用户等级
    level_name: string;              // 等级名称
    total_articles_read: number;     // 累计阅读文章数
    total_dictations_completed: number; // 累计完成默写次数
    total_check_ins: number;         // 累计签到天数
    continuous_check_ins: number;    // 当前连续签到天数
    max_continuous_check_ins: number; // 最大连续签到天数
    created_at: string;
    updated_at: string;
}

export interface PointRecord {
    id: number;
    user_id: number;
    points: number;                  // 积分变动数量（正数为增加，负数为扣除）
    type: string;                    // 积分类型
    description: string;             // 积分说明
    article_id?: number;             // 关联文章ID
    dictation_record_id?: number;    // 关联默写记录ID
    check_in_id?: number;            // 关联签到记录ID
    balance_before: number;          // 变动前积分余额
    balance_after: number;           // 变动后积分余额
    created_at: string;
}

export interface CheckInRecord {
    id: number;
    user_id: number;
    check_in_date: string;           // 签到日期
    points_awarded: number;          // 本次签到获得的积分
    continuous_days: number;         // 签到时的连续天数
    is_makeup: boolean;              // 是否补签
    created_at: string;
}

export interface CheckInStatus {
    today_checked_in: boolean;       // 今天是否已签到
    continuous_days: number;         // 连续签到天数
    month_check_in_count: number;    // 本月签到次数
    recent_check_ins: CheckInRecord[]; // 最近签到记录
    next_milestone: number;          // 下一个里程碑天数
    next_milestone_reward: number;   // 下一个里程碑奖励
    check_in_rewards: Record<number, number>; // 签到奖励配置
}

export interface TitleConfig {
    id: number;
    title_key: string;               // 称号唯一标识
    title_name: string;              // 称号名称
    title_icon: string;              // 称号图标
    description: string;             // 称号描述
    category: string;                // 称号分类
    condition_type: string;          // 条件类型
    condition_value: number;         // 条件数值
    condition_description: string;   // 条件说明
    rarity: string;                  // 稀有度
    sort_order: number;              // 排序
    is_active: boolean;              // 是否启用
}

export interface UserTitle {
    id: number;
    user_id: number;
    title_config_id: number;
    awarded_at: string;              // 获得时间
    is_equipped: boolean;            // 是否佩戴
    title_config?: TitleConfig;      // 称号配置详情
}

export interface TitleProgress {
    title_id: number;
    title_key: string;
    title_name: string;
    title_icon: string;
    description: string;
    category: string;
    condition_type: string;
    condition_value: number;
    condition_description: string;
    current_value: number;           // 当前进度值
    progress: number;                // 进度百分比
    owned: boolean;                  // 是否已拥有
    is_equipped: boolean;            // 是否正在穿戴
    rarity: string;                  // 稀有度
}

// ========== 排名系统相关类型 ==========

export interface UserRanking {
    rank: number;                    // 全站排名
    total_users: number;             // 总用户数
    score: number;                   // 综合分数 = 累积时长 * 0.5 + 累计积分 * 0.5
    total_duration: number;          // 累积时长（分钟）
    total_points: number;            // 累计积分
    nearby_users: NearbyUser[];       // 周围排名用户
}

export interface NearbyUser {
    rank: number;
    nickname: string;
    score: number;
    isCurrentUser?: boolean;
}
