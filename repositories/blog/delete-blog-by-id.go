package BlogRepository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) DeleteBlogByID(blogID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Delete Blog By ID")

	// Begin a transaction for this operation
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Rollback transaction on error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Update the blog status to deleted
	query := `
		UPDATE blog_posts
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err = tx.Exec(query, types.BlogStatusDeleted, time.Now(), blogID)
	if err != nil {
		return fmt.Errorf("failed to mark blog as deleted: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) HardDeleteBlogByID(blogID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Hard Delete Blog By ID")

	// Begin a transaction for this operation
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Rollback transaction on error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// The database schema has CASCADE DELETE constraints for these related tables:
	// - blog_metadata
	// - blog_content
	// - blog_stats
	// - blog_categories
	// - blog_tags
	// So we only need to delete from the main blog_posts table

	query := `DELETE FROM blog_posts WHERE id = $1`
	result, err := tx.Exec(query, blogID)
	if err != nil {
		return fmt.Errorf("failed to delete blog: %w", err)
	}

	// Check if any row was affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("blog with ID %s not found", blogID)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
