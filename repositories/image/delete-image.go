// repositories/image/delete-image.go
package ImageRepository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// DeleteImage bir resmi siler (veya durumunu 'deleted' olarak günceller)
func (r *Repository) DeleteImage(ctx context.Context, imageID, userID uuid.UUID) error {
	// Soft delete - durumu 'deleted' olarak güncelle
	query := `
		UPDATE images
		SET status = 'deleted', updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, imageID, userID)
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
