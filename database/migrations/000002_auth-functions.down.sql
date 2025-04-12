-- Trigger'ları kaldır
DROP TRIGGER IF EXISTS trigger_refresh_tokens_on_user_status ON users;

DROP TRIGGER IF EXISTS trigger_user_deleted_at ON users;

-- Fonksiyonları kaldır
DROP FUNCTION IF EXISTS update_refresh_tokens_on_user_status ();

DROP FUNCTION IF EXISTS update_user_deleted_at ();
