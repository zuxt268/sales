package auth

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zuxt268/sales/internal/config"
)

// Claims はJWTのクレーム情報
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(password string) (string, error) {
	p := []byte(password)

	sha := sha256.Sum256(p)
	hashStr := fmt.Sprintf("%x", sha)

	if hashStr != config.Env.Password {
		return "", fmt.Errorf("invalid password")
	}
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(365 * 24 * time.Hour)), // 一年
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Env.JWTSecret))
}
