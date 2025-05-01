package BlogRepository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectBlogByGroupID(request types.BlogSelectByGroupIDInput) (*types.BlogPostView, []*types.BlogPostView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog And Alternatives")

	// 1. Önce ana postu bul
	var blogID uuid.UUID
	var groupID string

	query := `
        SELECT id, group_id FROM blog_posts
        WHERE (group_id = $1 OR slug = $1)
          AND status != 'deleted'
        LIMIT 1
    `
	err := r.db.QueryRow(query, request.SlugOrGroupID).Scan(&blogID, &groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("no blog posts found for groupId or slug=%s", request.SlugOrGroupID)
		}
		return nil, nil, fmt.Errorf("error retrieving blog data: %w", err)
	}

	// 2. Ana postun tüm detaylarını çek
	post, err := r.SelectBlogByID(blogID)
	if err != nil {
		return nil, nil, err
	}

	// 3. Alternatifleri bul (aynı groupID'ye sahip tüm postlar)
	altQuery := `
        SELECT id FROM blog_posts
        WHERE group_id = $1
          AND status != 'deleted'
        ORDER BY language
    `
	rows, err := r.db.Query(altQuery, groupID)
	if err != nil {
		return post, nil, fmt.Errorf("error retrieving alternatives: %w", err)
	}
	defer rows.Close()

	var alternatives []*types.BlogPostView
	for rows.Next() {
		var altID uuid.UUID
		if err := rows.Scan(&altID); err != nil {
			continue // hata olursa atla
		}
		altPost, err := r.SelectBlogByID(altID)
		if err != nil {
			continue // hata olursa atla
		}
		alternatives = append(alternatives, altPost)
	}

	return post, alternatives, nil
}
