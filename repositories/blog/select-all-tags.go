package BlogRepository

import (
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectAllTags() ([]types.TagView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select All Tags")
	var tags []types.TagView

	query := `
		SELECT name, value FROM tags
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var tag types.TagView
		err := rows.Scan(&tag.Name, &tag.Value)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}
