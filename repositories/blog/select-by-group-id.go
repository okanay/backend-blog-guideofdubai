package BlogRepository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectBlogByGroupID(request types.BlogSelectByGroupIDInput) (*types.BlogPostView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog By GroupID")
	var blogID uuid.UUID

	// 1. Önce istenen dilde blog post'u arayalım
	query := `
		SELECT id FROM blog_posts
		WHERE group_id = $1 AND language = $2 AND status != 'deleted'
		LIMIT 1
	`
	err := r.db.QueryRow(query, request.GroupID, request.Language).Scan(&blogID)
	if err == nil {
		return r.SelectBlogByID(blogID)
	}

	// 2. İstenen dilde blog post bulunamadıysa, herhangi bir dildeki blog post'u ara
	query = `
		SELECT id FROM blog_posts
		WHERE group_id = $1 AND status != 'deleted'
		ORDER BY created_at DESC
		LIMIT 1
	`
	err = r.db.QueryRow(query, request.GroupID).Scan(&blogID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no blog posts found for group_id=%s", request.GroupID)
		}
		return nil, fmt.Errorf("error retrieving blog data: %w", err)
	}

	return r.SelectBlogByID(blogID)
}
