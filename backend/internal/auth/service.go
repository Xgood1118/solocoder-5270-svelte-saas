package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type Service struct {
	db        *sql.DB
	jwtSecret string
}

func NewService(db *sql.DB, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) GetUserByID(ctx context.Context, userID string) (*User, error) {
	var user User
	err := s.db.QueryRowContext(ctx,
		"SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.db.QueryRowContext(ctx,
		"SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) CreateUser(ctx context.Context, email, password, name string) (*User, error) {
	existing, _ := s.GetUserByEmail(ctx, email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(passwordHash),
		Name:         name,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		user.ID, user.Email, user.PasswordHash, user.Name, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

func (s *Service) GenerateToken(userID, orgID string, duration time.Duration) (string, error) {
	return GenerateToken(userID, orgID, s.jwtSecret, duration)
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	return ValidateToken(tokenString, s.jwtSecret)
}
