// go-bank-app/config/config.go
package config

import "os"

// JWTSecretKey adalah kunci rahasia untuk menandatangani JWT.
// DI LINGKUNGAN PRODUKSI, INI HARUS DIBAWA DARI ENVIRONMENT VARIABLE ATAU SISTEM KONFIGURASI YANG AMAN!
var JWTSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func init() {
	if len(JWTSecretKey) == 0 {
		// Default jika env var tidak diset. Ubah ini di produksi!
		JWTSecretKey = []byte("supersecretjwtreallystrongkey")
		// log.Println("WARNING: JWT_SECRET_KEY environment variable not set. Using default key. DO NOT USE IN PRODUCTION!")
	}
}
