package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/adeelkhan/code_diff/internal/auth"
	"github.com/adeelkhan/code_diff/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var testLog = logger.GetLogger()

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", HealthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestHelloWorld(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/hello", HelloWorld)

	req, _ := http.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "hello world" {
		t.Errorf("Expected message 'hello world', got '%s'", response["message"])
	}
}

func TestGetToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/get_token", GetToken)

	tests := []struct {
		name       string
		body       string
		wantCode   int
		wantToken  bool
		wantErrMsg string
	}{
		{
			name:      "valid credentials",
			body:      `{"username":"testuser","password":"testpass"}`,
			wantCode:  http.StatusOK,
			wantToken: true,
		},
		{
			name:       "missing username",
			body:       `{"password":"testpass"}`,
			wantCode:   http.StatusBadRequest,
			wantToken:  false,
			wantErrMsg: "Username",
		},
		{
			name:       "missing password",
			body:       `{"username":"testuser"}`,
			wantCode:   http.StatusBadRequest,
			wantToken:  false,
			wantErrMsg: "Password",
		},
		{
			name:       "empty body",
			body:       `{}`,
			wantCode:   http.StatusBadRequest,
			wantToken:  false,
			wantErrMsg: "",
		},
		{
			name:       "invalid json",
			body:       `invalid json`,
			wantCode:   http.StatusBadRequest,
			wantToken:  false,
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/get_token", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, w.Code)
			}

			if tt.wantToken {
				var response TokenResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal token response: %v", err)
					return
				}
				if response.Token == "" {
					t.Error("Expected token in response, got empty string")
				}
				if response.Type != "Bearer" {
					t.Errorf("Expected token type 'Bearer', got '%s'", response.Type)
				}
				if response.ExpiresIn != 10*60 {
					t.Errorf("Expected expires_in %d, got %d", 10*60, response.ExpiresIn)
				}
			}

			if tt.wantErrMsg != "" {
				var response map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
					return
				}
				if !strings.Contains(response["error"], tt.wantErrMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.wantErrMsg, response["error"])
				}
			}
		})
	}
}

