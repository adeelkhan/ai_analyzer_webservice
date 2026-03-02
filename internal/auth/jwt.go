package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adeelkhan/code_diff/internal/infra/redis"
	"github.com/adeelkhan/code_diff/internal/logger"

	"github.com/golang-jwt/jwt/v5"
	redislib "github.com/redis/go-redis/v9"
)

var (
	jwtSecret = []byte("your-secret-key-change-in-production")
	ctx       = context.Background()
	log       = logger.GetLogger()
)

// TokenExpiryDuration is the duration for which a JWT token is valid (10 minutes)
const TokenExpiryDuration = 10 * time.Minute

// TokenExpirySeconds is the expiry time in seconds for API responses
const TokenExpirySeconds = 600 // 10 minutes * 60 seconds

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token and stores it in Redis
func GenerateToken(userID, email string) (string, error) {
	log.Info("Generating token for user: %s", userID)

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiryDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Error("Failed to sign token for user %s: %v", userID, err)
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	// Store token in Redis with configured expiry duration
	client := redis.GetClient()
	if client != nil {
		err = client.Set(ctx, redis.TokenPrefix+userID, tokenString, TokenExpiryDuration).Err()
		if err != nil {
			log.Error("Failed to store token in Redis for user %s: %v", userID, err)
			return "", fmt.Errorf("failed to store token in redis: %w", err)
		}
		log.Info("Token stored in Redis for user %s with expiry %v", userID, TokenExpiryDuration)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and checks Redis
func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Info("Token expired, attempting cleanup")
			if claims, ok := token.Claims.(*JWTClaims); ok {
				client := redis.GetClient()
				if client != nil {
					client.Del(ctx, redis.TokenPrefix+claims.UserID)
					log.Info("Cleaned up expired token for user %s from Redis", claims.UserID)
				}
			}
		}
		log.Error("Token validation failed: %v", err)
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		log.Error("Invalid token claims")
		return nil, errors.New("invalid token")
	}

	client := redis.GetClient()
	if client != nil {
		storedToken, err := client.Get(ctx, redis.TokenPrefix+claims.UserID).Result()
		if errors.Is(err, redislib.Nil) {
			log.Error("Token not found in Redis for user %s", claims.UserID)
			return nil, errors.New("token not found in redis")
		} else if err != nil {
			log.Error("Redis error while validating token: %v", err)
			return nil, fmt.Errorf("redis error: %w", err)
		}

		if storedToken != tokenString {
			log.Error("Token mismatch for user %s", claims.UserID)
			return nil, errors.New("token mismatch")
		}
	}

	log.Info("Token validated successfully for user %s", claims.UserID)
	return claims, nil
}

// InvalidateToken removes token from Redis
func InvalidateToken(userID string) error {
	client := redis.GetClient()
	if client != nil {
		err := client.Del(ctx, redis.TokenPrefix+userID).Err()
		if err != nil {
			log.Error("Failed to invalidate token for user %s: %v", userID, err)
			return err
		}
		log.Info("Token invalidated for user %s", userID)
	}
	return nil
}

// RefreshToken generates a new token and replaces the old one in Redis
func RefreshToken(userID, email string) (string, error) {
	log.Info("Refreshing token for user: %s", userID)

	client := redis.GetClient()
	if client != nil {
		err := client.Del(ctx, redis.TokenPrefix+userID).Err()
		if err != nil {
			log.Error("Failed to invalidate old token for user %s: %v", userID, err)
			return "", fmt.Errorf("failed to invalidate old token: %w", err)
		}
		log.Info("Old token invalidated for user %s before refresh", userID)
	}

	token, err := GenerateToken(userID, email)
	if err != nil {
		log.Error("Failed to generate refreshed token for user %s: %v", userID, err)
		return "", err
	}

	log.Info("Token refreshed successfully for user %s", userID)
	return token, nil
}
