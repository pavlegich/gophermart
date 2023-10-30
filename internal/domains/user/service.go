package user

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	errs "github.com/pavlegich/gophermart/internal/errors"
)

type UserService struct {
	repo Repository
}

func NewUserService(repo Repository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// Register проверяет и сохраняет данные нового пользователя в хранилище
func (s *UserService) Register(ctx context.Context, user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Register: hash generate failed %w", err)
	}
	user.Password = string(hashedPassword)
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("Register: save user failed %w", err)
	}
	return nil
}

// Login проверяет корректность полученных данных пользователя
func (s *UserService) Login(ctx context.Context, user *User) (*User, error) {
	storedUser, err := s.repo.GetUserByLogin(ctx, user.Login)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
		return nil, errs.ErrPasswordNotMatch
	}
	return storedUser, nil
}
