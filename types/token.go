package types

import "time"

// Table Model (database/migrations/00001.auth.up.sql)
type RefreshToken struct {
	ID            int64     `db:"id" json:"id"`
	UserID        int64     `db:"user_id" json:"userId"`
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

// Information to be carried in JWT
type TokenClaims struct {
	UniqueID string `json:"uniqueId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     Role   `json:"role"`
}
