package BlogRepository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) UpdateBlogPost(input types.BlogUpdateInput) (*types.BlogPostView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Update Blog Post")

	// Blog ID'sini kontrol et
	blogID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid blog ID: %w", err)
	}

	// İşlem için bir transaction başlat
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to initiate transaction: %w", err)
	}

	// Herhangi bir hata durumunda transaction'ı geri al
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Blog post bilgilerini güncelle
	err = r.updateBlogPostDetails(tx, blogID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update blog post details: %w", err)
	}

	// 2. Metadata bilgilerini güncelle
	err = r.updateBlogMetadata(tx, blogID, input.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to update blog metadata: %w", err)
	}

	// 3. İçerik bilgilerini güncelle
	err = r.updateBlogContent(tx, blogID, input.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to update blog content: %w", err)
	}

	// 4. Kategorileri güncelle
	err = r.updateBlogCategories(tx, blogID, input.Categories)
	if err != nil {
		return nil, fmt.Errorf("failed to update blog categories: %w", err)
	}

	// 5. Etiketleri güncelle
	err = r.updateBlogTags(tx, blogID, input.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to update blog tags: %w", err)
	}

	// Transaction'ı commit et
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Güncellenmiş blog post bilgilerini getir
	blogPost, err := r.SelectBlogByID(blogID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated blog post: %w", err)
	}

	return blogPost, nil
}

func (r *Repository) updateBlogPostDetails(tx *sql.Tx, blogID uuid.UUID, input types.BlogUpdateInput) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Update Blog Post Details")

	query := `
		UPDATE blog_posts
		SET group_id = $1, slug = $2, language = $3, featured = $4, updated_at = $5
		WHERE id = $6
	`

	now := time.Now()
	_, err := tx.Exec(
		query,
		input.GroupID,
		input.Slug,
		input.Language,
		input.Featured,
		now,
		blogID,
	)

	if err != nil {
		return fmt.Errorf("error updating blog post details: %w", err)
	}

	return nil
}

func (r *Repository) updateBlogMetadata(tx *sql.Tx, blogID uuid.UUID, metadata types.MetadataInput) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Update Blog Metadata")

	query := `
		UPDATE blog_metadata
		SET title = $1, description = $2, image = $3
		WHERE id = $4
	`

	_, err := tx.Exec(
		query,
		metadata.Title,
		metadata.Description,
		metadata.Image,
		blogID,
	)

	if err != nil {
		return fmt.Errorf("error updating blog metadata: %w", err)
	}

	return nil
}

func (r *Repository) updateBlogContent(tx *sql.Tx, blogID uuid.UUID, content types.ContentInput) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Update Blog Content")

	query := `
		UPDATE blog_content
		SET title = $1, description = $2, image = $3, read_time = $4, html = $5, json = $6
		WHERE id = $7
	`

	_, err := tx.Exec(
		query,
		content.Title,
		content.Description,
		content.Image,
		content.ReadTime,
		content.HTML,
		content.JSON,
		blogID,
	)

	if err != nil {
		return fmt.Errorf("error updating blog content: %w", err)
	}

	return nil
}

func (r *Repository) updateBlogCategories(tx *sql.Tx, blogID uuid.UUID, categories []string) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Update Blog Categories")

	// 1. Önce mevcut kategorileri temizle
	deleteQuery := `DELETE FROM blog_categories WHERE blog_id = $1`
	_, err := tx.Exec(deleteQuery, blogID)
	if err != nil {
		return fmt.Errorf("error deleting existing blog categories: %w", err)
	}

	// 2. Yeni kategorileri ekle (eğer liste boş değilse)
	if len(categories) > 0 {
		for _, categoryValue := range categories {
			insertQuery := `
				INSERT INTO blog_categories (blog_id, category_name)
				VALUES ($1, $2)
				ON CONFLICT (blog_id, category_name) DO NOTHING
			`

			_, err := tx.Exec(insertQuery, blogID, categoryValue)
			if err != nil {
				return fmt.Errorf("error associating blog category (%s): %w", categoryValue, err)
			}
		}
	}

	return nil
}

func (r *Repository) updateBlogTags(tx *sql.Tx, blogID uuid.UUID, tags []string) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Update Blog Tags")

	// 1. Önce mevcut etiketleri temizle
	deleteQuery := `DELETE FROM blog_tags WHERE blog_id = $1`
	_, err := tx.Exec(deleteQuery, blogID)
	if err != nil {
		return fmt.Errorf("error deleting existing blog tags: %w", err)
	}

	// 2. Yeni etiketleri ekle (eğer liste boş değilse)
	if len(tags) > 0 {
		for _, tagValue := range tags {
			insertQuery := `
				INSERT INTO blog_tags (blog_id, tag_name)
				VALUES ($1, $2)
				ON CONFLICT (blog_id, tag_name) DO NOTHING
			`

			_, err := tx.Exec(insertQuery, blogID, tagValue)
			if err != nil {
				return fmt.Errorf("error associating blog tag (%s): %w", tagValue, err)
			}
		}
	}

	return nil
}
