package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	appMiddleware "pmfoodcourt/internal/middleware"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/service"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	svc       *service.AdminService
	jwtSecret string
}

func NewAdminHandler(svc *service.AdminService, jwtSecret string) *AdminHandler {
	return &AdminHandler{svc: svc, jwtSecret: jwtSecret}
}

func (h *AdminHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(appMiddleware.Auth(h.jwtSecret))

	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.RequireRole("STAFF", "ADMIN"))
		r.Post("/refunds/initiate", h.initiateRefund)
		r.Get("/staff-history", h.getStaffHistory)
	})

	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.RequireRole("ADMIN"))
		r.Get("/summary", h.getDashboardSummary)
		r.Get("/users", h.listUsers)
		r.Patch("/users/{id}/suspend", h.suspendUser)
		r.Patch("/users/{id}/activate", h.activateUser)
		r.Get("/refunds", h.listRefunds)
		r.Post("/refunds/approve", h.approveRefund)
		r.Get("/gp-vat", h.getGpVat)
		r.Put("/gp-vat", h.updateGpVat)
		r.Get("/reports/revenue", h.getStallRevenueReport)
		r.Put("/stalls/{id}", h.updateStall)
		r.Patch("/stalls/{id}/status", h.updateStallStatus)
		r.Get("/stalls/{id}", h.getStallByID)
		r.Post("/stalls", h.createStall)
		r.Post("/staff", h.createStaff)
		r.Put("/staff/{id}", h.updateStaff)
		r.Put("/change-password", h.changePassword)

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

func (h *AdminHandler) getDashboardSummary(w http.ResponseWriter, r *http.Request) {
	// 🚩 ดึงวันที่จาก URL (?start=...&end=...)
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	// ถ้าหน้าบ้านไม่ส่งมา ให้ Default เป็นวันนี้
	if start == "" {
		start = time.Now().Format("2006-01-02")
	}
	if end == "" {
		end = start
	}

	// 🚩 ส่ง start และ end เข้าไปใน Service (ต้องแก้ Parameter ที่ Service ด้วยตามที่คุยกัน)
	summary, err := h.svc.GetDashboardSummary(r.Context(), start, end)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, summary)
}

func (h *AdminHandler) getStallRevenueReport(w http.ResponseWriter, r *http.Request) {
	// 🚩 ดึงวันที่เหมือนกัน
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	if start == "" {
		start = time.Now().Format("2006-01-02")
	}
	if end == "" {
		end = start
	}

	// 🚩 เปลี่ยนไปเรียก GetStallRevenueReport และส่ง Context + วันที่เข้าไป
	reports, err := h.svc.GetStallRevenueReport(r.Context(), start, end)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, reports)
}

// ── Helpers ──

func jsonError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *AdminHandler) getStaffHistory(w http.ResponseWriter, r *http.Request) {
	staffID := appMiddleware.GetUserID(r.Context())

	history, err := h.svc.GetStaffWorkHistory(r.Context(), staffID)
	if err != nil {
		// 🚩 ถ้าพังตรงนี้ ให้ Print ดูใน Terminal ว่าพังเพราะอะไร
		fmt.Println("❌ AdminHandler Error:", err)
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonOK(w, history)
}

func (h *AdminHandler) updateStall(w http.ResponseWriter, r *http.Request) {
	stallID := chi.URLParam(r, "id")
	var req model.UpdateStallRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 🚩 เพิ่ม Status: ปกติถ้าแก้ไขร้านควรส่ง Status "ACTIVE" ไปด้วยถ้าไม่ได้เลือกปิดร้าน
	active := "ACTIVE"
	if req.Status == nil {
		req.Status = &active
	}

	if err := h.svc.UpdateStall(r.Context(), stallID, req); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, map[string]string{"message": "stall updated successfully"})
}

func (h *AdminHandler) updateStallStatus(w http.ResponseWriter, r *http.Request) {
	stallID := chi.URLParam(r, "id")
	var req struct {
		StallName string `json:"stall_name"`
		OwnerName string `json:"owner_name"`
		Phone     string `json:"phone"`
		Category  string `json:"category"`
		Status    string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// เรียกผ่าน Service
	if err := h.svc.UpdateStallStatus(r.Context(), stallID, req.Status); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonOK(w, map[string]string{"status": req.Status})
}

func (h *AdminHandler) getStallByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" || id == "undefined" {
		jsonError(w, http.StatusBadRequest, "invalid stall id")
		return
	}

	stall, err := h.svc.GetStallByID(r.Context(), id)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, stall)
}

func (h *AdminHandler) createStall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StallName string `json:"stall_name"`
		OwnerName string `json:"owner_name"` // 🚩 รับมาแล้ว
		Phone     string `json:"phone"`      // 🚩 รับมาแล้ว
		Category  string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 🚩 แก้ไข: ส่งค่า req.OwnerName และ req.Phone เข้าไปด้วย
	stall, err := h.svc.CreateStall(r.Context(), req.StallName, req.OwnerName, req.Phone, req.Category)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, stall)
}

func (h *AdminHandler) createStaff(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FullName string `json:"full_name"`
		Phone    string `json:"phone"` // 🚩 รับเบอร์โทรเพิ่ม
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	// Hash รหัสผ่าน 'password' เป็นค่าเริ่มต้น
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	id := uuid.NewString()
	// 🚩 ส่ง phone เข้าไปใน CreateStaff ด้วย
	err := h.svc.CreateStaff(r.Context(), id, req.FullName, req.Phone, string(hashedPassword))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, map[string]string{"id": id, "status": "STAFF_CREATED"})
}

func (h *AdminHandler) updateStaff(w http.ResponseWriter, r *http.Request) {
	staffID := chi.URLParam(r, "id")
	var req struct {
		FullName string `json:"full_name"`
		Phone    string `json:"phone"` // 🚩 รับเบอร์โทรเพิ่ม
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 🚩 ส่ง phone เข้าไปใน UpdateStaff ด้วย
	if err := h.svc.UpdateStaff(r.Context(), staffID, req.FullName, req.Phone); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, map[string]string{"message": "staff updated"})
}

func (h *AdminHandler) changePassword(w http.ResponseWriter, r *http.Request) {
	adminID := appMiddleware.GetUserID(r.Context())
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if err := h.svc.ChangeAdminPassword(r.Context(), adminID, req.OldPassword, req.NewPassword); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, map[string]string{"message": "password updated successfully"})
}
