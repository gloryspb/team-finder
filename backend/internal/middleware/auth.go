package middleware

import (
	"strings"

	"team-finder/backend/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	ContextUserID   = "user_id"
	ContextUserRole = "user_role"
)

func JWT(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				return domain.ErrUnauthorized
			}
			tokenString := strings.TrimPrefix(header, "Bearer ")
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return domain.ErrUnauthorized
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return domain.ErrUnauthorized
			}
			sub, _ := claims["sub"].(string)
			role, _ := claims["role"].(string)
			userID, err := uuid.Parse(sub)
			if err != nil {
				return domain.ErrUnauthorized
			}
			c.Set(ContextUserID, userID)
			c.Set(ContextUserRole, role)
			return next(c)
		}
	}
}

func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if Role(c) != "admin" {
			return domain.ErrForbidden
		}
		return next(c)
	}
}

func UserID(c echo.Context) uuid.UUID {
	value, _ := c.Get(ContextUserID).(uuid.UUID)
	return value
}

func Role(c echo.Context) string {
	value, _ := c.Get(ContextUserRole).(string)
	return value
}
