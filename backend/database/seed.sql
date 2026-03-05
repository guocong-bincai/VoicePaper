-- VoicePaper Seed Data
-- Initial data for a functional application

SET NAMES utf8mb4;

-- 1. Initialize Categories
INSERT INTO `vp_categories` (`id`, `name`, `description`, `icon`, `sort`, `is_active`) VALUES
(1, '精读', '双语精读课，深入理解文章内容', '📖', 1, 1),
(2, 'Tech', 'Technology News and Updates', '💻', 2, 1),
(3, 'Science', 'Scientific Discoveries', '🔬', 3, 1),
(4, 'Life', 'Daily Life and Culture', '🌟', 4, 1);

-- 2. Initialize Title Configs (Gamification)
INSERT INTO `vp_title_configs` 
(`title_key`, `title_name`, `title_icon`, `description`, `category`, `condition_type`, `condition_value`, `condition_description`, `rarity`, `sort_order`, `is_active`) 
VALUES
('beginner', 'Newbie', '🌱', 'Welcome to VoicePaper!', 'special', 'custom', 0, 'Register an account', 'common', 1, 1),
('reader_novice', 'Reader Novice', '📖', 'Start your reading journey', 'reading', 'articles_read', 5, 'Read 5 articles', 'common', 10, 1),
('dictation_beginner', 'Dictation Apprentice', '✏️', 'Start dictation practice', 'dictation', 'dictations_completed', 10, 'Complete 10 dictations', 'common', 20, 1),
('check_in_week', 'Persistence Star', '⭐', 'Keep it up!', 'check_in', 'continuous_check_ins', 7, 'Check-in 7 days in a row', 'rare', 30, 1);

-- 3. Initialize Sample Article
-- Using the real Parthenon Marbles article provided by the user
INSERT INTO `vp_articles` (`id`, `title`, `online`, `category_id`, `publish_date`, `is_daily`, `audio_url`, `timeline_url`, `article_url`, `created_at`, `updated_at`) VALUES
(1, '《卫报》双语精读课：帕特农神庙大理石之争 —— 一场跨越 200 年的文化博弈', '1', 1, '2025-12-03', 1, 
'https://voicepaper.oss-cn-chengdu.aliyuncs.com/articles/audio/20251203/20251203_134536_audio.mp3', 
'https://voicepaper.oss-cn-chengdu.aliyuncs.com/articles/timeline/20251203/20251203_134536_timeline.json', 
'https://voicepaper.oss-cn-chengdu.aliyuncs.com/articles/content/20251203/20251203_134536_article.md', 
'2025-12-03 13:45:42', '2025-12-03 13:45:42');

-- 4. Initialize Sentences for the Sample Article
INSERT INTO `vp_sentences` (article_id, `text`, translation, `order`, created_at, updated_at) VALUES
(1, 'Time is money.', '时间就是金钱。', 1, NOW(), NOW()),
(1, 'Knowledge is power.', '知识就是力量。', 2, NOW(), NOW()),
(1, 'The museum operates independently of government.', '博物馆独立于政府运作。', 3, NOW(), NOW()),
(1, 'Practice makes perfect.', '熟能生巧。', 4, NOW(), NOW()),
(1, 'No pain, no gain.', '一分耕耘，一分收获。', 5, NOW(), NOW()),
(1, 'They were legally acquired by Elgin.', '它们是埃尔金合法获得的。', 6, NOW(), NOW()),
(1, 'Better late than never.', '迟做总比不做好。', 7, NOW(), NOW()),
(1, 'Actions speak louder than words.', '行动胜于言语。', 8, NOW(), NOW()),
(1, 'He was a true philhellene.', '他是真正的希腊文化爱好者。', 9, NOW(), NOW()),
(1, 'Honesty is the best policy.', '诚实为上策。', 10, NOW(), NOW()),
(1, 'The early bird catches the worm.', '早起的鸟儿有虫吃。', 11, NOW(), NOW()),
(1, 'Museums could repatriate disputed bronzes.', '博物馆可能归还有争议的青铜器。', 12, NOW(), NOW()),
(1, 'Every cloud has a silver lining.', '黑暗中总有一线光明。', 13, NOW(), NOW()),
(1, 'All that glitters is not gold.', '发光的未必都是金子。', 14, NOW(), NOW()),
(1, 'Johnson understood the Greek people.', '约翰逊理解希腊人民。', 15, NOW(), NOW()),
(1, 'Rome was not built in a day.', '罗马不是一天建成的。', 16, NOW(), NOW()),
(1, 'Live and learn.', '活到老，学到老。', 17, NOW(), NOW()),
(1, 'UK reversed its longstanding opposition.', '英国扭转了长期反对态度。', 18, NOW(), NOW()),
(1, 'Easy come, easy go.', '来得容易，去得快。', 19, NOW(), NOW()),
(1, 'Think before you speak.', '三思而后言。', 20, NOW(), NOW());

-- 5. Initialize Words for the Sample Article
INSERT INTO `vp_words` (article_id, `text`, phonetic, meaning, example, example_translation, `level`, frequency, `order`, is_key_word, created_at, updated_at) VALUES
(1, 'world', '/wɜːld/', '世界', 'The world is changing rapidly.', '世界正在迅速变化。', 1, 95, 1, 1, NOW(), NOW()),
(1, 'time', '/taɪm/', '时间', 'Time is money.', '时间就是金钱。', 1, 98, 2, 1, NOW(), NOW()),
(1, 'people', '/ˈpiːpl/', '人们', 'Many people visit this museum.', '许多人参观这个博物馆。', 1, 97, 3, 1, NOW(), NOW()),
(1, 'year', '/jɪə(r)/', '年', 'This year has been challenging.', '今年充满挑战。', 1, 96, 4, 1, NOW(), NOW()),
(1, 'repatriate', '/riːˈpætrieɪt/', '遣返', 'They decided to repatriate artifacts.', '他们决定遣返文物。', 5, 19, 5, 1, NOW(), NOW()),
(1, 'opposition', '/ˌɒpəˈzɪʃn/', '反对', 'There is strong opposition to the plan.', '该计划遭到强烈反对。', 3, 44, 6, 1, NOW(), NOW()),
(1, 'independently', '/ˌɪndɪˈpɛndəntli/', '独立地', 'The museum operates independently.', '博物馆独立运作。', 3, 42, 7, 1, NOW(), NOW());