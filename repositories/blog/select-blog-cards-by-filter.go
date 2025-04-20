package BlogRepository

import (
	"time"

	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (r *Repository) SelectBlogCards(options types.BlogCardQueryOptions) ([]types.BlogPostCardView, error) {
	defer utils.TimeTrack(time.Now(), "Blog -> Select Blog Cards")
	var blogCards []types.BlogPostCardView
	return blogCards, nil
}
