package service

import (
	"context"
	"errors"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/repository"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo       *repository.UserRepository
	walletRepo *repository.WalletRepository
}

func NewUserService(repo *repository.UserRepository, walletRepo *repository.WalletRepository) *UserService {
	return &UserService{repo: repo, walletRepo: walletRepo}
}

func (s *UserService) GetAll(ctx context.Context) ([]model.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (s *UserService) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	// 1. ดึงข้อมูล User จากเบอร์โทร
	user, err := s.repo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 🚩 2. จุดสำคัญ: ต้องไปดึง Balance มาใส่ใน User model ด้วย
	// ต้องมั่นใจว่าใน UserService struct มี walletRepo ให้เรียกใช้นะครับ
	balance, err := s.walletRepo.GetBalance(ctx, user.ID)
	if err != nil {
		// ถ้าดึง balance ไม่ได้ อาจจะให้เป็น 0 ไปก่อน
		user.Balance = 0
	} else {
		user.Balance = balance
	}

	return user, nil
}

func (s *UserService) Create(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 🚩 วิธีแก้แดงบรรทัด 57-58: ใช้ & นำหน้าเพื่อเปลี่ยน string เป็น *string
	u := model.User{
		ID:           uuid.New().String(),
		FullName:     req.FullName,
		Email:        &req.Email, // ✅ เติม & เข้าไป
		Phone:        &req.Phone, // ✅ เติม & เข้าไป
		PasswordHash: string(hash),
		Role:         req.Role,
		Status:       "ACTIVE",
		CreatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) ChangePassword(ctx context.Context, userID string, req model.ChangePasswordRequest) error {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil || u == nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.OldPassword)); err != nil {
		return errors.New("รหัสผ่านเดิมไม่ถูกต้อง")
	}

	newHash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	return s.repo.UpdatePassword(ctx, userID, string(newHash))
}
