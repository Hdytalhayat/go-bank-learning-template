// go-bank-app/handlers/transaction_handler.go
package handlers

import (
	"log"
	"net/http"
	"strconv" // Pastikan ada
	"strings" // Pastikan ada

	"go-bank-app/models"
	"go-bank-app/services" // Import package services kita

	"github.com/gin-gonic/gin"
)

// TransactionHandler struct untuk dependensi service
type TransactionHandler struct {
	TransactionService services.TransactionService
	AccountService     services.AccountService // Untuk otorisasi
}

// NewTransactionHandler membuat instance baru dari TransactionHandler
func NewTransactionHandler(transactionService services.TransactionService, accountService services.AccountService) *TransactionHandler {
	return &TransactionHandler{TransactionService: transactionService, AccountService: accountService}
}

// Transfer handles POST /transactions/transfer
func (h *TransactionHandler) Transfer(c *gin.Context) {
	loggedInUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	var req models.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.FromAccountID == req.ToAccountID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot transfer to the same account"})
		return
	}

	// Otorisasi: Pastikan user yang login adalah pemilik akun pengirim
	fromAccount, err := h.AccountService.GetAccountByNumber(req.FromAccountID)
	if err != nil {
		if strings.Contains(err.Error(), "akun tidak ditemukan") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Source account not found"})
		} else {
			log.Printf("Error checking source account ownership for transfer: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify source account"})
		}
		return
	}
	if fromAccount.UserID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to initiate transfer from this account"})
		return
	}

	err = h.TransactionService.Transfer(&req)
	if err != nil {
		log.Printf("Error during transfer via service: %v", err)
		if strings.Contains(err.Error(), "akun tidak ditemukan") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "saldo tidak cukup") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds in source account"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process transfer"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}

// GetAccountTransactions handles GET /accounts/:id/transactions
func (h *TransactionHandler) GetAccountTransactions(c *gin.Context) {
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

	// Otorisasi: Pastikan user yang login adalah pemilik akun ini
	account, err := h.AccountService.GetAccountByID(accountID)
	if err != nil {
		if strings.Contains(err.Error(), "akun tidak ditemukan") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			log.Printf("Error checking account ownership for transactions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify account ownership"})
		}
		return
	}
	if account.UserID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view transactions for this account"})
		return
	}

	transactions, err := h.TransactionService.GetAccountTransactions(accountID)
	if err != nil {
		log.Printf("Error getting account transactions via service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
