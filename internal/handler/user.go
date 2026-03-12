package handler

import (
	"encoding/json"
	"net/http"

	"pmfoodcourt/internal/middleware"
	"pmfoodcourt/internal/model"
	"pmfoodcourt/internal/service"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	svc       *service.UserService
	jwtSecret string
}

func NewUserHandler(svc *service.UserService, secret string) *UserHandler {
	return &UserHandler{svc: svc, jwtSecret: secret}
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Delete("/{id}", h.delete)
	r.Get("/search", h.SearchByPhone)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(h.jwtSecret))
		r.Post("/change-password", h.changePassword)
	})

	return r
}

func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.GetAll(r.Context())
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, users)
}

func (h *UserHandler) get(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	if id == "" {
		jsonError(w, http.StatusBadRequest, "invalid id")
		return
	}
	u, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, http.StatusNotFound, err.Error())
		return
	}
	jsonOK(w, u)
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid body")
		return
	}
	u, err := h.svc.Create(r.Context(), req)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, u)
}

func (h *UserHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	if id == "" {
		jsonError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		jsonError(w, http.StatusNotFound, err.Error())
		return
	}
	jsonOK(w, map[string]string{"message": "deleted"})
}

func (h *UserHandler) changePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	var req model.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "ข้อมูลไม่ถูกต้อง")
		return
	}

	if err := h.svc.ChangePassword(r.Context(), userID, req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, map[string]string{"message": "เปลี่ยนรหัสผ่านสำเร็จ"})
}

func (h *UserHandler) SearchByPhone(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")

	if phone == "" {
		jsonError(w, http.StatusBadRequest, "phone required")
		return
	}

	user, err := h.svc.GetByPhone(r.Context(), phone)
	if err != nil {
		jsonError(w, http.StatusNotFound, "user not found")
		return
	}

	jsonOK(w, user)
}

// ── helpers ────────────────────────────────────────────────────

func parseID(r *http.Request) string {
	return chi.URLParam(r, "id")
}
