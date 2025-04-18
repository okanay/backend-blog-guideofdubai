package BlogHandler

import (
	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (h *Handler) SelectBlogByGroupID(c *gin.Context) {
	var request types.BlogSelectByGroupIDInput

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

}
