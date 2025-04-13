-- İndeksleri kaldır
DROP INDEX IF EXISTS idx_refresh_tokens_token;

DROP INDEX IF EXISTS idx_users_username;

-- Tabloları kaldır
DROP TABLE IF EXISTS refresh_tokens;

DROP TABLE IF EXISTS users;

-- Enum tipleri kaldır
DROP TYPE IF EXISTS role;

DROP TYPE IF EXISTS user_status;
