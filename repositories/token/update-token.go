package TokenRepository

import (
	"time"

	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) UpdateRefreshTokenLastUsed(token string) error {
	defer utils.TimeTrack(time.Now(), "Token -> Update Refresh Token Last Used")

	now := time.Now()
	query := `UPDATE refresh_tokens SET last_used_at = $1 WHERE token = $2`

	_, err := r.db.Exec(query, now, token)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) ExtendRefreshTokenExpiry(token string, duration time.Duration) error {
	defer utils.TimeTrack(time.Now(), "Token -> Extend Refresh Token Expiry")

	now := time.Now()
	newExpiryTime := now.Add(duration)

	query := `UPDATE refresh_tokens
              SET expires_at = $1, last_used_at = $2
              WHERE token = $3 AND is_revoked = FALSE`

	_, err := r.db.Exec(query, newExpiryTime, now, token)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateRefreshToken(oldToken, newToken string, expiresAt time.Time) error {
	defer utils.TimeTrack(time.Now(), "Token -> Update Refresh Token")

	now := time.Now()

	query := `UPDATE refresh_tokens
              SET token = $1, expires_at = $2, last_used_at = $3
              WHERE token = $4 AND is_revoked = FALSE`

	_, err := r.db.Exec(query, newToken, expiresAt, now, oldToken)
	if err != nil {
		return err
	}

	return nil
}
