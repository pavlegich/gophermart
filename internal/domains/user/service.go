package user

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	errs "github.com/pavlegich/gophermart/internal/errors"
)

// Service содержит указатель на базу данных
type UserService struct {
	repository Repository
}

func NewUserService(repo Repository) *UserService {
	return &UserService{
		repository: repo,
	}
}

func (s *UserService) List(ctx context.Context) ([]*User, error) {
	return nil, nil
}

func (s *UserService) Register(ctx context.Context, user *User) error {
	if err := s.repository.SaveUser(ctx, user); err != nil {
		return err
	}
	return nil
}

func (s *UserService) Login(ctx context.Context, user *User) error {
	storedUser, err := s.repository.GetUserByID(ctx, user.Login)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
		return errs.ErrPasswordNotMatch
	}
	return nil
}
