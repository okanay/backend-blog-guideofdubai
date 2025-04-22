// repositories/image/get-user-images.go
package ImageRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

// GetImagesByUserID kullanıcıya ait resimleri getirir
func (r *Repository) GetImagesByUserID(ctx context.Context, userID uuid.UUID) ([]types.Image, error) {
	query := `
		SELECT id, user_id, url, filename, alt_text, file_type, size_in_bytes, width, height, status, created_at, updated_at
		FROM images
		WHERE user_id = $1 AND status = 'active'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []types.Image
	for rows.Next() {
		var img types.Image
		err := rows.Scan(
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
			return nil, err
		}
		images = append(images, img)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}
