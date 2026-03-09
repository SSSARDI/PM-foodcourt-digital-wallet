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
}

func NewWalletService(wr *repository.WalletRepository, sr *repository.StallRepository, qr *repository.QRRepository) *WalletService {
	return &WalletService{wr, sr, qr}
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
