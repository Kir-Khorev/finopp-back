package auth

import (
	"time"

	apperrors "github.com/Kir-Khorev/finopp-back/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      *Repository
	jwtSecret string
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Register(req RegisterRequest) (*AuthResponse, error) {
	// Валидация
	if req.Email == "" || req.Password == "" {
		return nil, apperrors.ErrBadRequest
	}
	if len(req.Password) < 6 {
		return nil, apperrors.ErrWeakPassword
	}

	// Проверка существования email
	exists, err := s.repo.EmailExists(req.Email)
	if err != nil {
		return nil, apperrors.Wrap(err, "Ошибка проверки email")
	}
	if exists {
		return nil, apperrors.ErrEmailExists
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Wrap(err, "Ошибка хеширования пароля")
	}

	// Создание пользователя
	user, err := s.repo.CreateUser(req.Email, string(hashedPassword), req.Name)
	if err != nil {
		return nil, apperrors.Wrap(err, "Ошибка создания пользователя")
	}

	// Генерация токена
	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, apperrors.Wrap(err, "Ошибка генерации токена")
	}

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *Service) Login(req LoginRequest) (*AuthResponse, error) {
	// Валидация
	if req.Email == "" || req.Password == "" {
		return nil, apperrors.ErrBadRequest
	}

	// Получение пользователя
	user, passwordHash, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	// Генерация токена
	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, apperrors.Wrap(err, "Ошибка генерации токена")
	}

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *Service) generateToken(userID int, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 дней
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

