package BlogRepository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectBlogByID(blogID uuid.UUID) (*types.BlogPostView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog By ID")

	query := `
        SELECT
            -- Blog Post primary data
            bp.id,
            bp.group_id,
            bp.slug,
            bp.language,
            bp.status,
            bp.created_at,
            bp.updated_at,
            bp.published_at,

            -- Featured status
            CASE WHEN bf.blog_id IS NOT NULL THEN true ELSE false END as featured,

            -- Metadata
            bm.title as meta_title,
            bm.description as meta_description,
            bm.image as meta_image,

            -- Content
            bc.title as content_title,
            bc.description as content_description,
            bc.image as content_image,
            bc.read_time,
            bc.html,
            bc.json,

            -- Statistics
            bs.views,
            bs.likes,
            bs.shares,
            bs.comments,
            bs.last_viewed_at,

            -- Categories as JSON array
            (
                SELECT COALESCE(json_agg(json_build_object('name', c.name, 'value', c.value)), '[]'::json)
                FROM blog_categories bc2
                JOIN categories c ON bc2.category_name = c.name
                WHERE bc2.blog_id = bp.id
            ) AS categories,

            -- Tags as JSON array
            (
                SELECT COALESCE(json_agg(json_build_object('name', t.name, 'value', t.value)), '[]'::json)
                FROM blog_tags bt
                JOIN tags t ON bt.tag_name = t.name
                WHERE bt.blog_id = bp.id
            ) AS tags

        FROM blog_posts bp
        LEFT JOIN blog_metadata bm ON bp.id = bm.id
        LEFT JOIN blog_content bc ON bp.id = bc.id
        LEFT JOIN blog_stats bs ON bp.id = bs.id
        LEFT JOIN blog_featured bf ON bp.id = bf.blog_id AND bf.language = bp.language
        WHERE bp.id = $1`

	var blog types.BlogPostView
	var metadata types.MetadataView
	var content types.ContentView
	var stats types.StatsView
	var publishedAt, lastViewedAt sql.NullTime
	var metaDesc, metaImage, contentDesc sql.NullString
	var categoriesJSON, tagsJSON []byte

	err := r.db.QueryRow(query, blogID).Scan(
		&blog.ID,
		&blog.GroupID,
		&blog.Slug,
		&blog.Language,
		&blog.Status,
		&blog.CreatedAt,
		&blog.UpdatedAt,
		&publishedAt,
		&blog.Featured,

		&metadata.Title,
		&metaDesc,
		&metaImage,

		&content.Title,
		&contentDesc,
		&content.Image,
		&content.ReadTime,
		&content.HTML,
		&content.JSON,

		&stats.Views,
		&stats.Likes,
		&stats.Shares,
		&stats.Comments,
		&lastViewedAt,

		&categoriesJSON,
		&tagsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog post not found: %w", err)
		}
		return nil, fmt.Errorf("error retrieving blog data: %w", err)
	}

	// Nullable alanları işle
	if publishedAt.Valid {
		blog.PublishedAt = publishedAt.Time
	}
	if metaDesc.Valid {
		metadata.Description = metaDesc.String
	}
	if metaImage.Valid {
		metadata.Image = metaImage.String
	}
	if contentDesc.Valid {
		content.Description = contentDesc.String
	}
	if lastViewedAt.Valid {
		stats.LastViewedAt = &lastViewedAt.Time
	}

	// JSON kategorileri ve etiketleri çöz
	var categories []types.CategoryView
	var tags []types.TagView

	if err := json.Unmarshal(categoriesJSON, &categories); err != nil {
		return nil, fmt.Errorf("error unmarshalling categories: %w", err)
	}
	if err := json.Unmarshal(tagsJSON, &tags); err != nil {
		return nil, fmt.Errorf("error unmarshalling tags: %w", err)
	}

	// Ana yapıya alt yapıları ata
	blog.Metadata = metadata
	blog.Content = content
	blog.Stats = stats
	blog.Categories = categories
	blog.Tags = tags

	return &blog, nil
}

