package dto

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pendeploy-simple/models"
)

// TokenClaims represents our custom JWT claims
type TokenClaims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	Username *string `json:"username"`
	Name     *string `json:"name"`
}

// AuthResponse represents the response after authentication
type AuthResponse struct {
	Token     string      `json:"token"`
	User      models.User `json:"user"`
	ExpiresAt time.Time   `json:"expiresAt"`
}
