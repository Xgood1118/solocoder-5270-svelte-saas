package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"saas-system/internal/auth"
	"saas-system/internal/audit"
	"saas-system/internal/billing"
	"saas-system/internal/middleware"
	"saas-system/internal/org"
	"saas-system/internal/webhook"
)

func NewRouter(
	authService *auth.Service,
	orgService *org.Service,
	billingService *billing.Service,
	auditService *audit.Service,
	webhookService *webhook.Service,
) http.Handler {
	handler := NewAPIHandler(authService, orgService, billingService, auditService, webhookService)

	r := chi.NewRouter()

	r.Use(middleware.CORS)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recover)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/register", handler.Register)
		r.Post("/auth/login", handler.Login)
		r.Post("/invitations/{token}/accept", handler.AcceptInvitation)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(authService))

			r.Get("/auth/me", handler.Me)
			r.Post("/auth/logout", handler.Logout)

			r.Route("/orgs", func(r chi.Router) {
				r.Get("/", handler.ListOrgs)
				r.Post("/", handler.CreateOrg)
				r.Get("/{id}", handler.GetOrg)
				r.Put("/{id}", handler.UpdateOrg)
				r.Delete("/{id}", handler.DeleteOrg)
			})

			r.Get("/plans", handler.ListPlans)

			r.Group(func(r chi.Router) {
				r.Use(middleware.OrgMiddleware(orgService))

				r.Route("/members", func(r chi.Router) {
					r.Get("/", handler.ListMembers)
					r.Patch("/{user_id}/role", handler.UpdateMemberRole)
					r.Delete("/{user_id}", handler.RemoveMember)
				})

				r.Route("/invitations", func(r chi.Router) {
					r.Get("/", handler.ListInvitations)
					r.Post("/", handler.InviteMember)
					r.Delete("/{id}", handler.CancelInvitation)
				})

				r.Route("/billing", func(r chi.Router) {
					r.Get("/subscription", handler.GetSubscription)
					r.Post("/subscription/upgrade", handler.UpgradeSubscription)
					r.Post("/subscription/downgrade", handler.DowngradeSubscription)
					r.Post("/subscription/cancel", handler.CancelSubscription)
					r.Get("/invoices", handler.ListInvoices)
					r.Get("/quotas", handler.ListQuotas)
					r.Post("/invoices/{id}/pay", handler.PayInvoice)
				})

				r.Route("/audit-logs", func(r chi.Router) {
					r.Get("/", handler.ListAuditLogs)
					r.With(middleware.RequireRole("admin", "owner")).Post("/archive", handler.ArchiveLogs)
				})

				r.Route("/webhooks", func(r chi.Router) {
					r.Get("/endpoints", handler.ListWebhookEndpoints)
					r.Post("/endpoints", handler.CreateWebhookEndpoint)
					r.Put("/endpoints/{id}", handler.UpdateWebhookEndpoint)
					r.Delete("/endpoints/{id}", handler.DeleteWebhookEndpoint)
				})
			})
		})
	})

	return r
}
