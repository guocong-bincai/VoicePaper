-- VoicePaper Database Schema
-- Combined and Cleaned for Open Source Initialization

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. Core Content Tables
-- ----------------------------

-- Categories
CREATE TABLE IF NOT EXISTS `vp_categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分类名称',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '分类描述',
  `icon` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分类图标',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序',
  `is_active` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_categories_name` (`name`),
  KEY `idx_categories_deleted_at` (`deleted_at`),
  KEY `idx_categories_sort` (`sort`),
  KEY `idx_categories_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章分类表';

-- Articles
CREATE TABLE IF NOT EXISTS `vp_articles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文章标题',
  `online` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '是否上线0:否，1:是',
  `category_id` bigint unsigned DEFAULT NULL COMMENT '分类ID',
  `publish_date` date DEFAULT NULL COMMENT '发布日期（用于每日文章）',
  `is_daily` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否为每日文章',
  `audio_url` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '音频完整访问URL',
  `timeline_url` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '时间轴完整访问URL',
  `article_url` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文章URL',
  `original_article_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '英文原文 OSS 地址',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_articles_deleted_at` (`deleted_at`),
  KEY `idx_articles_category_id` (`category_id`),
  KEY `idx_articles_publish_date` (`publish_date`),
  KEY `idx_articles_is_daily` (`is_daily`),
  KEY `idx_articles_audio_url` (`audio_url`),
  KEY `idx_articles_timeline_url` (`timeline_url`),
  KEY `idx_articles_article_url` (`article_url`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章表';

-- Sentences
CREATE TABLE IF NOT EXISTS `vp_sentences` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `article_id` bigint unsigned NOT NULL COMMENT '关联articles.id',
  `text` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '句子文本',
  `translation` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '句子中文翻译',
  `order` int NOT NULL COMMENT '排序',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_sentences_article_id` (`article_id`),
  KEY `idx_sentences_order` (`order`),
  KEY `idx_sentences_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='句子表（用于听写和高亮）';

-- Words
CREATE TABLE IF NOT EXISTS `vp_words` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `article_id` bigint unsigned NOT NULL COMMENT '关联articles.id',
  `text` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '单词文本',
  `phonetic` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '音标，如 /wɜːld/',
  `meaning` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '中文释义',
  `example` text COLLATE utf8mb4_unicode_ci COMMENT '例句（4级/6级/考研真题）',
  `example_translation` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '例句中文翻译',
  `level` int NOT NULL DEFAULT '0' COMMENT '难度等级 0-5',
  `frequency` int NOT NULL DEFAULT '0' COMMENT '词频（可用于推荐）',
  `order` int NOT NULL COMMENT '排序（在文章中的出现顺序）',
  `is_key_word` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否为关键词（补充字段，匹配原有索引）',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_words_article_id` (`article_id`),
  KEY `idx_words_text` (`text`),
  KEY `idx_words_is_key_word` (`is_key_word`),
  KEY `idx_words_level` (`level`),
  KEY `idx_words_frequency` (`frequency`),
  KEY `idx_words_order` (`order`),
  KEY `idx_words_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='单词表（用于听写练习）';

-- ----------------------------
-- 2. User & Auth Tables
-- ----------------------------

-- Users
CREATE TABLE IF NOT EXISTS `vp_users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `nickname` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '昵称',
  `avatar` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '头像URL',
  `bio` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '个人简介',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机号（加密存储）',
  `phone_verified` tinyint(1) NOT NULL DEFAULT '0' COMMENT '手机号是否已验证',
  `phone_verified_at` datetime(3) DEFAULT NULL COMMENT '手机号验证时间',
  `email` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '邮箱地址',
  `email_verified` tinyint(1) NOT NULL DEFAULT '0' COMMENT '邮箱是否已验证',
  `email_verified_at` datetime(3) DEFAULT NULL COMMENT '邮箱验证时间',
  `github_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'GitHub用户ID',
  `github_username` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'GitHub用户名',
  `google_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Google用户ID',
  `google_email` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Google邮箱',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active' COMMENT '状态：active/inactive/banned',
  `role` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'user' COMMENT '用户角色：user(普通用户), admin(管理员)',
  `password_hash` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '密码哈希（bcrypt加密，仅超管账户使用）',
  `invite_code` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户邀请码（基于用户ID加密生成）',
  `invited_by_id` bigint unsigned DEFAULT NULL COMMENT '邀请人ID（谁邀请的当前用户）',
  `last_login_at` datetime(3) DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '最后登录IP',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_phone` (`phone`),
  UNIQUE KEY `idx_users_email` (`email`),
  UNIQUE KEY `idx_users_github_id` (`github_id`),
  UNIQUE KEY `idx_users_google_id` (`google_id`),
  UNIQUE KEY `idx_users_invite_code` (`invite_code`),
  KEY `idx_users_deleted_at` (`deleted_at`),
  KEY `idx_users_status` (`status`),
  KEY `idx_users_role` (`role`),
  KEY `idx_users_invited_by_id` (`invited_by_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- User Sessions
CREATE TABLE IF NOT EXISTS `vp_user_sessions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '关联users.id',
  `session_token` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '会话令牌',
  `ip_address` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'IP地址',
  `user_agent` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'User Agent',
  `device_info` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '设备信息',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `expires_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_sessions_token` (`session_token`),
  KEY `idx_user_sessions_user_id` (`user_id`),
  KEY `idx_user_sessions_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';

-- OAuth Bindings
CREATE TABLE IF NOT EXISTS `vp_oauth_bindings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '关联users.id',
  `provider` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '提供商：github/google',
  `provider_user_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '第三方用户ID',
  `provider_username` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '第三方用户名',
  `provider_email` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '第三方邮箱',
  `access_token` text COLLATE utf8mb4_unicode_ci COMMENT '访问令牌（加密存储）',
  `refresh_token` text COLLATE utf8mb4_unicode_ci COMMENT '刷新令牌（加密存储）',
  `token_expires_at` datetime(3) DEFAULT NULL COMMENT '令牌过期时间',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_oauth_bindings_provider_user` (`provider`,`provider_user_id`),
  KEY `idx_oauth_bindings_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OAuth绑定表';

-- Verification Codes
CREATE TABLE IF NOT EXISTS `vp_verification_codes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '类型：phone/email',
  `receiver` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '接收方：手机号或邮箱',
  `code` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '验证码（6位数字）',
  `expires_at` datetime(3) NOT NULL COMMENT '过期时间',
  `used` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否已使用',
  `used_at` datetime(3) DEFAULT NULL COMMENT '使用时间',
  `purpose` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'login' COMMENT '用途：login/register/reset_password',
  `ip_address` varchar(45) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '请求IP',
  `user_agent` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'User Agent',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_verification_codes_receiver` (`receiver`,`code_type`),
  KEY `idx_verification_codes_code` (`code`),
  KEY `idx_verification_codes_expires_at` (`expires_at`),
  KEY `idx_verification_codes_used` (`used`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='验证码表';

-- ----------------------------
-- 3. Progress & Learning Records
-- ----------------------------

-- Dictation Records
CREATE TABLE IF NOT EXISTS `vp_dictation_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL COMMENT '关联users.id（允许NULL，支持未登录用户）',
  `article_id` bigint unsigned NOT NULL COMMENT '关联articles.id',
  `dictation_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '默写类型：word/sentence',
  `sentence_id` bigint unsigned DEFAULT NULL COMMENT '如果是句子默写，关联sentences.id',
  `word_id` bigint unsigned DEFAULT NULL COMMENT '如果是单词默写，关联words.id',
  `user_answer` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户输入的答案',
  `is_correct` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否正确',
  `score` int NOT NULL DEFAULT '0' COMMENT '得分（0-100）',
  `attempt_count` int NOT NULL DEFAULT '1' COMMENT '尝试次数',
  `time_spent` int NOT NULL DEFAULT '0' COMMENT '花费时间（秒）',
  `last_attempt` datetime(3) NOT NULL COMMENT '最后尝试时间',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_dictation_records_user_id` (`user_id`),
  KEY `idx_dictation_records_article_id` (`article_id`),
  KEY `idx_dictation_records_dictation_type` (`dictation_type`),
  KEY `idx_dictation_records_sentence_id` (`sentence_id`),
  KEY `idx_dictation_records_word_id` (`word_id`),
  KEY `idx_dictation_records_is_correct` (`is_correct`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='默写记录表';

-- User Dictation Progress
CREATE TABLE IF NOT EXISTS `vp_user_dictation_progress` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '关联users.id',
  `article_id` bigint unsigned NOT NULL COMMENT '关联articles.id',
  `dictation_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '默写类型：word/sentence',
  `current_index` int NOT NULL DEFAULT '0' COMMENT '当前学习到的索引位置（从0开始）',
  `total_items` int NOT NULL DEFAULT '0' COMMENT '总共的单词/句子数量',
  `score` int NOT NULL DEFAULT '0' COMMENT '当前得分（正确的数量）',
  `completed` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否已完成本轮练习',
  `last_practice_at` datetime(3) NOT NULL COMMENT '最后练习时间',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_progress_user_article_type` (`user_id`,`article_id`,`dictation_type`),
  KEY `idx_progress_user_id` (`user_id`),
  KEY `idx_progress_article_id` (`article_id`),
  KEY `idx_progress_completed` (`completed`),
  KEY `idx_progress_last_practice` (`last_practice_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户默写练习进度表';

-- ----------------------------
-- 4. Points & Gamification System
-- ----------------------------

-- User Points
CREATE TABLE IF NOT EXISTS `vp_user_points` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID，关联vp_users.id',
  `total_points` INT NOT NULL DEFAULT 0 COMMENT '累计总积分（历史所有获得的积分）',
  `current_points` INT NOT NULL DEFAULT 0 COMMENT '当前可用积分（可用于兑换等）',
  `level` INT NOT NULL DEFAULT 1 COMMENT '用户等级（根据总积分计算）',
  `level_name` VARCHAR(50) DEFAULT '初学者' COMMENT '等级名称',
  `total_articles_read` INT NOT NULL DEFAULT 0 COMMENT '累计阅读文章数',
  `total_dictations_completed` INT NOT NULL DEFAULT 0 COMMENT '累计完成默写次数',
  `total_check_ins` INT NOT NULL DEFAULT 0 COMMENT '累计签到天数',
  `continuous_check_ins` INT NOT NULL DEFAULT 0 COMMENT '当前连续签到天数',
  `max_continuous_check_ins` INT NOT NULL DEFAULT 0 COMMENT '最大连续签到天数',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`),
  KEY `idx_level` (`level`),
  KEY `idx_total_points` (`total_points`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户积分表';

-- Point Records
CREATE TABLE IF NOT EXISTS `vp_point_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID，关联vp_users.id',
  `points` INT NOT NULL COMMENT '积分变动数量（正数为增加，负数为扣除）',
  `type` VARCHAR(50) NOT NULL COMMENT '积分类型',
  `description` VARCHAR(255) DEFAULT NULL COMMENT '积分说明',
  `article_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联文章ID（如果是阅读相关）',
  `dictation_record_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联默写记录ID',
  `check_in_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '关联签到记录ID',
  `balance_before` INT NOT NULL DEFAULT 0 COMMENT '变动前积分余额',
  `balance_after` INT NOT NULL DEFAULT 0 COMMENT '变动后积分余额',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_type` (`type`),
  KEY `idx_article_id` (`article_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分记录表';

-- User Check Ins
CREATE TABLE IF NOT EXISTS `vp_user_check_ins` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID，关联vp_users.id',
  `check_in_date` DATE NOT NULL COMMENT '签到日期',
  `points_awarded` INT NOT NULL DEFAULT 0 COMMENT '本次签到获得的积分',
  `continuous_days` INT NOT NULL DEFAULT 1 COMMENT '连续签到天数（签到时的连续天数）',
  `is_补签` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否补签（预留字段）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间（签到时间）',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_date` (`user_id`, `check_in_date`),
  KEY `idx_check_in_date` (`check_in_date`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户签到记录表';

-- Title Configs
CREATE TABLE IF NOT EXISTS `vp_title_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `title_key` VARCHAR(50) NOT NULL COMMENT '称号唯一标识',
  `title_name` VARCHAR(100) NOT NULL COMMENT '称号名称',
  `title_icon` VARCHAR(255) DEFAULT NULL COMMENT '称号图标（emoji或图片URL）',
  `description` VARCHAR(500) DEFAULT NULL COMMENT '称号描述',
  `category` VARCHAR(50) NOT NULL COMMENT '称号分类',
  `condition_type` VARCHAR(50) NOT NULL COMMENT '条件类型',
  `condition_value` INT NOT NULL DEFAULT 0 COMMENT '条件数值',
  `condition_description` VARCHAR(255) DEFAULT NULL COMMENT '条件说明',
  `rarity` VARCHAR(20) DEFAULT 'common' COMMENT '稀有度',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `is_active` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_title_key` (`title_key`),
  KEY `idx_category` (`category`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_sort_order` (`sort_order`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='称号配置表';

-- User Titles
CREATE TABLE IF NOT EXISTS `vp_user_titles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID，关联vp_users.id',
  `title_config_id` BIGINT UNSIGNED NOT NULL COMMENT '称号配置ID，关联vp_title_configs.id',
  `awarded_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '获得时间',
  `is_equipped` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否佩戴',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_title` (`user_id`, `title_config_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_title_config_id` (`title_config_id`),
  KEY `idx_awarded_at` (`awarded_at`),
  KEY `idx_is_equipped` (`is_equipped`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户称号表';

SET FOREIGN_KEY_CHECKS = 1;