func (r *Repository) SelectBlogBody(blogID uuid.UUID) (*types.BlogPostView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Body Skeleton")

	query := `
        SELECT
            -- Blog Post primary data (featured kaldırıldı)
            bp.id,
            bp.group_id,
            bp.slug,
            bp.language,
            bp.status,
            bp.created_at,
            bp.updated_at,
            bp.published_at,

            -- Featured durumu artık blog_featured tablosundan alınacak
            CASE WHEN bf.blog_id IS NOT NULL THEN true ELSE false END as featured,

            -- Metadata
            bm.title as meta_title,
            bm.description as meta_description,
            bm.image as meta_image,

            -- Content
            bc.title as content_title,
            bc.description as content_description,
            bc.image as content_image,
            bc.read_time,
            bc.html,
            bc.json,

            -- Statistics
            bs.views,
            bs.likes,
            bs.shares,
            bs.comments,
            bs.last_viewed_at
        FROM blog_posts bp
        LEFT JOIN blog_metadata bm ON bp.id = bm.id
        LEFT JOIN blog_content bc ON bp.id = bc.id
        LEFT JOIN blog_stats bs ON bp.id = bs.id
        LEFT JOIN blog_featured bf ON bp.id = bf.blog_id AND bf.language = bp.language
        WHERE bp.id = $1`

	var blog types.BlogPostView
	var metadata types.MetadataView
	var content types.ContentView
	var stats types.StatsView
	var publishedAt, lastViewedAt sql.NullTime
	var metaDesc, metaImage, contentDesc sql.NullString

	err := r.db.QueryRow(query, blogID).Scan(
		&blog.ID,
		&blog.GroupID,
		&blog.Slug,
		&blog.Language,
		&blog.Status,
		&blog.CreatedAt,
		&blog.UpdatedAt,
		&publishedAt,
		&blog.Featured, // Bu artık hesaplanıyor

		&metadata.Title,
		&metaDesc,
		&metaImage,

		&content.Title,
		&contentDesc,
		&content.Image,
		&content.ReadTime,
		&content.HTML,
		&content.JSON,

		&stats.Views,
		&stats.Likes,
		&stats.Shares,
		&stats.Comments,
		&lastViewedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog post not found: %w", err)
		}
		return nil, fmt.Errorf("error retrieving blog data: %w", err)
	}

	// Handle nullable fields
	if publishedAt.Valid {
		blog.PublishedAt = publishedAt.Time
	}

	if metaDesc.Valid {
		metadata.Description = metaDesc.String
	}
	if metaImage.Valid {
		metadata.Image = metaImage.String
	}

	if contentDesc.Valid {
		content.Description = contentDesc.String
	}

	if lastViewedAt.Valid {
		stats.LastViewedAt = &lastViewedAt.Time
	}

	// Assign sub-structures to main structure
	blog.Metadata = metadata
	blog.Content = content
	blog.Stats = stats

	return &blog, nil
}

