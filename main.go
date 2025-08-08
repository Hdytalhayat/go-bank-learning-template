package main

import (
	"log"
	"os"

	"go-bank-app/config"
	"go-bank-app/handlers"
	"go-bank-app/repositories"
	"go-bank-app/routes"
	"go-bank-app/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration (if using Viper as suggested in Phase 5 Step 3)
	// config.LoadConfig()

	// Ensure JWT_SECRET_KEY is set. This is important for production.
	if os.Getenv("JWT_SECRET_KEY") == "" {
		log.Println("WARNING: JWT_SECRET_KEY environment variable not set. Using default key from config/config.go. DO NOT USE IN PRODUCTION!")
	}
	// Or if using Viper:
	// if config.AppCfg.JWTSecret == "" {
	// 	log.Println("WARNING: JWT_SECRET_KEY (or jwt_secret in config) not set. Using hardcoded fallback. DO NOT USE IN PRODUCTION!")
	// }

	// Initialize database connection
	config.InitDB()
	defer func() {
		if config.DB != nil {
			log.Println("Closing database connection.")
			config.DB.Close()
		}
	}()

	// Initialize Repositories
	userRepo := repositories.NewUserRepository(config.DB)
	accountRepo := repositories.NewAccountRepository(config.DB)
	transactionRepo := repositories.NewTransactionRepository(config.DB)

	// Initialize Services
	userService := services.NewUserService(userRepo)
	accountService := services.NewAccountService(accountRepo, transactionRepo)
	transactionService := services.NewTransactionService(accountRepo, transactionRepo) // transactionService also requires accountRepo for transfer logic

	// Initialize Handlers
	routes.AuthHandler = handlers.NewAuthHandler(userService)
	routes.UserHandler = handlers.NewUserHandler(userService)
	routes.AccountHandler = handlers.NewAccountHandler(accountService)
	routes.TransactionHandler = handlers.NewTransactionHandler(transactionService, accountService) // TransactionHandler also needs AccountService for transfer authorization

	// Initialize Gin router
	router := gin.Default()

	// Setup all routes
	routes.SetupRoutes(router)

	// Run the server on port 8080
	log.Printf("Server running on port %s", "8080") // Replace with config.AppCfg.ServerPort if using Viper
	router.Run(":8080")                             // Replace with ":" + config.AppCfg.ServerPort if using Viper
}
