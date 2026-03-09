package repository

import (
	"context"
	"database/sql"
	"pmfoodcourt/internal/model"
)

// ── QR Payments ───────────────────────────────────────────────

type QRRepository struct{ db *sql.DB }

func NewQRRepo(db *sql.DB) *QRRepository { return &QRRepository{db} }

func (r *QRRepository) Create(ctx context.Context, q model.QRPayment) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO qr_payments (id, stall_id, qr_token, amount, expires_at, status)
		 VALUES (?, ?, ?, ?, ?, 'PENDING')`,
		q.ID, q.StallID, q.QRToken, q.Amount, q.ExpiresAt,
	)
	return err
}

func (r *QRRepository) GetByToken(ctx context.Context, token string) (*model.QRPayment, error) {
	q := &model.QRPayment{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, stall_id, qr_token, amount, expires_at, status, created_at
		 FROM qr_payments WHERE qr_token = ?`, token,
	).Scan(&q.ID, &q.StallID, &q.QRToken, &q.Amount, &q.ExpiresAt, &q.Status, &q.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return q, err
}

func (r *QRRepository) ExpireStale(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE qr_payments SET status = 'EXPIRED'
		 WHERE status = 'PENDING' AND expires_at < NOW()`,
	)
	return err
}

// ── Cash Refunds ──────────────────────────────────────────────

type RefundRepository struct{ db *sql.DB }

func NewRefundRepo(db *sql.DB) *RefundRepository { return &RefundRepository{db} }

func (r *RefundRepository) Create(ctx context.Context, ref model.PendingCashRefund) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO pending_cash_refunds (id, wallet_id, amount, status, initiated_by, expires_at)
		 VALUES (?, ?, ?, 'PENDING', ?, ?)`,
		ref.ID, ref.WalletID, ref.Amount, ref.InitiatedBy, ref.ExpiresAt,
	)
	return err
}

func (r *RefundRepository) GetByID(ctx context.Context, id string) (*model.PendingCashRefund, error) {
	ref := &model.PendingCashRefund{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, wallet_id, amount, status, initiated_by, expires_at, created_at
		 FROM pending_cash_refunds WHERE id = ?`, id,
	).Scan(&ref.ID, &ref.WalletID, &ref.Amount, &ref.Status,
		&ref.InitiatedBy, &ref.ExpiresAt, &ref.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return ref, err
}

func (r *RefundRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE pending_cash_refunds SET status = ? WHERE id = ?`, status, id,
	)
	return err
}

func (r *RefundRepository) ListPending(ctx context.Context) ([]model.PendingCashRefund, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, wallet_id, amount, status, initiated_by, expires_at, created_at
		 FROM pending_cash_refunds WHERE status = 'PENDING' ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refs []model.PendingCashRefund
	for rows.Next() {
		var ref model.PendingCashRefund
		if err := rows.Scan(&ref.ID, &ref.WalletID, &ref.Amount, &ref.Status,
			&ref.InitiatedBy, &ref.ExpiresAt, &ref.CreatedAt); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	if refs == nil {
		refs = []model.PendingCashRefund{}
	}
	return refs, rows.Err()
}

// ── GP/VAT ────────────────────────────────────────────────────

type GpVatRepository struct{ db *sql.DB }

func NewGpVatRepo(db *sql.DB) *GpVatRepository { return &GpVatRepository{db} }

func (r *GpVatRepository) Get(ctx context.Context) (*model.GpVat, error) {
	gv := &model.GpVat{}
	err := r.db.QueryRowContext(ctx, `SELECT gp, vat FROM gp_vat LIMIT 1`).Scan(&gv.GP, &gv.VAT)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return gv, err
}

func (r *GpVatRepository) Upsert(ctx context.Context, gv model.GpVat) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO gp_vat (gp, vat) VALUES (?, ?)
		 ON DUPLICATE KEY UPDATE vat = VALUES(vat)`,
		gv.GP, gv.VAT,
	)
	return err
}
