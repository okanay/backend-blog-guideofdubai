-- İndeksleri kaldır
DROP INDEX IF EXISTS idx_refresh_tokens_user_email;

DROP INDEX IF EXISTS idx_refresh_tokens_user_username;

DROP INDEX IF EXISTS idx_refresh_tokens_user_id_is_revoked;

DROP INDEX IF EXISTS idx_refresh_tokens_user_id;

DROP INDEX IF EXISTS idx_refresh_tokens_token;

DROP INDEX IF EXISTS idx_users_status;

DROP INDEX IF EXISTS idx_users_username;

DROP INDEX IF EXISTS idx_users_email;

-- Tabloları kaldır
DROP TABLE IF EXISTS refresh_tokens;

DROP TABLE IF EXISTS users;

-- Enum tipleri kaldır
DROP TYPE IF EXISTS role;

DROP TYPE IF EXISTS user_status;
