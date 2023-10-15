package hash

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	ID int
}

const tokenExp = time.Hour * 3
const secretKey = "supersecretkey"

// BuildJWTString создаёт токен и возвращает его в виде строки
func BuildJWTString(ctx context.Context, id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		ID: id,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetCredentials возвращает полученные из токена данные для аутентификации
func GetCredentials(tokenString string) (int, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("GetCredentials: unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return -1, fmt.Errorf("GetCredentials: parse token failed %w", err)
	}
	if !token.Valid {
		return -1, fmt.Errorf("GetCredentials: token is not valid")
	}

	return claims.ID, nil
}
