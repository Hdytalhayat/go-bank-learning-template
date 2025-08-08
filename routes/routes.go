// go-bank-app/routes/routes.go (Modifikasi)
package routes

import (
	"net/http"

	"go-bank-app/handlers"
	"go-bank-app/middleware"

	"github.com/gin-gonic/gin"
)

// InitHandlers and Services (akan diinisialisasi di main.go)
var (
	AuthHandler        *handlers.AuthHandler
	UserHandler        *handlers.UserHandler
	AccountHandler     *handlers.AccountHandler     // Belum dibuat, tapi placeholder
	TransactionHandler *handlers.TransactionHandler // Belum dibuat, tapi placeholder
)

// SetupRoutes mengatur semua rute API untuk aplikasi
func SetupRoutes(router *gin.Engine) {
	// Rute Publik
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	router.GET("/hello/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gin.H{"message": "Hello, " + name + "!"})
	})

	// Rute Autentikasi
	router.POST("/auth/register", AuthHandler.RegisterUser)
	router.POST("/auth/login", AuthHandler.LoginUser)

	// Rute yang Dilindungi (memerlukan autentikasi JWT)
	authenticated := router.Group("/")
	authenticated.Use(middleware.AuthMiddleware())
	{
		authenticated.GET("/users/:id", UserHandler.GetUserByID)
		authenticated.GET("/users", UserHandler.GetAllUsers) // Akan memerlukan otorisasi peran admin

		// Account
		authenticated.POST("/accounts", AccountHandler.CreateAccount)
		authenticated.GET("/accounts/:id", AccountHandler.GetAccountByID)
		authenticated.POST("/accounts/:id/deposit", AccountHandler.Deposit)
		authenticated.POST("/accounts/:id/withdraw", AccountHandler.Withdraw)

		// Transaction
		authenticated.POST("/transactions/transfer", TransactionHandler.Transfer)
		authenticated.GET("/accounts/:id/transactions", TransactionHandler.GetAccountTransactions)
	}
}
