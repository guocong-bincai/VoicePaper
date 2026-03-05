-- ============================================
-- VoicePaper 称号系统升级 SQL
-- 科举官职制 + 多分类体系
-- 生成日期: 2025-01-27
-- ============================================

-- 先清空旧数据
TRUNCATE TABLE vp_title_configs;

-- ============================================
-- 📚 阅读系列（6级）- category: reading
-- ============================================
INSERT INTO vp_title_configs (title_key, title_name, title_icon, description, category, condition_type, condition_value, condition_description, rarity, sort_order, is_active) VALUES
('reader_lv1', '白丁', '📄', '踏入书海的第一步', 'reading', 'articles_read', 1, '阅读1篇文章', 'common', 101, 1),
('reader_lv2', '书童', '📖', '开始你的阅读之旅', 'reading', 'articles_read', 5, '阅读5篇文章', 'common', 102, 1),
('reader_lv3', '秀才', '📚', '阅读让你成长', 'reading', 'articles_read', 20, '阅读20篇文章', 'rare', 103, 1),
('reader_lv4', '举人', '🎓', '博览群书的你', 'reading', 'articles_read', 50, '阅读50篇文章', 'rare', 104, 1),
('reader_lv5', '进士', '🏅', '学富五车', 'reading', 'articles_read', 150, '阅读150篇文章', 'epic', 105, 1),
('reader_lv6', '状元', '👑', '天下第一读书人', 'reading', 'articles_read', 500, '阅读500篇文章', 'legendary', 106, 1);

-- ============================================
-- ✍️ 默写系列（6级）- category: dictation
-- ============================================
INSERT INTO vp_title_configs (title_key, title_name, title_icon, description, category, condition_type, condition_value, condition_description, rarity, sort_order, is_active) VALUES
('dictation_lv1', '抄书郎', '✏️', '提笔开始练习', 'dictation', 'dictations_completed', 5, '完成5次默写', 'common', 201, 1),
('dictation_lv2', '笔吏', '✒️', '笔耕不辍', 'dictation', 'dictations_completed', 20, '完成20次默写', 'common', 202, 1),
('dictation_lv3', '编修', '✍️', '文思初成', 'dictation', 'dictations_completed', 50, '完成50次默写', 'rare', 203, 1),
('dictation_lv4', '侍读', '📝', '下笔如有神', 'dictation', 'dictations_completed', 150, '完成150次默写', 'rare', 204, 1),
('dictation_lv5', '翰林', '🖋️', '才高八斗', 'dictation', 'dictations_completed', 300, '完成300次默写', 'epic', 205, 1),
('dictation_lv6', '大学士', '💯', '文曲星下凡', 'dictation', 'dictations_completed', 600, '完成600次默写', 'legendary', 206, 1);

-- ============================================
-- ⭐ 签到系列（6级）- category: check_in
-- ============================================
INSERT INTO vp_title_configs (title_key, title_name, title_icon, description, category, condition_type, condition_value, condition_description, rarity, sort_order, is_active) VALUES
('checkin_lv1', '初心', '⭐', '坚持的开始', 'check_in', 'continuous_check_ins', 3, '连续签到3天', 'common', 301, 1),
('checkin_lv2', '立志', '🌟', '持之以恒', 'check_in', 'continuous_check_ins', 7, '连续签到7天', 'common', 302, 1),
('checkin_lv3', '守心', '💫', '不忘初心', 'check_in', 'continuous_check_ins', 14, '连续签到14天', 'rare', 303, 1),
('checkin_lv4', '恒心', '🏆', '锲而不舍', 'check_in', 'continuous_check_ins', 30, '连续签到30天', 'epic', 304, 1),
('checkin_lv5', '铁心', '💎', '铁杵成针', 'check_in', 'continuous_check_ins', 60, '连续签到60天', 'epic', 305, 1),
('checkin_lv6', '金心', '👑', '传说级的坚持', 'check_in', 'continuous_check_ins', 90, '连续签到90天', 'legendary', 306, 1);

-- ============================================
-- 💰 积分系列（6级）- category: points
-- ============================================
INSERT INTO vp_title_configs (title_key, title_name, title_icon, description, category, condition_type, condition_value, condition_description, rarity, sort_order, is_active) VALUES
('points_lv1', '温饱', '🪙', '积分初积累', 'points', 'total_points', 500, '累计获得500积分', 'common', 401, 1),
('points_lv2', '小康', '💰', '小有积蓄', 'points', 'total_points', 1000, '累计获得1000积分', 'common', 402, 1),
('points_lv3', '殷实', '💵', '积分稳步增长', 'points', 'total_points', 3000, '累计获得3000积分', 'rare', 403, 1),
('points_lv4', '富足', '💴', '积分丰厚', 'points', 'total_points', 6000, '累计获得6000积分', 'rare', 404, 1),
('points_lv5', '豪绅', '💎', '积分大户', 'points', 'total_points', 10000, '累计获得10000积分', 'epic', 405, 1),
('points_lv6', '首富', '👑', '积分富可敌国', 'points', 'total_points', 20000, '累计获得20000积分', 'legendary', 406, 1);

