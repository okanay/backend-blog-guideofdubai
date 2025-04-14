package UserRepository

import (
	"fmt"
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectByUsername(username string) (types.User, error) {
	defer utils.TimeTrack(time.Now(), "User -> Select User By Username")

	var user types.User

	query := `SELECT * FROM users WHERE username = $1 LIMIT 1`
	rows, err := r.db.Query(query, username)
	defer rows.Close()
	if err != nil {
		return user, err
	}

	if !rows.Next() {
		return user, fmt.Errorf("No rows returned after select")
	}

	err = utils.ScanStructByDBTags(rows, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}
