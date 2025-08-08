// go-bank-app/handlers/account_handler.go
package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"go-bank-app/config"
	"go-bank-app/models"

	"github.com/gin-gonic/gin"
)

// CreateAccount handles POST /accounts
func CreateAccount(c *gin.Context) {
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek apakah user_id valid
	var userExists bool
	err := config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", req.UserID).Scan(&userExists)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
		return
	}
	if !userExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with specified ID does not exist"})
		return
	}

	query := "INSERT INTO accounts (user_id, account_number, balance) VALUES (?, ?, ?)"
	result, err := config.DB.Exec(query, req.UserID, req.AccountNumber, 0.00) // Saldo awal 0
	if err != nil {
		log.Printf("Error creating account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account. Account number might already exist."})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID for account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve account ID"})
		return
	}

	var newAccount models.Account
	err = config.DB.QueryRow("SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE id = ?", id).
		Scan(&newAccount.ID, &newAccount.UserID, &newAccount.AccountNumber, &newAccount.Balance, &newAccount.CreatedAt, &newAccount.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching new account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve newly created account"})
		return
	}

	c.JSON(http.StatusCreated, newAccount)
}

// GetAccountByID handles GET /accounts/:id
func GetAccountByID(c *gin.Context) {
	id := c.Param("id")

	var account models.Account
	query := "SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE id = ?"
	err := config.DB.QueryRow(query, id).
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

	c.JSON(http.StatusOK, account)
}

// Deposit handles POST /accounts/:id/deposit
func Deposit(c *gin.Context) {
	accountID := c.Param("id")
	var req models.DepositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction for deposit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process deposit"})
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", req.Amount, accountID)
	if err != nil {
		log.Printf("Error updating account balance for deposit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
		return
	}

	_, err = tx.Exec("INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)",
		accountID, "deposit", req.Amount, "Deposit funds")
	if err != nil {
		log.Printf("Error logging deposit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing deposit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit deposit"})
		return
	}

	var updatedAccount models.Account
	err = config.DB.QueryRow("SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE id = ?", accountID).
		Scan(&updatedAccount.ID, &updatedAccount.UserID, &updatedAccount.AccountNumber, &updatedAccount.Balance, &updatedAccount.CreatedAt, &updatedAccount.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching updated account after deposit: %v", err)
		c.JSON(http.StatusOK, gin.H{"message": "Deposit successful, but failed to retrieve updated account details."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful", "account": updatedAccount})
}

// Withdraw handles POST /accounts/:id/withdraw
func Withdraw(c *gin.Context) {
	accountID := c.Param("id")
	var req models.DepositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction for withdraw: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process withdrawal"})
		return
	}
	defer tx.Rollback()

	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM accounts WHERE id = ? FOR UPDATE", accountID).Scan(&currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			log.Printf("Error getting account balance for withdraw: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve account balance"})
		}
		return
	}

	if currentBalance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds"})
		return
	}

	_, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", req.Amount, accountID)
	if err != nil {
		log.Printf("Error updating account balance for withdraw: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
		return
	}

	_, err = tx.Exec("INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)",
		accountID, "withdraw", req.Amount, "Withdrawal funds")
	if err != nil {
		log.Printf("Error logging withdraw transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing withdraw transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit withdrawal"})
		return
	}

	var updatedAccount models.Account
	err = config.DB.QueryRow("SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE id = ?", accountID).
		Scan(&updatedAccount.ID, &updatedAccount.UserID, &updatedAccount.AccountNumber, &updatedAccount.Balance, &updatedAccount.CreatedAt, &updatedAccount.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching updated account after withdraw: %v", err)
		c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful, but failed to retrieve updated account details."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful", "account": updatedAccount})
}
