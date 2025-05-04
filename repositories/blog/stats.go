// repositories/blog/stats.go
package BlogRepository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

// IncrementViewCount blog görüntülenme sayısını artırır
func (r *Repository) IncrementViewCount(blogID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Increment View Count")

	query := `
		UPDATE blog_stats
		SET views = views + 1,
		    last_viewed_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(query, blogID)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("blog stats not found for ID: %s", blogID)
	}

	return nil
}
