package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"saas-system/internal/db"
	"saas-system/internal/middleware"
)

type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

func (h *APIHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	members, err := h.OrgService.ListMembers(orgID)
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list members")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, members)
}

func (h *APIHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Email == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Email is required")
		return
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	invitation, err := h.OrgService.InviteMember(orgID, userID, req.Email, role)
	if err != nil {
		switch err.Error() {
		case "invalid role":
			middleware.JSONError(w, http.StatusBadRequest, "INVALID_ROLE", "Invalid role")
		case "permission denied":
			middleware.JSONError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
		default:
			middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to invite member")
		}
		return
	}

	middleware.JSONResponse(w, http.StatusCreated, invitation)
}

func (h *APIHandler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	invitations, err := h.OrgService.ListInvitations(orgID)
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list invitations")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, invitations)
}

func (h *APIHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invitation token is required")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	member, err := h.OrgService.AcceptInvitation(token, userID)
	if err != nil {
		switch err.Error() {
		case "invitation not found":
			middleware.JSONError(w, http.StatusNotFound, "INVITATION_NOT_FOUND", "Invitation not found")
		case "invitation expired":
			middleware.JSONError(w, http.StatusGone, "INVITATION_EXPIRED", "Invitation has expired")
		case "invitation already accepted":
			middleware.JSONError(w, http.StatusConflict, "INVITATION_ACCEPTED", "Invitation already accepted")
		default:
			middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to accept invitation")
		}
		return
	}

	middleware.JSONResponse(w, http.StatusOK, member)
}

func (h *APIHandler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	invitationID := chi.URLParam(r, "id")
	if invitationID == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invitation ID is required")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	err := h.OrgService.CancelInvitation(orgID, invitationID, userID)
	if err != nil {
		switch err.Error() {
		case "invitation not found":
			middleware.JSONError(w, http.StatusNotFound, "INVITATION_NOT_FOUND", "Invitation not found")
		case "permission denied":
			middleware.JSONError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
		default:
			middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel invitation")
		}
		return
	}

	middleware.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Invitation cancelled successfully",
	})
}

func (h *APIHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	userIDToUpdate := chi.URLParam(r, "user_id")
	if userIDToUpdate == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	var req UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Role == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Role is required")
		return
	}

	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	err := h.OrgService.UpdateMemberRole(orgID, userIDToUpdate, req.Role, actorID)
	if err != nil {
		switch err.Error() {
		case "member not found":
			middleware.JSONError(w, http.StatusNotFound, "MEMBER_NOT_FOUND", "Member not found")
		case "invalid role":
			middleware.JSONError(w, http.StatusBadRequest, "INVALID_ROLE", "Invalid role")
		case "permission denied":
			middleware.JSONError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
		default:
			middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update member role")
		}
		return
	}

	middleware.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Member role updated successfully",
	})
}

func (h *APIHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID, ok := db.OrgIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
		return
	}

	userIDToRemove := chi.URLParam(r, "user_id")
	if userIDToRemove == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
		return
	}

	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	err := h.OrgService.RemoveMember(orgID, userIDToRemove, actorID)
	if err != nil {
		switch err.Error() {
		case "member not found":
			middleware.JSONError(w, http.StatusNotFound, "MEMBER_NOT_FOUND", "Member not found")
		case "permission denied":
			middleware.JSONError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
		default:
			middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove member")
		}
		return
	}

	middleware.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Member removed successfully",
	})
}
