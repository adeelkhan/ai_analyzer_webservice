package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adeelkhan/code_diff/internal/auth"
	"github.com/adeelkhan/code_diff/internal/middleware"
	"github.com/adeelkhan/code_diff/internal/models"

	"github.com/gin-gonic/gin"
)

func TestUserRepository_LoadData(t *testing.T) {
	repo := models.DefaultUserRepository()
	users, err := repo.LoadData()
	if err != nil {
		t.Fatalf("LoadData() error = %v", err)
	}
	if len(users) == 0 {
		t.Error("Expected users to be loaded")
	}
}

func TestUserRepository_CalculateSumAge(t *testing.T) {
	repo := models.DefaultUserRepository()
	response, err := repo.CalculateSumAge()
	if err != nil {
		t.Fatalf("CalculateSumAge() error = %v", err)
	}
	if response.TotalUsers == 0 {
		t.Error("Expected users to be present")
	}
	if response.SumOfAges <= 0 {
		t.Error("Expected sum of ages to be positive")
	}
}

func TestUserRepository_FindUsersByCountry(t *testing.T) {
	repo := models.DefaultUserRepository()
	response, err := repo.FindUsersByCountry("USA")
	if err != nil {
		t.Fatalf("FindUsersByCountry() error = %v", err)
	}
	if response.Country != "USA" {
		t.Errorf("Expected country USA, got %s", response.Country)
	}
}

func TestUserRepository_FindUsersOlderThan(t *testing.T) {
	repo := models.DefaultUserRepository()
	response, err := repo.FindUsersOlderThan(30)
	if err != nil {
		t.Fatalf("FindUsersOlderThan() error = %v", err)
	}
	if response.MinAge != 30 {
		t.Errorf("Expected min_age 30, got %d", response.MinAge)
	}
}

func TestSumAgeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.JWTMiddleware())
	router.GET("/analytics/sum_age", SumAge)

	token, _ := auth.GenerateToken("testuser", "test@example.com")
	req, _ := http.NewRequest("GET", "/analytics/sum_age", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.SumAgeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
		return
	}

	if response.TotalUsers == 0 {
		t.Error("Expected users to be loaded")
	}
}

func TestUsersByCountryHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.JWTMiddleware())
	router.POST("/analytics/users_by_country", UsersByCountry)

	tests := []struct {
		name      string
		country   string
		wantCode  int
		wantUsers bool
	}{
		{
			name:      "valid country",
			country:   "USA",
			wantCode:  http.StatusOK,
			wantUsers: true,
		},
		{
			name:      "country with no users",
			country:   "NonExistent",
			wantCode:  http.StatusOK,
			wantUsers: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(UsersByCountryRequest{Country: tt.country})
			token, _ := auth.GenerateToken("testuser", "test@example.com")

			req, _ := http.NewRequest("POST", "/analytics/users_by_country", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, w.Code)
			}

			if tt.wantUsers {
				var response models.UsersByCountryResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}
				if response.Country != tt.country {
					t.Errorf("Expected country '%s', got '%s'", tt.country, response.Country)
				}
			}
		})
	}
}

func TestUsersOlderThanHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.JWTMiddleware())
	router.POST("/analytics/users_older_than", UsersOlderThan)

	tests := []struct {
		name     string
		age      int
		wantCode int
	}{
		{
			name:     "age 30",
			age:      30,
			wantCode: http.StatusOK,
		},
		{
			name:     "age 50",
			age:      50,
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(UsersOlderThanRequest{Age: tt.age})
			token, _ := auth.GenerateToken("testuser", "test@example.com")

			req, _ := http.NewRequest("POST", "/analytics/users_older_than", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, w.Code)
			}

			var response models.UsersOlderThanResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to unmarshal response: %v", err)
				return
			}

			if response.MinAge != tt.age {
				t.Errorf("Expected min_age %d, got %d", tt.age, response.MinAge)
			}
		})
	}
}
