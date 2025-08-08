// go-bank-app/models/user.go
package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name" binding:"required"`
	Email        string    `json:"email" binding:"required,email"`
	PasswordHash string    `json:"-"` // "-" agar tidak disertakan dalam JSON response
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Struct untuk request membuat user baru (tanpa ID dan timestamp)
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}
