package service

import (
	"context"
	"errors"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/repository"
	"time"

	"github.com/google/uuid"
)

type AdminService struct {
	userRepo   *repository.UserRepository
	walletRepo *repository.WalletRepository
	refundRepo *repository.RefundRepository
	gpvatRepo  *repository.GpVatRepository
}

func NewAdminService(
	ur *repository.UserRepository,
	wr *repository.WalletRepository,
	rr *repository.RefundRepository,
	gv *repository.GpVatRepository,
) *AdminService {
	return &AdminService{ur, wr, rr, gv}
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
