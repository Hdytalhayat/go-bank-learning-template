// go-bank-app/handlers/auth_handler.go (Modifikasi)
package handlers

import (
	"net/http"

	"go-bank-app/models"
	"go-bank-app/services" // Import service

	"github.com/gin-gonic/gin"
)

// AuthHandler struct untuk dependensi service
type AuthHandler struct {
	UserService services.UserService
}

// NewAuthHandler membuat instance baru dari AuthHandler
func NewAuthHandler(userService services.UserService) *AuthHandler {
	return &AuthHandler{UserService: userService}
}

// RegisterUser handles POST /auth/register
func (h *AuthHandler) RegisterUser(c *gin.Context) { // Perhatikan receiver 'h *AuthHandler'
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newUser, err := h.UserService.RegisterUser(&req)
	if err != nil {
		if err.Error() == "email sudah terdaftar" { // Contoh penanganan error spesifik dari service
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, newUser)
}

// LoginUser handles POST /auth/login
func (h *AuthHandler) LoginUser(c *gin.Context) { // Perhatikan receiver 'h *AuthHandler'
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.UserService.LoginUser(req.Email, req.Password)
	if err != nil {
		if err.Error() == "kredensial tidak valid" { // Contoh penanganan error spesifik dari service
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process login"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token, "user_id": user.ID})
}
