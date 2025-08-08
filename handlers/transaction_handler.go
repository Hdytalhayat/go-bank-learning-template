// go-bank-app/handlers/transaction_handler.go
package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"go-bank-app/config"
	"go-bank-app/models"

	"github.com/gin-gonic/gin"
)

// Transfer handles POST /transactions/transfer
// This function processes a fund transfer from one account to another.
func Transfer(c *gin.Context) {
	var req models.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Return bad request if JSON is invalid
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent transfer to the same account
	if req.FromAccountID == req.ToAccountID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot transfer to the same account"})
		return
	}

	// Start database transaction
	tx, err := config.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction for transfer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process transfer"})
		return
	}
	defer tx.Rollback()

	var fromBalance, toBalance float64
	var fromAccountIDInt, toAccountIDInt int

	// Lock and check source account
	err = tx.QueryRow("SELECT id, balance FROM accounts WHERE account_number = ? FOR UPDATE", req.FromAccountID).
		Scan(&fromAccountIDInt, &fromBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Source account not found"})
		} else {
			log.Printf("Error fetching source account for transfer: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve source account"})
		}
		return
	}

	// Lock and check destination account
	err = tx.QueryRow("SELECT id, balance FROM accounts WHERE account_number = ? FOR UPDATE", req.ToAccountID).
		Scan(&toAccountIDInt, &toBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Destination account not found"})
		} else {
			log.Printf("Error fetching destination account for transfer: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve destination account"})
		}
		return
	}

	// Ensure sufficient balance
	if fromBalance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds in source account"})
		return
	}

	// Deduct from source account
	_, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", req.Amount, fromAccountIDInt)
	if err != nil {
		log.Printf("Error updating source account balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update source account balance"})
		return
	}

	// Add to destination account
	_, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", req.Amount, toAccountIDInt)
	if err != nil {
		log.Printf("Error updating destination account balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update destination account balance"})
		return
	}

	// Log the transfer out transaction
	_, err = tx.Exec("INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)",
		fromAccountIDInt, "transfer_out", req.Amount, "Transfer to "+req.ToAccountID+": "+req.Description)
	if err != nil {
		log.Printf("Error logging transfer_out transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log outbound transaction"})
		return
	}

	// Log the transfer in transaction
	_, err = tx.Exec("INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)",
		toAccountIDInt, "transfer_in", req.Amount, "Transfer from "+req.FromAccountID+": "+req.Description)
	if err != nil {
		log.Printf("Error logging transfer_in transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log inbound transaction"})
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transfer transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transfer"})
		return
	}

	// Respond success
	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}

// GetAccountTransactions handles GET /accounts/:id/transactions
// This function retrieves all transactions for a specific account.
func GetAccountTransactions(c *gin.Context) {
	accountID := c.Param("id")

	var transactions []models.Transaction
	query := "SELECT id, account_id, transaction_type, amount, description, transaction_date FROM transactions WHERE account_id = ? ORDER BY transaction_date DESC"
	rows, err := config.DB.Query(query, accountID)
	if err != nil {
		log.Printf("Error getting account transactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}
	defer rows.Close()

	// Iterate through the result set
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(&t.ID, &t.AccountID, &t.TransactionType, &t.Amount, &t.Description, &t.TransactionDate)
		if err != nil {
			log.Printf("Error scanning transaction row: %v", err)
			continue
		}
		transactions = append(transactions, t)
	}

	// Check for iteration errors
	if err = rows.Err(); err != nil {
		log.Printf("Error during transactions rows iteration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving transactions during iteration"})
		return
	}

	// Return the list of transactions
	c.JSON(http.StatusOK, transactions)
}
