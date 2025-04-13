// middlewares/rbac.go
package middlewares

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/configs"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func PermissionMiddleware(permission configs.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Kullanıcı rolünü context'ten al
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
				"message": "You must be logged in to access this resource.",
			})
			c.Abort()
			return
		}

		// İzin kontrolü
		if !HasPermission(role.(types.Role), permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "forbidden",
				"message": "You don't have permission to perform this action.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func HasPermission(role types.Role, permission configs.Permission) bool {
	permissions, exists := configs.RolePermissions[role]
	if !exists {
		return false
	}

	return slices.Contains(permissions, permission)
}
