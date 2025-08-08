// go-bank-app/main.go
package main

import (
	"go-bank-app/config" // Import package config kita
	"go-bank-app/routes" // Import package routes kita

	"github.com/gin-gonic/gin"
)

func main() {
	// Inisialisasi koneksi database
	config.InitDB()
	// Pastikan koneksi ditutup saat aplikasi berhenti
	defer func() {
		if config.DB != nil {
			config.DB.Close()
		}
	}()

	// Inisialisasi Gin router
	router := gin.Default()

	// Setup semua rute
	routes.SetupRoutes(router)

	// Jalankan server di port 8080
	router.Run(":8080")
}
