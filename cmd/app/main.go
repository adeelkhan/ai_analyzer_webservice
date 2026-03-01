package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adeelkhan/code_diff/cmd/app/handlers"
	"github.com/adeelkhan/code_diff/logger"
	"github.com/adeelkhan/code_diff/middleware"
	"github.com/adeelkhan/code_diff/utils"
	"github.com/gin-gonic/gin"
)

var log = logger.GetLogger()

func main() {
	log.Info("Starting application...")

	// Initialize Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	util.InitRedis(redisURL)

	// Test Redis connection
	client := util.GetRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Error("Redis connection failed: %v", err)
		log.Warn("Continuing without Redis - token storage will be disabled")
	} else {
		log.Info("Connected to Redis successfully at %s", redisURL)
	}

	r := gin.Default()

	// Public routes (no JWT required)
	r.GET("/health", handlers.HealthCheck)
	r.GET("/hello", handlers.HelloWorld)
	r.POST("/get_token", handlers.GetToken)

	// Protected routes (JWT required)
	protected := r.Group("/")
	protected.Use(middleware.JWTMiddleware())
	protected.POST("/process_request", handlers.ProcessRequest)
	protected.POST("/refresh_token", handlers.RefreshToken)

	// CSV Analytics endpoints (protected)
	protected.GET("/analytics/sum_age", handlers.SumAge)
	protected.POST("/analytics/users_by_country", handlers.UsersByCountry)
	protected.POST("/analytics/users_older_than", handlers.UsersOlderThan)

	// Start server in goroutine
	go func() {
		log.Info("Server starting on port 9991")
		if err := r.Run(":9991"); err != nil {
			log.Error("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
}
