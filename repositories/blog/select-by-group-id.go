package BlogRepository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectBlogByGroupID(request types.BlogSelectByGroupIDInput) (*types.BlogPostView, int, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog By GroupID or Slug")
	var blogID uuid.UUID
	var priority int

	query := `
        SELECT id,
            CASE
                WHEN slug = $1 AND language = $2 THEN 1
                WHEN group_id = $1 AND language = $2 THEN 2
                WHEN slug = $1 THEN 3
                WHEN group_id = $1 THEN 4
                ELSE 5
            END AS priority
        FROM blog_posts
        WHERE (slug = $1 OR group_id = $1)
          AND status != 'deleted'
        ORDER BY priority
        LIMIT 1
    `
	err := r.db.QueryRow(query, request.SlugOrGroupID, request.Language).Scan(&blogID, &priority)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, fmt.Errorf("no blog posts found for slugOrGroupID=%s", request.SlugOrGroupID)
		}
		return nil, 0, fmt.Errorf("error retrieving blog data: %w", err)
	}

	blog, err := r.SelectBlogByID(blogID)
	if err != nil {
		return nil, 0, err
	}

	return blog, priority, nil
}
