package services

import (
	"fmt"
	"go-bank-app/config" // For accessing config.DB.Begin()
	"go-bank-app/models"
	"go-bank-app/repositories"
)

// TransactionService defines the interface for transaction-related business logic.
type TransactionService interface {
	Transfer(req *models.TransferRequest) error
	GetAccountTransactions(accountID int) ([]models.Transaction, error)
}

// transactionServiceImpl is the concrete implementation of TransactionService.
type transactionServiceImpl struct {
	accountRepo     repositories.AccountRepository
	transactionRepo repositories.TransactionRepository
}

// NewTransactionService creates a new instance of TransactionService.
func NewTransactionService(accountRepo repositories.AccountRepository, transactionRepo repositories.TransactionRepository) TransactionService {
	return &transactionServiceImpl{accountRepo: accountRepo, transactionRepo: transactionRepo}
}

func (s *transactionServiceImpl) Transfer(req *models.TransferRequest) error {
	tx, err := config.DB.Begin() // Start transaction at the service layer
	if err != nil {
		return fmt.Errorf("failed to begin transfer transaction: %w", err)
	}
	defer tx.Rollback() // Ensure rollback in case of error

	// Get sender and receiver accounts (use tx to ensure row-level locking if needed)
	fromAccount, err := s.accountRepo.GetAccountByNumber(req.FromAccountID)
	if err != nil {
		return fmt.Errorf("sender account not found: %w", err)
	}
	toAccount, err := s.accountRepo.GetAccountByNumber(req.ToAccountID)
	if err != nil {
		return fmt.Errorf("receiver account not found: %w", err)
	}

	// Check sufficient balance
	if fromAccount.Balance < req.Amount {
		return fmt.Errorf("insufficient balance in sender's account")
	}

	// Debit sender's account
	err = s.accountRepo.UpdateAccountBalance(tx, fromAccount.ID, -req.Amount)
	if err != nil {
		return fmt.Errorf("failed to update sender's account balance: %w", err)
	}

	// Credit receiver's account
	err = s.accountRepo.UpdateAccountBalance(tx, toAccount.ID, req.Amount)
	if err != nil {
		return fmt.Errorf("failed to update receiver's account balance: %w", err)
	}

	// Record outbound transaction for sender
	outboundTransaction := &models.Transaction{
		AccountID:       fromAccount.ID,
		TransactionType: "transfer_out",
		Amount:          req.Amount,
		Description:     fmt.Sprintf("Transfer to %s: %s", toAccount.AccountNumber, req.Description),
	}
	_, err = s.transactionRepo.CreateTransaction(tx, outboundTransaction)
	if err != nil {
		return fmt.Errorf("failed to record outbound transaction: %w", err)
	}

	// Record inbound transaction for receiver
	inboundTransaction := &models.Transaction{
		AccountID:       toAccount.ID,
		TransactionType: "transfer_in",
		Amount:          req.Amount,
		Description:     fmt.Sprintf("Transfer from %s: %s", fromAccount.AccountNumber, req.Description),
	}
	_, err = s.transactionRepo.CreateTransaction(tx, inboundTransaction)
	if err != nil {
		return fmt.Errorf("failed to record inbound transaction: %w", err)
	}

	err = tx.Commit() // Commit transaction
	if err != nil {
		return fmt.Errorf("failed to commit transfer transaction: %w", err)
	}

	return nil
}

func (s *transactionServiceImpl) GetAccountTransactions(accountID int) ([]models.Transaction, error) {
	transactions, err := s.transactionRepo.GetTransactionsByAccountID(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account transactions: %w", err)
	}
	return transactions, nil
}
