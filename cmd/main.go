package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adeelkhan/code_diff/internal/handlers/analytics"
	"github.com/adeelkhan/code_diff/internal/handlers/auth"
	"github.com/adeelkhan/code_diff/internal/infra/redis"
	"github.com/adeelkhan/code_diff/internal/logger"
	"github.com/adeelkhan/code_diff/internal/middleware"

	"github.com/gin-gonic/gin"
)

var log = logger.GetLogger()

func main() {
	log.Info("Starting application...")

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	redis.Init(redisURL)

	client := redis.GetClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Error("Redis connection failed: %v", err)
		log.Warn("Continuing without Redis - token storage will be disabled")
	} else {
		log.Info("Connected to Redis successfully at %s", redisURL)
	}

	r := gin.Default()

	r.GET("/health", auth.HealthCheckHandler)
	r.POST("/get_token", auth.GetTokenHandler)

	protected := r.Group("/")
	protected.Use(middleware.JWTMiddleware())
	protected.POST("/refresh_token", auth.RefreshTokenHandler)

	protected.GET("/analytics/sum_age", analytics.SumAgeHandler)
	protected.POST("/analytics/users_by_country", analytics.UsersByCountryHandler)
	protected.POST("/analytics/users_older_than", analytics.UsersOlderThanHandler)

	go func() {
		log.Info("Server starting on port 9991")
		if err := r.Run(":9991"); err != nil {
			log.Error("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
}
