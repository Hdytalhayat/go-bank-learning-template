// go-bank-app/handlers/auth_handler.go
package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"go-bank-app/auth"   // Import package auth kita
	"go-bank-app/config" // Import config untuk DB
	"go-bank-app/models" // Import models

	"github.com/gin-gonic/gin"
)

// RegisterUser handles POST /auth/register
func RegisterUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek apakah email sudah terdaftar
	var existingUserID int
	err := config.DB.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&existingUserID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	} else if err != sql.ErrNoRows {
		log.Printf("Error checking existing user email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Simpan user baru ke database
	query := "INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)"
	result, err := config.DB.Exec(query, req.Name, req.Email, hashedPassword)
	if err != nil {
		log.Printf("Error inserting new user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID for new user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user ID"})
		return
	}

	var newUser models.User
	err = config.DB.QueryRow("SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?", id).
		Scan(&newUser.ID, &newUser.Name, &newUser.Email, &newUser.CreatedAt, &newUser.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching new user after registration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve newly registered user"})
		return
	}

	c.JSON(http.StatusCreated, newUser)
}

// LoginUser handles POST /auth/login
func LoginUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cari user berdasarkan email
	var user models.User
	query := "SELECT id, name, email, password_hash FROM users WHERE email = ?"
	err := config.DB.QueryRow(query, req.Email).
		Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		} else {
			log.Printf("Error fetching user for login: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process login"})
		}
		return
	}

	// Verifikasi password
	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateJWTToken(user.ID)
	if err != nil {
		log.Printf("Error generating JWT token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token})
}
