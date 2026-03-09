package repository

import (
	"context"
	"database/sql"
	"fmt"
	"pmfoodcourt/internal/model"
)

type WalletRepository struct{ db *sql.DB }

func NewWalletRepo(db *sql.DB) *WalletRepository { return &WalletRepository{db} }

func (r *WalletRepository) Create(ctx context.Context, w model.Wallet) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO wallets (id, public_id, user_id, balance) VALUES (?, ?, ?, 0.00)`,
		w.ID, w.PublicID, w.UserID,
	)
	return err
}

func (r *WalletRepository) GetByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	return r.scan(r.db.QueryRowContext(ctx,
		`SELECT id, public_id, user_id, balance, created_at FROM wallets WHERE user_id = ?`, userID,
	))
}

func (r *WalletRepository) GetByPublicID(ctx context.Context, publicID string) (*model.Wallet, error) {
	return r.scan(r.db.QueryRowContext(ctx,
		`SELECT id, public_id, user_id, balance, created_at FROM wallets WHERE public_id = ?`, publicID,
	))
}

func (r *WalletRepository) GetByID(ctx context.Context, id string) (*model.Wallet, error) {
	return r.scan(r.db.QueryRowContext(ctx,
		`SELECT id, public_id, user_id, balance, created_at FROM wallets WHERE id = ?`, id,
	))
}

func (r *WalletRepository) TopUpCash(ctx context.Context, walletID, txnID, staffID string, amount float64) (*model.Wallet, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx,
		`UPDATE wallets SET balance = balance + ? WHERE id = ?`, amount, walletID,
	); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO wallet_transactions (id, wallet_id, type, amount, created_by)
		 VALUES (?, ?, 'TOPUP_CASH', ?, ?)`,
		txnID, walletID, amount, staffID,
	); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, walletID)
}

func (r *WalletRepository) TopUpQR(ctx context.Context, walletID, txnID, qrID string, amount float64) (*model.Wallet, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx,
		`UPDATE wallets SET balance = balance + ? WHERE id = ?`, amount, walletID,
	); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO wallet_transactions (id, wallet_id, type, amount, reference_id)
		 VALUES (?, ?, 'TOPUP_QR', ?, ?)`,
		txnID, walletID, amount, qrID,
	); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx,
		`UPDATE qr_payments SET status = 'PAID' WHERE id = ?`, qrID,
	); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, walletID)
}

func (r *WalletRepository) Payment(ctx context.Context, walletID, txnID, saleID string, amount float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var balance float64
	if err = tx.QueryRowContext(ctx,
		`SELECT balance FROM wallets WHERE id = ? FOR UPDATE`, walletID,
	).Scan(&balance); err != nil {
		return err
	}
	if balance < amount {
		return fmt.Errorf("insufficient balance")
	}
	if _, err = tx.ExecContext(ctx,
		`UPDATE wallets SET balance = balance - ? WHERE id = ?`, amount, walletID,
	); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO wallet_transactions (id, wallet_id, type, amount, reference_id)
		 VALUES (?, ?, 'PAYMENT', ?, ?)`,
		txnID, walletID, amount, saleID,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *WalletRepository) Refund(ctx context.Context, walletID, txnID, refundID string, amount float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx,
		`UPDATE wallets SET balance = balance + ? WHERE id = ?`, amount, walletID,
	); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO wallet_transactions (id, wallet_id, type, amount, reference_id)
		 VALUES (?, ?, 'REFUND', ?, ?)`,
		txnID, walletID, amount, refundID,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *WalletRepository) GetTransactions(ctx context.Context, walletID string) ([]model.WalletTransaction, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, wallet_id, type, amount, reference_id, created_by, created_at
		 FROM wallet_transactions WHERE wallet_id = ? ORDER BY created_at DESC`, walletID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var txns []model.WalletTransaction
	for rows.Next() {
		var t model.WalletTransaction
		if err := rows.Scan(&t.ID, &t.WalletID, &t.Type, &t.Amount,
			&t.ReferenceID, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	if txns == nil {
		txns = []model.WalletTransaction{}
	}
	return txns, rows.Err()
}

func (r *WalletRepository) scan(row *sql.Row) (*model.Wallet, error) {
	w := &model.Wallet{}
	err := row.Scan(&w.ID, &w.PublicID, &w.UserID, &w.Balance, &w.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return w, err
}
