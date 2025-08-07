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
	// Inisialisasi koneksi database saat aplikasi dimulai
	initDB()
	defer db.Close() // Pastikan koneksi ditutup saat main() selesai

	router := gin.Default()

	// 1. Endpoint GET /ping
	router.GET("/ping", func(c *gin.Context) {
		// Mengirimkan respons JSON dengan status HTTP 200 OK
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// 2. Endpoint GET /hello/:name
	// Parameter :name akan diambil dari URL
	router.GET("/hello/:name", func(c *gin.Context) {
		// Mengambil nilai parameter "name" dari URL
		name := c.Param("name")
		// Mengirimkan respons JSON
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, " + name + "!",
		})
	})
	// Rute untuk User API
	router.POST("/users", createUser)
	router.GET("/users/:id", getUserByID)
	router.GET("/users", getAllUsers)
	// Menjalankan server di port 8080
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
