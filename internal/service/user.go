package service

import (
	"context"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetAll(ctx context.Context) ([]model.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) Create(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	user := model.User{
		Name:  req.Name,
		Email: req.Email,
	}
	err := s.repo.Create(ctx, user)
	return &user, err
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}