package constants

import (
	"time"
)

const (
	// TokenExpiryDuration is the duration for which a JWT token is valid (10 minutes)
	TokenExpiryDuration = 10 * time.Minute

	// TokenExpirySeconds is the expiry time in seconds for API responses
	TokenExpirySeconds = 600 // 10 minutes * 60 seconds

	// RedisTokenPrefix is the prefix used for token keys in Redis
	RedisTokenPrefix = "token:"
)
