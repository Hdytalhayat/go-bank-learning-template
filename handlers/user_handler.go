// go-bank-app/handlers/user_handler.go
package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"go-bank-app/config" // Import package config kita
	"go-bank-app/models" // Import package models kita

	"github.com/gin-gonic/gin"
)

// createUser handles POST /users
func CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passwordHash := req.Password // Placeholder for hashing (will be implemented in Phase 4)

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

// getUserByID handles GET /users/:id
func GetUserByID(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?"
	err := config.DB.QueryRow(query, id).
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

// getAllUsers handles GET /users
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
