package main

import (
	"database/sql" // Package standar untuk berinteraksi dengan database
	"fmt"
	"log"      // Untuk logging error
	"net/http" // Untuk konstanta HTTP status codes
	"time"     // Untuk tipe data waktu

	"github.com/gin-gonic/gin"         // Gin framework
	_ "github.com/go-sql-driver/mysql" // Driver MySQL, pakai underscore karena kita hanya butuh side effect-nya (registrasi driver)
)

// Di atas fungsi main, atau di file models/user.go
type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name" binding:"required"`        // `binding:"required"` untuk validasi input Gin
	Email        string    `json:"email" binding:"required,email"` // `email` juga untuk validasi format email
	PasswordHash string    `json:"-"`                              // "-" agar tidak disertakan dalam JSON response
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Struct untuk request membuat user baru (tanpa ID dan timestamp)
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"` // Password langsung, akan di-hash nanti
}

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

// Global variable untuk koneksi database
var db *sql.DB

const dsn = "root@tcp(127.0.0.1:3306)/bank_app_db?parseTime=true"

func initDB() {
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Kesalahan saat membuka koneksi database: %v", err)
	}

	// Ping database untuk memastikan koneksi berhasil
	err = db.Ping()
	if err != nil {
		log.Fatalf("Kesalahan saat ping database: %v", err)
	}

	fmt.Println("Koneksi ke database MySQL berhasil!")

	// Set pengaturan koneksi (opsional, untuk performa)
	db.SetMaxOpenConns(10)                 // Maksimal jumlah koneksi yang terbuka ke database
	db.SetMaxIdleConns(5)                  // Maksimal jumlah koneksi idle dalam pool
	db.SetConnMaxLifetime(5 * time.Minute) // Waktu maksimal koneksi dapat digunakan kembali
}

func main() {
	initDB()
	defer db.Close()

	router := gin.Default()

	// Rute dari Fase 1
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	router.GET("/hello/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gin.H{"message": "Hello, " + name + "!"})
	})

	// Rute untuk User API (dari Fase 2)
	router.POST("/users", createUser)
	router.GET("/users/:id", getUserByID)
	router.GET("/users", getAllUsers)

	// Rute untuk Account & Transaction API (Fase 3)
	router.POST("/accounts", createAccount)
	router.GET("/accounts/:id", getAccountByID)
	router.POST("/accounts/:id/deposit", deposit)
	router.POST("/accounts/:id/withdraw", withdraw)
	router.POST("/transactions/transfer", transfer)
	router.GET("/accounts/:id/transactions", getAccountTransactions)

	router.Run(":8080")
}

// Handler untuk membuat user baru (POST /users)
func createUser(c *gin.Context) {
	var req CreateUserRequest
	// Bind JSON request body ke struct CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Placeholder untuk hashing password (akan diimplementasikan di Fase 4)
	// Untuk saat ini, kita simpan password mentah (TIDAK AMAN DI PRODUKSI!)
	passwordHash := req.Password

	// Query SQL untuk memasukkan data user baru
	query := "INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)"
	result, err := db.Exec(query, req.Name, req.Email, passwordHash)
	if err != nil {
		// Tangani error, misalnya email sudah ada (UNIQUE constraint violation)
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Dapatkan ID dari user yang baru dibuat
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user ID"})
		return
	}

	// Ambil user yang baru dibuat dari database untuk respons yang lengkap
	var newUser User
	err = db.QueryRow("SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?", id).
		Scan(&newUser.ID, &newUser.Name, &newUser.Email, &newUser.CreatedAt, &newUser.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching new user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve newly created user"})
		return
	}

	c.JSON(http.StatusCreated, newUser)
}

// Handler untuk mendapatkan detail user berdasarkan ID (GET /users/:id)
func getUserByID(c *gin.Context) {
	id := c.Param("id") // Ambil ID dari URL parameter

	var user User
	// Query SQL untuk mengambil user berdasarkan ID
	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?"
	err := db.QueryRow(query, id).
		Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			log.Printf("Error getting user by ID: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// Handler untuk mendapatkan daftar semua user (GET /users)
func getAllUsers(c *gin.Context) {
	var users []User
	// Query SQL untuk mengambil semua user
	rows, err := db.Query("SELECT id, name, email, created_at, updated_at FROM users")
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	defer rows.Close() // Penting: Pastikan rows ditutup untuk melepaskan koneksi

	for rows.Next() { // Iterasi setiap baris hasil
		var user User
		// Scan nilai dari baris ke struct User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning user row: %v", err)
			// Lanjutkan ke baris berikutnya atau tangani error lebih lanjut
			continue
		}
		users = append(users, user) // Tambahkan user ke slice
	}

	if err = rows.Err(); err != nil { // Cek error setelah iterasi selesai
		log.Printf("Error during rows iteration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving users during iteration"})
		return
	}

	c.JSON(http.StatusOK, users)
}
func createAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek apakah user_id valid
	var userExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", req.UserID).Scan(&userExists)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
		return
	}
	if !userExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with specified ID does not exist"})
		return
	}

	// Masukkan akun baru ke database
	query := "INSERT INTO accounts (user_id, account_number, balance) VALUES (?, ?, ?)"
	result, err := db.Exec(query, req.UserID, req.AccountNumber, 0.00) // Saldo awal 0
	if err != nil {
		log.Printf("Error creating account: %v", err)
		// Tangani error jika account_number sudah ada (UNIQUE constraint)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account. Account number might already exist."})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID for account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve account ID"})
		return
	}

	var newAccount Account
	err = db.QueryRow("SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE id = ?", id).
		Scan(&newAccount.ID, &newAccount.UserID, &newAccount.AccountNumber, &newAccount.Balance, &newAccount.CreatedAt, &newAccount.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching new account: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve newly created account"})
		return
	}

	c.JSON(http.StatusCreated, newAccount)
}

