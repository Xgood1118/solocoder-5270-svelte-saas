package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"saas-system/internal/middleware"
)

type UpgradePlanRequest struct {
	PlanID string `json:"plan_id"`
}

type DowngradePlanRequest struct {
	PlanID string `json:"plan_id"`
}

func (h *APIHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.BillingService.ListPlans(r.Context())
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list plans")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, plans)
}

func (h *APIHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	sub, err := h.BillingService.GetSubscription(r.Context())
	if err != nil {
		if err.Error() == "subscription not found" {
			middleware.JSONError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Subscription not found")
			return
		}
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get subscription")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, sub)
}

func (h *APIHandler) UpgradeSubscription(w http.ResponseWriter, r *http.Request) {
	var req UpgradePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.PlanID == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Plan ID is required")
		return
	}

	sub, err := h.BillingService.UpgradePlan(r.Context(), req.PlanID)
	if err != nil {
		switch err.Error() {
		case "plan not found":
			middleware.JSONError(w, http.StatusNotFound, "PLAN_NOT_FOUND", "Plan not found")
		case "invalid plan":
			middleware.JSONError(w, http.StatusBadRequest, "INVALID_PLAN", "Invalid plan")
		default:
			middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to upgrade subscription")
		}
		return
	}

	middleware.JSONResponse(w, http.StatusOK, sub)
}

func (h *APIHandler) DowngradeSubscription(w http.ResponseWriter, r *http.Request) {
	var req DowngradePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.PlanID == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Plan ID is required")
		return
	}

	sub, err := h.BillingService.DowngradePlan(r.Context(), req.PlanID)
	if err != nil {
		switch err.Error() {
		case "plan not found":
			middleware.JSONError(w, http.StatusNotFound, "PLAN_NOT_FOUND", "Plan not found")
		case "invalid plan":
			middleware.JSONError(w, http.StatusBadRequest, "INVALID_PLAN", "Invalid plan")
		default:
			middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to downgrade subscription")
		}
		return
	}

	middleware.JSONResponse(w, http.StatusOK, sub)
}

func (h *APIHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	err := h.BillingService.CancelSubscription(r.Context())
	if err != nil {
		if err.Error() == "subscription not found" {
			middleware.JSONError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND", "Subscription not found")
			return
		}
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel subscription")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Subscription cancelled successfully",
	})
}

func (h *APIHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.BillingService.ListInvoices(r.Context())
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list invoices")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, invoices)
}

func (h *APIHandler) ListQuotas(w http.ResponseWriter, r *http.Request) {
	quotas, err := h.BillingService.ListQuotas(r.Context())
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list quotas")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, quotas)
}

func (h *APIHandler) PayInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID := chi.URLParam(r, "id")
	if invoiceID == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invoice ID is required")
		return
	}

	err := h.BillingService.PayInvoice(r.Context(), invoiceID)
	if err != nil {
		switch err.Error() {
		case "invoice not found":
			middleware.JSONError(w, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice not found")
		case "payment failed":
			middleware.JSONError(w, http.StatusPaymentRequired, "PAYMENT_FAILED", "Payment failed")
		default:
			middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to pay invoice")
		}
		return
	}

	middleware.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Invoice paid successfully",
	})
}
