package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/repository"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	adminRepo  *repository.AdminRepository
	userRepo   *repository.UserRepository
	walletRepo *repository.WalletRepository
	refundRepo *repository.RefundRepository
	gpvatRepo  *repository.GpVatRepository
	db         *sql.DB
}

func NewAdminService(
	adminRepo *repository.AdminRepository,
	ur *repository.UserRepository,
	wr *repository.WalletRepository,
	rr *repository.RefundRepository,
	gv *repository.GpVatRepository,
	db *sql.DB,
) *AdminService {
	return &AdminService{
		adminRepo:  adminRepo,
		userRepo:   ur,
		walletRepo: wr,
		refundRepo: rr,
		gpvatRepo:  gv,
		db:         db,
	}
}

func (s *AdminService) ListUsers(ctx context.Context, role string) ([]model.User, error) {
	return s.userRepo.ListByRole(ctx, role)
}

func (s *AdminService) SuspendUser(ctx context.Context, id string) error {
	return s.userRepo.UpdateStatus(ctx, id, "SUSPENDED")

}

func (s *AdminService) ActivateUser(ctx context.Context, id string) error {
	return s.userRepo.UpdateStatus(ctx, id, "ACTIVE")
}

func (s *AdminService) InitiateRefund(ctx context.Context, staffID string, req model.InitiateRefundRequest) (*model.PendingCashRefund, error) {
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	expires := time.Now().Add(24 * time.Hour)
	ref := model.PendingCashRefund{
		ID:          uuid.New().String(),
		WalletID:    req.WalletID,
		Amount:      req.Amount,
		InitiatedBy: &staffID,
		ExpiresAt:   &expires,
	}
	if err := s.refundRepo.Create(ctx, ref); err != nil {
		return nil, err
	}
	return &ref, nil
}

func (s *AdminService) ListPendingRefunds(ctx context.Context) ([]model.PendingCashRefund, error) {
	return s.refundRepo.ListPending(ctx)
}

func (s *AdminService) ApproveRefund(ctx context.Context, req model.ApproveRefundRequest) error {
	ref, err := s.refundRepo.GetByID(ctx, req.RefundID)
	if err != nil || ref == nil {
		return errors.New("refund not found")
	}
	if ref.Status != "PENDING" {
		return errors.New("refund is no longer pending")
	}
	switch req.Action {
	case "APPROVE":
		if err := s.walletRepo.Refund(ctx, ref.WalletID, uuid.New().String(), ref.ID, ref.Amount); err != nil {
			return err
		}
		return s.refundRepo.UpdateStatus(ctx, ref.ID, "APPROVED")
	case "REJECT":
		return s.refundRepo.UpdateStatus(ctx, ref.ID, "REJECTED")
	default:
		return errors.New("action must be APPROVE or REJECT")
	}
}

func (s *AdminService) GetGpVat(ctx context.Context) (*model.GpVat, error) {
	return s.gpvatRepo.Get(ctx)
}

func (s *AdminService) UpdateGpVat(ctx context.Context, gv model.GpVat) error {
	if gv.GP < 0 || gv.VAT < 0 {
		return errors.New("gp and vat must be non-negative")
	}
	return s.gpvatRepo.Upsert(ctx, gv)
}

