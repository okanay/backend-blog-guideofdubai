package BlogRepository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) AddToFeaturedList(blogID uuid.UUID, language string) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Add To Featured List")

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. Blog'un var olduğunu ve published durumunda olduğunu kontrol et
	var exists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM blog_posts
			WHERE id = $1 AND status = 'published'
		)
	`
	err = tx.QueryRow(checkQuery, blogID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check blog existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("blog not found or not published")
	}

	// 2. Mevcut en büyük position değerini bul
	var maxPosition sql.NullInt64
	positionQuery := `
		SELECT MAX(position)
		FROM blog_featured
		WHERE language = $1
	`
	err = tx.QueryRow(positionQuery, language).Scan(&maxPosition)
	if err != nil {
		return fmt.Errorf("failed to get max position: %w", err)
	}

	// 3. Yeni position değerini hesapla (100'er artışla)
	newPosition := 0
	if maxPosition.Valid {
		newPosition = int(maxPosition.Int64) + 100
	} else {
		newPosition = 100
	}

	// 4. Featured tablosuna ekle
	insertQuery := `
		INSERT INTO blog_featured (blog_id, language, position)
		VALUES ($1, $2, $3)
	`
	_, err = tx.Exec(insertQuery, blogID, language, newPosition)
	if err != nil {
		return fmt.Errorf("failed to insert into featured list: %w", err)
	}

	return tx.Commit()
}

func (r *Repository) RemoveFromFeaturedList(blogID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Remove From All Featured Lists")

	query := `
		DELETE FROM blog_featured
		WHERE blog_id = $1
	`

	_, err := r.db.Exec(query, blogID)
	if err != nil {
		return fmt.Errorf("failed to remove from all featured lists: %w", err)
	}

	return nil
}

func (r *Repository) UpdateFeaturedOrdering(language string, orderedBlogIDs []uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Update Featured Ordering (Negative Temp)")

	if len(orderedBlogIDs) == 0 {
		// Güncellenecek bir şey yoksa boşuna işlem yapma.
		// İsteğe bağlı: Bu dildeki tüm featured postları silmek isteyebilirsiniz.
		// Veya sadece başarılı dönün.
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// defer tx.Rollback() // Hata durumunda otomatik Rollback için helper kullanmak daha iyi
	// Helper fonksiyon örneği:
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-panic after Rollback
		} else if err != nil {
			// log.Printf("Rolling back transaction due to error: %v", err)
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // Commit if no error occurred
			// if err != nil {
			//  log.Printf("Error during transaction commit: %v", err)
			// }
		}
	}()

	// --- Adım 1: Geçici Negatif Pozisyonlara Güncelle ---

	// UPDATE FROM VALUES için VALUES listesini ve argümanları hazırla
	valuesStatement := strings.Builder{}
	args := []interface{}{language} // İlk argüman dil
	paramIndex := 2                 // $1 dil için kullanıldı

	for i, blogID := range orderedBlogIDs {
		tempPosition := -(i + 1) // -1, -2, -3...
		if i > 0 {
			valuesStatement.WriteString(", ")
		}
		// ($2::uuid, $3::integer)
		valuesStatement.WriteString(fmt.Sprintf("($%d::uuid, $%d::integer)", paramIndex, paramIndex+1))
		args = append(args, blogID, tempPosition)
		paramIndex += 2
	}

	updateToNegativeQuery := fmt.Sprintf(`
			UPDATE blog_featured AS bf
			SET position = v.temp_position, updated_at = NOW()
			FROM (VALUES %s) AS v(blog_id_val, temp_position)
			WHERE bf.blog_id = v.blog_id_val AND bf.language = $1
		`, valuesStatement.String())

	result, err := tx.Exec(updateToNegativeQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update to temporary negative positions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	// Önemli: Eğer listedeki bazı ID'ler featured değilse burada rowsAffected < len(orderedBlogIDs) olabilir.
	// Bu bir hata durumu mu, yoksa kabul edilebilir mi? Karar vermelisiniz.
	// Şimdilik sadece devam ediyoruz. Eksik ID'ler sonraki adımda güncellenmeyecektir.
	log.Printf("Rows affected in negative update: %d", rowsAffected)

	// --- Adım 2: Nihai Pozitif Pozisyonlara Güncelle ---
	// Negatif pozisyonları pozitif ve aralıklı hale getir (örn: -1 -> 100, -2 -> 200)
	updateToPositiveQuery := `
			UPDATE blog_featured
			SET position = ABS(position) * 100, updated_at = NOW()
			WHERE language = $1 AND position < 0
		`
	_, err = tx.Exec(updateToPositiveQuery, language)
	if err != nil {
		return fmt.Errorf("failed to update to final positive positions: %w", err)
	}

	// Hata yoksa defer içindeki Commit çalışacak
	return nil
}

func (r *Repository) GetFeaturedBlogs(language string) ([]types.BlogPostCardView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Get Featured Blogs")

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
			true as featured,
			bf.position
		FROM blog_posts bp
		JOIN blog_content bc ON bp.id = bc.id
		JOIN blog_featured bf ON bp.id = bf.blog_id
		WHERE bf.language = $1
		  AND bp.status = 'published'
		ORDER BY bf.position ASC
	`

	rows, err := r.db.Query(query, language)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured blogs: %w", err)
	}
	defer rows.Close()

	var blogs []types.BlogPostCardView
	for rows.Next() {
		var blog types.BlogPostCardView
		var content types.ContentCardView
		var position int // Sadece sıralama için kullanılacak, response'a dahil edilmeyecek

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
			&position,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blog: %w", err)
		}

		blog.Content = content

		// Kategorileri ekle
		blogUUID, err := uuid.Parse(blog.ID)
		if err == nil {
			categories, err := r.SelectBlogCategories(blogUUID)
			if err == nil && len(categories) > 0 {
				blog.Categories = categories
			}
		}

		blogs = append(blogs, blog)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return blogs, nil
}
