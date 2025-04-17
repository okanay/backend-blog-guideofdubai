package BlogRepository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

// CreateBlogPost, blog postunu ve ilişkili bileşenlerini (metadata, content, stats) oluşturur
func (r *Repository) CreateBlogPost(input types.BlogPostCreateInput, userID uuid.UUID) (*types.BlogPostView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Post")

	// Transaction başlat
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("transaction başlatılamadı: %w", err)
	}

	// İşlem başarısız olursa rollback yap
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Blog iskeletini oluştur
	blogID, err := r.CreateBlogSkeleton(tx, input, userID)
	if err != nil {
		return nil, fmt.Errorf("blog iskeleti oluşturulamadı: %w", err)
	}

	// 2. Blog metadata'sını oluştur
	err = r.CreateBlogMetadata(tx, blogID, input.Metadata)
	if err != nil {
		return nil, fmt.Errorf("blog metadata oluşturulamadı: %w", err)
	}

	// 3. Blog içeriğini oluştur
	err = r.CreateBlogContent(tx, blogID, input.Content)
	if err != nil {
		return nil, fmt.Errorf("blog içeriği oluşturulamadı: %w", err)
	}

	// 4. Blog istatistiklerini başlat
	err = r.InitializeBlogStatsForSkeleton(tx, blogID)
	if err != nil {
		return nil, fmt.Errorf("blog istatistikleri başlatılamadı: %w", err)
	}

	// 5. Kategorileri oluştur ve blog ile ilişkilendir
	if len(input.Categories) > 0 {
		err = r.CreateBlogCategories(tx, blogID, input.Categories)
		if err != nil {
			return nil, fmt.Errorf("blog kategorileri oluşturulamadı: %w", err)
		}
	}

	// 6. Etiketleri oluştur ve blog ile ilişkilendir
	if len(input.Tags) > 0 {
		err = r.CreateBlogTags(tx, blogID, input.Tags)
		if err != nil {
			return nil, fmt.Errorf("blog etiketleri oluşturulamadı: %w", err)
		}
	}

	// Transaction'ı commit et
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit edilemedi: %w", err)
	}

	// Oluşturulan blog postunu getir
	blogPost, err := r.SelectBlogByID(blogID)
	if err != nil {
		return nil, fmt.Errorf("oluşturulan blog postu getirilemedi: %w", err)
	}

	return blogPost, nil
}

func (r *Repository) CreateBlogSkeleton(tx *sql.Tx, input types.BlogPostCreateInput, userID uuid.UUID) (uuid.UUID, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Skeleton")

	var blogID uuid.UUID

	// Blog durumu kontrol et - varsayılan olarak taslak
	status := types.BlogStatusDraft

	query := `
		INSERT INTO blog_posts (
			user_id, group_id, slug, language, featured, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id
	`

	now := time.Now()
	err := tx.QueryRow(
		query,
		userID,
		input.GroupID,
		input.Slug,
		input.Language,
		input.Featured,
		status,
		now,
		now,
	).Scan(&blogID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("blog iskeleti eklenirken hata: %w", err)
	}

	return blogID, nil
}

func (r *Repository) CreateBlogMetadata(tx *sql.Tx, blogID uuid.UUID, metadata types.MetadataInput) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Metadata")

	query := `
		INSERT INTO blog_metadata (
			id, title, description, image, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`

	now := time.Now()
	_, err := tx.Exec(
		query,
		blogID,
		metadata.Title,
		metadata.Description,
		metadata.Image,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("blog metadata eklenirken hata: %w", err)
	}

	return nil
}

func (r *Repository) CreateBlogContent(tx *sql.Tx, blogID uuid.UUID, content types.ContentInput) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Content")

	query := `
		INSERT INTO blog_content (
			id, title, description, read_time, html, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	now := time.Now()
	_, err := tx.Exec(
		query,
		blogID,
		content.Title,
		content.Description,
		content.ReadTime,
		content.HTML,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("blog içeriği eklenirken hata: %w", err)
	}

	return nil
}

func (r *Repository) CreateBlogCategories(tx *sql.Tx, blogID uuid.UUID, categories []string) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Categories")

	// Her kategori için ilişkilendirme ekle
	for _, categoryValue := range categories {
		query := `
			INSERT INTO blog_categories (blog_id, category_name)
			VALUES ($1, $2)
			ON CONFLICT (blog_id, category_name) DO NOTHING
		`

		_, err := tx.Exec(query, blogID, categoryValue)
		if err != nil {
			return fmt.Errorf("blog kategorisi ilişkilendirilirken hata (%s): %w", categoryValue, err)
		}
	}

	return nil
}

func (r *Repository) CreateBlogTags(tx *sql.Tx, blogID uuid.UUID, tags []string) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Tags")

	// Her etiket için ilişkilendirme ekle
	for _, tagValue := range tags {
		query := `
			INSERT INTO blog_tags (blog_id, tag_name)
			VALUES ($1, $2)
			ON CONFLICT (blog_id, tag_name) DO NOTHING
		`

		_, err := tx.Exec(query, blogID, tagValue)
		if err != nil {
			return fmt.Errorf("blog etiketi ilişkilendirilirken hata (%s): %w", tagValue, err)
		}
	}

	return nil
}

func (r *Repository) InitializeBlogStatsForSkeleton(tx *sql.Tx, blogID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Initialize Blog Stats")

	query := `
		INSERT INTO blog_stats (
			id, views, likes, shares, comments, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	now := time.Now()
	_, err := tx.Exec(
		query,
		blogID,
		0, // views
		0, // likes
		0, // shares
		0, // comments
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("blog istatistikleri başlatılırken hata: %w", err)
	}

	return nil
}