func (r *Repository) SelectBlogCategories(blogID uuid.UUID) ([]types.CategoryView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog Categories")

	var categories []types.CategoryView

	query := `
		SELECT c.name, c.value
		FROM categories c
		JOIN blog_categories bc ON c.name = bc.category_name
		WHERE bc.blog_id = $1
	`

	rows, err := r.db.Query(query, blogID)
	if err != nil {
		return categories, fmt.Errorf("error retrieving blog categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category types.CategoryView
		if err := rows.Scan(&category.Name, &category.Value); err != nil {
			return categories, fmt.Errorf("error scanning category data: %w", err)
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return categories, fmt.Errorf("error processing category rows: %w", err)
	}

	return categories, nil
}

func (r *Repository) SelectBlogTags(blogID uuid.UUID) ([]types.TagView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog Tags")

	var tags []types.TagView

	query := `
		SELECT t.name, t.value
		FROM tags t
		JOIN blog_tags bt ON t.name = bt.tag_name
		WHERE bt.blog_id = $1
	`

	rows, err := r.db.Query(query, blogID)
	if err != nil {
		return tags, fmt.Errorf("error retrieving blog tags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tag types.TagView
		if err := rows.Scan(&tag.Name, &tag.Value); err != nil {
			return tags, fmt.Errorf("error scanning tag data: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return tags, fmt.Errorf("error processing tag rows: %w", err)
	}

	return tags, nil
}

// Deprecated: Use SelectBodySkeleton instead.
func (r *Repository) SelectBlogBase(blogID uuid.UUID) (types.BlogPost, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog Base")

	var blog types.BlogPost
	var publishedAt sql.NullTime

	query := `
		SELECT id, group_id, slug, language, status, created_at, updated_at, published_at
		FROM blog_posts
		WHERE id = $1
	`

	err := r.db.QueryRow(query, blogID).Scan(
		&blog.ID,
		&blog.GroupID,
		&blog.Slug,
		&blog.Language,
		&blog.Status,
		&blog.CreatedAt,
		&blog.UpdatedAt,
		&publishedAt,
	)
	if err != nil {
		return blog, fmt.Errorf("error retrieving blog post base data: %w", err)
	}

	if publishedAt.Valid {
		blog.PublishedAt = publishedAt.Time
	} else {
		blog.PublishedAt = time.Time{}
	}

	return blog, nil
}

// Deprecated: Use SelectBodySkeleton instead.
func (r *Repository) SelectBlogMetadata(blogID uuid.UUID) (types.MetadataView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog Metadata")

	var metadata types.MetadataView
	var metaDesc, metaImage sql.NullString

	query := `
		SELECT title, description, image
		FROM blog_metadata
		WHERE id = $1
	`

	err := r.db.QueryRow(query, blogID).Scan(
		&metadata.Title,
		&metaDesc,
		&metaImage,
	)
	if err != nil {
		return metadata, fmt.Errorf("error retrieving blog metadata: %w", err)
	}

	if metaDesc.Valid {
		metadata.Description = metaDesc.String
	}
	if metaImage.Valid {
		metadata.Image = metaImage.String
	}

	return metadata, nil
}

// Deprecated: Use SelectBodySkeleton instead.
func (r *Repository) SelectBlogContent(blogID uuid.UUID) (types.ContentView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog Content")

	var content types.ContentView
	var contentDesc sql.NullString

	query := `
		SELECT title, description, read_time, html, json
		FROM blog_content
		WHERE id = $1
	`

	err := r.db.QueryRow(query, blogID).Scan(
		&content.Title,
		&contentDesc,
		&content.ReadTime,
		&content.HTML,
		&content.JSON,
	)
	if err != nil {
		return content, fmt.Errorf("error retrieving blog content: %w", err)
	}

	if contentDesc.Valid {
		content.Description = contentDesc.String
	}

	return content, nil
}

// Deprecated: Use SelectBodySkeleton instead.
func (r *Repository) SelectBlogStats(blogID uuid.UUID) (types.StatsView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog Stats")

	var stats types.StatsView
	var lastViewedAt sql.NullTime

	query := `
		SELECT views, likes, shares, comments, last_viewed_at
		FROM blog_stats
		WHERE id = $1
	`

	err := r.db.QueryRow(query, blogID).Scan(
		&stats.Views,
		&stats.Likes,
		&stats.Shares,
		&stats.Comments,
		&lastViewedAt,
	)
	if err != nil {
		return stats, fmt.Errorf("error retrieving blog statistics: %w", err)
	}

	if lastViewedAt.Valid {
		stats.LastViewedAt = &lastViewedAt.Time
	} else {
		stats.LastViewedAt = nil
	}

	return stats, nil
}
