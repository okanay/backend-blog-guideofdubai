package UserRepository

import (
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectUserByUsername(username string) (types.User, error) {
	defer utils.TimeTrack(time.Now(), "User -> Select User By Username")

	var user types.User

	query := `SELECT * FROM users WHERE username = $1`

	row := r.db.QueryRow(query, username)
	err := utils.ScanStructByDBTags(row, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}
