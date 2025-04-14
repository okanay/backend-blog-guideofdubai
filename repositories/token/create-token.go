package TokenRepository

import (
	"fmt"
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) CreateRefreshToken(request types.TokenCreateRequest) (types.RefreshToken, error) {
	defer utils.TimeTrack(time.Now(), "Token -> Create Token")

	var token types.RefreshToken

	query := `INSERT INTO refresh_tokens (user_id, user_email, user_username, token, ip_address, user_agent, expires_at, revoked_reason) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *`
	rows, err := r.db.Query(query, request.UserID, request.UserEmail, request.UserUsername, request.Token, request.IPAddress, request.UserAgent, request.ExpiresAt, "")
	defer rows.Close()

	if err != nil {
		return token, err
	}

	if !rows.Next() {
		return token, fmt.Errorf("No rows returned after insert")
	}

	err = utils.ScanStructByDBTags(rows, &token)
	if err != nil {
		return token, err
	}

	return token, nil
}
