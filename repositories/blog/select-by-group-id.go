// repositories/blog/select-by-group-id.go
package BlogRepository

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectBlogByGroupID(request types.BlogSelectByGroupIDInput) (*types.BlogPostView, []*types.BlogPostView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog And Alternatives")

	// 1. Slug ile doğrudan blog post arama (slug primary öncelik, groupID sonra bakılacak)
	// Önce sadece slug ile ara (her zaman slug öncelikli)
	var mainPostID uuid.UUID
	var groupID string

	query := `
        SELECT id, group_id
        FROM blog_posts
        WHERE slug = $1
          AND status = 'published'
        LIMIT 1
    `
	err := r.db.QueryRow(query, request.SlugOrGroupID).Scan(&mainPostID, &groupID)

	// Slug ile bulunamadıysa, groupID olarak dene
	if err == sql.ErrNoRows {
		query = `
			SELECT id, group_id
			FROM blog_posts
			WHERE group_id = $1
			  AND status != 'deleted'
			LIMIT 1
		`
		err = r.db.QueryRow(query, request.SlugOrGroupID).Scan(&mainPostID, &groupID)
	}

	// Hala bulunamadıysa hata döndür
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("blog post not found with slug or groupId=%s", request.SlugOrGroupID)
		}
		return nil, nil, fmt.Errorf("error retrieving blog data: %w", err)
	}

	// 2. Ana postun tüm detaylarını çek
	var wg sync.WaitGroup
	var mainPost *types.BlogPostView
	var mainPostErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		mainPost, mainPostErr = r.SelectBlogByID(mainPostID)
	}()

	// 3. Aynı groupID'ye sahip tüm alternatifleri bul (ana post hariç)
	var alternativeIDs []uuid.UUID

	altQuery := `
		SELECT id
		FROM blog_posts
		WHERE group_id = $1
		  AND id != $2
		  AND status != 'deleted'
	`

	rows, err := r.db.Query(altQuery, groupID, mainPostID)
	if err != nil {
		wg.Wait() // Ana post için wait et
		if mainPostErr != nil {
			return nil, nil, fmt.Errorf("error retrieving main blog: %w", mainPostErr)
		}
		return mainPost, nil, fmt.Errorf("error retrieving alternative posts: %w", err)
	}
	defer rows.Close()

	// Alternatif ID'leri topla
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue // Hatalı ID'leri atla
		}
		alternativeIDs = append(alternativeIDs, id)
	}

	// 4. Alternatif blogları paralel olarak çek
	var alternativeMutex sync.Mutex
	var alternatives []*types.BlogPostView

	if len(alternativeIDs) > 0 {
		var altWg sync.WaitGroup

		for _, id := range alternativeIDs {
			altWg.Add(1)

			go func(blogID uuid.UUID) {
				defer altWg.Done()

				blog, err := r.SelectBlogByID(blogID)
				if err != nil {
					// Hata loglanabilir ama işlemi durdurmak istemiyoruz
					return
				}

				alternativeMutex.Lock()
				alternatives = append(alternatives, blog)
				alternativeMutex.Unlock()
			}(id)
		}

		// Bütün alternatif blogların yüklenmesini bekle
		altWg.Wait()
	}

	// Ana postun yüklenmesini bekle
	wg.Wait()

	// Ana post için hata kontrolü
	if mainPostErr != nil {
		return nil, nil, fmt.Errorf("error retrieving main blog: %w", mainPostErr)
	}

	return mainPost, alternatives, nil
}
