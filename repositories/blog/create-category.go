package BlogRepository

import (
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) CreateBlogCategory(request types.CategoryInput, userID uuid.UUID) (types.CategoryView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Category")

	query := `
		INSERT INTO categories (
			name, value, user_id
		) VALUES (
			$1, $2, $3
		) RETURNING name, value
	`
	var categoryView types.CategoryView
	err := r.db.QueryRow(query, request.Name, request.Value, userID).Scan(
		&categoryView.Name,
		&categoryView.Value,
	)
	if err != nil {
		return types.CategoryView{}, err
	}

	return categoryView, nil
}
