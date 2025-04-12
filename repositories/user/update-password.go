package UserRepository

import (
	"time"

	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) UpdatePassword(email string, password string) error {
	defer utils.TimeTrack(time.Now(), "User -> Update Password")

	hash, err := utils.EncryptPassword(password)
	if err != nil {
		return err
	}

	query := `UPDATE users SET password=$1, updated_at=$2 WHERE email=$3`

	_, err = r.db.Exec(query, hash, time.Now(), email)
	if err != nil {
		return err
	}

	return nil
}
