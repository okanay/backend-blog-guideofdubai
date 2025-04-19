package types

import (
	"time"

	"github.com/google/uuid"
)

// Table Model (database/migrations/00001.auth.up.sql)
type RefreshToken struct {
	ID            uuid.UUID `db:"id" json:"id"`
	UserID        uuid.UUID `db:"user_id" json:"userId"`
	UserEmail     string    `db:"user_email" json:"userEmail"`
	UserUsername  string    `db:"user_username" json:"userUsername"`
	Token         string    `db:"token" json:"token"`
	IPAddress     string    `db:"ip_address" json:"ipAddress"`
	UserAgent     string    `db:"user_agent" json:"userAgent"`
	ExpiresAt     time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
	LastUsedAt    time.Time `db:"last_used_at" json:"lastUsedAt"`
	IsRevoked     bool      `db:"is_revoked" json:"isRevoked"`
	RevokedReason string    `db:"revoked_reason,omitempty" json:"revokedReason,omitempty"`
}

type TokenCreateRequest struct {
	UserID       uuid.UUID `db:"user_id" json:"userId"`
	UserEmail    string    `db:"user_email" json:"userEmail"`
	UserUsername string    `db:"user_username" json:"userUsername"`
	Token        string    `db:"token" json:"token"`
	IPAddress    string    `db:"ip_address" json:"ipAddress"`
	UserAgent    string    `db:"user_agent" json:"userAgent"`
	ExpiresAt    time.Time `db:"expires_at" json:"expiresAt"`
}

// Information to be carried in JWT
type TokenClaims struct {
	ID            uuid.UUID  `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	Role          Role       `json:"role"`
	EmailVerified bool       `json:"emailVerified"`
	Status        UserStatus `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastLogin     time.Time  `json:"lastLogin"`
}
