package BlogRepository

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectBlogCards(options types.BlogCardQueryOptions) ([]types.BlogPostCardView, int, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog Cards")

	// 1. Önce toplam sayı için COUNT sorgusu oluşturalım
	countQuery := `
		SELECT COUNT(DISTINCT bp.id)
		FROM blog_posts bp
	`

	// 2. Ana veri sorgusu (mevcut kodunuzdan)
	dataQuery := `
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

            -- Etiketleri JSON dizisi olarak al
            (
                SELECT COALESCE(json_agg(json_build_object('name', t.name, 'value', t.value)), '[]'::json)
                FROM blog_tags bt
                JOIN tags t ON bt.tag_name = t.name
                WHERE bt.blog_id = bp.id
            ) AS tags
        FROM blog_posts bp
        LEFT JOIN blog_content bc ON bp.id = bc.id
        LEFT JOIN blog_featured bf ON bp.id = bf.blog_id AND bf.language = bp.language
    `

	// Her iki sorgu için filtreleri ve join'leri hazırla
	var joins []string
	var conditions []string
	var params []any
	paramCounter := 1

	// Silinen blogları dahil etme
	conditions = append(conditions, "bp.status != 'deleted'")

	// ID filtresi
	if options.ID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("bp.id = $%d", paramCounter))
		params = append(params, options.ID)
		paramCounter++
	}

	// Çoklu ID desteği
	if len(options.IDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("bp.id = ANY($%d)", paramCounter))
		params = append(params, pq.Array(options.IDs))
		paramCounter++
	}

	// Title filtresi
	if options.Title != "" {
		countQuery += " LEFT JOIN blog_content bc ON bp.id = bc.id"
		cleanSearchTerm := cleanSearchQuery(options.Title)
		searchWords := strings.Split(cleanSearchTerm, " ")

		var titleConditions []string
		var descConditions []string

		for _, word := range searchWords {
			if len(word) >= 3 { // Çok kısa kelimeleri atla
				// Başlık için koşul
				titleConditions = append(titleConditions,
					fmt.Sprintf("lower(bc.title) LIKE $%d", paramCounter))
				params = append(params, "%"+word+"%")
				paramCounter++

				// Açıklama için koşul
				descConditions = append(descConditions,
					fmt.Sprintf("lower(bc.description) LIKE $%d", paramCounter))
				params = append(params, "%"+word+"%")
				paramCounter++
			}
		}

		// Sorgu koşullarını ekle
		var orConditions []string
		if len(titleConditions) > 0 {
			orConditions = append(orConditions, "("+strings.Join(titleConditions, " AND ")+")")
		}
		if len(descConditions) > 0 {
			orConditions = append(orConditions, "("+strings.Join(descConditions, " AND ")+")")
		}
		if len(orConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(orConditions, " OR ")+")")
		}
	}

	// Language filtresi
	if options.Language != "" {
		conditions = append(conditions, fmt.Sprintf("bp.language = $%d", paramCounter))
		params = append(params, options.Language)
		paramCounter++
	}

	// Featured filtresi
	if options.Featured {
		conditions = append(conditions, "bf.blog_id IS NOT NULL")
	}

	// Status filtresi
	if options.Status != "" {
		conditions = append(conditions, fmt.Sprintf("bp.status = $%d", paramCounter))
		params = append(params, options.Status)
		paramCounter++
	}

	// Kategori filtresi
	if options.CategoryValue != "" {
		joins = append(joins, "JOIN blog_categories bc_rel ON bp.id = bc_rel.blog_id")
		joins = append(joins, "JOIN categories c ON bc_rel.category_name = c.name")
		conditions = append(conditions, fmt.Sprintf("c.name = $%d", paramCounter))
		params = append(params, options.CategoryValue)
		paramCounter++
	}

	// Tag filtresi
	if options.TagValue != "" {
		if !strings.Contains(dataQuery, "JOIN blog_tags") {
			joins = append(joins, "JOIN blog_tags bt_rel ON bp.id = bt_rel.blog_id")
			joins = append(joins, "JOIN tags t ON bt_rel.tag_name = t.name")
		}
		conditions = append(conditions, fmt.Sprintf("t.name = $%d", paramCounter))
		params = append(params, options.TagValue)
		paramCounter++
	}

	// Tarih aralığı filtresi
	if options.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("bp.created_at >= $%d", paramCounter))
		params = append(params, options.StartDate)
		paramCounter++
	}
	if options.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("bp.created_at <= $%d", paramCounter))
		params = append(params, options.EndDate)
		paramCounter++
	}

	// 3. Count sorgusu için join'leri ekle
	if len(joins) > 0 {
		// blog_content join'i COUNT sorgusu için de gerekli olabilir (title ve description filtresi için)
		if options.Title != "" && !strings.Contains(countQuery, "JOIN blog_content") {
			countQuery += " LEFT JOIN blog_content bc ON bp.id = bc.id"
		}

		// Featured filtresi için count sorgusuna da bu join'i ekle
		if options.Featured && !strings.Contains(countQuery, "LEFT JOIN blog_featured") {
			countQuery += " LEFT JOIN blog_featured bf ON bp.id = bf.blog_id AND bf.language = bp.language"
		}

		for _, join := range joins {
			countQuery += " " + join
		}
	}

	// 4. Ana sorgu için join'leri ekle
	if len(joins) > 0 {
		for _, join := range joins {
			dataQuery += " " + join
		}
	}

	// 5. Her iki sorguya da WHERE koşullarını ekle
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		countQuery += whereClause
		dataQuery += whereClause
	}

	// 6. Kategori ve etiket birden fazla eşleşme gerektiriyorsa, HAVING ile grup filtrelemesi
	if options.CategoryValue != "" && options.TagValue != "" {
		groupByClause := " GROUP BY bp.id, bp.group_id, bp.slug, bp.language, bp.status, bp.created_at, bp.updated_at, bc.title, bc.description, bc.image, bc.read_time, bf.blog_id"
		countQuery += groupByClause
		dataQuery += groupByClause
	}

	// 7. Toplam kayıt sayısını çek
	// COUNT sorgusu için parametrelerin kopyası (limit ve offset olmadan)
	countParams := make([]any, len(params))
	copy(countParams, params)

	var total int
	err := r.db.QueryRow(countQuery, countParams...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count query failed: %w", err)
	}

	// 8. Sıralama seçenekleri (sadece ana sorgu için)
	if options.SortBy != "" {
		allowedSortColumns := map[string]bool{
			"created_at": true,
			"updated_at": true,
			"title":      true,
			"views":      true,
			"likes":      true,
		}

		sortColumn := "created_at"
		if allowedSortColumns[options.SortBy] {
			if options.SortBy == "created_at" || options.SortBy == "updated_at" {
				sortColumn = "bp." + options.SortBy
			} else if options.SortBy == "title" {
				sortColumn = "bc." + options.SortBy
			} else if options.SortBy == "views" || options.SortBy == "likes" {
				if !strings.Contains(dataQuery, "JOIN blog_stats") {
					dataQuery = strings.Replace(dataQuery, "LEFT JOIN blog_content bc ON bp.id = bc.id",
						"LEFT JOIN blog_content bc ON bp.id = bc.id LEFT JOIN blog_stats bs ON bp.id = bs.id", 1)
				}
				sortColumn = "bs." + options.SortBy
			}
		}

		sortDirection := "DESC"
		if options.SortDirection == types.SortAsc {
			sortDirection = "ASC"
		}

		dataQuery += fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortDirection)
	} else {
		dataQuery += " ORDER BY bp.created_at DESC"
	}

	// 9. Limit ve Offset (sadece ana sorgu için)
	if options.Limit > 0 {
		dataQuery += fmt.Sprintf(" LIMIT $%d", paramCounter)
		params = append(params, options.Limit)
		paramCounter++

		if options.Offset > 0 {
			dataQuery += fmt.Sprintf(" OFFSET $%d", paramCounter)
			params = append(params, options.Offset)
		}
	}

	// 10. Ana sorguyu çalıştır
	rows, err := r.db.Query(dataQuery, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("blog card query failed: %w", err)
	}
	defer rows.Close()

	var blogCards []types.BlogPostCardView

	for rows.Next() {
		var card types.BlogPostCardView
		var content types.ContentCardView
		var categoriesJSON, tagsJSON []byte

		// Scan sırası SQL SELECT sırasıyla aynı olmalı
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
			&tagsJSON,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("error scanning blog card row: %w", err)
		}

		card.Content = content

		// JSON kategorileri çöz
		var categories []types.CategoryView
		if err := json.Unmarshal(categoriesJSON, &categories); err == nil {
			card.Categories = categories
		}

		// JSON etiketleri çöz
		var tags []types.TagView
		if err := json.Unmarshal(tagsJSON, &tags); err == nil {
			card.Tags = tags
		}

		blogCards = append(blogCards, card)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating through blog cards: %w", err)
	}

	return blogCards, total, nil
}

func cleanSearchQuery(input string) string {
	// Tüm özel karakterleri kaldır
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	cleaned := re.ReplaceAllString(input, " ")

	// Küçük harfe çevir
	cleaned = strings.ToLower(cleaned)

	// Çoklu boşlukları tek boşluğa dönüştür
	re = regexp.MustCompile(`\s+`)
	cleaned = re.ReplaceAllString(cleaned, " ")

	return strings.TrimSpace(cleaned)
}
