package repository

import (
	"context"
	"database/sql"
	"pmfoodcourt/internal/model"
	"time"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepo(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
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

func (r *AuthRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, full_name, email, phone, password_hash, role, status, created_at
		 FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.FullName, &u.Email, &u.Phone,
		&u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *AuthRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, full_name, email, phone, password_hash, role, status, created_at
		 FROM users WHERE phone = ?`, phone,
	).Scan(&u.ID, &u.FullName, &u.Email, &u.Phone,
		&u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *AuthRepository) Create(ctx context.Context, user model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, full_name, email, phone, password_hash, role, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.FullName, user.Email, user.Phone,
		user.PasswordHash, user.Role, user.Status,
	)
	return err
}

func (r *AuthRepository) CreateResetToken(ctx context.Context, id, userID, token string, expires time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (id, user_id, token, expires_at, used)
		 VALUES (?, ?, ?, ?, ?)`,
		id, userID, token, expires, false,
	)
	return err
}

func (r *AuthRepository) GetResetToken(ctx context.Context, token string) (string, bool, error) {
	var userID string
	var used bool
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id, used FROM password_reset_tokens WHERE token = ? AND expires_at > ?`,
		token, time.Now(),
	).Scan(&userID, &used)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	return userID, used, err
}

func (r *AuthRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ? WHERE id = ?`,
		passwordHash, userID,
	)
	return err
}

func (r *AuthRepository) MarkResetTokenUsed(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used = ? WHERE token = ?`,
		true, token,
	)
	return err
}
