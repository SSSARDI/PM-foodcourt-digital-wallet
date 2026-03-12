package repository

import (
	"context"
	"database/sql"
)

type PaymentRepository struct {
	db *sql.DB
}

type SalesSummary struct {
	SaleDate    string  `json:"sale_date"`
	OrderCount  int     `json:"order_count"`
	TotalAmount float64 `json:"total_amount"`
}

type ActivityDetail struct {
	ID          string  `json:"id"`
	TotalAmount float64 `json:"total_amount"`
	CreatedAt   string  `json:"created_at"`
}

func NewPaymentRepo(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) GetMerchantMetrics(ctx context.Context, stallID string) (float64, int, error) {
	var totalSales float64
	var totalOrders int

	// SQL ตัวที่ฟลุ๊ครันใน MySQL แล้วผ่านนั่นแหละครับ
	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0), 
			COUNT(id) 
		FROM sale_transactions 
		WHERE stall_id = ?`

	err := r.db.QueryRowContext(ctx, query, stallID).Scan(&totalSales, &totalOrders)
	return totalSales, totalOrders, err
}

func (r *PaymentRepository) GetMetrics(ctx context.Context, stallID string) (float64, int, error) {
	var sales float64
	var orders int
	query := `
        SELECT COALESCE(SUM(total_amount), 0), COUNT(id) 
        FROM sale_transactions 
        WHERE stall_id = ? AND DATE(created_at) = CURDATE()`
	err := r.db.QueryRowContext(ctx, query, stallID).Scan(&sales, &orders)
	return sales, orders, err
}

func (r *PaymentRepository) GetSalesHistorySummary(ctx context.Context, stallID string) ([]SalesSummary, error) {
	// ใช้ GROUP BY DATE เพื่อรวมยอดเป็นรายวัน
	query := `
        SELECT 
            DATE(created_at) as sale_date,
            COUNT(id) as order_count,
            COALESCE(SUM(total_amount), 0) as total_amount
        FROM sale_transactions
        WHERE stall_id = ?
        GROUP BY DATE(created_at)
        ORDER BY sale_date DESC`

	rows, err := r.db.QueryContext(ctx, query, stallID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []SalesSummary
	for rows.Next() {
		var s SalesSummary
		if err := rows.Scan(&s.SaleDate, &s.OrderCount, &s.TotalAmount); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

func (r *PaymentRepository) GetActivityHistory(ctx context.Context, stallID string) ([]ActivityDetail, error) {
	// ดึงแค่ข้อมูลการขายล้วนๆ เรียงจากใหม่ไปเก่า
	query := `
        SELECT id, total_amount, created_at
        FROM sale_transactions
        WHERE stall_id = ?
        ORDER BY created_at DESC
        LIMIT 50`

	rows, err := r.db.QueryContext(ctx, query, stallID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []ActivityDetail
	for rows.Next() {
		var a ActivityDetail
		if err := rows.Scan(&a.ID, &a.TotalAmount, &a.CreatedAt); err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}
