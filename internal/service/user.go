package service

import (
	"context"
	"errors"
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

func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (s *UserService) Create(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	// basic creation via UserRepository (ไม่มี auth — ใช้ AuthService แทนสำหรับ register จริง)
	user := model.User{
		FullName: req.FullName,
		Role:     "CUSTOMER",
		Status:   "ACTIVE",
	}
	if req.Email != "" {
		user.Email = &req.Email
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
