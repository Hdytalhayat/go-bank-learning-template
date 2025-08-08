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
func Transfer(c *gin.Context) {
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

	tx, err := config.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction for transfer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process transfer"})
		return
	}
	defer tx.Rollback()

	var fromBalance, toBalance float64
	var fromAccountIDInt, toAccountIDInt int
	var fromAccountOwnerID int

	err = tx.QueryRow("SELECT id, user_id, balance FROM accounts WHERE account_number = ? FOR UPDATE", req.FromAccountID).Scan(&fromAccountIDInt, &fromAccountOwnerID, &fromBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Source account not found"})
		} else {
			log.Printf("Error fetching source account for transfer: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve source account"})
		}
		return
	}

	if fromAccountOwnerID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to initiate transfer from this account"})
		return
	}

	err = tx.QueryRow("SELECT id, balance FROM accounts WHERE account_number = ? FOR UPDATE", req.ToAccountID).Scan(&toAccountIDInt, &toBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Destination account not found"})
		} else {
			log.Printf("Error fetching destination account for transfer: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve destination account"})
		}
		return
	}

	if fromBalance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds in source account"})
		return
	}

	_, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", req.Amount, fromAccountIDInt)
	if err != nil {
		log.Printf("Error updating source account balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update source account balance"})
		return
	}

	_, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", req.Amount, toAccountIDInt)
	if err != nil {
		log.Printf("Error updating destination account balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update destination account balance"})
		return
	}

	_, err = tx.Exec(`INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)`,
		fromAccountIDInt, "transfer_out", req.Amount, "Transfer to "+req.ToAccountID)
	if err != nil {
		log.Printf("Error inserting transfer_out transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transfer_out transaction"})
		return
	}

	_, err = tx.Exec(`INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)`,
		toAccountIDInt, "transfer_in", req.Amount, "Transfer from "+req.FromAccountID)
	if err != nil {
		log.Printf("Error inserting transfer_in transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transfer_in transaction"})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transfer transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transfer transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}

// GetAccountTransactions handles GET /accounts/:id/transactions
func GetAccountTransactions(c *gin.Context) {
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
			log.Printf("Error checking account ownership for transactions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify account ownership"})
		}
		return
	}

	if ownerID != loggedInUserID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view transactions for this account"})
		return
	}

	var transactions []models.Transaction
	query := "SELECT id, account_id, transaction_type, amount, description, transaction_date FROM transactions WHERE account_id = ? ORDER BY transaction_date DESC"
	rows, err := config.DB.Query(query, accountID)
	if err != nil {
		log.Printf("Error getting account transactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(&t.ID, &t.AccountID, &t.TransactionType, &t.Amount, &t.Description, &t.TransactionDate)
		if err != nil {
			log.Printf("Error scanning transaction row: %v", err)
			continue
		}
		transactions = append(transactions, t)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during transactions rows iteration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving transactions during iteration"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
