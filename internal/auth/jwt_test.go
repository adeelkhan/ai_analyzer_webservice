package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
			token, err := GenerateToken(tt.userID, tt.email)
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
	validToken, _ := GenerateToken("user123", "test@example.com")

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
			claims, err := ValidateToken(tt.token)
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
	claims := JWTClaims{
		UserID: "user123",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("your-secret-key-change-in-production"))

	_, err := ValidateToken(tokenString)
	if err == nil {
		t.Error("ValidateToken() should fail for expired token")
	}
}
