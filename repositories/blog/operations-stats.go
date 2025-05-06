// repositories/blog/stats.go
package BlogRepository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

// IncrementViewCount blog görüntülenme sayısını artırır
func (r *Repository) IncrementViewCount(blogID uuid.UUID) error {
	defer utils.TimeTrack(time.Now(), "Blog -> Increment View Count")

	query := `
		UPDATE blog_stats
		SET views = views + 1,
		    last_viewed_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(query, blogID)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("blog stats not found for ID: %s", blogID)
	}

	return nil
}

func (r *Repository) GetBlogStats(language string, limit int, offset int) ([]types.BlogStatsDetailView, int, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Get Blog Stats")

	// Toplam kayıt sayısını al
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM blog_posts bp
		JOIN blog_stats bs ON bp.id = bs.id
		WHERE bp.status = 'published'
	`

	if language != "" {
		countQuery += " AND bp.language = $1"
		err := r.db.QueryRow(countQuery, language).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get total count: %w", err)
		}
	} else {
		err := r.db.QueryRow(countQuery).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get total count: %w", err)
		}
	}

	// İstatistikleri getir
	query := `
		SELECT
			bp.id,
			bc.title,
			bc.image,
			bp.language,
			bp.group_id,
			bp.slug,
			bs.views,
			bs.likes,
			bs.shares,
			bs.comments,
			bs.last_viewed_at,
			bp.created_at,
			bp.updated_at
		FROM blog_posts bp
		JOIN blog_content bc ON bp.id = bc.id
		JOIN blog_stats bs ON bp.id = bs.id
		WHERE bp.status = 'published'
	`

	var args []any
	argPosition := 1

	if language != "" {
		query += fmt.Sprintf(" AND bp.language = $%d", argPosition)
		args = append(args, language)
		argPosition++
	}

	query += " ORDER BY bs.views DESC, bs.last_viewed_at DESC NULLS LAST"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPosition)
		args = append(args, limit)
		argPosition++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPosition)
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get blog stats: %w", err)
	}
	defer rows.Close()

	var stats []types.BlogStatsDetailView
	for rows.Next() {
		var stat types.BlogStatsDetailView
		var lastViewedAt sql.NullTime

		err := rows.Scan(
			&stat.BlogID,
			&stat.Title,
			&stat.Image,
			&stat.Language,
			&stat.GroupID,
			&stat.Slug,
			&stat.Views,
			&stat.Likes,
			&stat.Shares,
			&stat.Comments,
			&lastViewedAt,
			&stat.CreatedAt,
			&stat.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan blog stat: %w", err)
		}

		if lastViewedAt.Valid {
			stat.LastViewedAt = &lastViewedAt.Time
		}

		stats = append(stats, stat)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return stats, total, nil
}

// GetBlogStatByID tek bir blog'un istatistiklerini getirir
func (r *Repository) GetBlogStatByID(blogID uuid.UUID) (*types.BlogStatsDetailView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Get Blog Stat By ID")

	query := `
		SELECT
			bp.id,
			bc.title,
			bc.image,
			bp.language,
			bp.group_id,
			bp.slug,
			bs.views,
			bs.likes,
			bs.shares,
			bs.comments,
			bs.last_viewed_at,
			bp.created_at,
			bp.updated_at
		FROM blog_posts bp
		JOIN blog_content bc ON bp.id = bc.id
		JOIN blog_stats bs ON bp.id = bs.id
		WHERE bp.id = $1
	`

	var stat types.BlogStatsDetailView
	var lastViewedAt sql.NullTime

	err := r.db.QueryRow(query, blogID).Scan(
		&stat.BlogID,
		&stat.Title,
		&stat.Image,
		&stat.Language,
		&stat.GroupID,
		&stat.Slug,
		&stat.Views,
		&stat.Likes,
		&stat.Shares,
		&stat.Comments,
		&lastViewedAt,
		&stat.CreatedAt,
		&stat.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blog not found: %s", blogID)
		}
		return nil, fmt.Errorf("failed to get blog stat: %w", err)
	}

	if lastViewedAt.Valid {
		stat.LastViewedAt = &lastViewedAt.Time
	}

	return &stat, nil
}
