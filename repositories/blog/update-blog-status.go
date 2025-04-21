package BlogRepository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) UpdateBlogStatus(blogID uuid.UUID, status types.BlogStatus) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Update Blog Status")

	query := `
		UPDATE blog_posts
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	// Blog yayınlanıyorsa published_at tarihini güncelle
	var result any
	var err error

	if status == types.BlogStatusPublished {
		queryWithPublished := `
			UPDATE blog_posts
			SET status = $1, updated_at = $2, published_at = $2
			WHERE id = $3
		`
		result, err = r.db.Exec(queryWithPublished, status, time.Now(), blogID)
	} else {
		result, err = r.db.Exec(query, status, time.Now(), blogID)
	}

	if err != nil {
		return fmt.Errorf("blog durumu güncellenirken hata oluştu: %w", err)
	}

	// Satır etkilenme kontrolü
	rowsAffected, err := result.(interface{ RowsAffected() (int64, error) }).RowsAffected()
	if err != nil {
		return fmt.Errorf("etkilenen satır sayısı kontrol edilirken hata: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("blog bulunamadı veya durum değişikliği yapılamadı")
	}

	return nil
}
