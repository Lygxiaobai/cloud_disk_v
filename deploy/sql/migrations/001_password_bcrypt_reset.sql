-- ========================================================
-- 001_password_bcrypt_reset.sql
-- 日期：2026-04-21
-- 说明：将 user_basic.password 从 MD5（32 位 hex）迁移到 bcrypt。
--       bcrypt 哈希以 $2a$/$2b$ 开头，长度约 60。
--       所有 MD5 老密码一律清空（字符串长度为 32），
--       登录时 bcrypt.CompareHashAndPassword 对空字符串会失败，
--       效果等同于"账号被锁定"，用户需要管理员重置或未来的"忘记密码"流程。
-- ========================================================

-- 安全提示：执行前务必备份 user_basic 表。
-- 建议执行：
--   mysqldump cloud_disk user_basic > backup_user_basic_$(date +%Y%m%d).sql

UPDATE user_basic
SET    password = ''
WHERE  LENGTH(password) = 32
   AND password REGEXP '^[0-9a-f]{32}$';

-- 校验：确认没有遗留 MD5 记录
-- SELECT COUNT(*) FROM user_basic WHERE LENGTH(password) = 32;
