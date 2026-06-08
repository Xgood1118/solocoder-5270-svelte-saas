package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"saas-system/internal/db"
	"saas-system/internal/middleware"
)

type CreateEndpointRequest struct {
	URL    string `json:"url"`
	Active bool   `json:"active"`
}

type UpdateEndpointRequest struct {
	URL    string `json:"url"`
	Active bool   `json:"active"`
}

func (h *APIHandler) ListWebhookEndpoints(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	endpoints, err := h.WebhookService.ListEndpoints(r.Context(), orgID)
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list webhook endpoints")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, endpoints)
}

func (h *APIHandler) CreateWebhookEndpoint(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	var req CreateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.URL == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "URL is required")
		return
	}

	endpoint, err := h.WebhookService.CreateEndpoint(r.Context(), orgID, req.URL, req.Active)
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create webhook endpoint")
		return
	}

	middleware.JSONResponse(w, http.StatusCreated, endpoint)
}

func (h *APIHandler) UpdateWebhookEndpoint(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	endpointID := chi.URLParam(r, "id")
	if endpointID == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Endpoint ID is required")
		return
	}

	var req UpdateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.URL == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "URL is required")
		return
	}

	endpoint, err := h.WebhookService.UpdateEndpoint(r.Context(), orgID, endpointID, req.URL, req.Active)
	if err != nil {
		if err.Error() == "webhook endpoint not found" {
			middleware.JSONError(w, http.StatusNotFound, "ENDPOINT_NOT_FOUND", "Webhook endpoint not found")
			return
		}
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update webhook endpoint")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, endpoint)
}

func (h *APIHandler) DeleteWebhookEndpoint(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	endpointID := chi.URLParam(r, "id")
	if endpointID == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Endpoint ID is required")
		return
	}

	err := h.WebhookService.DeleteEndpoint(r.Context(), orgID, endpointID)
	if err != nil {
		if err.Error() == "webhook endpoint not found" {
			middleware.JSONError(w, http.StatusNotFound, "ENDPOINT_NOT_FOUND", "Webhook endpoint not found")
			return
		}
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete webhook endpoint")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Webhook endpoint deleted successfully",
	})
}
