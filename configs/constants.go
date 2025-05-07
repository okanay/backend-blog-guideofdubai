package configs

import (
	"time"
)

const (
	// Project Rules
	PROJECT_NAME = "Guide Of Dubai - Blog"

	// STATS RULES
	VIEW_CACHE_EXPIRATION = 1 * time.Minute

	AI_RATE_LIMIT_WINDOW         = 5 * time.Minute // Zaman penceresi (5 dakika)
	AI_RATE_LIMIT_MAX_TOKENS     = 50000           // Zaman penceresi içinde kullanılabilecek maksimum token sayısı
	AI_RATE_LIMIT_MAX_REQUESTS   = 25              // Zaman penceresinde maksimum istek sayısı
	AI_RATE_LIMIT_REQ_PER_MINUTE = 5               // Dakika başına maksimum istek sayısı

	// Session Rules
	REFRESH_TOKEN_LENGTH   = 32
	REFRESH_TOKEN_DURATION = 30 * 24 * time.Hour
	REFRESH_TOKEN_NAME     = "guideofdubai_blog_refresh_token"
	ACCESS_TOKEN_NAME      = "guideofdubai_blog_access_token"
	ACCESS_TOKEN_DURATION  = 1 * time.Minute
	JWT_ISSUER             = "guideofdubai-blog"
)