func TestGetToken_ValidTokenWorks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/get_token", GetToken)

	req, _ := http.NewRequest("POST", "/get_token", bytes.NewBufferString(`{"username":"testuser","password":"testpass"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get token: expected status %d, got %d", http.StatusOK, w.Code)
	}

	var tokenResponse TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &tokenResponse); err != nil {
		t.Fatalf("Failed to unmarshal token response: %v", err)
	}

	// Use token to access protected endpoint
	protectedRouter := gin.New()
	protectedRouter.Use(func(c *gin.Context) {
		claims, err := auth.ValidateToken(strings.Split(c.GetHeader("Authorization"), " ")[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	})
	protectedRouter.POST("/process_request", ProcessRequest)

	req2, _ := http.NewRequest("POST", "/process_request", bytes.NewBufferString("{}"))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+tokenResponse.Token)

	w2 := httptest.NewRecorder()
	protectedRouter.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w2.Code)
	}
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		email   string
		wantErr bool
	}{
		{
			name:    "valid token",
			userID:  "user123",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "empty userID",
			userID:  "",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			userID:  "user123",
			email:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.GenerateToken(tt.userID, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	validToken, _ := auth.GenerateToken("user123", "test@example.com")

	tests := []struct {
		name       string
		token      string
		wantErr    bool
		wantUserID string
		wantEmail  string
	}{
		{
			name:       "valid token",
			token:      validToken,
			wantErr:    false,
			wantUserID: "user123",
			wantEmail:  "test@example.com",
		},
		{
			name:    "invalid token",
			token:   "invalid.token.here",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := auth.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if claims.UserID != tt.wantUserID {
					t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, tt.wantUserID)
				}
				if claims.Email != tt.wantEmail {
					t.Errorf("ValidateToken() Email = %v, want %v", claims.Email, tt.wantEmail)
				}
			}
		})
	}
}

func TestValidateToken_Expired(t *testing.T) {
	claims := auth.JWTClaims{
		UserID: "user123",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("your-secret-key-change-in-production"))

	_, err := auth.ValidateToken(tokenString)
	if err == nil {
		t.Error("ValidateToken() should fail for expired token")
	}
}

func TestJWTMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		authHeader string
		wantCode   int
		wantErr    bool
	}{
		{
			name:       "missing authorization header",
			authHeader: "",
			wantCode:   http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:       "invalid format",
			authHeader: "InvalidFormat",
			wantCode:   http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid.token.here",
			wantCode:   http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				authHeader := c.GetHeader("Authorization")
				if authHeader == "" {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
					c.Abort()
					return
				}

				bearerToken := strings.Split(authHeader, " ")
				if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
					c.Abort()
					return
				}

				claims, err := auth.ValidateToken(bearerToken[1])
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
					c.Abort()
					return
				}

				c.Set("user_id", claims.UserID)
				c.Set("email", claims.Email)
				c.Next()
			})
			router.GET("/protected", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, w.Code)
			}

			if tt.wantErr {
				var response map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
				}
				if response["error"] == "" {
					t.Error("Expected error in response")
				}
			}
		})
	}
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(bearerToken[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	})
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"email":   email,
		})
	})

	token, _ := auth.GenerateToken("user123", "test@example.com")
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["user_id"] != "user123" {
		t.Errorf("Expected user_id 'user123', got '%s'", response["user_id"])
	}

	if response["email"] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", response["email"])
	}
}

func TestProcessRequest_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}
		c.Next()
	})
	router.POST("/process_request", ProcessRequest)

	req, _ := http.NewRequest("POST", "/process_request", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if !strings.Contains(response["error"], "authorization header required") {
		t.Errorf("Expected error containing 'authorization header required', got '%s'", response["error"])
	}
}

func TestProcessRequest_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		bearerToken := strings.Split(c.GetHeader("Authorization"), " ")
		if len(bearerToken) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid"})
			c.Abort()
			return
		}
		_, err := auth.ValidateToken(bearerToken[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Next()
	})
	router.POST("/process_request", ProcessRequest)

	req, _ := http.NewRequest("POST", "/process_request", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestProcessRequest_ValidToken_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		bearerToken := strings.Split(c.GetHeader("Authorization"), " ")
		claims, err := auth.ValidateToken(bearerToken[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	})
	router.POST("/process_request", ProcessRequest)

	token, _ := auth.GenerateToken("user123", "test@example.com")
	req, _ := http.NewRequest("POST", "/process_request", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
}

func TestProcessRequest_ValidToken_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		bearerToken := strings.Split(c.GetHeader("Authorization"), " ")
		claims, err := auth.ValidateToken(bearerToken[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	})
	router.POST("/process_request", ProcessRequest)

	token, _ := auth.GenerateToken("user123", "test@example.com")
	req, _ := http.NewRequest("POST", "/process_request", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupAuth  bool
		wantCode   int
		wantToken  bool
		wantErrMsg string
	}{
		{
			name:      "valid token refresh",
			setupAuth: true,
			wantCode:  http.StatusOK,
			wantToken: true,
		},
		{
			name:       "missing authorization",
			setupAuth:  false,
			wantCode:   http.StatusUnauthorized,
			wantToken:  false,
			wantErrMsg: "authorization",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if !tt.setupAuth {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					c.Abort()
					return
				}
				token := strings.Split(c.GetHeader("Authorization"), " ")[1]
				claims, _ := auth.ValidateToken(token)
				c.Set("user_id", claims.UserID)
				c.Set("email", claims.Email)
				c.Next()
			})
			router.POST("/refresh_token", RefreshToken)

			req, _ := http.NewRequest("POST", "/refresh_token", nil)
			req.Header.Set("Content-Type", "application/json")

			if tt.setupAuth {
				token, _ := auth.GenerateToken("user123", "test@example.com")
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, w.Code)
			}

			if tt.wantToken {
				var response TokenResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal token response: %v", err)
					return
				}
				if response.Token == "" {
					t.Error("Expected token in response, got empty string")
				}
				if response.Type != "Bearer" {
					t.Errorf("Expected token type 'Bearer', got '%s'", response.Type)
				}
			}

			if tt.wantErrMsg != "" {
				var response map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
					return
				}
				if response["error"] == "" {
					t.Error("Expected error in response")
				}
			}
		})
	}
}

func TestValidateToken_AutoCleanupExpired(t *testing.T) {
	// Generate a token
	token, err := auth.GenerateToken("cleanupuser", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Manually set token expiry in Redis to past
	redisClient := auth.GetRedisClient()
	ctx := context.Background()
	if redisClient != nil {
		// Set token with very short TTL that expires
		redisClient.Set(ctx, auth.RedisTokenPrefix+"cleanupuser", token, 1*time.Millisecond)
		time.Sleep(2 * time.Millisecond) // Wait for expiry

		// Try to validate - should fail and clean up
		_, err := auth.ValidateToken(token)
		if err == nil {
			t.Error("Expected error for expired token")
		}

		// Check that token was removed from Redis
		_, err = redisClient.Get(ctx, auth.RedisTokenPrefix+"cleanupuser").Result()
		if err != redis.Nil {
			t.Error("Expected token to be cleaned up from Redis")
		}
	}
}
