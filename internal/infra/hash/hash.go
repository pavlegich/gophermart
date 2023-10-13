package hash

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	login string
}

const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "supersecretkey"

// BuildJWTString создаёт токен и возвращает его в виде строки
func BuildJWTString(ctx context.Context) (string, error) {
	login, ok := ctx.Value("Login").(string)
	if !ok {
		return "", fmt.Errorf("BuildJWTString: convert context value failed")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		login: login,
	})

	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetCredentials возвращает полученные из токена данные для аутентификации
func GetCredentials(tokenString string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("GetCredentials: unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SECRET_KEY), nil
		})
	if err != nil {
		return "", fmt.Errorf("GetCredentials: parse token failed %w", err)
	}
	if !token.Valid {
		return "", fmt.Errorf("GetCredentials: token is not valid")
	}

	return claims.login, nil
}
