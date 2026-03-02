package redis

import (
	"context"

	"github.com/adeelkhan/code_diff/internal/logger"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	ctx    = context.Background()
	log    = logger.GetLogger()
)

// TokenPrefix is the prefix used for token keys in Redis
const TokenPrefix = "token:"

// Init initializes the Redis client
func Init(addr string) {
	log.Info("Initializing Redis client at %s", addr)
	client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
}

// GetClient returns the Redis client
func GetClient() *redis.Client {
	return client
}
