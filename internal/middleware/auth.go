package middleware

import (
	"net/http"
	"strings"

	"github.com/Kir-Khorev/finopp-back/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware проверяет JWT токен
func AuthMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(errors.ErrUnauthorized.Code, errors.ErrUnauthorized)
			}

			// Проверяем формат "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(errors.ErrInvalidToken.Code, errors.ErrInvalidToken)
			}

			tokenString := parts[1]

			// Парсим токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Проверяем метод подписи
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.ErrInvalidToken
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				return c.JSON(errors.ErrInvalidToken.Code, errors.ErrInvalidToken)
			}

			// Извлекаем claims
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("user_id", int(claims["user_id"].(float64)))
				c.Set("email", claims["email"].(string))
			} else {
				return c.JSON(errors.ErrInvalidToken.Code, errors.ErrInvalidToken)
			}

			return next(c)
		}
	}
}

// OptionalAuthMiddleware проверяет токен если он есть, но не требует его
func OptionalAuthMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return next(c)
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString := parts[1]
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.ErrInvalidToken
					}
					return []byte(jwtSecret), nil
				})

				if err == nil && token.Valid {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						c.Set("user_id", int(claims["user_id"].(float64)))
						c.Set("email", claims["email"].(string))
					}
				}
			}

			return next(c)
		}
	}
}

