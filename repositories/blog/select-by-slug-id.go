// repositories/blog/select-by-group-id.go
package BlogRepository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectBlogBySlugID(request types.BlogSelectByGroupIDInput) (*types.BlogPostView, []*types.BlogPostView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog And Alternatives")

	// 1. Slug veya groupID ile doğrudan blog post arama
	query := `
        WITH target_post AS (
            SELECT id, group_id
            FROM blog_posts
            WHERE (slug = $1 OR group_id = $1)
              AND status = 'published'
            ORDER BY
                CASE WHEN slug = $1 THEN 0 ELSE 1 END, -- Slug eşleşmesine öncelik ver
                CASE WHEN $2 != '' AND language = $2 THEN 0 ELSE 1 END, -- Dil eşleşmesine sonra öncelik ver
                created_at DESC
            LIMIT 1
        )
        SELECT
            -- Ana blog bilgileri
            bp.id,
            bp.group_id,
            bp.slug,
            bp.language,
            bp.status,
            bp.created_at,
            bp.updated_at,
            bp.published_at,

            -- Featured durumu
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

            -- Kategorileri JSON dizisi olarak al
            (
                SELECT COALESCE(json_agg(json_build_object('name', c.name, 'value', c.value)), '[]'::json)
                FROM blog_categories bc2
                JOIN categories c ON bc2.category_name = c.name
                WHERE bc2.blog_id = bp.id
            ) AS categories,

            -- Etiketleri JSON dizisi olarak al
            (
                SELECT COALESCE(json_agg(json_build_object('name', t.name, 'value', t.value)), '[]'::json)
                FROM blog_tags bt
                JOIN tags t ON bt.tag_name = t.name
                WHERE bt.blog_id = bp.id
            ) AS tags

        FROM target_post tp
        JOIN blog_posts bp ON tp.id = bp.id
        LEFT JOIN blog_metadata bm ON bp.id = bm.id
        LEFT JOIN blog_content bc ON bp.id = bc.id
        LEFT JOIN blog_stats bs ON bp.id = bs.id
        LEFT JOIN blog_featured bf ON bp.id = bf.blog_id AND bf.language = bp.language
    `

	// Ana post için veritabanı sorgusunu yap
	var mainPost types.BlogPostView
	var metadata types.MetadataView
	var content types.ContentView
	var stats types.StatsView
	var publishedAt, lastViewedAt sql.NullTime
	var metaDesc, metaImage, contentDesc sql.NullString
	var categoriesJSON, tagsJSON []byte
	var groupID string

	err := r.db.QueryRow(query, request.SlugOrGroupID, request.Language).Scan(
		&mainPost.ID,
		&groupID,
		&mainPost.Slug,
		&mainPost.Language,
		&mainPost.Status,
		&mainPost.CreatedAt,
		&mainPost.UpdatedAt,
		&publishedAt,
		&mainPost.Featured,

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

	// Hata kontrolü
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("blog post not found with slug or groupId=%s", request.SlugOrGroupID)
		}
		return nil, nil, fmt.Errorf("error retrieving blog data: %w", err)
	}

	// Nullable alanları işle
	if publishedAt.Valid {
		mainPost.PublishedAt = publishedAt.Time
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

	// JSON verileri çöz
	var categories []types.CategoryView
	var tags []types.TagView

	if err := json.Unmarshal(categoriesJSON, &categories); err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling categories: %w", err)
	}
	if err := json.Unmarshal(tagsJSON, &tags); err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling tags: %w", err)
	}

	// Ana yapıya alt yapıları ata
	mainPost.Metadata = metadata
	mainPost.Content = content
	mainPost.Stats = stats
	mainPost.Categories = categories
	mainPost.Tags = tags
	mainPost.GroupID = groupID

	// 2. Aynı groupID'ye sahip alternatif blog yazılarını al
	alternativeQuery := `
        WITH alt_posts AS (
            SELECT id
            FROM blog_posts
            WHERE group_id = $1
              AND id != $2
              AND status = 'published'
        )
        SELECT
            bp.id,
            bp.group_id,
            bp.slug,
            bp.language,
            bp.status,
            bp.created_at,
            bp.updated_at,
            bp.published_at,

            CASE WHEN bf.blog_id IS NOT NULL THEN true ELSE false END as featured,

            bm.title as meta_title,
            bm.description as meta_description,
            bm.image as meta_image,

            bc.title as content_title,
            bc.description as content_description,
            bc.image as content_image,
            bc.read_time,
            bc.html,
            bc.json,

            bs.views,
            bs.likes,
            bs.shares,
            bs.comments,
            bs.last_viewed_at,

            (
                SELECT COALESCE(json_agg(json_build_object('name', c.name, 'value', c.value)), '[]'::json)
                FROM blog_categories bc2
                JOIN categories c ON bc2.category_name = c.name
                WHERE bc2.blog_id = bp.id
            ) AS categories,

            (
                SELECT COALESCE(json_agg(json_build_object('name', t.name, 'value', t.value)), '[]'::json)
                FROM blog_tags bt
                JOIN tags t ON bt.tag_name = t.name
                WHERE bt.blog_id = bp.id
            ) AS tags

        FROM alt_posts ap
        JOIN blog_posts bp ON ap.id = bp.id
        LEFT JOIN blog_metadata bm ON bp.id = bm.id
        LEFT JOIN blog_content bc ON bp.id = bc.id
        LEFT JOIN blog_stats bs ON bp.id = bs.id
        LEFT JOIN blog_featured bf ON bp.id = bf.blog_id AND bf.language = bp.language
    `

	// Alternatifler için sorguyu çalıştır
	rows, err := r.db.Query(alternativeQuery, groupID, mainPost.ID)
	if err != nil {
		// Ana post bulunmuşsa ama alternatifler bulunamazsa, sadece ana postu döndür
		return &mainPost, nil, nil
	}
	defer rows.Close()

	var alternatives []*types.BlogPostView

	// Alternatif blogları işle
	for rows.Next() {
		var alt types.BlogPostView
		var altMetadata types.MetadataView
		var altContent types.ContentView
		var altStats types.StatsView
		var altPublishedAt, altLastViewedAt sql.NullTime
		var altMetaDesc, altMetaImage, altContentDesc sql.NullString
		var altCategoriesJSON, altTagsJSON []byte

		err := rows.Scan(
			&alt.ID,
			&alt.GroupID,
			&alt.Slug,
			&alt.Language,
			&alt.Status,
			&alt.CreatedAt,
			&alt.UpdatedAt,
			&altPublishedAt,
			&alt.Featured,

			&altMetadata.Title,
			&altMetaDesc,
			&altMetaImage,

			&altContent.Title,
			&altContentDesc,
			&altContent.Image,
			&altContent.ReadTime,
			&altContent.HTML,
			&altContent.JSON,

			&altStats.Views,
			&altStats.Likes,
			&altStats.Shares,
			&altStats.Comments,
			&altLastViewedAt,

			&altCategoriesJSON,
			&altTagsJSON,
		)

		if err != nil {
			// Hatalı satırı atla
			continue
		}

		// Nullable alanları işle
		if altPublishedAt.Valid {
			alt.PublishedAt = altPublishedAt.Time
		}
		if altMetaDesc.Valid {
			altMetadata.Description = altMetaDesc.String
		}
		if altMetaImage.Valid {
			altMetadata.Image = altMetaImage.String
		}
		if altContentDesc.Valid {
			altContent.Description = altContentDesc.String
		}
		if altLastViewedAt.Valid {
			altStats.LastViewedAt = &altLastViewedAt.Time
		}

		// JSON verileri çöz
		var altCategories []types.CategoryView
		var altTags []types.TagView

		if err := json.Unmarshal(altCategoriesJSON, &altCategories); err == nil {
			alt.Categories = altCategories
		}
		if err := json.Unmarshal(altTagsJSON, &altTags); err == nil {
			alt.Tags = altTags
		}

		// Alt yapıları ata
		alt.Metadata = altMetadata
		alt.Content = altContent
		alt.Stats = altStats

		alternatives = append(alternatives, &alt)
	}

	// Satır hata kontrolü
	if err = rows.Err(); err != nil {
		// Ana post bulunmuşsa ama satır işlemede hata oluştuysa, yine de ana postu döndür
		return &mainPost, alternatives, nil
	}

	return &mainPost, alternatives, nil
}
