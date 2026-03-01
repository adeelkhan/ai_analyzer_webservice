package models

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

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
