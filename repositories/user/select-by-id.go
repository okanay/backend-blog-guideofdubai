package UserRepository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectByID(id uuid.UUID) (types.User, error) {
	defer utils.TimeTrack(time.Now(), "User -> Select User By ID")

	var user types.User

	query := `SELECT * FROM users WHERE id = $1 LIMIT 1`
	rows, err := r.db.Query(query, id)
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
