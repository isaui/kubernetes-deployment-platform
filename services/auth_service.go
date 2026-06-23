package services

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pendeploy-simple/database"
	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// This file uses dto.TokenClaims, dto.LoginRequest, dto.RegisterRequest, and dto.AuthResponse
// which are defined in the dto/auth.go file

// EnsureAdminExists creates or promotes the default admin user when configured.
func EnsureAdminExists() error {
	email := strings.TrimSpace(os.Getenv("DEFAULT_ADMIN_EMAIL"))
	password := os.Getenv("DEFAULT_ADMIN_PASSWORD")

	if email == "" && password == "" {
		log.Println("Default admin bootstrap skipped: DEFAULT_ADMIN_EMAIL and DEFAULT_ADMIN_PASSWORD are not set")
		return nil
	}

	if email == "" || password == "" {
		return errors.New("DEFAULT_ADMIN_EMAIL and DEFAULT_ADMIN_PASSWORD must both be set")
	}

	var user models.User
	result := database.DB.Where("email = ?", email).First(&user)
	if result.Error == nil {
		if user.Role != models.RoleAdmin {
			if err := database.DB.Model(&user).Update("role", models.RoleAdmin).Error; err != nil {
				return err
			}
			log.Printf("Default admin user promoted: %s", email)
		}
		return nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	username := optionalEnvString("DEFAULT_ADMIN_USERNAME")
	name := optionalEnvString("DEFAULT_ADMIN_NAME")
	admin := models.User{
		Email:    email,
		Password: string(hashedPassword),
		Username: username,
		Name:     name,
		Role:     models.RoleAdmin,
	}

	if err := database.DB.Create(&admin).Error; err != nil {
		return err
	}

	log.Printf("Default admin user created: %s", email)
	return nil
}

func optionalEnvString(key string) *string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}
	return &value
}

// Register creates a new user account
func Register(req dto.RegisterRequest) (*models.User, error) {
	// Check if email already exists
	var existingUser models.User
	result := database.DB.Where("email = ?", req.Email).First(&existingUser)
	if result.RowsAffected > 0 {
		return nil, errors.New("email already registered")
	}

	// Check if username exists if provided
	if req.Username != nil && *req.Username != "" {
		result = database.DB.Where("username = ?", req.Username).First(&existingUser)
		if result.RowsAffected > 0 {
			return nil, errors.New("username already taken")
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create new user
	user := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Username: req.Username,
		Name:     req.Name,
		Role:     models.RoleUser,
	}

	// Save user to database
	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUser retrieves a user by ID
func GetUser(id string) (*models.User, error) {
	var user models.User
	result := database.DB.Where("id = ?", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// Login authenticates a user and returns a token
func Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Find user by email
	var user models.User
	result := database.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate token
	token, expiresAt, err := GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, err
	}

	// Clear password from response
	responseUser := user
	responseUser.Password = ""

	return &dto.AuthResponse{
		Token:     token,
		User:      responseUser,
		ExpiresAt: expiresAt,
	}, nil
}

// GenerateToken generates a new JWT token for a user
func GenerateToken(userID, email, role string) (string, time.Time, error) {
	// Get secret key from environment
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", time.Time{}, errors.New("JWT_SECRET not set in environment")
	}

	// Set expiration time
	expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	// Create claims with expiry time
	claims := dto.TokenClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns claims if valid
func ValidateToken(tokenString string) (*dto.TokenClaims, error) {
	// Get secret key from environment
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, errors.New("JWT_SECRET not set in environment")
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &dto.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Get claims
	claims, ok := token.Claims.(*dto.TokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
