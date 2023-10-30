package user

import (
	"context"
)

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Service interface {
	Register(ctx context.Context, user *User) error
	Login(ctx context.Context, user *User) (*User, error)
}

type Repository interface {
	GetUserByLogin(ctx context.Context, login string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
}
