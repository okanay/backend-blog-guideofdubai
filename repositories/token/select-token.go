package TokenRepository

import (
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectRefreshTokenByToken(token string) (types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> Select Refresh Token By Token")

	var refreshToken types.RefreshToken

	query := `SELECT * FROM refresh_tokens WHERE token = $1 AND is_revoked = FALSE`

	row := r.db.QueryRow(query, token)
	err := utils.ScanStructByDBTags(row, &refreshToken)
	if err != nil {
		return refreshToken, err
	}

	return refreshToken, nil
}

func (r *Repository) SelectActiveTokensByUserID(userID int64) ([]types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> Select Active Tokens By User ID")

	var tokens []types.RefreshToken

	query := `SELECT * FROM refresh_tokens WHERE user_id = $1 AND is_revoked = FALSE AND expires_at > NOW()`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return tokens, err
	}
	defer rows.Close()

	for rows.Next() {
		var token types.RefreshToken
		if err := utils.ScanStructByDBTagsForRows(rows, &token); err != nil {
			return tokens, err
		}
		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return tokens, err
	}

	return tokens, nil
}
