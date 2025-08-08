// go-bank-app/repositories/user_repository.go
package repositories

import (
	"database/sql" // Untuk akses ke config.DB
	"go-bank-app/models"
)

// UserRepository adalah interface untuk operasi User di database.
type UserRepository interface {
	CreateUser(user *models.User) (int64, error)
	GetUserByID(id int) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
}

// userRepositoryImpl adalah implementasi konkrit dari UserRepository.
type userRepositoryImpl struct {
	db *sql.DB
}

// NewUserRepository membuat instance baru dari UserRepository.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) CreateUser(user *models.User) (int64, error) {
	query := "INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)"
	result, err := r.db.Exec(query, user.Name, user.Email, user.PasswordHash)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *userRepositoryImpl) GetUserByID(id int) (*models.User, error) {
	var user models.User
	query := "SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = ?"
	err := r.db.QueryRow(query, id).
		Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := "SELECT id, name, email, password_hash FROM users WHERE email = ?"
	err := r.db.QueryRow(query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) GetAllUsers() ([]models.User, error) {
	var users []models.User
	rows, err := r.db.Query("SELECT id, name, email, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err // Atau log dan continue
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
