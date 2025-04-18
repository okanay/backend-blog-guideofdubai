package BlogRepository

import (
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) CreateBlogTag(request types.TagInput, userID uuid.UUID) (types.TagView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Category")
	var tagsView types.TagView

	query := `
		INSERT INTO tags (
			name, value, user_id
		) VALUES (
			$1, $2, $3
		) RETURNING name, value
	`
	err := r.db.QueryRow(query, request.Name, request.Value, userID).Scan(
		&tagsView.Name,
		&tagsView.Value,
	)
	if err != nil {
		return types.TagView{}, err
	}

	return tagsView, nil
}
