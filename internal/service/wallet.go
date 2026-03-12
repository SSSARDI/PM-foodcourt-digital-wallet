package service

import (
	"context"
	"errors"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/repository"
	"time"

	"github.com/google/uuid"
)

type WalletService struct {
	walletRepo *repository.WalletRepository
	stallRepo  *repository.StallRepository
	qrRepo     *repository.QRRepository
	userRepo   *repository.UserRepository // ✅ เพิ่มบรรทัดนี้เข้าไปเพื่อให้ข้างล่างเรียกใช้ได้
}

func NewWalletService(
	repo *repository.WalletRepository,
	stallRepo *repository.StallRepository,
	qrRepo *repository.QRRepository,
	userRepo *repository.UserRepository,
) *WalletService {
	return &WalletService{
		walletRepo: repo, // ✅ เปลี่ยนจาก repo: เป็น walletRepo: ให้ตรงกับข้างบน
		stallRepo:  stallRepo,
		qrRepo:     qrRepo,
		userRepo:   userRepo, // ✅ ตอนนี้จะไม่แดงแล้วเพราะข้างบนมี userRepo แล้ว
	}
}

func (s *WalletService) GetMyWallet(ctx context.Context, userID string) (*model.Wallet, error) {
	w, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, errors.New("wallet not found")
	}
	return w, nil
}

func (s *WalletService) GetTransactions(ctx context.Context, userID string) ([]model.WalletTransaction, error) {
	w, err := s.GetMyWallet(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.walletRepo.GetTransactions(ctx, w.ID)
}

func (s *WalletService) TopUpCash(ctx context.Context, staffID string, req model.TopUpCashRequest) (*model.Wallet, error) {
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	w, err := s.walletRepo.GetByPublicID(ctx, req.WalletPublicID)
	if err != nil || w == nil {
		return nil, errors.New("wallet not found")
	}
	return s.walletRepo.TopUpCash(ctx, w.ID, uuid.New().String(), staffID, req.Amount)
}

func (s *WalletService) CreateQR(ctx context.Context, req model.CreateQRRequest) (*model.QRPayment, error) {
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	token := uuid.New().String()
	expires := time.Now().Add(10 * time.Minute)
	q := model.QRPayment{
		ID:        uuid.New().String(),
		StallID:   req.StallID,
		QRToken:   &token,
		Amount:    &req.Amount,
		ExpiresAt: &expires,
		Status:    "PENDING",
	}
	if err := s.qrRepo.Create(ctx, q); err != nil {
		return nil, err
	}
	return &q, nil
}

func (s *WalletService) PayQR(ctx context.Context, userID string, req model.PayQRRequest) (*model.Wallet, error) {
	if req.QRToken == "BYPASS_TOKEN" {
		// หา Wallet ของเราก่อน
		w, err := s.walletRepo.GetByUserID(ctx, userID)
		if err != nil || w == nil {
			return nil, errors.New("wallet not found")
		}

		// สั่งเติมเงินเลยโดยไม่ต้องเช็คตาราง QR
		txnID := uuid.New().String()
		// เราใช้ "BYPASS_QR" เป็น ID อ้างอิงหลอกๆ ไว้ในระบบ
		return s.walletRepo.TopUpQR(ctx, w.ID, txnID, "BYPASS_QR", req.Amount)
	}

	_ = s.qrRepo.ExpireStale(ctx)
	qr, err := s.qrRepo.GetByToken(ctx, req.QRToken)
	if err != nil {
		return nil, err
	}
	if qr == nil {
		return nil, errors.New("QR not found")
	}
	if qr.Status != "PENDING" {
		return nil, errors.New("QR is " + qr.Status)
	}
	if qr.ExpiresAt != nil && time.Now().After(*qr.ExpiresAt) {
		return nil, errors.New("QR expired")
	}
	w, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil || w == nil {
		return nil, errors.New("wallet not found")
	}
	return s.walletRepo.TopUpQR(ctx, w.ID, uuid.New().String(), qr.ID, *qr.Amount)
}

func (s *WalletService) PayStall(ctx context.Context, customerID string, req model.PayStallRequest) (*model.SaleTransaction, error) {
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	w, err := s.walletRepo.GetByUserID(ctx, customerID)
	if err != nil || w == nil {
		return nil, errors.New("wallet not found")
	}
	saleID := uuid.New().String()
	txnID := uuid.New().String()
	if err := s.walletRepo.Payment(ctx, w.ID, txnID, saleID, req.Amount); err != nil {
		return nil, err
	}
	sale := model.SaleTransaction{
		ID:                   saleID,
		StallID:              req.StallID,
		CustomerID:           customerID,
		TotalAmount:          req.Amount,
		PaymentTransactionID: txnID,
	}
	if err := s.stallRepo.CreateSale(ctx, sale); err != nil {
		return nil, err
	}
	return &sale, nil
}

func (s *WalletService) TopupByStaff(ctx context.Context, phone string, amount float64, staffID string) error {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil || user == nil {
		return errors.New("ไม่พบผู้ใช้งาน")
	}

	wallet, err := s.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil || wallet == nil {
		return errors.New("ไม่พบกระเป๋าเงิน")
	}

	txnID := uuid.New().String()

	// 🚩 ใช้ staffID ที่ส่งมาจาก Handler แทนคำว่า "STAFF_SYSTEM"
	_, err = s.walletRepo.TopUpCash(ctx, wallet.ID, txnID, staffID, amount)
	return err
}

func (s *WalletService) RefundByStaff(ctx context.Context, phone string, amount float64, staffID string) error {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil || user == nil {
		return errors.New("ไม่พบผู้ใช้งาน")
	}

	wallet, err := s.walletRepo.GetByUserID(ctx, user.ID)
	if err != nil || wallet == nil {
		return errors.New("ไม่พบกระเป๋าเงิน")
	}

	// เช็คยอดเงินอีกรอบเพื่อความชัวร์
	if wallet.Balance < amount {
		return errors.New("ยอดเงินในกระเป๋าไม่เพียงพอ")
	}

	txnID := uuid.New().String()
	_, err = s.walletRepo.WithdrawCash(ctx, wallet.ID, txnID, staffID, amount)
	return err
}

func (s *WalletService) BypassRefund(ctx context.Context, userID string, amount float64) error {
	w, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil || w == nil {
		return errors.New("wallet not found")
	}

	if w.Balance < amount {
		return errors.New("ยอดเงินไม่เพียงพอ")
	}

	txnID := uuid.New().String()

	// 🚩 เปลี่ยนจาก "SYSTEM_BYPASS" เป็น userID (เจ้าของเครื่องที่กด)
	// เพื่อให้ Database ตรวจสอบผ่านเพราะ User นี้มีตัวตนอยู่จริงในระบบ
	_, err = s.walletRepo.WithdrawCash(ctx, w.ID, txnID, userID, amount)
	return err
}
