package middleware

import (
	"log"
	"net/http"

	apperrors "github.com/Kir-Khorev/finopp-back/pkg/errors"
	"github.com/labstack/echo/v4"
)

// ErrorHandler обрабатывает ошибки приложения
func ErrorHandler(err error, c echo.Context) {
	// Если это наша кастомная ошибка
	if appErr, ok := err.(*apperrors.AppError); ok {
		if appErr.Code >= 500 {
			log.Printf("Internal error: %v", appErr)
		}
		_ = c.JSON(appErr.Code, appErr)
		return
	}

	// Если это Echo HTTPError
	if he, ok := err.(*echo.HTTPError); ok {
		_ = c.JSON(he.Code, map[string]interface{}{
			"error": he.Message,
		})
		return
	}

	// Для всех остальных ошибок
	log.Printf("Unexpected error: %v", err)
	_ = c.JSON(http.StatusInternalServerError, map[string]string{
		"error": "Внутренняя ошибка сервера",
	})
}

// RequestLogger логирует входящие запросы с полезной информацией
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			log.Printf("%s %s from %s", req.Method, req.URL.Path, req.RemoteAddr)
			return next(c)
		}
	}
}

