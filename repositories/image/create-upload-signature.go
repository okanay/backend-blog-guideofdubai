package ImageRepository

import (
	"context"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (r *Repository) CreateUploadSignature(ctx context.Context, userID uuid.UUID, input types.UploadSignatureInput) (uuid.UUID, error) {

	var id uuid.UUID

	query := `
		INSERT INTO upload_signatures (
			user_id, presigned_url, upload_url, filename, file_type, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		) RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
		input.PresignedURL,
		input.UploadURL,
		input.Filename,
		input.FileType,
		input.ExpiresAt,
	).Scan(&id)

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
