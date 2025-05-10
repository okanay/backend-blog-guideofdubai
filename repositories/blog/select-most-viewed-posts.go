// repositories/blog/select-most-viewed-posts.go
package BlogRepository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectMostViewedPosts(language string, limit int, period string) ([]types.BlogPostCardView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Most Viewed Posts")

	// Başlangıç tarihini belirle (period'a göre)
	var startDate time.Time
	now := time.Now()

	switch period {
	case "day":
		startDate = now.AddDate(0, 0, -1)
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default: // all time
		startDate = time.Time{} // Unix epoch başlangıcı
	}

	// Sorguyu hazırla
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
			CASE WHEN bf.blog_id IS NOT NULL THEN true ELSE false END as featured,
			bs.views,

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
		FROM blog_posts bp
		JOIN blog_content bc ON bp.id = bc.id
		JOIN blog_stats bs ON bp.id = bs.id
		LEFT JOIN blog_featured bf ON bp.id = bf.blog_id AND bf.language = bp.language
		WHERE bp.status = 'published'
	`

	// Filtreleri ekle
	var args []any
	var paramIndex = 1

	// Dil filtresi
	if language != "" {
		query += fmt.Sprintf(" AND bp.language = $%d", paramIndex)
		args = append(args, language)
		paramIndex++
	}

	// Dönem filtresi (eğer "all" değilse)
	if period != "all" {
		query += fmt.Sprintf(" AND (bs.last_viewed_at IS NULL OR bs.last_viewed_at >= $%d)", paramIndex)
		args = append(args, startDate)
		paramIndex++
	}

	// Sıralama ve limit
	query += " ORDER BY bs.views DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIndex)
		args = append(args, limit)
	}

	// Sorguyu çalıştır
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get most viewed posts: %w", err)
	}
	defer rows.Close()

	var blogs []types.BlogPostCardView

	for rows.Next() {
		var blog types.BlogPostCardView
		var content types.ContentCardView
		var views int
		var categoriesJSON, tagsJSON []byte

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
			&views,
			&categoriesJSON,
			&tagsJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning blog row: %w", err)
		}

		blog.Content = content

		// JSON'dan kategori ve tag bilgilerini çözümle
		if err := json.Unmarshal(categoriesJSON, &blog.Categories); err != nil {
			return nil, fmt.Errorf("error unmarshalling categories: %w", err)
		}

		if err := json.Unmarshal(tagsJSON, &blog.Tags); err != nil {
			return nil, fmt.Errorf("error unmarshalling tags: %w", err)
		}

		blogs = append(blogs, blog)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return blogs, nil
}
