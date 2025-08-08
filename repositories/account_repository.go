package repositories

import (
	"database/sql"
	"fmt"
	"go-bank-app/models"
)

// AccountRepository defines the interface for account operations in the database.
type AccountRepository interface {
	CreateAccount(account *models.Account) (int64, error)
	GetAccountByID(id int) (*models.Account, error)
	GetAccountByNumber(accountNumber string) (*models.Account, error)
	UpdateAccountBalance(tx *sql.Tx, accountID int, amount float64) error // Accepts *sql.Tx
}

// accountRepositoryImpl is the concrete implementation of AccountRepository.
type accountRepositoryImpl struct {
	db *sql.DB
}

// NewAccountRepository creates a new instance of AccountRepository.
func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepositoryImpl{db: db}
}

// CreateAccount inserts a new account into the database.
func (r *accountRepositoryImpl) CreateAccount(account *models.Account) (int64, error) {
	query := "INSERT INTO accounts (user_id, account_number, balance) VALUES (?, ?, ?)"
	result, err := r.db.Exec(query, account.UserID, account.AccountNumber, account.Balance)
	if err != nil {
		return 0, fmt.Errorf("failed to create account in database: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve new account ID: %w", err)
	}
	return id, nil
}

// GetAccountByID retrieves an account from the database using its ID.
func (r *accountRepositoryImpl) GetAccountByID(id int) (*models.Account, error) {
	var account models.Account
	query := "SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE id = ?"
	err := r.db.QueryRow(query, id).
		Scan(&account.ID, &account.UserID, &account.AccountNumber, &account.Balance, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to retrieve account by ID: %w", err)
	}
	return &account, nil
}

// GetAccountByNumber retrieves an account from the database using its account number.
func (r *accountRepositoryImpl) GetAccountByNumber(accountNumber string) (*models.Account, error) {
	var account models.Account
	query := "SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE account_number = ?"
	err := r.db.QueryRow(query, accountNumber).
		Scan(&account.ID, &account.UserID, &account.AccountNumber, &account.Balance, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to retrieve account by number: %w", err)
	}
	return &account, nil
}

// UpdateAccountBalance updates the balance of an account within the given transaction.
func (r *accountRepositoryImpl) UpdateAccountBalance(tx *sql.Tx, accountID int, amount float64) error {
	_, err := tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, accountID)
	if err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}
	return nil
}
