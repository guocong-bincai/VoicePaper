-- 清理指定邮箱的验证码记录（用于测试）
DELETE FROM vp_verification_codes 
WHERE receiver = '1752936267@qq.com' 
  AND code_type = 'email';

-- 或者清理所有过期和已使用的验证码
DELETE FROM vp_verification_codes 
WHERE expires_at < NOW() 
   OR (used = 1 AND used_at < DATE_SUB(NOW(), INTERVAL 24 HOUR));