func (s *AdminService) GetDashboardSummary(ctx context.Context, start, end string) (*model.DashboardSummary, error) {
	query := `
        SELECT 
            (SELECT COALESCE(SUM(total_amount), 0) FROM sale_transactions WHERE DATE(created_at) BETWEEN ? AND ?) as total_sales_amount,
            (SELECT COUNT(id) FROM sale_transactions WHERE DATE(created_at) BETWEEN ? AND ?) as total_sales_count,
            (SELECT COALESCE(SUM(amount), 0) FROM wallet_transactions WHERE type = 'TOPUP_QR' AND DATE(created_at) BETWEEN ? AND ?) as qr_topup_amount,
            (SELECT COALESCE(SUM(amount), 0) FROM wallet_transactions WHERE type = 'TOPUP_CASH' AND DATE(created_at) BETWEEN ? AND ?) as cash_topup_amount,
            (SELECT COALESCE(SUM(amount), 0) FROM wallet_transactions WHERE type = 'REFUND' AND DATE(created_at) BETWEEN ? AND ?) as cash_refund_amount
    `

	summary := &model.DashboardSummary{}
	err := s.db.QueryRowContext(ctx, query,
		start, end, start, end, start, end, start, end, start, end,
	).Scan(
		&summary.TotalSalesAmount,
		&summary.TotalSalesCount,
		&summary.QrTopupAmount,
		&summary.CashTopupAmount,
		&summary.CashRefundAmount,
	)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func (s *AdminService) GetStallRevenueReport(ctx context.Context, start, end string) ([]model.StallRevenue, error) {
	query := `
        SELECT 
            fs.stall_name, 
            fs.id, 
            COALESCE(fs.category, 'Uncategorized'),
            COALESCE(SUM(st.total_amount), 0)
        FROM food_stalls fs
        LEFT JOIN sale_transactions st ON fs.id = st.stall_id AND DATE(st.created_at) BETWEEN ? AND ?
        GROUP BY fs.id, fs.stall_name, fs.category
        ORDER BY COALESCE(SUM(st.total_amount), 0) DESC
    `

	rows, err := s.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []model.StallRevenue
	for rows.Next() {
		var r model.StallRevenue
		err := rows.Scan(&r.ShopName, &r.ShopID, &r.ShopType, &r.PaymentAmt)
		if err != nil {
			return nil, err
		}

		r.GPAmount = r.PaymentAmt * 0.20
		r.VatAmount = r.PaymentAmt * 0.07
		r.NetIncome = r.PaymentAmt - r.GPAmount - r.VatAmount

		reports = append(reports, r)
	}
	return reports, nil
}

func (s *AdminService) GetRevenueReport() ([]model.StallRevenue, error) {
	query := `
        SELECT 
            fs.stall_name, 
            COALESCE(SUM(st.total_amount), 0) as payment_amt
        FROM food_stalls fs
        LEFT JOIN sale_transactions st ON fs.id = st.stall_id 
            AND DATE(st.created_at) = CURRENT_DATE
        WHERE fs.id != 'SYSTEM-TOPUP'
        GROUP BY fs.stall_name
    `
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []model.StallRevenue
	for rows.Next() {
		var r model.StallRevenue

		if err := rows.Scan(&r.ShopName, &r.PaymentAmt); err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, nil
}

func (s *AdminService) GetStaffWorkHistory(ctx context.Context, staffID string) (interface{}, error) {
	return s.walletRepo.GetTransactionsByStaff(ctx, staffID)
}

func (s *AdminService) UpdateStall(ctx context.Context, id string, req model.UpdateStallRequest) error {
	return s.adminRepo.UpdateStall(id, req)
}

func (s *AdminService) UpdateStallStatus(ctx context.Context, id string, status string) error {
	return s.adminRepo.UpdateStallStatus(ctx, id, status)
}

func (s *AdminService) GetStallByID(ctx context.Context, id string) (*model.FoodStall, error) {
	return s.adminRepo.GetStallByID(ctx, id)
}

func (s *AdminService) CreateStall(ctx context.Context, name, ownerName, phone, stallCategory string) (*model.FoodStall, error) {
	defaultPassword := "password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	vendorUserID := uuid.NewString()

	queryUser := `INSERT INTO users (id, full_name, role, status, password_hash) VALUES (?, ?, ?, ?, ?)`
	if _, err := s.db.ExecContext(ctx, queryUser, vendorUserID, ownerName, "VENDOR", "ACTIVE", string(hashedPassword)); err != nil {
		return nil, fmt.Errorf("failed to create vendor user: %v", err)
	}

	newStall := &model.FoodStall{
		ID:        uuid.NewString(),
		StallName: name,
		OwnerID:   vendorUserID,
		OwnerName: ownerName,
		Phone:     phone,
		Category:  &stallCategory,
		Status:    "ACTIVE",
	}

	if err := s.adminRepo.CreateStall(ctx, newStall); err != nil {
		return nil, fmt.Errorf("failed to create stall: %v", err)
	}

	return newStall, nil
}

func (s *AdminService) CreateStaff(ctx context.Context, id, name, phone, passwordHash string) error {
	query := `INSERT INTO users (id, full_name, phone, role, status, password_hash) 
              VALUES (?, ?, ?, 'STAFF', 'ACTIVE', ?)`

	_, err := s.db.ExecContext(ctx, query, id, name, phone, passwordHash)
	return err
}

func (s *AdminService) UpdateStaff(ctx context.Context, id, name, phone string) error {
	query := `UPDATE users SET full_name = ?, phone = ? WHERE id = ? AND role = 'STAFF'`
	_, err := s.db.ExecContext(ctx, query, name, phone, id)
	return err
}

func (s *AdminService) ChangeAdminPassword(ctx context.Context, id, oldPwd, newPwd string) error {
	var hashed string
	err := s.db.QueryRowContext(ctx, "SELECT password_hash FROM users WHERE id = ?", id).Scan(&hashed)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(oldPwd)); err != nil {
		return fmt.Errorf("รหัสผ่านเดิมไม่ถูกต้อง")
	}

	newHashed, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, "UPDATE users SET password_hash = ? WHERE id = ?", string(newHashed), id)
	return err
}
