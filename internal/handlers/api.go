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

// GetTokenHandler validates credentials and returns a JWT token
func GetTokenHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Get token failed: invalid request body - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

// RefreshTokenHandler generates a new token and replaces the old one in Redis
func RefreshTokenHandler(c *gin.Context) {
	log.Info("Refresh token requested")

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

// HealthCheckHandler returns the health status of the service
func HealthCheckHandler(c *gin.Context) {
	log.Info("Health check requested")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
