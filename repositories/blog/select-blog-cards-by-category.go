package BlogRepository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectBlogCards(options types.BlogCardQueryOptions) ([]types.BlogPostCardView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog Cards")

	var params []interface{}
	paramCount := 1

	// Base sorgu
	baseQuery := `
        SELECT
            bp.id,
            bp.group_id,
            bp.slug,
            bp.language,
            bp.featured,
            bp.status,
            bc.title,
            bc.description,
            bc.read_time,
            bp.created_at,
            bp.updated_at
        FROM blog_posts bp
        JOIN blog_content bc ON bp.id = bc.id
    `

	// Kategori için JOIN
	if options.CategoryValue != "" {
		baseQuery += `
        JOIN blog_categories bcat ON bp.id = bcat.blog_id
        JOIN categories c ON bcat.category_name = c.name
        `
	}

	// Etiket için JOIN
	if options.TagValue != "" {
		baseQuery += `
            JOIN blog_tags bt ON bp.id = bt.blog_id
            JOIN tags t ON bt.tag_name = t.name
        `
	}

	// WHERE koşullarını oluştur
	var conditions []string

	// ID filtreleri için
	if options.ID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("bp.id = $%d", paramCount))
		params = append(params, options.ID)
		paramCount++
	} else if len(options.IDs) > 0 {
		placeholders := make([]string, len(options.IDs))
		for i, id := range options.IDs {
			placeholders[i] = fmt.Sprintf("$%d", paramCount)
			params = append(params, id)
			paramCount++
		}
		conditions = append(conditions, fmt.Sprintf("bp.id IN (%s)", strings.Join(placeholders, ",")))
	}

	// Kategori filtresi
	if options.CategoryValue != "" {
		conditions = append(conditions, fmt.Sprintf("c.name = $%d", paramCount))
		params = append(params, options.CategoryValue)
		paramCount++
	}

	// Etiket filtresi
	if options.TagValue != "" {
		conditions = append(conditions, fmt.Sprintf("t.name = $%d", paramCount))
		params = append(params, options.TagValue)
		paramCount++
	}

	// Dil filtresi
	if options.Language != "" {
		conditions = append(conditions, fmt.Sprintf("bp.language = $%d", paramCount))
		params = append(params, options.Language)
		paramCount++
	}

	// Öne çıkanlar filtresi
	if options.Featured {
		conditions = append(conditions, "bp.featured = true")
	}

	// Status filtresi
	if options.Status != "" {
		conditions = append(conditions, fmt.Sprintf("bp.status = $%d", paramCount))
		params = append(params, options.Status)
		paramCount++
	} else {
		// Varsayılan olarak sadece published olanları getir
		conditions = append(conditions, "bp.status = 'published'")
	}

	if options.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("bp.created_at >= $%d", paramCount))
		params = append(params, options.StartDate)
		paramCount++
	}

	// Bitiş tarihi
	if options.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("bp.created_at <= $%d", paramCount))
		params = append(params, options.EndDate)
		paramCount++
	}

	// Son X gün
	if options.LastDays > 0 {
		conditions = append(conditions, fmt.Sprintf("bp.created_at >= NOW() - INTERVAL '$%d days'", paramCount))
		params = append(params, options.LastDays)
		paramCount++
	}

	// Son X hafta
	if options.LastWeeks > 0 {
		conditions = append(conditions, fmt.Sprintf("bp.created_at >= NOW() - INTERVAL '$%d weeks'", paramCount))
		params = append(params, options.LastWeeks)
		paramCount++
	}

	// Son X ay
	if options.LastMonths > 0 {
		conditions = append(conditions, fmt.Sprintf("bp.created_at >= NOW() - INTERVAL '$%d months'", paramCount))
		params = append(params, options.LastMonths)
		paramCount++
	}

	// Koşulları WHERE kısmına ekle
	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	sortField := "bp.created_at"
	if options.SortBy != "" {
		// Güvenli sıralama alanları
		allowedSortFields := map[string]string{
			"created_at": "bp.created_at",
			"updated_at": "bp.updated_at",
			"title":      "bc.title",
			"views":      "bs.views", // eğer views bilgisi de çekiliyorsa
		}

		if field, ok := allowedSortFields[options.SortBy]; ok {
			sortField = field
		}
	}

	// Sıralama
	sortDirection := "DESC"
	if options.SortDirection == types.SortAsc {
		sortDirection = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortField, sortDirection)

	// Limit
	if options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramCount)
		params = append(params, options.Limit)
		paramCount++
	}

	// Offset
	if options.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", paramCount)
		params = append(params, options.Offset)
	}

	// Sorguyu çalıştır
	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("error executing blog cards query: %w", err)
	}
	defer rows.Close()

	var blogCards []types.BlogPostCardView
	for rows.Next() {
		var card types.BlogPostCardView
		var content types.ContentCardView
		var description sql.NullString

		err := rows.Scan(
			&card.ID,
			&card.GroupID,
			&card.Slug,
			&card.Language,
			&card.Featured,
			&card.Status,
			&content.Title,
			&description,
			&content.ReadTime,
			&card.CreatedAt,
			&card.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning blog card data: %w", err)
		}

		if description.Valid {
			content.Description = description.String
		}

		card.Content = content
		blogCards = append(blogCards, card)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error processing blog cards: %w", err)
	}

	// Eğer tek bir ID istendiyse ve bulunamadıysa hata döndür
	if options.ID != uuid.Nil && len(blogCards) == 0 {
		return nil, fmt.Errorf("blog card with ID %s not found", options.ID)
	}

	return blogCards, nil
}
