package configs

import (
	"github.com/okanay/backend-blog-guideofdubai/types"
)

type Permission string

const (
	PermissionView       Permission = "view-post"
	PermissionCreatePost Permission = "create-post"

	PermissionEditOwnPost Permission = "edit-own-post"
	PermissionEditAnyPost Permission = "edit-any-post"

	PermissionDeleteOwnPost Permission = "delete-own-post"
	PermissionDeleteAnyPost Permission = "delete-any-post"
)

var RolePermissions = map[types.Role][]Permission{
	types.RoleUser: {
		PermissionView,
	},
	types.RoleEditor: {
		PermissionView,
		PermissionCreatePost,
		PermissionEditOwnPost,
		PermissionDeleteOwnPost,
	},
	types.RoleAdmin: {
		PermissionView,
		PermissionCreatePost,
		PermissionEditAnyPost,
		PermissionDeleteAnyPost,
	},
}

// Bu fonksiyon faydalı değil ve gelecekteki senaryolara uygun değil.
func HasPermission(role types.Role, requiredPermissions []Permission) bool {
	grantedPermissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	if len(requiredPermissions) == 0 {
		return true
	}

	grantedSet := make(map[Permission]bool)
	for _, p := range grantedPermissions {
		grantedSet[p] = true
	}

	for _, required := range requiredPermissions {
		if !grantedSet[required] {
			// Eğer GEREKLİ izinlerden BİR TANESİ BİLE role atanmamışsa, false dön.
			return false
		}
	}

	return true
}
