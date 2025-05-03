package BlogRepository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) AddToFeaturedList(blogID uuid.UUID, language string) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Add To Featured List")

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Blog'un var olduğunu ve published durumunda olduğunu kontrol et
	var exists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM blog_posts
			WHERE id = $1 AND status = 'published'
		)
	`
	err = tx.QueryRow(checkQuery, blogID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check blog existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("blog not found or not published")
	}

	// 2. Mevcut en büyük position değerini bul
	var maxPosition sql.NullInt64
	positionQuery := `
		SELECT MAX(position)
		FROM blog_featured
		WHERE language = $1
	`
	err = tx.QueryRow(positionQuery, language).Scan(&maxPosition)
	if err != nil {
		return fmt.Errorf("failed to get max position: %w", err)
	}

	// 3. Yeni position değerini hesapla (100'er artışla)
	newPosition := 0
	if maxPosition.Valid {
		newPosition = int(maxPosition.Int64) + 100
	}

	// 4. Featured tablosuna ekle
	insertQuery := `
		INSERT INTO blog_featured (blog_id, language, position)
		VALUES ($1, $2, $3)
		ON CONFLICT (blog_id, language)
		DO UPDATE SET position = EXCLUDED.position, updated_at = NOW()
	`
	_, err = tx.Exec(insertQuery, blogID, language, newPosition)
	if err != nil {
		return fmt.Errorf("failed to insert into featured list: %w", err)
	}

	return tx.Commit()
}

func (r *Repository) RemoveFromFeaturedList(blogID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Remove From All Featured Lists")

	query := `
		DELETE FROM blog_featured
		WHERE blog_id = $1
	`

	_, err := r.db.Exec(query, blogID)
	if err != nil {
		return fmt.Errorf("failed to remove from all featured lists: %w", err)
	}

	return nil
}

func (r *Repository) UpdateFeaturedOrdering(language string, orderedBlogIDs []uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Update Featured Ordering")

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Her blog için pozisyonu güncelle
	updateQuery := `
		UPDATE blog_featured
		SET position = $1, updated_at = NOW()
		WHERE blog_id = $2 AND language = $3
	`

	for i, blogID := range orderedBlogIDs {
		position := i * 100 // 0, 100, 200, 300...

		result, err := tx.Exec(updateQuery, position, blogID, language)
		if err != nil {
			return fmt.Errorf("failed to update position for blog %s: %w", blogID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("blog %s not found in featured list for language %s", blogID, language)
		}
	}

	return tx.Commit()
}

func (r *Repository) GetFeaturedBlogs(language string, limit int) ([]types.BlogPostCardView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Get Featured Blogs")

	query := `
		SELECT
			bp.id,
			bp.group_id,
			bp.slug,
			bp.language,
			bp.status,
			bp.created_at,
			bp.updated_at,
			bc.title,
			bc.description,
			bc.image,
			bc.read_time,
			true as featured,
			bf.position
		FROM blog_posts bp
		JOIN blog_content bc ON bp.id = bc.id
		JOIN blog_featured bf ON bp.id = bf.blog_id
		WHERE bf.language = $1
		  AND bp.status = 'published'
		ORDER BY bf.position ASC
		LIMIT $2
	`

	rows, err := r.db.Query(query, language, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured blogs: %w", err)
	}
	defer rows.Close()

	var blogs []types.BlogPostCardView
	for rows.Next() {
		var blog types.BlogPostCardView
		var content types.ContentCardView
		var position int // Sadece sıralama için kullanılacak, response'a dahil edilmeyecek

		err := rows.Scan(
			&blog.ID,
			&blog.GroupID,
			&blog.Slug,
			&blog.Language,
			&blog.Status,
			&blog.CreatedAt,
			&blog.UpdatedAt,
			&content.Title,
			&content.Description,
			&content.Image,
			&content.ReadTime,
			&blog.Featured,
			&position,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blog: %w", err)
		}

		blog.Content = content

		// Kategorileri ekle
		blogUUID, err := uuid.Parse(blog.ID)
		if err == nil {
			categories, err := r.SelectBlogCategories(blogUUID)
			if err == nil && len(categories) > 0 {
				blog.Categories = categories
			}
		}

		blogs = append(blogs, blog)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return blogs, nil
}

func (r *Repository) IsBlogFeatured(blogID uuid.UUID, language string) (bool, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Is Blog Featured")

	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM blog_featured
			WHERE blog_id = $1 AND language = $2
		)
	`

	err := r.db.QueryRow(query, blogID, language).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check featured status: %w", err)
	}

	return exists, nil
}
