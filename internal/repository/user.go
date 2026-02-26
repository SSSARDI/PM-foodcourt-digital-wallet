package repository

import (
	"context"
	"database/sql"
	"pmfoodcourt/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetAll(ctx context.Context) ([]model.User, error) {
	return []model.User{}, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return &model.User{ID: id}, nil
}

func (r *UserRepository) Create(ctx context.Context, user model.User) error {
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return nil
}