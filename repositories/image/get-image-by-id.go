// repositories/image/get-image-by-id.go
package ImageRepository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

// GetImageByID bir resmi ID'ye göre getirir
func (r *Repository) GetImageByID(ctx context.Context, imageID uuid.UUID) (*types.Image, error) {
	query := `
		SELECT id, user_id, url, filename, alt_text, file_type, size_in_bytes, width, height, status, created_at, updated_at
		FROM images
		WHERE id = $1 AND status = 'active'
	`

	var img types.Image
	err := r.db.QueryRowContext(ctx, query, imageID).Scan(
		&img.ID,
		&img.UserID,
		&img.URL,
		&img.Filename,
		&img.AltText,
		&img.FileType,
		&img.SizeInBytes,
		&img.Width,
		&img.Height,
		&img.Status,
		&img.CreatedAt,
		&img.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Resim bulunamadı
		}
		return nil, err
	}

	return &img, nil
}
