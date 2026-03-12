package handler

import (
	"encoding/json"
	"net/http"

	appMiddleware "pmfoodcourt/internal/middleware"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/service"

	"github.com/go-chi/chi/v5"
)

type StallHandler struct {
	svc       *service.StallService
	jwtSecret string
}

func NewStallHandler(svc *service.StallService, jwtSecret string) *StallHandler {
	return &StallHandler{svc, jwtSecret}
}

func (h *StallHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// public
	r.Get("/", h.list)
	r.Get("/{id}", h.get)

	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.Auth(h.jwtSecret))

		r.Group(func(r chi.Router) {
			r.Use(appMiddleware.RequireRole("CUSTOMER"))
			r.Get("/my-purchases", h.myPurchases)
		})
		r.Group(func(r chi.Router) {
			r.Use(appMiddleware.RequireRole("VENDOR", "ADMIN"))
			r.Get("/summary", h.getSummary)
			r.Get("/my", h.myStalls)
			r.Get("/my", h.myStalls)
			r.Post("/", h.create)
			r.Patch("/{id}", h.update)
			r.Delete("/{id}", h.delete)
			r.Get("/{id}/sales", h.saleHistory)
		})
	})

	return r
}

func (h *StallHandler) list(w http.ResponseWriter, r *http.Request) {
	stalls, err := h.svc.List(r.Context())
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, stalls)
}

func (h *StallHandler) get(w http.ResponseWriter, r *http.Request) {
	stall, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusNotFound, err.Error())
		return
	}
	jsonOK(w, stall)
}

func (h *StallHandler) myStalls(w http.ResponseWriter, r *http.Request) {
	stalls, err := h.svc.MyStalls(r.Context(), appMiddleware.GetUserID(r.Context()))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, stalls)
}

func (h *StallHandler) myPurchases(w http.ResponseWriter, r *http.Request) {
	sales, err := h.svc.MySales(r.Context(), appMiddleware.GetUserID(r.Context()))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, sales)
}

func (h *StallHandler) create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateStallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	stall, err := h.svc.Create(r.Context(), appMiddleware.GetUserID(r.Context()), req)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(stall)
}

func (h *StallHandler) update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateStallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	stall, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, stall)
}

func (h *StallHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, map[string]string{"message": "deleted"})
}

func (h *StallHandler) saleHistory(w http.ResponseWriter, r *http.Request) {
	sales, err := h.svc.SaleHistory(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, sales)
}

func (h *StallHandler) getSummary(w http.ResponseWriter, r *http.Request) {
	userID := appMiddleware.GetUserID(r.Context())

	// 1. หา StallID ของป้าจงก่อน
	stalls, err := h.svc.MyStalls(r.Context(), userID)
	if err != nil || len(stalls) == 0 {
		jsonError(w, http.StatusNotFound, "stall not found")
		return
	}

	// 2. ดึงสรุปยอดจาก Service (ฟังก์ชันที่เราเพิ่ม GetDailySummary เข้าไป)
	total, count, err := h.svc.GetDailySummary(r.Context(), stalls[0].ID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonOK(w, map[string]interface{}{
		"today_sales": total,
		"order_count": count,
	})
}
