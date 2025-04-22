// repositories/image/save-image.go
package ImageRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (r *Repository) SaveImage(ctx context.Context, userID uuid.UUID, input types.SaveImageInput) (uuid.UUID, error) {
	var id uuid.UUID

	query := `
		INSERT INTO images (
			user_id, url, filename, alt_text, file_type, size_in_bytes, width, height
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
		input.URL,
		input.Filename,
		input.AltText,
		input.FileType,
		input.SizeInBytes,
		input.Width,
		input.Height,
	).Scan(&id)

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
