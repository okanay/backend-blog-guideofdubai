package types

import "time"

// UserStatus - user status enum type
type UserStatus string

const (
	UserStatusActive    UserStatus = "Active"
	UserStatusSuspended UserStatus = "Suspended"
	UserStatusDeleted   UserStatus = "Deleted"
)

// Role - user role enum type
type Role string

const (
	RoleUser   Role = "User"
	RoleEditor Role = "Editor"
	RoleAdmin  Role = "Admin"
)

// Table Model (database/migrations/00001.auth.up.sql)
type User struct {
	ID             int64      `db:"id" json:"id"`
	UniqueID       string     `db:"unique_id" json:"uniqueId"`
	Email          string     `db:"email" json:"email"`
	Username       string     `db:"username" json:"username"`
	HashedPassword string     `db:"hashed_password" json:"-"`
	Membership     Role       `db:"membership" json:"membership"`
	EmailVerified  bool       `db:"email_verified" json:"emailVerified"`
	Status         UserStatus `db:"status" json:"status"`
	DeletedAt      *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"createdAt"`
	LastLogin      time.Time  `db:"last_login" json:"lastLogin"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updatedAt"`
}

// UserProfileResponse - secure model to return user profile
type UserProfileResponse struct {
	ID            string     `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	Membership    Role       `json:"membership"`
	EmailVerified bool       `json:"emailVerified"`
	Status        UserStatus `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastLogin     time.Time  `json:"lastLogin"`
}

// UserCreateRequest - user creation request
type UserCreateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserLoginRequest - user login request
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
