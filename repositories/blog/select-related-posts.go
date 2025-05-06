package BlogRepository

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectRelatedPosts(
	excludeBlogID uuid.UUID,
	categories []string,
	tags []string,
	language string,
	limit int,
) ([]types.BlogPostCardView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Get Related Posts")

	// İlgili blog yazılarını almak için temel sorgu
	baseQuery := `
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

            -- Kategorileri JSON dizisi olarak al
            (
                SELECT COALESCE(json_agg(json_build_object('name', c.name, 'value', c.value)), '[]'::json)
                FROM blog_categories bc2
                JOIN categories c ON bc2.category_name = c.name
                WHERE bc2.blog_id = bp.id
            ) AS categories,

            -- Benzerlik skoru hesapla
            (
                CASE WHEN ARRAY_LENGTH($1::text[], 1) > 0 THEN
                    (SELECT COUNT(*) FROM blog_categories bc_match
                     WHERE bc_match.blog_id = bp.id
                     AND bc_match.category_name = ANY($1::text[])) * 10
                ELSE 0 END
                +
                CASE WHEN ARRAY_LENGTH($2::text[], 1) > 0 THEN
                    (SELECT COUNT(*) FROM blog_tags bt_match
                     WHERE bt_match.blog_id = bp.id
                     AND bt_match.tag_name = ANY($2::text[])) * 5
                ELSE 0 END
            ) AS match_score
        FROM blog_posts bp
        LEFT JOIN blog_content bc ON bp.id = bc.id
        LEFT JOIN blog_featured bf ON bp.id = bf.blog_id AND bf.language = bp.language
        WHERE bp.id != $3
        AND bp.status = 'published'
    `

	var query string
	var params []any
	var relatedPosts []types.BlogPostCardView

	// İlk sorgu: Kategori veya etiketlerle eşleşen bloglar
	if len(categories) > 0 || len(tags) > 0 {
		query = baseQuery + `
            AND (
                ($1::text[] IS NOT NULL AND ARRAY_LENGTH($1::text[], 1) > 0 AND EXISTS (
                    SELECT 1 FROM blog_categories bc_match
                    WHERE bc_match.blog_id = bp.id
                    AND bc_match.category_name = ANY($1::text[])
                ))
                OR
                ($2::text[] IS NOT NULL AND ARRAY_LENGTH($2::text[], 1) > 0 AND EXISTS (
                    SELECT 1 FROM blog_tags bt_match
                    WHERE bt_match.blog_id = bp.id
                    AND bt_match.tag_name = ANY($2::text[])
                ))
            )
        `

		// Dil filtresi
		if language != "" {
			query += " AND bp.language = $4"
			query += " ORDER BY match_score DESC, bp.created_at DESC LIMIT $5"
			params = []any{pq.Array(categories), pq.Array(tags), excludeBlogID, language, limit}
		} else {
			query += " ORDER BY match_score DESC, bp.created_at DESC LIMIT $4"
			params = []any{pq.Array(categories), pq.Array(tags), excludeBlogID, limit}
		}

		// İlk sorguyu çalıştır
		matchedPosts, err := r.runRelatedPostsQuery(query, params)
		if err != nil {
			return nil, err
		}
		relatedPosts = append(relatedPosts, matchedPosts...)

		// Eğer yeterli sonuç bulduysak direkt döndür
		if len(relatedPosts) >= limit {
			return relatedPosts[:limit], nil
		}
	}

	// İkinci sorgu: Sadece dil bazında eşleşen bloglar (eğer ilk sorguda yeterli sonuç bulunamadıysa)
	if language != "" && len(relatedPosts) < limit {
		remainingLimit := limit - len(relatedPosts)
		query = baseQuery + " AND bp.language = $4"

		// İlk sorguda bulunan blog ID'lerini hariç tut
		if len(relatedPosts) > 0 {
			var excludeIDs []uuid.UUID
			excludeIDs = append(excludeIDs, excludeBlogID)

			for _, post := range relatedPosts {
				postID, err := uuid.Parse(post.ID)
				if err == nil {
					excludeIDs = append(excludeIDs, postID)
				}
			}

			query += " AND bp.id != ALL($5::uuid[])"
			query += " ORDER BY bp.created_at DESC LIMIT $6"
			params = []any{pq.Array([]string{}), pq.Array([]string{}), excludeBlogID, language, pq.Array(excludeIDs), remainingLimit}
		} else {
			query += " ORDER BY bp.created_at DESC LIMIT $5"
			params = []any{pq.Array([]string{}), pq.Array([]string{}), excludeBlogID, language, remainingLimit}
		}

		// İkinci sorguyu çalıştır
		languageMatchedPosts, err := r.runRelatedPostsQuery(query, params)
		if err == nil {
			relatedPosts = append(relatedPosts, languageMatchedPosts...)
		}

		// Eğer yeterli sonuç bulduysak direkt döndür
		if len(relatedPosts) >= limit {
			return relatedPosts[:limit], nil
		}
	}

	// Üçüncü sorgu: Herhangi bir dildeki en son bloglar (eğer hala yeterli sonuç bulunamadıysa)
	if len(relatedPosts) < limit {
		remainingLimit := limit - len(relatedPosts)
		query = baseQuery

		// Daha önce bulunan blog ID'lerini hariç tut
		if len(relatedPosts) > 0 {
			var excludeIDs []uuid.UUID
			excludeIDs = append(excludeIDs, excludeBlogID)

			for _, post := range relatedPosts {
				postID, err := uuid.Parse(post.ID)
				if err == nil {
					excludeIDs = append(excludeIDs, postID)
				}
			}

			query += " AND bp.id != ALL($4::uuid[])"
			query += " ORDER BY bp.created_at DESC LIMIT $5"
			params = []any{pq.Array([]string{}), pq.Array([]string{}), excludeBlogID, pq.Array(excludeIDs), remainingLimit}
		} else {
			query += " ORDER BY bp.created_at DESC LIMIT $4"
			params = []any{pq.Array([]string{}), pq.Array([]string{}), excludeBlogID, remainingLimit}
		}

		// Üçüncü sorguyu çalıştır
		anyLanguagePosts, err := r.runRelatedPostsQuery(query, params)
		if err == nil {
			relatedPosts = append(relatedPosts, anyLanguagePosts...)
		}
	}

	// Limit'e göre sonuçları kırp
	if len(relatedPosts) > limit {
		relatedPosts = relatedPosts[:limit]
	}

	return relatedPosts, nil
}

// Yardımcı fonksiyon: İlgili blog sorgusunu çalıştır ve sonuçları işle
func (r *Repository) runRelatedPostsQuery(query string, params []any) ([]types.BlogPostCardView, error) {
	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relatedPosts []types.BlogPostCardView

	for rows.Next() {
		var card types.BlogPostCardView
		var content types.ContentCardView
		var categoriesJSON []byte
		var matchScore int

		err := rows.Scan(
			&card.ID,
			&card.GroupID,
			&card.Slug,
			&card.Language,
			&card.Status,
			&card.CreatedAt,
			&card.UpdatedAt,
			&content.Title,
			&content.Description,
			&content.Image,
			&content.ReadTime,
			&card.Featured,
			&categoriesJSON,
			&matchScore,
		)
		if err != nil {
			continue // Hatalı satırı atla
		}

		card.Content = content

		// JSON kategorileri çöz
		var categories []types.CategoryView
		if err := json.Unmarshal(categoriesJSON, &categories); err == nil {
			card.Categories = categories
		}

		relatedPosts = append(relatedPosts, card)
	}

	return relatedPosts, nil
}
