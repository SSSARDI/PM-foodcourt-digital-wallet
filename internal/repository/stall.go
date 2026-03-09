package repository

import (
	"context"
	"database/sql"
	"pmfoodcourt/internal/model"
)

type StallRepository struct{ db *sql.DB }

func NewStallRepo(db *sql.DB) *StallRepository { return &StallRepository{db} }

func (r *StallRepository) GetAll(ctx context.Context) ([]model.FoodStall, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, stall_name, owner_id, category, location, status, created_at
		 FROM food_stalls ORDER BY stall_name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stalls []model.FoodStall
	for rows.Next() {
		var s model.FoodStall
		if err := rows.Scan(&s.ID, &s.StallName, &s.OwnerID,
			&s.Category, &s.Location, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		stalls = append(stalls, s)
	}
	if stalls == nil {
		stalls = []model.FoodStall{}
	}
	return stalls, rows.Err()
}

func (r *StallRepository) GetByID(ctx context.Context, id string) (*model.FoodStall, error) {
	s := &model.FoodStall{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, stall_name, owner_id, category, location, status, created_at
		 FROM food_stalls WHERE id = ?`, id,
	).Scan(&s.ID, &s.StallName, &s.OwnerID, &s.Category, &s.Location, &s.Status, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *StallRepository) GetByOwner(ctx context.Context, ownerID string) ([]model.FoodStall, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, stall_name, owner_id, category, location, status, created_at
		 FROM food_stalls WHERE owner_id = ?`, ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stalls []model.FoodStall
	for rows.Next() {
		var s model.FoodStall
		if err := rows.Scan(&s.ID, &s.StallName, &s.OwnerID,
			&s.Category, &s.Location, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		stalls = append(stalls, s)
	}
	if stalls == nil {
		stalls = []model.FoodStall{}
	}
	return stalls, rows.Err()
}

func (r *StallRepository) Create(ctx context.Context, s model.FoodStall) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO food_stalls (id, stall_name, owner_id, category, location, status)
		 VALUES (?, ?, ?, ?, ?, 'OPEN')`,
		s.ID, s.StallName, s.OwnerID, s.Category, s.Location,
	)
	return err
}

func (r *StallRepository) Update(ctx context.Context, id string, req model.UpdateStallRequest) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE food_stalls SET
			stall_name = COALESCE(?, stall_name),
			category   = COALESCE(?, category),
			location   = COALESCE(?, location),
			status     = COALESCE(?, status)
		 WHERE id = ?`,
		req.StallName, req.Category, req.Location, req.Status, id,
	)
	return err
}

func (r *StallRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM food_stalls WHERE id = ?`, id)
	return err
}

func (r *StallRepository) CreateSale(ctx context.Context, s model.SaleTransaction) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sale_transactions (id, stall_id, customer_id, total_amount, payment_transaction_id)
		 VALUES (?, ?, ?, ?, ?)`,
		s.ID, s.StallID, s.CustomerID, s.TotalAmount, s.PaymentTransactionID,
	)
	return err
}

func (r *StallRepository) GetSalesByStall(ctx context.Context, stallID string) ([]model.SaleTransaction, error) {
	return r.querySales(ctx,
		`SELECT id, stall_id, customer_id, total_amount, payment_transaction_id, created_at
		 FROM sale_transactions WHERE stall_id = ? ORDER BY created_at DESC`, stallID,
	)
}

func (r *StallRepository) GetSalesByCustomer(ctx context.Context, customerID string) ([]model.SaleTransaction, error) {
	return r.querySales(ctx,
		`SELECT id, stall_id, customer_id, total_amount, payment_transaction_id, created_at
		 FROM sale_transactions WHERE customer_id = ? ORDER BY created_at DESC`, customerID,
	)
}

func (r *StallRepository) querySales(ctx context.Context, query, arg string) ([]model.SaleTransaction, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sales []model.SaleTransaction
	for rows.Next() {
		var s model.SaleTransaction
		if err := rows.Scan(&s.ID, &s.StallID, &s.CustomerID,
			&s.TotalAmount, &s.PaymentTransactionID, &s.CreatedAt); err != nil {
			return nil, err
		}
		sales = append(sales, s)
	}
	if sales == nil {
		sales = []model.SaleTransaction{}
	}
	return sales, rows.Err()
}
