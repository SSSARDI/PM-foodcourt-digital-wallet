package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	appMiddleware "pmfoodcourt/internal/middleware"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/service"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	svc       *service.AuthService
	jwtSecret string
}

func NewAuthHandler(svc *service.AuthService, jwtSecret string) *AuthHandler {
	return &AuthHandler{svc, jwtSecret}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.Post("/forgot-password", h.forgotPassword)
	r.Post("/reset-password", h.resetPassword)

	r.Group(func(r chi.Router) {
		// 🚩 ใช้ h.jwtSecret ที่เราเก็บไว้ใน struct
		r.Use(appMiddleware.Auth(h.jwtSecret))
		r.Get("/me", h.getMe)
	})

	return r
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.FullName == "" || req.Password == "" || (req.Email == "" && req.Phone == "") {
		jsonError(w, http.StatusBadRequest, "full_name, password, and email or phone are required")
		return
	}
	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) || errors.Is(err, service.ErrPhoneTaken) {
			jsonError(w, http.StatusConflict, err.Error())
			return
		}
		jsonError(w, http.StatusInternalServerError, "registration failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCreds) {
			jsonError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if errors.Is(err, service.ErrUserSuspended) {
			jsonError(w, http.StatusForbidden, err.Error())
			return
		}
		jsonError(w, http.StatusInternalServerError, "login failed")
		return
	}
	jsonOK(w, resp)
}

func (h *AuthHandler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var req model.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	token, err := h.svc.ForgotPassword(r.Context(), req.Email)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed")
		return
	}
	// production: ส่ง token ทาง email แทน — ตอนนี้ return มาใน dev
	jsonOK(w, map[string]string{"message": "if email exists, reset link sent", "dev_token": token})
}

func (h *AuthHandler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req model.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.svc.ResetPassword(r.Context(), req); err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		jsonError(w, http.StatusInternalServerError, "reset failed")
		return
	}
	jsonOK(w, map[string]string{"message": "password updated"})
}

func (h *AuthHandler) getMe(w http.ResponseWriter, r *http.Request) {
	userID := appMiddleware.GetUserID(r.Context())
	// ไปดึงข้อมูล User จาก Service (ที่รวมยอดเงิน balance มาแล้ว)
	user, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		jsonError(w, http.StatusNotFound, "user not found")
		return
	}
	jsonOK(w, map[string]interface{}{"user": user})
}
