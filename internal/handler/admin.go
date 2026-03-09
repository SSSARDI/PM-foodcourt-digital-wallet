package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appMiddleware "pmfoodcourt/internal/middleware"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/service"
)

type AdminHandler struct {
	svc       *service.AdminService
	jwtSecret string
}

func NewAdminHandler(svc *service.AdminService, jwtSecret string) *AdminHandler {
	return &AdminHandler{svc, jwtSecret}
}

func (h *AdminHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(appMiddleware.Auth(h.jwtSecret))

	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.RequireRole("STAFF", "ADMIN"))
		r.Post("/refunds/initiate", h.initiateRefund)
	})
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.RequireRole("ADMIN"))
		r.Get("/users",                 h.listUsers)
		r.Patch("/users/{id}/suspend",  h.suspendUser)
		r.Patch("/users/{id}/activate", h.activateUser)
		r.Get("/refunds",               h.listRefunds)
		r.Post("/refunds/approve",      h.approveRefund)
		r.Get("/gp-vat",                h.getGpVat)
		r.Put("/gp-vat",                h.updateGpVat)
	})

	return r
}

func (h *AdminHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	if role == "" {
		role = "CUSTOMER"
	}
	users, err := h.svc.ListUsers(r.Context(), role)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, users)
}

func (h *AdminHandler) suspendUser(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.SuspendUser(r.Context(), chi.URLParam(r, "id")); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, map[string]string{"status": "SUSPENDED"})
}

func (h *AdminHandler) activateUser(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.ActivateUser(r.Context(), chi.URLParam(r, "id")); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, map[string]string{"status": "ACTIVE"})
}

func (h *AdminHandler) initiateRefund(w http.ResponseWriter, r *http.Request) {
	staffID := appMiddleware.GetUserID(r.Context())
	var req model.InitiateRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	ref, err := h.svc.InitiateRefund(r.Context(), staffID, req)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ref)
}

func (h *AdminHandler) listRefunds(w http.ResponseWriter, r *http.Request) {
	refs, err := h.svc.ListPendingRefunds(r.Context())
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, refs)
}

func (h *AdminHandler) approveRefund(w http.ResponseWriter, r *http.Request) {
	var req model.ApproveRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.svc.ApproveRefund(r.Context(), req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, map[string]string{"message": "done"})
}

func (h *AdminHandler) getGpVat(w http.ResponseWriter, r *http.Request) {
	gv, err := h.svc.GetGpVat(r.Context())
	if err != nil || gv == nil {
		jsonError(w, http.StatusNotFound, "gp_vat not configured")
		return
	}
	jsonOK(w, gv)
}

func (h *AdminHandler) updateGpVat(w http.ResponseWriter, r *http.Request) {
	var gv model.GpVat
	if err := json.NewDecoder(r.Body).Decode(&gv); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.svc.UpdateGpVat(r.Context(), gv); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, map[string]string{"message": "updated"})
}
