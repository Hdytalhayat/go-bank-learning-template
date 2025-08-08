// go-bank-app/models/account.go
package models

import "time"

type Account struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"` // Gunakan float64 untuk kenyamanan, tapi ingat DECIMAL di DB
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateAccountRequest struct {
	UserID        int    `json:"user_id" binding:"required"`
	AccountNumber string `json:"account_number" binding:"required,min=10,max=20"` // Contoh validasi panjang akun
}

type DepositWithdrawRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"` // Jumlah harus positif (greater than 0)
}
