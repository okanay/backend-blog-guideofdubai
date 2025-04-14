package UserRepository

import (
	"fmt"
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) CreateUser(request types.UserCreateRequest) (types.User, error) {
	defer utils.TimeTrack(time.Now(), "User -> Create User")

	var user types.User
	hashedPassword, err := utils.EncryptPassword(request.Password)
	if err != nil {
		return user, err
	}

	// QueryRow yerine Query kullanarak Rows elde et
	query := `INSERT INTO users (email, username, hashed_password) VALUES ($1, $2, $3) RETURNING *`
	rows, err := r.db.Query(query, request.Email, request.Username, hashedPassword)
	if err != nil {
		return user, err
	}
	defer rows.Close() // Connection'ı kapatmayı unutmayın

	if !rows.Next() {
		return user, fmt.Errorf("No rows returned after insert")
	}

	err = utils.ScanStructByDBTags(rows, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}
