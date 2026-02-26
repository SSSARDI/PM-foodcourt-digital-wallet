package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/joho/godotenv/autoload"

	"pmfoodcourt/config"
    "pmfoodcourt/internal/handler"
    "pmfoodcourt/internal/repository"
    "pmfoodcourt/internal/service"
    "pmfoodcourt/pkg/database"
)
func main() {
	cfg := config.Load()
	db := database.New(cfg.DSN)
	defer db.Close()

	// ── wire up ──────────────────────────────
	userRepo := repository.NewUserRepo(db)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	// ── router ───────────────────────────────
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(cors.AllowAll().Handler)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})
	r.Mount("/api/v1/users", userHandler.Routes())

	// ── start ────────────────────────────────
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 running on http://localhost%s  [%s]", addr, cfg.Env)
	log.Fatal(http.ListenAndServe(addr, r))
}
