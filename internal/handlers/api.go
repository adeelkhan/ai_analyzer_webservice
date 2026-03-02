package handlers

import (
	"net/http"

	"github.com/adeelkhan/code_diff/internal/auth"
	"github.com/adeelkhan/code_diff/internal/logger"

	"github.com/gin-gonic/gin"
)

var log = logger.GetLogger()

// LoginRequest represents the login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
	Type      string `json:"type"`
}

// Request represents the request body for processing
type Request struct {
	// TODO: Add your request fields here
}

// Response represents the response body for processing
type Response struct {
	// TODO: Add your response fields here
}

// GetToken validates credentials and returns a JWT token
func GetToken(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Get token failed: invalid request body - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement proper user validation (check against database)
	// For now, accept any username/password combination
	// In production, validate against a database or external service
	if req.Username == "" || req.Password == "" {
		log.Warn("Get token failed: empty credentials for username %s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	log.Info("Login attempt for user: %s", req.Username)
	token, err := auth.GenerateToken(req.Username, req.Username+"@example.com")
	if err != nil {
		log.Error("Failed to generate token for user %s: %v", req.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info("Token generated successfully for user: %s", req.Username)
	c.JSON(http.StatusOK, TokenResponse{
		Token:     token,
		ExpiresIn: auth.TokenExpirySeconds,
		Type:      "Bearer",
	})
}

// RefreshToken generates a new token and replaces the old one in Redis
func RefreshToken(c *gin.Context) {
	log.Info("Refresh token requested")

	// Get user info from context (set by JWTMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Warn("Refresh token attempted without authenticated user")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	email, _ := c.Get("email")

	userIDStr, ok := userID.(string)
	if !ok {
		log.Error("Invalid user ID type in refresh token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	emailStr, _ := email.(string)

	// Use auth.RefreshToken
	token, err := auth.RefreshToken(userIDStr, emailStr)
	if err != nil {
		log.Error("Failed to refresh token for user %s: %v", userIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info("Token refreshed successfully for user %s", userIDStr)
	c.JSON(http.StatusOK, TokenResponse{
		Token:     token,
		ExpiresIn: auth.TokenExpirySeconds,
		Type:      "Bearer",
	})
}

// ProcessRequest handles the main request processing
func ProcessRequest(c *gin.Context) {
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Process request failed: invalid request body - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")

	userIDStr, _ := userID.(string)
	log.Info("Processing request for user: %s", userIDStr)
	_ = email

	c.JSON(http.StatusOK, Response{})
}

// HealthCheck returns the health status of the service
func HealthCheck(c *gin.Context) {
	log.Info("Health check requested")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// HelloWorld returns a simple hello world message
func HelloWorld(c *gin.Context) {
	log.Info("Hello world endpoint requested")
	c.JSON(http.StatusOK, gin.H{"message": "hello world"})
}
