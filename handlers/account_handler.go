package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"go-bank-app/config"
	"go-bank-app/models"

	"github.com/gin-gonic/gin"
)

// CreateAccount handles POST /accounts
func CreateAccount(c *gin.Context) {
	loggedInUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to create account for another user"})
		return
	}

	query := "INSERT INTO accounts (user_id, account_number, balance) VALUES (?, ?, ?)"
	result, err := config.DB.Exec(query, req.UserID, req.AccountNumber, 0.00)
	if err != nil {
		log.Printf("Error creating account: %v", err)
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "account_number") {
			c.JSON(http.StatusConflict, gin.H{"error": "Account number already exists."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account."})
		}
		return
	}

	accountID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Account created successfully",
		"account_id":  accountID,
		"user_id":     req.UserID,
		"balance":     0.00,
		"account_num": req.AccountNumber,
	})
}

// GetAccountByID handles GET /accounts/:id
func GetAccountByID(c *gin.Context) {
	accountID := c.Param("id")

	loggedInUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	var account models.Account
	query := "SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE id = ?"
	err := config.DB.QueryRow(query, accountID).
		Scan(&account.ID, &account.UserID, &account.AccountNumber, &account.Balance, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			log.Printf("Error getting account by ID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve account"})
		}
		return
	}

	if account.UserID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to access this account"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// Deposit handles POST /accounts/:id/deposit
func Deposit(c *gin.Context) {
	accountID := c.Param("id")
	loggedInUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	var ownerID int
	err := config.DB.QueryRow("SELECT user_id FROM accounts WHERE id = ?", accountID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			log.Printf("Error checking account ownership for deposit: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify account ownership"})
		}
		return
	}

	if ownerID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to deposit to this account"})
		return
	}

	var req models.DepositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Deposit amount must be greater than zero"})
		return
	}

	_, err = config.DB.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", req.Amount, accountID)
	if err != nil {
		log.Printf("Error updating balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deposit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful"})
}

// Withdraw handles POST /accounts/:id/withdraw
func Withdraw(c *gin.Context) {
	accountID := c.Param("id")
	loggedInUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	var ownerID int
	err := config.DB.QueryRow("SELECT user_id FROM accounts WHERE id = ?", accountID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			log.Printf("Error checking account ownership for withdrawal: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify account ownership"})
		}
		return
	}

	if ownerID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to withdraw from this account"})
		return
	}

	var req models.DepositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Withdrawal amount must be greater than zero"})
		return
	}

	var currentBalance float64
	err = config.DB.QueryRow("SELECT balance FROM accounts WHERE id = ?", accountID).Scan(&currentBalance)
	if err != nil {
		log.Printf("Error getting balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current balance"})
		return
	}

	if currentBalance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds"})
		return
	}

	_, err = config.DB.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", req.Amount, accountID)
	if err != nil {
		log.Printf("Error updating balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to withdraw"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful"})
}
