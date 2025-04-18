package BlogRepository

import (
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectAllCategories() ([]types.CategoryView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select All Categories")
	var categories []types.CategoryView

	query := `
		SELECT name, value FROM categories
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var category types.CategoryView
		err := rows.Scan(&category.Name, &category.Value)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}
