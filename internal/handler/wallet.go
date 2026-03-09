package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appMiddleware "pmfoodcourt/internal/middleware"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/service"
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

	r.Get("/me",           h.getMyWallet)
	r.Get("/transactions", h.getTransactions)

	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.RequireRole("CUSTOMER"))
		r.Post("/topup/qr", h.payQR)
		r.Post("/pay",      h.payStall)
	})
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.RequireRole("STAFF", "ADMIN"))
		r.Post("/topup/cash", h.topUpCash)
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
	staffID := appMiddleware.GetUserID(r.Context())
	var req model.TopUpCashRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	wallet, err := h.svc.TopUpCash(r.Context(), staffID, req)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, wallet)
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
