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
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, full_name, email, phone, password_hash, role, status, created_at
		 FROM users ORDER BY full_name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.FullName, &u.Email, &u.Phone,
			&u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if users == nil {
		users = []model.User{}
	}
	return users, rows.Err()
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, full_name, email, phone, password_hash, role, status, created_at
		 FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.FullName, &u.Email, &u.Phone,
		&u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) Create(ctx context.Context, user model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, full_name, email, phone, password_hash, role, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.FullName, user.Email, user.Phone,
		user.PasswordHash, user.Role, user.Status,
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}

func (r *UserRepository) ListByRole(ctx context.Context, role string) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, full_name, email, phone, password_hash, role, status, created_at
		 FROM users WHERE role = ? ORDER BY full_name`,
		role,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.FullName, &u.Email, &u.Phone,
			&u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if users == nil {
		users = []model.User{}
	}
	return users, rows.Err()
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET status = ? WHERE id = ?`,
		status, id,
	)
	return err
}
