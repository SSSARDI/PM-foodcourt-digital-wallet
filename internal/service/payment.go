package service

import (
	"context"
	"database/sql"
	"log"
	"pmfoodcourt/internal/repository"

	"github.com/google/uuid"
)

type PaymentService struct {
	db          *sql.DB
	paymentRepo *repository.PaymentRepository
	repo        *repository.PaymentRepository
}

func NewPaymentService(db *sql.DB, paymentRepo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{db: db, paymentRepo: paymentRepo}
}

// SimulateReceiveMoney: ฟังก์ชันโกงเงินเข้าและเพิ่มยอดขายแบบทันใจ
func (s *PaymentService) SimulateReceiveMoney(ctx context.Context, userID string, amount float64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 🚩 1. หา stall_id
	var stallID string
	err = tx.QueryRowContext(ctx, "SELECT id FROM food_stalls WHERE owner_id = ? LIMIT 1", userID).Scan(&stallID)
	if err != nil {
		log.Printf("❌ ERROR 1 (StallID): %v for UserID: %s", err, userID)
		return err
	}

	// 🚩 2. หา wallet_id
	var walletID string
	err = tx.QueryRowContext(ctx, "SELECT id FROM wallets WHERE user_id = ? LIMIT 1", userID).Scan(&walletID)
	if err != nil {
		log.Printf("❌ ERROR 2 (WalletID): %v for UserID: %s", err, userID)
		return err
	}

	// 🚩 3. อัปเดตเงินใน Wallet
	_, err = tx.ExecContext(ctx, "UPDATE wallets SET balance = balance + ? WHERE id = ?", amount, walletID)
	if err != nil {
		log.Printf("❌ ERROR 3 (Update Wallet): %v", err)
		return err
	}

	// 🚩 4. สร้างรายการใน wallet_transactions (ตัวแม่)
	walletTxID := uuid.New().String()

	walletSQL := `
        INSERT INTO wallet_transactions (id, wallet_id, type, amount, created_at) 
        VALUES (?, ?, 'PAYMENT', ?, NOW())`

	if _, err := tx.ExecContext(ctx, walletSQL, walletTxID, walletID, amount); err != nil {
		log.Printf("❌ ERROR 4 (Insert WalletTx): %v", err)
		return err
	}

	// 🚩 5. บันทึกยอดขาย (sale_transactions - ตัวลูก)
	saleSQL := `
        INSERT INTO sale_transactions 
        (id, stall_id, customer_id, total_amount, payment_transaction_id, created_at) 
        VALUES (?, ?, ?, ?, ?, NOW())`

	if _, err := tx.ExecContext(ctx, saleSQL, uuid.New().String(), stallID, userID, amount, walletTxID); err != nil {
		log.Printf("❌ ERROR 5 (Insert Sale): %v", err)
		return err
	}

	log.Printf("✅ SUCCESS: Payment simulated for User: %s, Amount: %f", userID, amount)
	return tx.Commit()
}

func (s *PaymentService) GetMetricsByOwner(ctx context.Context, userID string) (float64, int, error) {
	var stallID string
	// หา StallID ก่อน
	err := s.db.QueryRowContext(ctx, "SELECT id FROM food_stalls WHERE owner_id = ? LIMIT 1", userID).Scan(&stallID)
	if err != nil {
		return 0, 0, err
	}

	// เรียกใช้ฟังก์ชัน Repo ที่สร้างมะกี้
	return s.paymentRepo.GetMetrics(ctx, stallID)
}

func (s *PaymentService) GetPastSalesSummary(ctx context.Context, userID string) ([]repository.SalesSummary, error) {
	var stallID string
	// 🚩 1. หา StallID ของป้าจงก่อน
	err := s.db.QueryRowContext(ctx, "SELECT id FROM food_stalls WHERE owner_id = ? LIMIT 1", userID).Scan(&stallID)
	if err != nil {
		log.Printf("❌ ERROR: GetPastSalesSummary - Stall not found for User: %s", userID)
		return nil, err
	}

	// 🚩 2. เรียกใช้ Repo ตัวใหม่ที่เราเพิ่งเพิ่ม GetSalesHistorySummary เข้าไป
	return s.paymentRepo.GetSalesHistorySummary(ctx, stallID)
}

func (s *PaymentService) GetActivityHistory(ctx context.Context, userID string) ([]repository.ActivityDetail, error) {
	var stallID string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM food_stalls WHERE owner_id = ? LIMIT 1", userID).Scan(&stallID)
	if err != nil {
		return nil, err
	}

	// 🚩 เปลี่ยนจาก s.repo เป็น s.paymentRepo ครับ!
	return s.paymentRepo.GetActivityHistory(ctx, stallID)
}
