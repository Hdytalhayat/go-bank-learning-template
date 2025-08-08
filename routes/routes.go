// go-bank-app/routes/routes.go
package routes

import (
	"go-bank-app/handlers" // Import package handlers kita
	"net/http"             // Digunakan untuk konstanta HTTP status codes

	"github.com/gin-gonic/gin"
)

// SetupRoutes mengatur semua rute API untuk aplikasi
func SetupRoutes(router *gin.Engine) {
	// Rute dari Fase 1
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	router.GET("/hello/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gin.H{"message": "Hello, " + name + "!"})
	})

	// Rute untuk User API (dari Fase 2)
	router.POST("/users", handlers.CreateUser)
	router.GET("/users/:id", handlers.GetUserByID)
	router.GET("/users", handlers.GetAllUsers)

	// Rute untuk Account & Transaction API (Fase 3)
	router.POST("/accounts", handlers.CreateAccount)
	router.GET("/accounts/:id", handlers.GetAccountByID)
	router.POST("/accounts/:id/deposit", handlers.Deposit)
	router.POST("/accounts/:id/withdraw", handlers.Withdraw)
	router.POST("/transactions/transfer", handlers.Transfer)
	router.GET("/accounts/:id/transactions", handlers.GetAccountTransactions)
}
