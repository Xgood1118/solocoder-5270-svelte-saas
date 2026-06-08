package main

import (
	"fmt"
	"log"
	"net/http"

	"saas-system/internal/api"
	"saas-system/internal/auth"
	"saas-system/internal/audit"
	"saas-system/internal/billing"
	"saas-system/internal/config"
	"saas-system/internal/db"
	"saas-system/internal/org"
	"saas-system/internal/webhook"
)

func main() {
	cfg := config.Load()

	database, err := db.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database.DB); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	if err := db.Seed(database.DB); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	authService := auth.NewService(database.DB, cfg.JWTSecret)
	orgService := org.NewService(database)
	billingService := billing.NewService(database.DB)
	auditService := audit.NewService(database.DB)
	webhookService := webhook.NewService(database.DB, cfg.WebhookSecret)

	r := api.NewRouter(authService, orgService, billingService, auditService, webhookService)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
