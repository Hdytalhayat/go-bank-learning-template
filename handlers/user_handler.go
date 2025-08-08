// go-bank-app/handlers/user_handler.go (Modifikasi)
package handlers

import (
	"database/sql"
	"net/http"
	"strconv" // Tambahkan untuk strconv.Atoi

	"go-bank-app/services" // Import service

	"github.com/gin-gonic/gin"
)

// UserHandler struct untuk dependensi service
type UserHandler struct {
	UserService services.UserService
}

// NewUserHandler membuat instance baru dari UserHandler
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

// GetUserByID handles GET /users/:id
func (h *UserHandler) GetUserByID(c *gin.Context) { // Perhatikan receiver 'h *UserHandler'
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

	user, err := h.UserService.GetUserByID(requestedUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetAllUsers handles GET /users
func (h *UserHandler) GetAllUsers(c *gin.Context) { // Perhatikan receiver 'h *UserHandler'
	users, err := h.UserService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	c.JSON(http.StatusOK, users)
}
