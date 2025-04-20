package BlogRepository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectBlogCards(options types.BlogCardQueryOptions) ([]types.BlogPostCardView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog Cards")

	// Ana sorgu ve WHERE koşulları için diziler oluştur
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
            bc.read_time
        FROM blog_posts bp
        LEFT JOIN blog_content bc ON bp.id = bc.id
    `

	// WHERE koşulları ve parametreleri için slice'lar oluştur
	var conditions []string
	var params []any
	paramCounter := 1 // PostgreSQL'de parametre indeksi 1'den başlar

	// Temel WHERE koşulu: Silinen blogları dahil etme
	conditions = append(conditions, "bp.status != 'deleted'")

	// Belirli bir ID için filtreleme
	if options.ID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("bp.id = $%d", paramCounter))
		params = append(params, options.ID)
		paramCounter++
	}

	// Title filtresi (case-insensitive)
	if options.Title != "" {
		conditions = append(conditions, fmt.Sprintf("bc.title ILIKE $%d", paramCounter))
		params = append(params, "%"+options.Title+"%") // ILIKE için % ile wildcard
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
		conditions = append(conditions, "bp.featured = true")
	}

	// Status filtresi
	if options.Status != "" {
		conditions = append(conditions, fmt.Sprintf("bp.status = $%d", paramCounter))
		params = append(params, options.Status)
		paramCounter++
	}

	// Category filtresi
	if options.CategoryValue != "" {
		query += `
            JOIN blog_categories bc_rel ON bp.id = bc_rel.blog_id
            JOIN categories c ON bc_rel.category_name = c.name
        `
		conditions = append(conditions, fmt.Sprintf("c.name = $%d", paramCounter))
		params = append(params, options.CategoryValue)
		paramCounter++
	}

	// Tag filtresi
	if options.TagValue != "" {
		query += `
            JOIN blog_tags bt_rel ON bp.id = bt_rel.blog_id
            JOIN tags t ON bt_rel.tag_name = t.name
        `
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

	// WHERE koşullarını sorguya ekle
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Sıralama seçenekleri
	if options.SortBy != "" {
		// SQL injection'ı önlemek için izin verilen sütunları kontrol et
		allowedSortColumns := map[string]bool{
			"created_at": true,
			"updated_at": true,
			"title":      true,
			"views":      true,
			"likes":      true,
		}

		sortColumn := "created_at" // varsayılan
		if allowedSortColumns[options.SortBy] {
			// blog_posts tablosundaki alanlar için
			if options.SortBy == "created_at" || options.SortBy == "updated_at" {
				sortColumn = "bp." + options.SortBy
			} else if options.SortBy == "title" {
				sortColumn = "bc." + options.SortBy
			} else if options.SortBy == "views" || options.SortBy == "likes" {
				// İstatistikleri katmak için JOIN ekle (eğer henüz eklenmemişse)
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
		// Varsayılan sıralama: En yeni blogları göster
		query += " ORDER BY bp.created_at DESC"
	}

	// Limit ve Offset
	if options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramCounter)
		params = append(params, options.Limit)
		paramCounter++

		// Offset sadece limit belirtilmişse anlamlıdır
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

	// Sonuçları işle
	var blogCards []types.BlogPostCardView

	for rows.Next() {
		var card types.BlogPostCardView
		var content types.ContentCardView

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
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning blog card row: %w", err)
		}

		card.Content = content
		blogCards = append(blogCards, card)
	}

	// rows.Next() döngüsü içindeki olası hataları kontrol et
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through blog cards: %w", err)
	}

	return blogCards, nil
}
