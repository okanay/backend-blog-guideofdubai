package BlogRepository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectBlogCards(options types.BlogCardQueryOptions) ([]types.BlogPostCardView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog Cards")

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
            CASE WHEN bf.blog_id IS NOT NULL THEN true ELSE false END as featured
        FROM blog_posts bp
        LEFT JOIN blog_content bc ON bp.id = bc.id
        LEFT JOIN blog_featured bf ON bp.id = bf.blog_id AND bf.language = bp.language
    `

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
		conditions = append(conditions, fmt.Sprintf("bc.title ILIKE $%d", paramCounter))
		params = append(params, "%"+options.Title+"%")
		paramCounter++
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

	// Kategori filtresi (AND mantığı: blogun sadece bu kategoriye sahip olması)
	if options.CategoryValue != "" {
		joins = append(joins, "JOIN blog_categories bc_rel ON bp.id = bc_rel.blog_id")
		joins = append(joins, "JOIN categories c ON bc_rel.category_name = c.name")
		query += " " + strings.Join(joins, " ")
		conditions = append(conditions, fmt.Sprintf("c.name = $%d", paramCounter))
		params = append(params, options.CategoryValue)
		paramCounter++

		// Blogun sahip olduğu kategori sayısı 1 olmalı (sadece bu kategori)
		query += `
            GROUP BY bp.id, bp.group_id, bp.slug, bp.language, bp.featured, bp.status, bp.created_at, bp.updated_at, bc.title, bc.description, bc.image, bc.read_time
            HAVING COUNT(DISTINCT c.name) = 1
        `
	}

	// Tag filtresi (AND mantığı: blogun sadece bu tag'e sahip olması)
	if options.TagValue != "" {
		// Eğer kategori de varsa, JOIN'ler zaten eklenmiş olabilir, tekrar eklememek için kontrol et
		if !strings.Contains(query, "JOIN blog_tags") {
			query += " JOIN blog_tags bt_rel ON bp.id = bt_rel.blog_id JOIN tags t ON bt_rel.tag_name = t.name"
		}
		conditions = append(conditions, fmt.Sprintf("t.name = $%d", paramCounter))
		params = append(params, options.TagValue)
		paramCounter++

		// Blogun sahip olduğu tag sayısı 1 olmalı (sadece bu tag)
		if !strings.Contains(query, "GROUP BY") {
			query += `
                GROUP BY bp.id, bp.group_id, bp.slug, bp.language, bp.featured, bp.status, bp.created_at, bp.updated_at, bc.title, bc.description, bc.image, bc.read_time
            `
		}
		query += `
            HAVING COUNT(DISTINCT t.name) = 1
        `
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

	// WHERE koşullarını sorguya ekle
	if len(conditions) > 0 {
		if !strings.Contains(query, "GROUP BY") {
			query += " WHERE " + strings.Join(conditions, " AND ")
		} else {
			// GROUP BY varsa, WHERE'i HAVING'den önce ekle
			query = strings.Replace(query, "GROUP BY", "WHERE "+strings.Join(conditions, " AND ")+" GROUP BY", 1)
		}
	}

	// Sıralama seçenekleri
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
				if !strings.Contains(query, "JOIN blog_stats") {
					query = strings.Replace(query, "LEFT JOIN blog_content bc ON bp.id = bc.id",
						"LEFT JOIN blog_content bc ON bp.id = bc.id LEFT JOIN blog_stats bs ON bp.id = bs.id", 1)
				}
				sortColumn = "bs." + options.SortBy
			}
		}

		sortDirection := "DESC"
		if options.SortDirection == types.SortAsc {
			sortDirection = "ASC"
		}

		query += fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortDirection)
	} else {
		query += " ORDER BY bp.created_at DESC"
	}

	// Limit ve Offset
	if options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramCounter)
		params = append(params, options.Limit)
		paramCounter++

		if options.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", paramCounter)
			params = append(params, options.Offset)
		}
	}

	// Sorguyu çalıştır
	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("blog card query failed: %w", err)
	}
	defer rows.Close()

	var blogCards []types.BlogPostCardView

	for rows.Next() {
		var card types.BlogPostCardView
		var content types.ContentCardView

		// Scan sırası SQL SELECT sırasıyla aynı olmalı
		err := rows.Scan(
			&card.ID,
			&card.GroupID,
			&card.Slug,
			&card.Language,
			&card.Status, // 5. sütun status
			&card.CreatedAt,
			&card.UpdatedAt,
			&content.Title,
			&content.Description,
			&content.Image,
			&content.ReadTime,
			&card.Featured, // 12. sütun featured (en son)
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning blog card row: %w", err)
		}

		card.Content = content
		cardUUID, err := uuid.Parse(card.ID)
		if err != nil {
			return nil, fmt.Errorf("error converting card ID to UUID: %w", err)
		}

		categories, err := r.SelectBlogCategories(cardUUID)
		if err == nil && len(categories) > 0 {
			card.Categories = categories
		}

		blogCards = append(blogCards, card)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through blog cards: %w", err)
	}

	return blogCards, nil
}
