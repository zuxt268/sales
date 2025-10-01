package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/domain"
)

// JWTMiddleware はBearerトークンを検証するミドルウェア
func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				slog.Warn("Missing Authorization header")
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "unauthorized",
					"message": "missing authorization header",
				})
			}

			// Bearer トークンの形式チェック
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				slog.Warn("Invalid Authorization header format", "header", authHeader)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "unauthorized",
					"message": "invalid authorization header format",
				})
			}

			tokenString := parts[1]

			// トークンの検証
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// 署名アルゴリズムの検証
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					slog.Warn("Unexpected signing method", "method", token.Header["alg"])
					return nil, domain.WrapUnauthorized("unexpected signing method")
				}
				return []byte(config.Env.JWTSecret), nil
			})

			if err != nil {
				slog.Warn("Failed to parse token", "error", err.Error())
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "unauthorized",
					"message": "invalid or expired token",
				})
			}

			if !token.Valid {
				slog.Warn("Invalid token")
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":   "unauthorized",
					"message": "invalid token",
				})
			}

			// トークンのクレームをコンテキストに保存
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("user_claims", claims)
			}

			return next(c)
		}
	}
}
