package handler

import (
	"encoding/json"
	"net/http"

	"fmt"
	appMiddleware "pmfoodcourt/internal/middleware"
	"pmfoodcourt/internal/service"

	"github.com/go-chi/chi/v5"
)

type PaymentHandler struct {
	svc       *service.PaymentService
	jwtSecret string
}

func NewPaymentHandler(svc *service.PaymentService, jwtSecret string) *PaymentHandler {
	return &PaymentHandler{svc: svc, jwtSecret: jwtSecret}
}

func (h *PaymentHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.Auth(h.jwtSecret)) // 🚩 ถ้าไม่ใส่บรรทัดนี้ userID จะเป็นค่าว่างแล้ว Error 500 ทันที
		r.Post("/simulate", h.SimulatePayment)
		r.Get("/metrics", h.GetMetrics)
		r.Get("/history", h.GetHistory)
		r.Get("/activities", h.GetActivities)
	})

	return r
}

func (h *PaymentHandler) SimulatePayment(w http.ResponseWriter, r *http.Request) {
	// 🚩 ดึง userID จาก Context (ที่ Middleware ถอดรหัสมาจาก Token)
	userID := appMiddleware.GetUserID(r.Context())
	fmt.Printf("DEBUG: Incoming UserID from Token: [%s]\n", userID) // 🚩 ดูใน Terminal ว่าขึ้นค่าไหม

	if userID == "" {
		http.Error(w, "Unauthorized: No UserID in context", http.StatusUnauthorized)
		return
	}

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}

	// ส่ง userID เข้าไปแทน stallID
	err := h.svc.SimulateReceiveMoney(r.Context(), userID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]string{"status": "success"})
}

func (h *PaymentHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	userID := appMiddleware.GetUserID(r.Context())

	sales, orders, err := h.svc.GetMetricsByOwner(r.Context(), userID)
	if err != nil {
		// 🚩 ถ้า Error ต้อง return ทันทีเพื่อหยุดโค้ด
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 🚩 ส่งข้อมูลก้อนเดียวเท่านั้น!
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"today_sales":  sales,
		"today_orders": orders,
	})
	// ❌ ห้ามมีคำสั่ง jsonOK หรือ Encode อื่นต่อท้ายตรงนี้
}

func (h *PaymentHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := appMiddleware.GetUserID(r.Context())

	// 🚩 เรียก Service ที่เราเพิ่งเพิ่ม GetPastSalesSummary เข้าไป
	history, err := h.svc.GetPastSalesSummary(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 🚩 พ่น JSON ออกไปให้หน้า Past Sales ใน React
	jsonOK(w, history)
}

func (h *PaymentHandler) GetActivities(w http.ResponseWriter, r *http.Request) {
	userID := appMiddleware.GetUserID(r.Context())
	activities, err := h.svc.GetActivityHistory(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonOK(w, activities)
}
