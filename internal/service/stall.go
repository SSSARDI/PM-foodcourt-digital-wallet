package service

import (
	"context"
	"errors"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/repository"

	"github.com/google/uuid"
)

type StallService struct {
	repo *repository.StallRepository
}

func NewStallService(repo *repository.StallRepository) *StallService {
	return &StallService{repo}
}

func (s *StallService) List(ctx context.Context) ([]model.FoodStall, error) {
	return s.repo.GetAll(ctx)
}

func (s *StallService) Get(ctx context.Context, id string) (*model.FoodStall, error) {
	stall, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if stall == nil {
		return nil, errors.New("stall not found")
	}
	return stall, nil
}

func (s *StallService) MyStalls(ctx context.Context, ownerID string) ([]model.FoodStall, error) {
	return s.repo.GetByOwner(ctx, ownerID)
}

func (s *StallService) Create(ctx context.Context, ownerID string, req model.CreateStallRequest) (*model.FoodStall, error) {
	if req.StallName == "" {
		return nil, errors.New("stall_name is required")
	}
	stall := model.FoodStall{
		ID:        uuid.New().String(),
		StallName: req.StallName,
		OwnerID:   ownerID,
		Category:  req.Category,
		Location:  req.Location,
	}
	if err := s.repo.Create(ctx, stall); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, stall.ID)
}

func (s *StallService) Update(ctx context.Context, id string, req model.UpdateStallRequest) (*model.FoodStall, error) {
	if err := s.repo.Update(ctx, id, req); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *StallService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *StallService) SaleHistory(ctx context.Context, stallID string) ([]model.SaleTransaction, error) {
	return s.repo.GetSalesByStall(ctx, stallID)
}

func (s *StallService) MySales(ctx context.Context, customerID string) ([]model.SaleTransaction, error) {
	return s.repo.GetSalesByCustomer(ctx, customerID)
}
