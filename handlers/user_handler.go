// go-bank-app/handlers/user_handler.go
package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"go-bank-app/config" // Import our config package
	"go-bank-app/models" // Import our models package

	"github.com/gin-gonic/gin"
)

// CreateUser handles POST /users
// This function creates a new user by reading JSON input, inserting it into the database,
// and returning the newly created user details.
func CreateUser(c *gin.Context) {
	var req models.CreateUserRequest

	// Bind JSON body to request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Note: password hashing should be implemented in a later phase
	passwordHash := req.Password // Placeholder for password hashing

	// Insert the new user into the database
	query := "INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)"
	result, err := config.DB.Exec(query, req.Name, req.Email, passwordHash)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Get the ID of the newly inserted user
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user ID"})
		return
	}

	// Retrieve the newly created user from the database
	var newUser models.User
	err = config.DB.QueryRow("SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?", id).
		Scan(&newUser.ID, &newUser.Name, &newUser.Email, &newUser.CreatedAt, &newUser.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching new user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve newly created user"})
		return
	}

	// Respond with the created user data
	c.JSON(http.StatusCreated, newUser)
}

// GetUserByID handles GET /users/:id
// This function retrieves a user by ID from the database and returns their details.
func GetUserByID(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?"
	err := config.DB.QueryRow(query, id).
		Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// User not found
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			log.Printf("Error getting user by ID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		}
		return
	}

	// Respond with user data
	c.JSON(http.StatusOK, user)
}

// GetAllUsers handles GET /users
// This function retrieves all users from the database and returns a list.
func GetAllUsers(c *gin.Context) {
	var users []models.User

	// Query all user records
	rows, err := config.DB.Query("SELECT id, name, email, created_at, updated_at FROM users")
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	defer rows.Close()

	// Iterate through result rows
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning user row: %v", err)
			continue // Skip this row and continue to the next
		}
		users = append(users, user)
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving users during iteration"})
		return
	}

	// Respond with the list of users
	c.JSON(http.StatusOK, users)
}
