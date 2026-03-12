package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	appMiddleware "pmfoodcourt/internal/middleware"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/service"

	"github.com/go-chi/chi/v5"
)

type WalletHandler struct {
	svc       *service.WalletService
	jwtSecret string
}

func NewWalletHandler(svc *service.WalletService, jwtSecret string) *WalletHandler {
	return &WalletHandler{svc, jwtSecret}
}

func (h *WalletHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(appMiddleware.Auth(h.jwtSecret))

	r.Get("/me", h.getMyWallet)
	r.Get("/transactions", h.getTransactions)
	r.Post("/refund/bypass", h.bypassRefund)
	r.Post("/refund/initiate", h.initiateRefund)

	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.RequireRole("CUSTOMER"))
		r.Post("/topup/qr", h.payQR)
		r.Post("/pay", h.payStall)
	})
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.RequireRole("STAFF", "ADMIN"))
		r.Post("/topup/cash", h.topUpCash)
		r.Post("/refund/cash", h.refundCash)
	})
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.RequireRole("VENDOR", "ADMIN"))
		r.Post("/qr", h.createQR)
	})

	return r
}

func (h *WalletHandler) getMyWallet(w http.ResponseWriter, r *http.Request) {
	userID := appMiddleware.GetUserID(r.Context())
	wallet, err := h.svc.GetMyWallet(r.Context(), userID)
	if err != nil {
		jsonError(w, http.StatusNotFound, err.Error())
		return
	}
	jsonOK(w, wallet)
}

func (h *WalletHandler) getTransactions(w http.ResponseWriter, r *http.Request) {
	userID := appMiddleware.GetUserID(r.Context())
	txns, err := h.svc.GetTransactions(r.Context(), userID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, txns)
}

func (h *WalletHandler) topUpCash(w http.ResponseWriter, r *http.Request) {
	// 🚩 ดึง ID พนักงานจริงๆ จาก Context (Middleware ทำไว้ให้แล้ว)
	staffID := appMiddleware.GetUserID(r.Context())

	var req struct {
		Phone  string  `json:"phone"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	// 🚩 ส่ง staffID เข้าไปด้วย
	err := h.svc.TopupByStaff(r.Context(), req.Phone, req.Amount, staffID)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, map[string]string{"message": "success"})
}

func (h *WalletHandler) payQR(w http.ResponseWriter, r *http.Request) {
	userID := appMiddleware.GetUserID(r.Context())
	var req model.PayQRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	wallet, err := h.svc.PayQR(r.Context(), userID, req)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, wallet)
}

func (h *WalletHandler) payStall(w http.ResponseWriter, r *http.Request) {
	customerID := appMiddleware.GetUserID(r.Context())
	var req model.PayStallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	sale, err := h.svc.PayStall(r.Context(), customerID, req)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, sale)
}

func (h *WalletHandler) createQR(w http.ResponseWriter, r *http.Request) {
	var req model.CreateQRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	qr, err := h.svc.CreateQR(r.Context(), req)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(qr)
}

func (h *WalletHandler) refundCash(w http.ResponseWriter, r *http.Request) {
	staffID := appMiddleware.GetUserID(r.Context())
	var req struct {
		Phone  string  `json:"phone"`
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "ข้อมูล Request ไม่ถูกต้อง")
		return
	}

	// 🚩 เรียกใช้ Service ที่เราทำไว้
	err := h.svc.RefundByStaff(r.Context(), req.Phone, req.Amount, staffID)
	if err != nil {
		// พิมพ์ Error ออกมาดูใน Terminal ของ Go
		fmt.Printf("❌ Refund Error: %v\n", err)
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonOK(w, map[string]string{"message": "Refund successful"})
}

func (h *WalletHandler) bypassRefund(w http.ResponseWriter, r *http.Request) {
	userID := appMiddleware.GetUserID(r.Context())
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// เรียก Service ไป Bypass (หักเงินทันที)
	err := h.svc.BypassRefund(r.Context(), userID, req.Amount)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, map[string]string{"message": "Bypass successful"})
}

func (h *WalletHandler) initiateRefund(w http.ResponseWriter, r *http.Request) {
	userID := appMiddleware.GetUserID(r.Context())
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request")
		return
	}

	fmt.Printf("📢 User %s ส่งคำร้องขอคืนเงินจำนวน %.2f (No Action Taken)\n", userID, req.Amount)

	jsonOK(w, map[string]string{
		"message": "ส่งคำร้องสำเร็จ! โปรดรับเงินสดที่เคาน์เตอร์",
	})
}
