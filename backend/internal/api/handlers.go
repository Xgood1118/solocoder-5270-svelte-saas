package api

import (
	"saas-system/internal/auth"
	"saas-system/internal/audit"
	"saas-system/internal/billing"
	"saas-system/internal/org"
	"saas-system/internal/webhook"
)

type APIHandler struct {
	AuthService    *auth.Service
	OrgService     *org.Service
	BillingService *billing.Service
	AuditService   *audit.Service
	WebhookService *webhook.Service
}

func NewAPIHandler(
	authService *auth.Service,
	orgService *org.Service,
	billingService *billing.Service,
	auditService *audit.Service,
	webhookService *webhook.Service,
) *APIHandler {
	return &APIHandler{
		AuthService:    authService,
		OrgService:     orgService,
		BillingService: billingService,
		AuditService:   auditService,
		WebhookService: webhookService,
	}
}
