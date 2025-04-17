package TokenRepository

import (
	"fmt"
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectRefreshTokenByToken(token string) (types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> Select Refresh Token By Token")

	var refreshToken types.RefreshToken
	query := `SELECT * FROM refresh_tokens WHERE token = $1 AND is_revoked = FALSE`

	rows, err := r.db.Query(query, token)
	if err != nil {
		return refreshToken, fmt.Errorf("token sorgulanırken hata: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return refreshToken, fmt.Errorf("token bulunamadı")
	}

	// Burada kritik hata vardı - &token yerine &refreshToken kullanılmalı
	err = utils.ScanStructByDBTags(rows, &refreshToken)
	if err != nil {
		return refreshToken, fmt.Errorf("token verileri okunurken hata: %w", err)
	}

	return refreshToken, nil
}

func (r *Repository) SelectActiveTokensByUserID(userID int64) ([]types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> Select Active Tokens By User ID")

	var tokens []types.RefreshToken

	query := `SELECT * FROM refresh_tokens WHERE user_id = $1 AND is_revoked = FALSE AND expires_at > NOW()`
	rows, err := r.db.Query(query, userID)
	defer rows.Close()

	if err != nil {
		return tokens, err
	}

	if !rows.Next() {
		return tokens, fmt.Errorf("No rows returned after select")
	}

	for rows.Next() {
		var token types.RefreshToken
		if err := utils.ScanStructByDBTags(rows, &token); err != nil {
			continue
		}
		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return tokens, err
	}

	return tokens, nil
}
