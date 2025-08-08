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
// This function creates a new bank account for a given user ID.
func CreateAccount(c *gin.Context) {
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Handle invalid JSON input
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the provided user ID exists
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

	// Insert the new account into the database with a starting balance of 0
	query := "INSERT INTO accounts (user_id, account_number, balance) VALUES (?, ?, ?)"
	result, err := config.DB.Exec(query, req.UserID, req.AccountNumber, 0.00)
	if err != nil {
		log.Printf("Error creating account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account. Account number might already exist."})
		return
	}

	// Retrieve the ID of the newly inserted account
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID for account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve account ID"})
		return
	}

	// Fetch the newly created account to return it in the response
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
// This function retrieves an account by its ID.
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
// This function adds funds to a specific account.
func Deposit(c *gin.Context) {
	accountID := c.Param("id")
	var req models.DepositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start a new transaction
	tx, err := config.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction for deposit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process deposit"})
		return
	}
	defer tx.Rollback()

	// Update the account balance
	_, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", req.Amount, accountID)
	if err != nil {
		log.Printf("Error updating account balance for deposit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
		return
	}

	// Insert a transaction log
	_, err = tx.Exec("INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)",
		accountID, "deposit", req.Amount, "Deposit funds")
	if err != nil {
		log.Printf("Error logging deposit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing deposit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit deposit"})
		return
	}

	// Return updated account info
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
// This function deducts funds from a specific account, ensuring sufficient balance.
func Withdraw(c *gin.Context) {
	accountID := c.Param("id")
	var req models.DepositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start a new transaction
	tx, err := config.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction for withdraw: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process withdrawal"})
		return
	}
	defer tx.Rollback()

	// Lock the account row for update and check balance
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

	// Check for sufficient funds
	if currentBalance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds"})
		return
	}

	// Deduct the amount
	_, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", req.Amount, accountID)
	if err != nil {
		log.Printf("Error updating account balance for withdraw: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
		return
	}

	// Log the transaction
	_, err = tx.Exec("INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)",
		accountID, "withdraw", req.Amount, "Withdrawal funds")
	if err != nil {
		log.Printf("Error logging withdraw transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing withdraw transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit withdrawal"})
		return
	}

	// Return updated account info
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
