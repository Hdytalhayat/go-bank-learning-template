// go-bank-app/config/database.go
package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // Driver MySQL
)

// DB adalah instance koneksi database yang dapat diakses dari package lain.
var DB *sql.DB

const DSN = "root:@tcp(127.0.0.1:3306)/bank_app_db?parseTime=true"

// InitDB menginisialisasi koneksi database.
func InitDB() {
	var err error
	DB, err = sql.Open("mysql", DSN)
	if err != nil {
		log.Fatalf("Kesalahan saat membuka koneksi database: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Kesalahan saat ping database: Pastikan MySQL server berjalan dan DSN benar. Error: %v", err)
	}

	fmt.Println("Koneksi ke database MySQL berhasil!")

	// Set pengaturan koneksi (opsional, untuk performa)
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)
}
