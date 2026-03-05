-- MySQL 5.7+ 兼容的表修改脚本
-- 为 vp_users 表添加累积学习时长字段
-- 注意：如果字段已存在，会报错，这是正常的
ALTER TABLE vp_users ADD COLUMN total_duration_minutes BIGINT DEFAULT 0 COMMENT '累积学习时长（分钟）';

-- 为 vp_user_points 表添加累积学习时长字段
ALTER TABLE vp_user_points ADD COLUMN total_duration_minutes BIGINT DEFAULT 0 COMMENT '累积学习时长（分钟）';

-- 为排名查询创建索引
-- 如果索引已存在会报错，这是正常的
CREATE INDEX idx_user_points_ranking ON vp_user_points(total_duration_minutes, total_points);
CREATE INDEX idx_user_points_user_id ON vp_user_points(user_id);