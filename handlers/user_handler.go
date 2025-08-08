package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"go-bank-app/config"
	"go-bank-app/models"

	"github.com/gin-gonic/gin"
)

// CreateUser handles POST /users
func CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passwordHash := req.Password // Placeholder: ganti dengan hash password di fase berikutnya

	query := "INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)"
	result, err := config.DB.Exec(query, req.Name, req.Email, passwordHash)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user ID"})
		return
	}

	var newUser models.User
	err = config.DB.QueryRow("SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?", id).
		Scan(&newUser.ID, &newUser.Name, &newUser.Email, &newUser.CreatedAt, &newUser.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching new user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve newly created user"})
		return
	}

	c.JSON(http.StatusCreated, newUser)
}

// GetUserByID handles GET /users/:id (Hanya boleh akses profil sendiri)
func GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	requestedUserID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	loggedInUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	if requestedUserID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to access this user's profile"})
		return
	}

	var user models.User
	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?"
	err = config.DB.QueryRow(query, requestedUserID).
		Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			log.Printf("Error getting user by ID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetAllUsers handles GET /users
// PERINGATAN: Endpoint ini mengembalikan seluruh daftar user.
// Disarankan membatasi hanya untuk admin (di fase 5 tambahkan role-based access).
func GetAllUsers(c *gin.Context) {
	var users []models.User
	rows, err := config.DB.Query("SELECT id, name, email, created_at, updated_at FROM users")
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning user row: %v", err)
			continue
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving users during iteration"})
		return
	}

	c.JSON(http.StatusOK, users)
}
