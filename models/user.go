package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" binding:"required"`
	Email     string    `json:"email" binding:"required,email" gorm:"uniqueIndex"`
	Password  string    `json:"-" binding:"required,min=6"`
	Role      string    `json:"role" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserResponse is the structure returned to clients (excludes password)
type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToUserResponse converts User to UserResponse
func (u *User) ToUserResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
