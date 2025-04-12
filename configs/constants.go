package configs

import (
	"time"
)

const (
	// Project Rules
	PROJECT_NAME = "Guide Of Dubai - Blog"

	// Session Rules
	REFRESH_TOKEN_LENGTH   = 32
	REFRESH_TOKEN_DURATION = 30 * 24 * time.Hour
	REFRESH_TOKEN_NAME     = "guideofdubai_blog_refresh_token"
	ACCESS_TOKEN_NAME      = "guideofdubai_blog_access_token"
	ACCESS_TOKEN_DURATION  = 30 * time.Minute
	JWT_ISSUER             = "guideofdubai-blog"
)
