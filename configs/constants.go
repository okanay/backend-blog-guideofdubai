package configs

import (
	"time"
)

const (
	// Project Rules
	PROJECT_NAME = "Guide Of Dubai - Blog"

	// Session Rules
	SESSION_DURATION           = 30 * 24 * time.Hour
	SESSION_REFRESH_TOKEN_NAME = "guideofdubai_refresh_token"
	SESSION_ACCESS_TOKEN_NAME  = "guideofdubai_access_token"
)
