package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
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
			token, err := verifyNextAuthToken(tokenString)
			if err != nil {
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

func verifyNextAuthToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// NEXTAUTH_SECRETで検証
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	// emailを取得
	//claims := token.Claims.(jwt.MapClaims)
	//email := claims["email"].(string)
	//
	//fmt.Println("email:", email)

	return token, nil
}
