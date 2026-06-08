package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"saas-system/internal/middleware"
)

type CreateOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type UpdateOrgRequest struct {
	Name string `json:"name"`
}

func (h *APIHandler) ListOrgs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	orgs, err := h.OrgService.ListUserOrgs(userID)
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list organizations")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, orgs)
}

func (h *APIHandler) CreateOrg(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Name == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Organization name is required")
		return
	}

	slug := req.Slug
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(req.Name), " ", "-"))
	}

	org, err := h.OrgService.CreateOrg(userID, req.Name, slug)
	if err != nil {
		if err.Error() == "org already exists" {
			middleware.JSONError(w, http.StatusConflict, "ORG_EXISTS", "Organization with this slug already exists")
			return
		}
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create organization")
		return
	}

	middleware.JSONResponse(w, http.StatusCreated, org)
}

func (h *APIHandler) GetOrg(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "id")
	if orgID == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Organization ID is required")
		return
	}

	org, err := h.OrgService.GetOrg(orgID)
	if err != nil {
		if err.Error() == "org not found" || err.Error() == "permission denied" {
			middleware.JSONError(w, http.StatusNotFound, "ORG_NOT_FOUND", "Organization not found")
			return
		}
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get organization")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, org)
}

func (h *APIHandler) UpdateOrg(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "id")
	if orgID == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Organization ID is required")
		return
	}

	var req UpdateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Name == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Organization name is required")
		return
	}

	org, err := h.OrgService.UpdateOrg(orgID, req.Name)
	if err != nil {
		if err.Error() == "org not found" || err.Error() == "permission denied" {
			middleware.JSONError(w, http.StatusNotFound, "ORG_NOT_FOUND", "Organization not found")
			return
		}
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update organization")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, org)
}

func (h *APIHandler) DeleteOrg(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "id")
	if orgID == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Organization ID is required")
		return
	}

	err := h.OrgService.DeleteOrg(orgID)
	if err != nil {
		if err.Error() == "org not found" || err.Error() == "permission denied" {
			middleware.JSONError(w, http.StatusNotFound, "ORG_NOT_FOUND", "Organization not found")
			return
		}
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete organization")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Organization deleted successfully",
	})
}
