package TokenRepository

import (
	"time"

	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) RevokeRefreshToken(token string, reason string) error {
	defer utils.TimeTrack(time.Now(), "Token -> Revoke Refresh Token")

	query := `UPDATE refresh_tokens SET is_revoked = TRUE, revoked_reason = $1 WHERE token = $2`

	_, err := r.db.Exec(query, reason, token)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) RevokeAllUserTokens(userID int64, reason string) error {
	defer utils.TimeTrack(time.Now(), "Token -> Revoke All User Tokens")

	query := `UPDATE refresh_tokens SET is_revoked = TRUE, revoked_reason = $1
              WHERE user_id = $2 AND is_revoked = FALSE`

	_, err := r.db.Exec(query, reason, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) RevokeExpiredTokens() error {
	defer utils.TimeTrack(time.Now(), "Token -> Revoke Expired Tokens")

	query := `UPDATE refresh_tokens SET is_revoked = TRUE, revoked_reason = 'Token expired'
              WHERE expires_at < NOW() AND is_revoked = FALSE`

	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
