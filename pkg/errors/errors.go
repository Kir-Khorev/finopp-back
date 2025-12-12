package errors

import (
	"fmt"
	"net/http"
)

// AppError представляет структурированную ошибку приложения
type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"error"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// Predefined errors
var (
	ErrBadRequest          = &AppError{Code: http.StatusBadRequest, Message: "Неверный запрос"}
	ErrUnauthorized        = &AppError{Code: http.StatusUnauthorized, Message: "Требуется авторизация"}
	ErrForbidden           = &AppError{Code: http.StatusForbidden, Message: "Доступ запрещён"}
	ErrNotFound            = &AppError{Code: http.StatusNotFound, Message: "Ресурс не найден"}
	ErrInternalServer      = &AppError{Code: http.StatusInternalServerError, Message: "Внутренняя ошибка сервера"}
	ErrInvalidCredentials  = &AppError{Code: http.StatusUnauthorized, Message: "Неверный email или пароль"}
	ErrEmailExists         = &AppError{Code: http.StatusConflict, Message: "Email уже зарегистрирован"}
	ErrWeakPassword        = &AppError{Code: http.StatusBadRequest, Message: "Пароль должен содержать минимум 6 символов"}
	ErrInvalidToken        = &AppError{Code: http.StatusUnauthorized, Message: "Невалидный токен"}
	ErrGroqAPIUnavailable  = &AppError{Code: http.StatusServiceUnavailable, Message: "AI сервис временно недоступен"}
)

// New создаёт новую ошибку приложения
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewWithDetails создаёт ошибку с дополнительными деталями
func NewWithDetails(code int, message, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Wrap оборачивает стандартную ошибку в AppError
func Wrap(err error, message string) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Details: err.Error(),
	}
}

