// go-bank-app/routes/routes.go
package routes

import (
	"go-bank-app/handlers" // Import package handlers kita
	"go-bank-app/middleware"
	"net/http" // Digunakan untuk konstanta HTTP status codes

	"github.com/gin-gonic/gin"
)

// SetupRoutes mengatur semua rute API untuk aplikasi
func SetupRoutes(router *gin.Engine) {
	// Rute Publik (tidak memerlukan autentikasi)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	router.GET("/hello/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gin.H{"message": "Hello, " + name + "!"})
	})

	// Rute Autentikasi
	router.POST("/auth/register", handlers.RegisterUser)
	router.POST("/auth/login", handlers.LoginUser)

	// Rute yang Dilindungi (memerlukan autentikasi JWT)
	// Buat group rute yang menggunakan middleware autentikasi
	authenticated := router.Group("/")
	authenticated.Use(middleware.AuthMiddleware())
	{
		// Rute untuk User API (dari Fase 2, sekarang dilindungi)
		// Catatan: GET /users/:id dan GET /users mungkin perlu otorisasi lebih lanjut (misal: hanya admin yang bisa lihat semua user)
		// Untuk saat ini, kita anggap hanya user terautentikasi bisa mengakses datanya sendiri
		authenticated.GET("/users/:id", handlers.GetUserByID)
		authenticated.GET("/users", handlers.GetAllUsers) // Biasanya GET /users untuk admin, atau pakai filter user_id

		// Rute untuk Account & Transaction API (Fase 3, sekarang dilindungi)
		authenticated.POST("/accounts", handlers.CreateAccount)
		authenticated.GET("/accounts/:id", handlers.GetAccountByID)
		authenticated.POST("/accounts/:id/deposit", handlers.Deposit)
		authenticated.POST("/accounts/:id/withdraw", handlers.Withdraw)
		authenticated.POST("/transactions/transfer", handlers.Transfer)
		authenticated.GET("/accounts/:id/transactions", handlers.GetAccountTransactions)
	}
}
