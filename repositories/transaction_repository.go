package repositories

import (
	"database/sql"
	"fmt"
	"go-bank-app/models"
)

// TransactionRepository defines the interface for transaction operations in the database.
type TransactionRepository interface {
	CreateTransaction(tx *sql.Tx, transaction *models.Transaction) (int64, error) // Accepts *sql.Tx
	GetTransactionsByAccountID(accountID int) ([]models.Transaction, error)
}

// transactionRepositoryImpl is the concrete implementation of TransactionRepository.
type transactionRepositoryImpl struct {
	db *sql.DB
}

// NewTransactionRepository creates a new instance of TransactionRepository.
func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepositoryImpl{db: db}
}

// CreateTransaction inserts a new transaction into the database within the given transaction context.
func (r *transactionRepositoryImpl) CreateTransaction(tx *sql.Tx, transaction *models.Transaction) (int64, error) {
	query := "INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)"
	result, err := tx.Exec(query, transaction.AccountID, transaction.TransactionType, transaction.Amount, transaction.Description)
	if err != nil {
		return 0, fmt.Errorf("failed to create transaction in the database: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve new transaction ID: %w", err)
	}
	return id, nil
}

// GetTransactionsByAccountID retrieves all transactions for a specific account, ordered by transaction date (descending).
func (r *transactionRepositoryImpl) GetTransactionsByAccountID(accountID int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := "SELECT id, account_id, transaction_type, amount, description, transaction_date FROM transactions WHERE account_id = ? ORDER BY transaction_date DESC"
	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(&t.ID, &t.AccountID, &t.TransactionType, &t.Amount, &t.Description, &t.TransactionDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}
		transactions = append(transactions, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating transaction rows: %w", err)
	}
	return transactions, nil
}
