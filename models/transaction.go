// go-bank-app/models/transaction.go
package models

import "time"

type Transaction struct {
	ID              int       `json:"id"`
	AccountID       int       `json:"account_id"`
	TransactionType string    `json:"transaction_type"` // deposit, withdraw, transfer_out, transfer_in
	Amount          float64   `json:"amount"`
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
}

type TransferRequest struct {
	FromAccountID string  `json:"from_account_id" binding:"required"`
	ToAccountID   string  `json:"to_account_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Description   string  `json:"description"`
}
