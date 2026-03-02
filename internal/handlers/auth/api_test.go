package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adeelkhan/code_diff/internal/auth"

	"github.com/gin-gonic/gin"
)

func TestHealthCheckHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", HealthCheckHandler)

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

func TestGetTokenHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/get_token", GetTokenHandler)

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

func TestJWTMiddlewareHandler(t *testing.T) {
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

func TestRefreshTokenHandler_MissingAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	})
	router.POST("/refresh_token", RefreshTokenHandler)

	req, _ := http.NewRequest("POST", "/refresh_token", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestJWTMiddlewareHandler_ValidToken(t *testing.T) {
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

func TestRefreshTokenHandler_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		token := strings.Split(c.GetHeader("Authorization"), " ")[1]
		claims, _ := auth.ValidateToken(token)
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	})
	router.POST("/refresh_token", RefreshTokenHandler)

	token, _ := auth.GenerateToken("user123", "test@example.com")
	req, _ := http.NewRequest("POST", "/refresh_token", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal token response: %v", err)
	}
	if response.Token == "" {
		t.Error("Expected token in response, got empty string")
	}
	if response.Type != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got '%s'", response.Type)
	}
}
