package user

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	errs "github.com/pavlegich/gophermart/internal/errors"
)

// UserService содержит методы и данные сервиса пользователя
type UserService struct {
	repo Repository
}

// NewUserService возвращает новый сервис для пользователя
func NewUserService(repo Repository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// List возвращает список пользователей
func (s *UserService) List(ctx context.Context) ([]*User, error) {
	return nil, nil
}

// Register проверяет и сохраняет данные нового пользователя в хранилище
func (s *UserService) Register(ctx context.Context, user *User) error {
	if err := s.repo.SaveUser(ctx, user); err != nil {
		return err
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
