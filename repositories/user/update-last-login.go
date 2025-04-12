package UserRepository

import (
	"time"

	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) UpdateLastLogin(email string, updateAt time.Time) error {
	defer utils.TimeTrack(time.Now(), "User -> Update Last Login User")

	query := `UPDATE users SET last_login=$1, updated_at=$2 WHERE email=$3`

	_, err := r.db.Exec(query, updateAt, updateAt, email)
	if err != nil {
		return err
	}

	return nil
}
