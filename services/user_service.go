// go-bank-app/services/user_service.go
package services

import (
	"database/sql"
	"fmt"
	"go-bank-app/auth" // Untuk hashing password
	"go-bank-app/models"
	"go-bank-app/repositories" // Untuk menggunakan repository
)

// UserService adalah interface untuk logika bisnis User.
type UserService interface {
	RegisterUser(req *models.CreateUserRequest) (*models.User, error)
	LoginUser(email, password string) (string, *models.User, error) // Mengembalikan token dan user
	GetUserByID(id int) (*models.User, error)
	GetAllUsers() ([]models.User, error)
}

// userServiceImpl adalah implementasi konkrit dari UserService.
type userServiceImpl struct {
	userRepo repositories.UserRepository
}

// NewUserService membuat instance baru dari UserService.
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userServiceImpl{userRepo: userRepo}
}

func (s *userServiceImpl) RegisterUser(req *models.CreateUserRequest) (*models.User, error) {
	existingUser, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("gagal memeriksa email user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("email sudah terdaftar")
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("gagal mengenkripsi password: %w", err)
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	id, err := s.userRepo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat user di database: %w", err)
	}
	user.ID = int(id) // Konversi int64 ke int

	// Ambil user yang baru dibuat untuk mendapatkan created_at/updated_at
	newUser, err := s.userRepo.GetUserByID(int(id))
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil user baru: %w", err)
	}

	return newUser, nil
}

func (s *userServiceImpl) LoginUser(email, password string) (string, *models.User, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil, fmt.Errorf("kredensial tidak valid")
		}
		return "", nil, fmt.Errorf("gagal mengambil user: %w", err)
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return "", nil, fmt.Errorf("kredensial tidak valid")
	}

	token, err := auth.GenerateJWTToken(user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("gagal menghasilkan token: %w", err)
	}

	return token, user, nil
}

func (s *userServiceImpl) GetUserByID(id int) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userServiceImpl) GetAllUsers() ([]models.User, error) {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}
