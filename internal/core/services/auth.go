package services

import (
	"context"
	"errors"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService реализация сервиса аутентификации.
type AuthService struct {
	userRepo  ports.UserRepository
	secretKey string
}

// NewAuthService создает новый экземпляр AuthService.
func NewAuthService(userRepo ports.UserRepository, secretKey string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		secretKey: secretKey,
	}
}

// Register регистрирует нового пользователя и возвращает JWT токен.
func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", domain.ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &domain.User{
		Login:        login,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", err
	}

	return s.generateToken(user.ID)
}

// Login проверяет учетные данные и возвращает JWT токен.
func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	return s.generateToken(user.ID)
}

func (s *AuthService) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}
