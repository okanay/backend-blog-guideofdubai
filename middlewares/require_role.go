// middlewares/require_role.go
package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

// RequireRole belirli bir role sahip olmayı gerektiren middleware
func RequireRole(requiredRole types.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Context'ten role bilgisini al (AuthMiddleware tarafından set edilmiş olmalı)
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
				"message": "Yetkilendirme bilgisi bulunamadı",
			})
			c.Abort()
			return
		}

		fmt.Println("Role:", role)
		if role != requiredRole || role != types.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "forbidden",
				"message": "Bu işlem için yetkiniz yok",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
