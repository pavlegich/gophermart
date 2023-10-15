package user

import (
	"context"
)

type User struct {
	ID       int    `json:"id,omitempty"`
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

type Service interface {
	List(ctx context.Context) ([]*User, error)
	Register(ctx context.Context, user *User) error
	Login(ctx context.Context, user *User) (*User, error)
}

type Repository interface {
	GetUserByLogin(ctx context.Context, login string) (*User, error)
	GetUsers(ctx context.Context) ([]*User, error)
	SaveUser(ctx context.Context, user *User) error
}
