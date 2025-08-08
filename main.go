// go-bank-app/main.go
package main

import (
	"go-bank-app/config"
	"go-bank-app/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// IMPORTANT: Set the environment variable JWT_SECRET_KEY before running the application
	// Example for terminal (PowerShell on Windows): $env:JWT_SECRET_KEY="supersecretjwtreallystrongkey"
	// Example for terminal (Bash/Zsh on Linux/macOS): export JWT_SECRET_KEY="supersecretjwtreallystrongkey"
	// Or you can set it directly in the code for development (NOT RECOMMENDED FOR PRODUCTION!)
	if os.Getenv("JWT_SECRET_KEY") == "" {
		log.Println("WARNING: JWT_SECRET_KEY environment variable not set. Using default key from config. DO NOT USE IN PRODUCTION!")
		// If you don't want to use os.Getenv() at all during development,
		// you can remove this block and rely on a default value from config/config.go.
		// However, it's best to get used to using environment variables.
	}

	// Initialize database connection
	config.InitDB()
	defer func() {
		if config.DB != nil {
			config.DB.Close() // Ensure the DB connection is closed when the app stops
		}
	}()

	// Initialize Gin HTTP router
	router := gin.Default()

	// Register all API routes
	routes.SetupRoutes(router)

	// Start the HTTP server on port 8080
	router.Run(":8080")
}
