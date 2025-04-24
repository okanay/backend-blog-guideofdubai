// repositories/image/get-image-by-id.go
package ImageRepository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (r *Repository) GetSignatureByID(ctx context.Context, signatureID uuid.UUID) (*types.UploadSignature, error) {
	query := `
		SELECT id, user_id, image_id, presigned_url, upload_url, filename, file_type, expires_at, completed, created_at
		FROM upload_signatures
		WHERE id = $1
	`

	var signature types.UploadSignature
	err := r.db.QueryRowContext(ctx, query, signatureID).Scan(
		&signature.ID,
		&signature.UserID,
		&signature.ImageID,
		&signature.PresignedURL,
		&signature.UploadURL,
		&signature.Filename,
		&signature.FileType,
		&signature.ExpiresAt,
		&signature.Completed,
		&signature.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Resim bulunamadı
		}
		return nil, err
	}

	return &signature, nil
}
