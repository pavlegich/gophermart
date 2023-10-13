package user

import (
	"context"
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
