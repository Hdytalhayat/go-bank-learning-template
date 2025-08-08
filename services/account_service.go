package services

import (
	"database/sql"
	"fmt"
	"go-bank-app/config" // For accessing config.DB.Begin()
	"go-bank-app/models"
	"go-bank-app/repositories"
)

// AccountService defines the interface for account-related business logic.
type AccountService interface {
	CreateAccount(req *models.CreateAccountRequest) (*models.Account, error)
	GetAccountByID(id int) (*models.Account, error)
	GetAccountByNumber(accountNumber string) (*models.Account, error)
	Deposit(accountID int, amount float64) (*models.Account, error)
	Withdraw(accountID int, amount float64) (*models.Account, error)
}

// accountServiceImpl is the concrete implementation of AccountService.
type accountServiceImpl struct {
	accountRepo     repositories.AccountRepository
	transactionRepo repositories.TransactionRepository // Needed for Deposit/Withdraw
}

// NewAccountService creates a new instance of AccountService.
func NewAccountService(accountRepo repositories.AccountRepository, transactionRepo repositories.TransactionRepository) AccountService {
	return &accountServiceImpl{accountRepo: accountRepo, transactionRepo: transactionRepo}
}

func (s *accountServiceImpl) CreateAccount(req *models.CreateAccountRequest) (*models.Account, error) {
	// You might also want to check if the user_id exists by calling userRepo,
	// or just rely on the foreign key constraint in the DB.
	// For now, we assume the handler validates the logged-in user ID.

	account := &models.Account{
		UserID:        req.UserID,
		AccountNumber: req.AccountNumber,
		Balance:       0.00, // Initial balance
	}

	id, err := s.accountRepo.CreateAccount(account)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	account.ID = int(id)

	// Retrieve the newly created account for a complete response
	newAccount, err := s.accountRepo.GetAccountByID(int(id))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch newly created account: %w", err)
	}
	return newAccount, nil
}

func (s *accountServiceImpl) GetAccountByID(id int) (*models.Account, error) {
	account, err := s.accountRepo.GetAccountByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account: %w", err)
	}
	return account, nil
}
func (s *accountServiceImpl) GetAccountByNumber(accountNumber string) (*models.Account, error) {
	account, err := s.accountRepo.GetAccountByNumber(accountNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("akun tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil akun berdasarkan nomor: %w", err)
	}
	return account, nil
}
func (s *accountServiceImpl) Deposit(accountID int, amount float64) (*models.Account, error) {
	tx, err := config.DB.Begin() // Begin transaction at the service layer
	if err != nil {
		return nil, fmt.Errorf("failed to begin deposit transaction: %w", err)
	}
	defer tx.Rollback() // Ensure rollback on error

	// Update account balance
	err = s.accountRepo.UpdateAccountBalance(tx, accountID, amount) // Positive amount for deposit
	if err != nil {
		return nil, fmt.Errorf("failed to update balance during deposit: %w", err)
	}

	// Record the transaction
	transaction := &models.Transaction{
		AccountID:       accountID,
		TransactionType: "deposit",
		Amount:          amount,
		Description:     "Deposit funds",
	}
	_, err = s.transactionRepo.CreateTransaction(tx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to record deposit transaction: %w", err)
	}

	err = tx.Commit() // Commit transaction
	if err != nil {
		return nil, fmt.Errorf("failed to commit deposit transaction: %w", err)
	}

	// Fetch and return the updated account
	updatedAccount, err := s.accountRepo.GetAccountByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("deposit succeeded, but failed to retrieve updated account: %w", err)
	}
	return updatedAccount, nil
}

func (s *accountServiceImpl) Withdraw(accountID int, amount float64) (*models.Account, error) {
	tx, err := config.DB.Begin() // Begin transaction at the service layer
	if err != nil {
		return nil, fmt.Errorf("failed to begin withdrawal transaction: %w", err)
	}
	defer tx.Rollback() // Ensure rollback on error

	// Retrieve account to check balance
	account, err := s.accountRepo.GetAccountByID(accountID) // Locks row if using FOR UPDATE
	if err != nil {
		return nil, fmt.Errorf("account not found or failed to fetch balance: %w", err)
	}

	if account.Balance < amount {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Update account balance (negative amount for withdrawal)
	err = s.accountRepo.UpdateAccountBalance(tx, accountID, -amount)
	if err != nil {
		return nil, fmt.Errorf("failed to update balance during withdrawal: %w", err)
	}

	// Record the transaction
	transaction := &models.Transaction{
		AccountID:       accountID,
		TransactionType: "withdraw",
		Amount:          amount,
		Description:     "Withdrawal funds",
	}
	_, err = s.transactionRepo.CreateTransaction(tx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to record withdrawal transaction: %w", err)
	}

	err = tx.Commit() // Commit transaction
	if err != nil {
		return nil, fmt.Errorf("failed to commit withdrawal transaction: %w", err)
	}

	// Fetch and return the updated account
	updatedAccount, err := s.accountRepo.GetAccountByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("withdrawal succeeded, but failed to retrieve updated account: %w", err)
	}
	return updatedAccount, nil
}
