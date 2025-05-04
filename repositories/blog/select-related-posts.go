package BlogRepository

import (
	"fmt"
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

	var relatedPosts []types.BlogPostCardView

	// 1. Kategori/Tag/Language ile eşleşenler
	query := makeRelatedPostQuery(true, true)
	params := []any{pq.Array(categories), pq.Array(tags), excludeBlogID, language, limit}
	posts, err := r.runRelatedPostQuery(query, params)
	if err != nil {
		return nil, err
	}
	relatedPosts = append(relatedPosts, posts...)
	if len(relatedPosts) >= limit {
		return relatedPosts[:limit], nil
	}

	// 2. Sadece Language ile eşleşenler
	kalanLimit := limit - len(relatedPosts)
	if kalanLimit > 0 {
		query = makeRelatedPostQuery(false, true)
		params = []any{pq.Array([]string{}), pq.Array([]string{}), excludeBlogID, language, kalanLimit}
		posts, err = r.runRelatedPostQuery(query, params)
		if err == nil {
			for _, p := range posts {
				alreadyExists := false
				for _, rp := range relatedPosts {
					if p.ID == rp.ID {
						alreadyExists = true
						break
					}
				}
				if !alreadyExists {
					relatedPosts = append(relatedPosts, p)
				}
			}
		}
	}
	if len(relatedPosts) >= limit {
		return relatedPosts[:limit], nil
	}

	// 3. Herhangi bir post (dil farketmez)
	kalanLimit = limit - len(relatedPosts)
	if kalanLimit > 0 {
		query = makeRelatedPostQuery(false, false)
		params = []any{pq.Array([]string{}), pq.Array([]string{}), excludeBlogID, kalanLimit}
		posts, err = r.runRelatedPostQuery(query, params)
		if err == nil {
			for _, p := range posts {
				alreadyExists := false
				for _, rp := range relatedPosts {
					if p.ID == rp.ID {
						alreadyExists = true
						break
					}
				}
				if !alreadyExists {
					relatedPosts = append(relatedPosts, p)
				}
			}
		}
	}

	if len(relatedPosts) > limit {
		relatedPosts = relatedPosts[:limit]
	}
	return relatedPosts, nil
}

func makeRelatedPostQuery(
	mustMatchCategoryOrTag bool,
	hasLanguage bool,
) string {
	// Query gövdesi
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
        AND bp.status = 'published'
    `
	paramIdx := 4

	if hasLanguage {
		query += fmt.Sprintf(" AND bp.language = $%d", paramIdx)
		paramIdx++
	}

	if mustMatchCategoryOrTag {
		query += `
            AND (
                (ARRAY_LENGTH($1::text[], 1) > 0 AND EXISTS (
                    SELECT 1 FROM blog_categories bc2
                    WHERE bc2.blog_id = bp.id
                    AND bc2.category_name = ANY($1::text[])
                ))
                OR
                (ARRAY_LENGTH($2::text[], 1) > 0 AND EXISTS (
                    SELECT 1 FROM blog_tags bt2
                    WHERE bt2.blog_id = bp.id
                    AND bt2.tag_name = ANY($2::text[])
                ))
            )
        `
	}

	query += fmt.Sprintf(`
        ORDER BY
            match_score DESC,
            bp.created_at DESC
        LIMIT $%d
    `, paramIdx)

	return query
}

func (r *Repository) runRelatedPostQuery(
	query string,
	params []any,
) ([]types.BlogPostCardView, error) {
	rows, err := r.db.Query(query, params...)
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

		cardID, _ := uuid.Parse(card.ID)
		cardCategories, _ := r.SelectBlogCategories(cardID)
		if len(cardCategories) > 0 {
			card.Categories = cardCategories
		}

		relatedPosts = append(relatedPosts, card)
	}

	return relatedPosts, nil
}
