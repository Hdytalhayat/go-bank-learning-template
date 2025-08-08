// go-bank-app/auth/auth_util.go
package auth

import (
	"fmt"
	"go-bank-app/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Define custom claims
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// HashPassword mengenkripsi password menggunakan bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("gagal mengenkripsi password: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash membandingkan password mentah dengan hash yang tersimpan.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWTToken membuat JWT baru untuk user yang diberikan.
func GenerateJWTToken(userID int) (string, error) {
	// Waktu kadaluarsa token (misal: 24 jam dari sekarang)
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.JWTSecretKey)
	if err != nil {
		return "", fmt.Errorf("gagal menandatangani token: %w", err)
	}

	return tokenString, nil
}
