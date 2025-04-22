// repositories/image/complete-upload-signature.go
package ImageRepository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// CompleteUploadSignature bir yükleme imzasını tamamlandı olarak işaretler ve resim ile ilişkilendirir
func (r *Repository) CompleteUploadSignature(ctx context.Context, signatureID, imageID uuid.UUID) error {
	query := `
		UPDATE upload_signatures
		SET completed = true, image_id = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, imageID, signatureID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