// Handler untuk mendapatkan detail akun berdasarkan ID (GET /accounts/:id)
func getAccountByID(c *gin.Context) {
	id := c.Param("id")

	var account Account
	query := "SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE id = ?"
	err := db.QueryRow(query, id).
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
func deposit(c *gin.Context) {
	accountID := c.Param("id")
	var req DepositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mulai transaksi database
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction for deposit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process deposit"})
		return
	}
	defer tx.Rollback() // Pastikan rollback jika ada error sebelum commit

	// 1. Update saldo akun
	_, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", req.Amount, accountID)
	if err != nil {
		log.Printf("Error updating account balance for deposit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
		return
	}

	// 2. Catat transaksi
	_, err = tx.Exec("INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)",
		accountID, "deposit", req.Amount, "Deposit funds")
	if err != nil {
		log.Printf("Error logging deposit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}

	err = tx.Commit() // Commit transaksi
	if err != nil {
		log.Printf("Error committing deposit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit deposit"})
		return
	}

	var updatedAccount Account
	err = db.QueryRow("SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE id = ?", accountID).
		Scan(&updatedAccount.ID, &updatedAccount.UserID, &updatedAccount.AccountNumber, &updatedAccount.Balance, &updatedAccount.CreatedAt, &updatedAccount.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching updated account after deposit: %v", err)
		// Ini bisa jadi notifikasi ke admin, tapi deposit sebenarnya sudah sukses
		c.JSON(http.StatusOK, gin.H{"message": "Deposit successful, but failed to retrieve updated account details."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful", "account": updatedAccount})
}
func withdraw(c *gin.Context) {
	accountID := c.Param("id")
	var req DepositWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mulai transaksi database
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction for withdraw: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process withdrawal"})
		return
	}
	defer tx.Rollback() // Pastikan rollback jika ada error sebelum commit

	// 1. Cek saldo saat ini
	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM accounts WHERE id = ? FOR UPDATE", accountID).Scan(&currentBalance) // FOR UPDATE untuk locking baris
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

	// 2. Update saldo akun
	_, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", req.Amount, accountID)
	if err != nil {
		log.Printf("Error updating account balance for withdraw: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
		return
	}

	// 3. Catat transaksi
	_, err = tx.Exec("INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)",
		accountID, "withdraw", req.Amount, "Withdrawal funds")
	if err != nil {
		log.Printf("Error logging withdraw transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log transaction"})
		return
	}

	err = tx.Commit() // Commit transaksi
	if err != nil {
		log.Printf("Error committing withdraw transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit withdrawal"})
		return
	}

	var updatedAccount Account
	err = db.QueryRow("SELECT id, user_id, account_number, balance, created_at, updated_at FROM accounts WHERE id = ?", accountID).
		Scan(&updatedAccount.ID, &updatedAccount.UserID, &updatedAccount.AccountNumber, &updatedAccount.Balance, &updatedAccount.CreatedAt, &updatedAccount.UpdatedAt)
	if err != nil {
		log.Printf("Error fetching updated account after withdraw: %v", err)
		c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful, but failed to retrieve updated account details."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful", "account": updatedAccount})
}
func transfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.FromAccountID == req.ToAccountID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot transfer to the same account"})
		return
	}

	// Mulai transaksi database
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction for transfer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process transfer"})
		return
	}
	defer tx.Rollback() // Pastikan rollback jika ada error

	// 1. Ambil saldo akun pengirim dan penerima (dan kunci barisnya)
	var fromBalance, toBalance float64
	var fromAccountIDInt, toAccountIDInt int

	// Ambil ID integer dari nomor akun string
	err = tx.QueryRow("SELECT id, balance FROM accounts WHERE account_number = ? FOR UPDATE", req.FromAccountID).Scan(&fromAccountIDInt, &fromBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Source account not found"})
		} else {
			log.Printf("Error fetching source account for transfer: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve source account"})
		}
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

	// 2. Cek saldo cukup
	if fromBalance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds in source account"})
		return
	}

	// 3. Update saldo akun pengirim (debit)
	_, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", req.Amount, fromAccountIDInt)
	if err != nil {
		log.Printf("Error updating source account balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update source account balance"})
		return
	}

	// 4. Update saldo akun penerima (kredit)
	_, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", req.Amount, toAccountIDInt)
	if err != nil {
		log.Printf("Error updating destination account balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update destination account balance"})
		return
	}

	// 5. Catat transaksi untuk akun pengirim
	_, err = tx.Exec("INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)",
		fromAccountIDInt, "transfer_out", req.Amount, "Transfer to "+req.ToAccountID+": "+req.Description)
	if err != nil {
		log.Printf("Error logging transfer_out transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log outbound transaction"})
		return
	}

	// 6. Catat transaksi untuk akun penerima
	_, err = tx.Exec("INSERT INTO transactions (account_id, transaction_type, amount, description) VALUES (?, ?, ?, ?)",
		toAccountIDInt, "transfer_in", req.Amount, "Transfer from "+req.FromAccountID+": "+req.Description)
	if err != nil {
		log.Printf("Error logging transfer_in transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log inbound transaction"})
		return
	}

	err = tx.Commit() // Commit transaksi
	if err != nil {
		log.Printf("Error committing transfer transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transfer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}
func getAccountTransactions(c *gin.Context) {
	accountID := c.Param("id")

	var transactions []Transaction
	query := "SELECT id, account_id, transaction_type, amount, description, transaction_date FROM transactions WHERE account_id = ? ORDER BY transaction_date DESC"
	rows, err := db.Query(query, accountID)
	if err != nil {
		log.Printf("Error getting account transactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var t Transaction
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
