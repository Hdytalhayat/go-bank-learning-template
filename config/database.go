// go-bank-app/config/database.go
package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// DB is a global database connection instance accessible from other packages.
var DB *sql.DB

const DSN = "root:@tcp(127.0.0.1:3306)/bank_app_db?parseTime=true"

// InitDB initializes the database connection.
func InitDB() {
	var err error

	// Open the database connection using the MySQL driver and DSN.
	DB, err = sql.Open("mysql", DSN)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	// Check if the database connection is alive.
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: Make sure the MySQL server is running and the DSN is correct. Error: %v", err)
	}

	fmt.Println("Successfully connected to the MySQL database!")

	// Set connection pool settings (optional but recommended for performance)
	DB.SetMaxOpenConns(10)                 // Maximum number of open connections to the database
	DB.SetMaxIdleConns(5)                  // Maximum number of idle connections in the pool
	DB.SetConnMaxLifetime(5 * time.Minute) // Maximum amount of time a connection may be reused
}
