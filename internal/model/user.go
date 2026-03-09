package model

import "time"

type User struct {
	ID           string    `json:"id"`
	FullName     string    `json:"full_name"`
	Email        *string   `json:"email,omitempty"`
	Phone        *string   `json:"phone,omitempty"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type Wallet struct {
	ID        string    `json:"id"`
	PublicID  string    `json:"public_id"`
	UserID    string    `json:"user_id"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type WalletTransaction struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	ReferenceID *string   `json:"reference_id,omitempty"`
	CreatedBy   *string   `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type TopUpCashRequest struct {
	WalletPublicID string  `json:"wallet_public_id"`
	Amount         float64 `json:"amount"`
}

type PayStallRequest struct {
	StallID string  `json:"stall_id"`
	Amount  float64 `json:"amount"`
}

type PayQRRequest struct {
	QRToken string `json:"qr_token"`
}

type FoodStall struct {
	ID        string    `json:"id"`
	StallName string    `json:"stall_name"`
	OwnerID   string    `json:"owner_id"`
	Category  *string   `json:"category,omitempty"`
	Location  *string   `json:"location,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type SaleTransaction struct {
	ID                   string    `json:"id"`
	StallID              string    `json:"stall_id"`
	CustomerID           string    `json:"customer_id"`
	TotalAmount          float64   `json:"total_amount"`
	PaymentTransactionID string    `json:"payment_transaction_id"`
	CreatedAt            time.Time `json:"created_at"`
}

type CreateStallRequest struct {
	StallName string  `json:"stall_name"`
	Category  *string `json:"category"`
	Location  *string `json:"location"`
}

type UpdateStallRequest struct {
	StallName *string `json:"stall_name"`
	Category  *string `json:"category"`
	Location  *string `json:"location"`
	Status    *string `json:"status"`
}

type QRPayment struct {
	ID        string     `json:"id"`
	StallID   string     `json:"stall_id"`
	QRToken   *string    `json:"qr_token,omitempty"`
	Amount    *float64   `json:"amount,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateQRRequest struct {
	StallID string  `json:"stall_id"`
	Amount  float64 `json:"amount"`
}

type PendingCashRefund struct {
	ID          string     `json:"id"`
	WalletID    string     `json:"wallet_id"`
	Amount      float64    `json:"amount"`
	Status      string     `json:"status"`
	InitiatedBy *string    `json:"initiated_by,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type InitiateRefundRequest struct {
	WalletID string  `json:"wallet_id"`
	Amount   float64 `json:"amount"`
}

type ApproveRefundRequest struct {
	RefundID string `json:"refund_id"`
	Action   string `json:"action"`
}

type GpVat struct {
	GP  float64 `json:"gp"`
	VAT float64 `json:"vat"`
}