package middleware

import (
	"net/http"
	"strings"

	"github.com/adeelkhan/code_diff/logger"
	"github.com/adeelkhan/code_diff/utils"

	"github.com/gin-gonic/gin"
)

var log = logger.GetLogger()

// JWTMiddleware validates JWT tokens
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("JWT validation failed: authorization header missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			log.Warn("JWT validation failed: invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := util.ValidateToken(bearerToken[1])
		if err != nil {
			log.Warn("JWT validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		log.Info("User %s authenticated successfully", claims.UserID)
		c.Next()
	}
}
