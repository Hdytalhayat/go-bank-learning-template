// go-bank-app/handlers/account_handler.go
package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"go-bank-app/models"
	"go-bank-app/services"

	"github.com/gin-gonic/gin"
)

// AccountHandler is a struct that contains the AccountService dependency
type AccountHandler struct {
	AccountService services.AccountService
}

// NewAccountHandler returns a new instance of AccountHandler
func NewAccountHandler(accountService services.AccountService) *AccountHandler {
	return &AccountHandler{AccountService: accountService}
}

// CreateAccount handles POST /accounts
// It creates a new account for the currently authenticated user
func (h *AccountHandler) CreateAccount(c *gin.Context) {
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

	// Authorization: Only allow users to create accounts for themselves
	if req.UserID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to create account for another user"})
		return
	}

	newAccount, err := h.AccountService.CreateAccount(&req)
	if err != nil {
		log.Printf("Error creating account via service: %v", err)
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "account_number") {
			c.JSON(http.StatusConflict, gin.H{"error": "Account number already exists."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account."})
		}
		return
	}

	c.JSON(http.StatusCreated, newAccount)
}

// GetAccountByID handles GET /accounts/:id
// It retrieves account details by account ID, only if the logged-in user owns the account
func (h *AccountHandler) GetAccountByID(c *gin.Context) {
	idParam := c.Param("id")
	accountID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID format"})
		return
	}

	loggedInUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	account, err := h.AccountService.GetAccountByID(accountID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "akun tidak ditemukan") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			log.Printf("Error getting account by ID via service: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve account"})
		}
		return
	}

	// Authorization: Only allow users to access their own accounts
	if account.UserID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to access this account"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// Deposit handles POST /accounts/:id/deposit
// It deposits an amount to the account, only if the user is the account owner
func (h *AccountHandler) Deposit(c *gin.Context) {
	idParam := c.Param("id")
	accountID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID format"})
		return
	}

	loggedInUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// Check account ownership before processing deposit
	account, err := h.AccountService.GetAccountByID(accountID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "akun tidak ditemukan") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			log.Printf("Error checking account ownership for deposit: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify account ownership"})
		}
		return
	}
	if account.UserID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to deposit to this account"})
		return
	}

	var req models.DepositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedAccount, err := h.AccountService.Deposit(accountID, req.Amount)
	if err != nil {
		log.Printf("Error during deposit via service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process deposit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful", "account": updatedAccount})
}

// Withdraw handles POST /accounts/:id/withdraw
// It withdraws an amount from the account, only if the user is the account owner and has sufficient funds
func (h *AccountHandler) Withdraw(c *gin.Context) {
	idParam := c.Param("id")
	accountID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID format"})
		return
	}

	loggedInUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	// Check account ownership before processing withdrawal
	account, err := h.AccountService.GetAccountByID(accountID)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "akun tidak ditemukan") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			log.Printf("Error checking account ownership for withdrawal: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify account ownership"})
		}
		return
	}
	if account.UserID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to withdraw from this account"})
		return
	}

	var req models.DepositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedAccount, err := h.AccountService.Withdraw(accountID, req.Amount)
	if err != nil {
		log.Printf("Error during withdrawal via service: %v", err)
		if strings.Contains(err.Error(), "saldo tidak cukup") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process withdrawal"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful", "account": updatedAccount})
}
