package BlogRepository

import (
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

	query := `
        SELECT
            bp.id,
            bp.group_id,
            bp.slug,
            bp.language,
            bp.featured,
            bp.status,
            bp.created_at,
            bp.updated_at,
            bc.title,
            bc.description,
            bc.image,
            bc.read_time,
            (
                CASE WHEN ARRAY_LENGTH($1::text[], 1) > 0 THEN
                    (SELECT COUNT(*) FROM blog_categories bc
                     WHERE bc.blog_id = bp.id
                     AND bc.category_name = ANY($1::text[])) * 10
                ELSE 0 END
                +
                CASE WHEN ARRAY_LENGTH($2::text[], 1) > 0 THEN
                    (SELECT COUNT(*) FROM blog_tags bt
                     WHERE bt.blog_id = bp.id
                     AND bt.tag_name = ANY($2::text[])) * 5
                ELSE 0 END
            ) AS match_score
        FROM blog_posts bp
        LEFT JOIN blog_content bc ON bp.id = bc.id
        WHERE bp.id != $3
        AND bp.language = $4
        AND bp.status = 'published'
        ORDER BY
            match_score DESC,
            bp.created_at DESC
        LIMIT $5
    `

	rows, err := r.db.Query(query, pq.Array(categories), pq.Array(tags), excludeBlogID, language, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relatedPosts []types.BlogPostCardView

	for rows.Next() {
		var card types.BlogPostCardView
		var content types.ContentCardView
		var matchScore int

		err := rows.Scan(
			&card.ID,
			&card.GroupID,
			&card.Slug,
			&card.Language,
			&card.Featured,
			&card.Status,
			&card.CreatedAt,
			&card.UpdatedAt,
			&content.Title,
			&content.Description,
			&content.Image,
			&content.ReadTime,
			&matchScore,
		)

		if err != nil {
			continue
		}

		card.Content = content

		// Kategori bilgilerini de ekle
		cardID, _ := uuid.Parse(card.ID)
		cardCategories, _ := r.SelectBlogCategories(cardID)
		if len(cardCategories) > 0 {
			card.Categories = cardCategories
		}

		relatedPosts = append(relatedPosts, card)
	}

	return relatedPosts, nil
}
