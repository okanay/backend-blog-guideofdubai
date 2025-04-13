package configs

import (
	"github.com/okanay/backend-blog-guideofdubai/types"
)

type Permission string

const (
	PermissionCreatePost Permission = "create-post"
	PermissionEditPost   Permission = "edit-post"
	PermissionDeletePost Permission = "delete-post"
)

var RolePermissions = map[types.Role][]Permission{
	types.RoleUser: {},
	types.RoleEditor: {
		PermissionCreatePost,
		PermissionEditPost,
	},
	types.RoleAdmin: {
		PermissionCreatePost,
		PermissionEditPost,
		PermissionDeletePost,
	},
}
