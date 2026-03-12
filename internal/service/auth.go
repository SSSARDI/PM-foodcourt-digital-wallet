package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken    = errors.New("email already registered")
	ErrPhoneTaken    = errors.New("phone already registered")
	ErrInvalidCreds  = errors.New("invalid credentials")
	ErrUserSuspended = errors.New("account suspended")
	ErrInvalidToken  = errors.New("invalid or expired token")
)

type AuthService struct {
	repo       *repository.AuthRepository
	walletRepo *repository.WalletRepository
	jwtSecret  string
	userRepo   *repository.UserRepository   // 🚩 เพิ่มบรรทัดนี้ (ต้องมีดอกจันถ้าใน main ใช้ NewUserRepo)
}

func NewAuthService(
    repo *repository.AuthRepository, 
    userRepo *repository.UserRepository, 
    walletRepo *repository.WalletRepository, 
    secret string,
) *AuthService {
    return &AuthService{
        repo:       repo,
        userRepo:   userRepo,
        walletRepo: walletRepo,
        jwtSecret:  secret, 
    }
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	role := "CUSTOMER"
	if req.Role == "VENDOR" || req.Role == "STAFF" || req.Role == "ADMIN" {
		role = req.Role
	}
	if req.Email != "" {
		if ex, _ := s.repo.GetByEmail(ctx, req.Email); ex != nil {
			return nil, ErrEmailTaken
		}
	}
	if req.Phone != "" {
		if ex, _ := s.repo.GetByPhone(ctx, req.Phone); ex != nil {
			return nil, ErrPhoneTaken
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	var emailPtr, phonePtr *string
	if req.Email != "" {
		emailPtr = &req.Email
	}
	if req.Phone != "" {
		phonePtr = &req.Phone
	}
	userID := uuid.New().String()
	u := model.User{
		ID:           userID,
		FullName:     req.FullName,
		Email:        emailPtr,
		Phone:        phonePtr,
		PasswordHash: string(hash),
		Role:         role,
		Status:       "ACTIVE",
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	if role == "CUSTOMER" || role == "VENDOR" {
		if err := s.walletRepo.Create(ctx, model.Wallet{
			ID:       uuid.New().String(),
			PublicID: uuid.New().String(),
			UserID:   userID,
		}); err != nil {
			return nil, err
		}
	}
	created, err := s.repo.GetByID(ctx, userID)
	if err != nil || created == nil {
		return nil, errors.New("failed to fetch created user")
	}
	token, err := s.makeToken(created)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{Token: token, User: *created}, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	var (
		u   *model.User
		err error
	)
	switch {
	case req.Email != "":
		u, err = s.repo.GetByEmail(ctx, req.Email)
	case req.Phone != "":
		u, err = s.repo.GetByPhone(ctx, req.Phone)
	default:
		return nil, errors.New("email or phone required")
	}
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrInvalidCreds
	}
	if u.Status == "SUSPENDED" {
		return nil, ErrUserSuspended
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCreds
	}
	token, err := s.makeToken(u)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{Token: token, User: *u}, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) (string, error) {
	u, _ := s.repo.GetByEmail(ctx, email)
	if u == nil {
		return "", nil
	}
	raw := make([]byte, 32)
	rand.Read(raw)
	token := hex.EncodeToString(raw)
	expires := time.Now().Add(1 * time.Hour)
	if err := s.repo.CreateResetToken(ctx, uuid.New().String(), u.ID, token, expires); err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req model.ResetPasswordRequest) error {
	userID, used, err := s.repo.GetResetToken(ctx, req.Token)
	if err != nil {
		return err
	}
	if userID == "" || used {
		return ErrInvalidToken
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return err
	}
	return s.repo.MarkResetTokenUsed(ctx, req.Token)
}

type jwtClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) makeToken(u *model.User) (string, error) {
	claims := jwtClaims{
		UserID: u.ID,
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) GetByID(ctx context.Context, id string) (*model.User, error) {
    // เรียกใช้ userRepo ที่เราเพิ่งเพิ่มเข้าไปใน NewAuthService เมื่อกี้
    u, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if u == nil {
        return nil, errors.New("user not found")
    }

    // 🚩 ถ้าต้องการดึงยอดเงิน balance มาโชว์ด้วย (ตามคอมเมนต์ในรูป)
    // ให้เรียก walletRepo มาดึง balance ต่อตรงนี้ได้เลยครับ
    balance, _ := s.walletRepo.GetBalance(ctx, u.ID)
    u.Balance = balance // มั่นใจว่าใน model.User มี field นี้

    return u, nil
}