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

	// ── repositories ──────────────────────────────────────────
	userRepo := repository.NewUserRepo(db)
	authRepo := repository.NewAuthRepo(db)
	walletRepo := repository.NewWalletRepo(db)
	stallRepo := repository.NewStallRepo(db)
	qrRepo := repository.NewQRRepo(db)
	refundRepo := repository.NewRefundRepo(db)
	gpvatRepo := repository.NewGpVatRepo(db)

	// ── services ──────────────────────────────────────────────
	userSvc := service.NewUserService(userRepo)
	authSvc := service.NewAuthService(authRepo, walletRepo, cfg.JWTSecret)
	walletSvc := service.NewWalletService(walletRepo, stallRepo, qrRepo)
	stallSvc := service.NewStallService(stallRepo)
	adminSvc := service.NewAdminService(userRepo, walletRepo, refundRepo, gpvatRepo)

	// ── handlers ──────────────────────────────────────────────
	userHandler := handler.NewUserHandler(userSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	walletHandler := handler.NewWalletHandler(walletSvc, cfg.JWTSecret)
	stallHandler := handler.NewStallHandler(stallSvc, cfg.JWTSecret)
	adminHandler := handler.NewAdminHandler(adminSvc, cfg.JWTSecret)

	// ── router ────────────────────────────────────────────────
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(cors.AllowAll().Handler)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Mount("/api/v1/users", userHandler.Routes())
	r.Mount("/api/v1/auth", authHandler.Routes())
	r.Mount("/api/v1/wallet", walletHandler.Routes())
	r.Mount("/api/v1/stalls", stallHandler.Routes())
	r.Mount("/api/v1/admin", adminHandler.Routes())

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 running on http://localhost%s  [%s]", addr, cfg.Env)
	log.Fatal(http.ListenAndServe(addr, r))
}
