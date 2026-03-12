package repository

import (
	"context"
	"database/sql"
	"pmfoodcourt/internal/model"
)

type AdminRepository struct {
	db *sql.DB
}

// ── QR Repository ──────────────────────────────────────────────
type QRRepository struct {
	db *sql.DB
}

func NewQRRepo(db *sql.DB) *QRRepository {
	return &QRRepository{db: db}
}

func NewAdminRepo(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

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

// ── Refund Repository ──────────────────────────────────────────
type RefundRepository struct {
	db *sql.DB
}

func NewRefundRepo(db *sql.DB) *RefundRepository {
	return &RefundRepository{db: db}
}

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

// ── GP/VAT Repository ──────────────────────────────────────────
type GpVatRepository struct {
	db *sql.DB
}

func NewGpVatRepo(db *sql.DB) *GpVatRepository {
	return &GpVatRepository{db: db}
}

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

func (r *AdminRepository) GetSummary(ctx context.Context, start, end string) (map[string]float64, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'PAYMENT' THEN amount ELSE 0 END), 0) as total_sales,
			COALESCE(SUM(CASE WHEN type = 'TOPUP_QR' THEN amount ELSE 0 END), 0) as qr_topup,
			COALESCE(SUM(CASE WHEN type = 'TOPUP_CASH' THEN amount ELSE 0 END), 0) as cash_topup,
			COALESCE(SUM(CASE WHEN type = 'REFUND' THEN amount ELSE 0 END), 0) as cash_refund
		FROM wallet_transactions 
		WHERE DATE(created_at) BETWEEN ? AND ?`

	var totalSales, qrTopup, cashTopup, cashRefund float64
	err := r.db.QueryRowContext(ctx, query, start, end).Scan(&totalSales, &qrTopup, &cashTopup, &cashRefund)
	if err != nil {
		return nil, err
	}

	return map[string]float64{
		"totalSalesAmount": totalSales,
		"qrTopupAmount":    qrTopup,
		"cashTopupAmount":  cashTopup,
		"cashRefundAmount": cashRefund,
	}, nil
}

func (r *AdminRepository) GetStallReports(ctx context.Context, start, end string) ([]map[string]interface{}, error) {
	query := `
        SELECT 
            s.id,
            s.stall_name,
            COALESCE(s.owner_name, 'Not Specified') as owner_name,
            COALESCE(s.phone, 'No Phone Info') as phone,
            COALESCE(s.category, 'General') as category,
            s.status,
            -- 🚩 เปลี่ยนมาดึงจาก sale_transactions แทน
            (SELECT COALESCE(SUM(total_amount), 0) FROM sale_transactions WHERE stall_id = s.id AND DATE(created_at) BETWEEN ? AND ?) as daily_revenue
        FROM food_stalls s
        WHERE s.stall_name IS NOT NULL AND s.stall_name != '' AND s.id != 'SYSTEM-TOPUP'
        GROUP BY s.id, s.stall_name, s.owner_name, s.phone, s.category, s.status`

	rows, err := r.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []map[string]interface{}
	for rows.Next() {
		var id, name, owner, phone, category, status string
		var revenue float64

		if err := rows.Scan(&id, &name, &owner, &phone, &category, &status, &revenue); err != nil {
			return nil, err
		}

		reports = append(reports, map[string]interface{}{
			"id":            id,
			"stall_name":    name,
			"owner_name":    owner,
			"phone":         phone,
			"category":      category,
			"status":        status,
			"daily_revenue": revenue,
		})
	}
	return reports, nil
}

func (r *AdminRepository) UpdateStall(id string, req model.UpdateStallRequest) error {
	query := `UPDATE food_stalls 
              SET stall_name = ?, 
                  owner_name = ?, 
                  phone = ?, 
                  category = ?, 
                  status = ? 
              WHERE id = ?`

	// 🚩 ตรวจสอบว่าใน model.UpdateStallRequest ฟิลด์พวกนี้เป็น *string
	// sql.DB จะจัดการค่าที่เป็น Pointer ให้เองถ้าเป็น nil จะกลายเป็น NULL ใน DB
	_, err := r.db.Exec(query,
		req.StallName,
		req.OwnerName,
		req.Phone,
		req.Category,
		req.Status,
		id, // WHERE id = ?
	)
	return err
}

func (r *AdminRepository) UpdateStallStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE food_stalls SET status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *AdminRepository) GetStallByID(ctx context.Context, id string) (*model.FoodStall, error) {
	s := &model.FoodStall{}
	query := `SELECT id, stall_name, owner_id, owner_name, phone, category, status FROM food_stalls WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.StallName, &s.OwnerID, &s.OwnerName, &s.Phone, &s.Category, &s.Status,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *AdminRepository) CreateStall(ctx context.Context, stall *model.FoodStall) error {
	query := `INSERT INTO food_stalls (id, stall_name, owner_id, owner_name, phone, category, status) 
              VALUES (?, ?, ?, ?, ?, ?, 'ACTIVE')`

	_, err := r.db.ExecContext(ctx, query,
		stall.ID,
		stall.StallName,
		stall.OwnerID,
		stall.OwnerName,
		stall.Phone,
		stall.Category,
	)
	return err
}

func (r *UserRepository) CreateStaff(ctx context.Context, u model.User) error {
	query := `INSERT INTO users (id, full_name, role, status, password_hash) VALUES (?, ?, 'STAFF', 'ACTIVE', ?)`
	_, err := r.db.ExecContext(ctx, query, u.ID, u.FullName, u.PasswordHash)
	return err
}

func (r *UserRepository) UpdateStaff(ctx context.Context, id string, name string) error {
	query := `UPDATE users SET full_name = ? WHERE id = ? AND role = 'STAFF'`
	_, err := r.db.ExecContext(ctx, query, name, id)
	return err
}

func (r *AdminRepository) GetAllStalls(ctx context.Context) ([]model.FoodStall, error) {
	query := `SELECT id, stall_name, owner_id, COALESCE(owner_name, '') as owner_name, 
                     COALESCE(phone, '') as phone, COALESCE(category, '') as category, 
                     status, created_at 
              FROM food_stalls 
              WHERE id != 'SYSTEM-TOPUP'`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stalls []model.FoodStall
	for rows.Next() {
		var s model.FoodStall
		// ต้อง Scan ให้ครบตามจำนวน Field ใน Model FoodStall นะครับ
		if err := rows.Scan(&s.ID, &s.StallName, &s.OwnerID, &s.OwnerName, &s.Phone, &s.Category, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		stalls = append(stalls, s)
	}

	if stalls == nil {
		stalls = []model.FoodStall{}
	}
	return stalls, nil
}
