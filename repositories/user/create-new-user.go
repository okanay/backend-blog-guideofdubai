package UserRepository

import (
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) CreateNewUser(request types.UserCreateRequest) (types.User, error) {
	defer utils.TimeTrack(time.Now(), "User -> Create User")

	var user types.User
	hashedPassword, err := utils.EncryptPassword(request.Password)

	if err != nil {
		return user, err
	}

	query := `INSERT INTO users (email, username, hashed_password) VALUES ($1, $2, $3) RETURNING *`

	row := r.db.QueryRow(query, request.Email, request.Username, hashedPassword)
	err = utils.ScanStructByDBTags(row, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}
