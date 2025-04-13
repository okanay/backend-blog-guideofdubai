package configs

import (
	"github.com/okanay/backend-blog-guideofdubai/types"
)

type Permission string
type Access string

const (
	CreatePost Permission = "create-post"
	EditPost   Permission = "edit-post"
	DeletePost Permission = "delete-post"
)

const (
	AccessFull Access = "full"
	AccessOwn  Access = "own"
	AccessNone Access = "none"
)

var RolePermissionConfig = map[types.Role]map[Permission]Access{
	types.RoleEditor: {
		CreatePost: AccessFull,
		EditPost:   AccessFull,
		DeletePost: AccessNone,
	},
	types.RoleUser: {
		CreatePost: AccessNone,
		EditPost:   AccessNone,
		DeletePost: AccessNone,
	},
}

func CheckPermission(role types.Role, permission Permission, userID string, resourceOwnerID string) bool {
	if role == types.RoleAdmin {
		return true
	}

	permConfig, exists := RolePermissionConfig[role]
	if !exists {
		return false
	}

	accessType, exists := permConfig[permission]
	if !exists {
		return false
	}

	switch accessType {
	case AccessFull:
		return true
	case AccessOwn:
		return userID == resourceOwnerID
	case AccessNone:
		return false
	default:
		return false
	}
}