-- ============================================
-- 📒 生词本系列（5级）- category: vocabulary
-- ============================================
INSERT INTO vp_title_configs (title_key, title_name, title_icon, description, category, condition_type, condition_value, condition_description, rarity, sort_order, is_active) VALUES
('vocab_lv1', '采词人', '📝', '开始收集生词', 'vocabulary', 'vocabulary_count', 10, '收藏10个生词', 'common', 501, 1),
('vocab_lv2', '集词客', '📒', '词汇渐丰', 'vocabulary', 'vocabulary_count', 50, '收藏50个生词', 'common', 502, 1),
('vocab_lv3', '藏词家', '📚', '词汇丰富', 'vocabulary', 'vocabulary_count', 150, '收藏150个生词', 'rare', 503, 1),
('vocab_lv4', '词库主', '🗃️', '坐拥词海', 'vocabulary', 'vocabulary_count', 300, '收藏300个生词', 'epic', 504, 1),
('vocab_lv5', '词海王', '👑', '词汇量惊人', 'vocabulary', 'vocabulary_count', 600, '收藏600个生词', 'legendary', 505, 1);

-- ============================================
-- 🔄 复习系列（5级）- category: review
-- ============================================
INSERT INTO vp_title_configs (title_key, title_name, title_icon, description, category, condition_type, condition_value, condition_description, rarity, sort_order, is_active) VALUES
('review_lv1', '温故', '🔄', '复习使人进步', 'review', 'review_count', 20, '完成20次复习', 'common', 601, 1),
('review_lv2', '知新', '📖', '温故而知新', 'review', 'review_count', 80, '完成80次复习', 'common', 602, 1),
('review_lv3', '融会', '🧠', '知识融会贯通', 'review', 'review_count', 200, '完成200次复习', 'rare', 603, 1),
('review_lv4', '贯通', '💡', '举一反三', 'review', 'review_count', 500, '完成500次复习', 'epic', 604, 1),
('review_lv5', '过目不忘', '👁️', '记忆力超群', 'review', 'review_count', 1000, '完成1000次复习', 'legendary', 605, 1);

-- ============================================
-- ⏱️ 学习时长系列（5级）- category: duration
-- ============================================
INSERT INTO vp_title_configs (title_key, title_name, title_icon, description, category, condition_type, condition_value, condition_description, rarity, sort_order, is_active) VALUES
('duration_lv1', '初学者', '⏰', '学习之路刚起步', 'duration', 'total_duration', 60, '累计学习1小时', 'common', 701, 1),
('duration_lv2', '勤学者', '⏱️', '勤奋好学', 'duration', 'total_duration', 300, '累计学习5小时', 'common', 702, 1),
('duration_lv3', '苦读生', '📚', '十年寒窗', 'duration', 'total_duration', 1200, '累计学习20小时', 'rare', 703, 1),
('duration_lv4', '学霸', '🎓', '学习达人', 'duration', 'total_duration', 3000, '累计学习50小时', 'epic', 704, 1),
('duration_lv5', '卷王', '👑', '无人能敌的勤奋', 'duration', 'total_duration', 6000, '累计学习100小时', 'legendary', 705, 1);

-- ============================================
-- 🎯 特殊成就系列 - category: achievement
-- ============================================
INSERT INTO vp_title_configs (title_key, title_name, title_icon, description, category, condition_type, condition_value, condition_description, rarity, sort_order, is_active) VALUES
('first_read', '开卷有益', '📖', '万事开头难', 'achievement', 'custom', 1, '首次阅读文章', 'common', 801, 1),
('first_dictation', '初试锋芒', '✏️', '迈出默写第一步', 'achievement', 'custom', 1, '首次完成默写', 'common', 802, 1),
('perfect_10', '十全十美', '🎯', '追求完美的你', 'achievement', 'perfect_streak', 10, '连续10次默写全对', 'rare', 803, 1),
('perfect_30', '完美主义', '💯', '完美是一种态度', 'achievement', 'perfect_streak', 30, '连续30次默写全对', 'epic', 804, 1),
('early_bird', '早起鸟儿', '🌅', '一日之计在于晨', 'achievement', 'custom', 1, '早上6点前学习', 'rare', 805, 1),
('night_owl', '夜猫子', '🌙', '夜深人静正读书', 'achievement', 'custom', 1, '凌晨12点后学习', 'rare', 806, 1),
('weekend_warrior', '周末战士', '⚔️', '周末也不忘学习', 'achievement', 'custom', 1, '周末学习超2小时', 'rare', 807, 1);

-- ============================================
-- 🌱 新手引导系列 - category: special
-- ============================================
INSERT INTO vp_title_configs (title_key, title_name, title_icon, description, category, condition_type, condition_value, condition_description, rarity, sort_order, is_active) VALUES
('beginner', '新手上路', '🌱', '欢迎加入VoicePaper！', 'special', 'custom', 0, '注册账号即可获得', 'common', 1, 1);

-- 查看插入结果
SELECT category, COUNT(*) as count FROM vp_title_configs GROUP BY category ORDER BY MIN(sort_order);
SELECT * FROM vp_title_configs ORDER BY sort_order;
