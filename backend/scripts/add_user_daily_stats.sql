-- 创建用户每日统计表
-- 用于记录小程序端的每日学习数据

CREATE TABLE IF NOT EXISTS `vp_user_daily_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `stat_date` date NOT NULL COMMENT '统计日期',
  `total_duration_seconds` int NOT NULL DEFAULT '0' COMMENT '今日总时长（秒）',
  `new_words` int NOT NULL DEFAULT '0' COMMENT '新学单词数',
  `reviewed_words` int NOT NULL DEFAULT '0' COMMENT '复习单词数',
  `correct_count` int NOT NULL DEFAULT '0' COMMENT '正确次数',
  `total_attempts` int NOT NULL DEFAULT '0' COMMENT '总尝试次数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_date` (`user_id`,`stat_date`),
  KEY `idx_vp_user_daily_stats_deleted_at` (`deleted_at`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_stat_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户每日统计表';
