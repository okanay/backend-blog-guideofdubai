package configs

import (
	"time"
)

const (
	// Project Rules
	PROJECT_NAME = "Guide Of Dubai - Blog"

	// STATS RULES
	VIEW_CACHE_EXPIRATION = 1 * time.Minute

	AI_RATE_LIMIT_WINDOW         = 10 * time.Minute
	AI_RATE_LIMIT_MAX_TOKENS     = 10000
	AI_RATE_LIMIT_MAX_REQUESTS   = 30
	AI_RATE_LIMIT_REQ_PER_MINUTE = 3

	// Session Rules
	REFRESH_TOKEN_LENGTH   = 32
	REFRESH_TOKEN_DURATION = 30 * 24 * time.Hour
	REFRESH_TOKEN_NAME     = "guideofdubai_blog_refresh_token"
	ACCESS_TOKEN_NAME      = "guideofdubai_blog_access_token"
	ACCESS_TOKEN_DURATION  = 1 * time.Minute
	JWT_ISSUER             = "guideofdubai-blog"
)
