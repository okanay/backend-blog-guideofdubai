-- Kullanıcı durumu değiştiğinde deleted_at alanını güncelleme
CREATE OR REPLACE FUNCTION update_user_deleted_at()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.status = 'Deleted' AND OLD.status != 'Deleted' THEN
    NEW.deleted_at = NOW();
  ELSIF NEW.status != 'Deleted' AND OLD.status = 'Deleted' THEN
    -- Eğer silinen kullanıcı tekrar aktifleştirilirse
    NEW.deleted_at = NULL;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_user_deleted_at
BEFORE UPDATE OF status ON users
FOR EACH ROW
EXECUTE FUNCTION update_user_deleted_at();

-- Kullanıcı durumu değiştiğinde refresh tokenları güncelleme.
CREATE OR REPLACE FUNCTION update_refresh_tokens_on_user_status()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.status = 'Deleted' THEN
    UPDATE refresh_tokens
    SET is_revoked = TRUE,
        revoked_reason = 'User deleted',
        user_email = OLD.email
    WHERE user_id = OLD.id AND is_revoked = FALSE;
  ELSIF NEW.status = 'Suspended' THEN
    UPDATE refresh_tokens
    SET is_revoked = TRUE,
        revoked_reason = 'User suspended'
    WHERE user_id = OLD.id AND is_revoked = FALSE;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_refresh_tokens_on_user_status
AFTER UPDATE OF status ON users
FOR EACH ROW
EXECUTE FUNCTION update_refresh_tokens_on_user_status();
