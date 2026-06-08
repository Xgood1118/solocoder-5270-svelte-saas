package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"saas-system/internal/audit"
	"saas-system/internal/db"
	"saas-system/internal/middleware"
)

type ArchiveLogsRequest struct {
	Before string `json:"before"`
}

func (h *APIHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	params := audit.ListAuditLogsParams{
		OrgID:      orgID,
		Action:     r.URL.Query().Get("action"),
		EntityType: r.URL.Query().Get("entity_type"),
		EntityID:   r.URL.Query().Get("entity_id"),
		UserID:     r.URL.Query().Get("user_id"),
		Limit:      limit,
		Offset:     offset,
	}

	logs, total, err := h.AuditService.ListAuditLogs(r.Context(), params)
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list audit logs")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":  logs,
		"total": total,
		"limit": params.Limit,
		"offset": params.Offset,
	})
}

func (h *APIHandler) ArchiveLogs(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	var req ArchiveLogsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	var before time.Time
	if req.Before != "" {
		var err error
		before, err = time.Parse(time.RFC3339, req.Before)
		if err != nil {
			middleware.JSONError(w, http.StatusBadRequest, "INVALID_DATE", "Invalid date format")
			return
		}
	} else {
		before = time.Now().AddDate(0, -3, 0)
	}

	count, err := h.AuditService.ArchiveLogs(r.Context(), orgID, before)
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to archive logs")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Logs archived successfully",
		"count":   count,
	})
}
