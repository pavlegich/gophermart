package hash

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	ID int
}

type JWT struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	tokenExp   time.Duration
}

func NewJWT(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, tokenExp time.Duration) *JWT {
	return &JWT{
		privateKey: privateKey,
		publicKey:  publicKey,
		tokenExp:   tokenExp,
	}
}

// Create создаёт токен и возвращает его в виде строки
func (j *JWT) Create(ctx context.Context, id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenExp)),
		},
		ID: id,
	})

	tokenString, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", fmt.Errorf("Create: sign string with key failed %w", err)
	}

	return tokenString, nil
}

// Validate возвращает полученные из токена данные для аутентификации
func (j *JWT) Validate(tokenString string) (int, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Validate: unexpected signing method: %v", t.Header["alg"])
			}
			return j.publicKey, nil
		})
	if err != nil {
		return -1, fmt.Errorf("Validate: parse token failed %w", err)
	}
	if !token.Valid {
		return -1, fmt.Errorf("Validate: token is not valid")
	}

	return claims.ID, nil
}
