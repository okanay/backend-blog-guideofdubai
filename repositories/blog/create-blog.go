package BlogRepository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) CreateBlogPost(input types.BlogPostCreateInput, userID uuid.UUID) (*types.BlogPostView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Post")

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to initiate transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Create blog skeleton
	blogID, err := r.CreateBlogSkeleton(tx, input, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create blog skeleton: %w", err)
	}

	// 2. Create blog metadata
	err = r.CreateBlogMetadata(tx, blogID, input.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create blog metadata: %w", err)
	}

	// 3. Create blog content
	err = r.CreateBlogContent(tx, blogID, input.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to create blog content: %w", err)
	}

	// 4. Initialize blog statistics
	err = r.InitializeBlogStatsForSkeleton(tx, blogID)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize blog statistics: %w", err)
	}

	// 5. Create and associate categories
	if len(input.Categories) > 0 {
		err = r.CreateBlogCategories(tx, blogID, input.Categories)
		if err != nil {
			return nil, fmt.Errorf("failed to create blog categories: %w", err)
		}
	}

	// 6. Create and associate tags
	if len(input.Tags) > 0 {
		err = r.CreateBlogTags(tx, blogID, input.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to create blog tags: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	blogPost, err := r.SelectBlogByID(blogID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created blog post: %w", err)
	}

	return blogPost, nil
}

func (r *Repository) CreateBlogSkeleton(tx *sql.Tx, input types.BlogPostCreateInput, userID uuid.UUID) (uuid.UUID, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Skeleton")

	var blogID uuid.UUID
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
		return uuid.Nil, fmt.Errorf("error creating blog skeleton: %w", err)
	}

	return blogID, nil
}

func (r *Repository) CreateBlogMetadata(tx *sql.Tx, blogID uuid.UUID, metadata types.MetadataInput) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Metadata")

	query := `
		INSERT INTO blog_metadata (
			id, title, description, image
		) VALUES (
			$1, $2, $3, $4
		)
	`

	_, err := tx.Exec(
		query,
		blogID,
		metadata.Title,
		metadata.Description,
		metadata.Image,
	)

	if err != nil {
		return fmt.Errorf("error creating blog metadata: %w", err)
	}

	return nil
}

func (r *Repository) CreateBlogContent(tx *sql.Tx, blogID uuid.UUID, content types.ContentInput) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Content")

	query := `
		INSERT INTO blog_content (
			id, title, description, image, read_time, json
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`

	_, err := tx.Exec(
		query,
		blogID,
		content.Title,
		content.Description,
		content.Image,
		content.ReadTime,
		content.JSON,
	)

	if err != nil {
		return fmt.Errorf("error creating blog content: %w", err)
	}

	return nil
}

func (r *Repository) CreateBlogCategories(tx *sql.Tx, blogID uuid.UUID, categories []string) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Categories")
	fmt.Println(categories)

	for _, categoryValue := range categories {
		query := `
			INSERT INTO blog_categories (blog_id, category_name)
			VALUES ($1, $2)
			ON CONFLICT (blog_id, category_name) DO NOTHING
		`

		_, err := tx.Exec(query, blogID, categoryValue)
		if err != nil {
			return fmt.Errorf("error associating blog category (%s): %w", categoryValue, err)
		}
	}

	return nil
}

func (r *Repository) CreateBlogTags(tx *sql.Tx, blogID uuid.UUID, tags []string) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Create Blog Tags")

	for _, tagValue := range tags {
		query := `
			INSERT INTO blog_tags (blog_id, tag_name)
			VALUES ($1, $2)
			ON CONFLICT (blog_id, tag_name) DO NOTHING
		`

		_, err := tx.Exec(query, blogID, tagValue)
		if err != nil {
			return fmt.Errorf("error associating blog tag (%s): %w", tagValue, err)
		}
	}

	return nil
}

func (r *Repository) InitializeBlogStatsForSkeleton(tx *sql.Tx, blogID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Initialize Blog Stats")

	query := `
		INSERT INTO blog_stats (
			id, views, likes, shares, comments
		) VALUES (
			$1, $2, $3, $4, $5
		)
	`

	_, err := tx.Exec(
		query,
		blogID,
		0, // views
		0, // likes
		0, // shares
		0, // comments
	)

	if err != nil {
		return fmt.Errorf("error initializing blog statistics: %w", err)
	}

	return nil
}
