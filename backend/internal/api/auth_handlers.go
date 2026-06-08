package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"saas-system/internal/middleware"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	OrgName  string `json:"org_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

func (h *APIHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Email, password, and name are required")
		return
	}

	if len(req.Password) < 6 {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Password must be at least 6 characters")
		return
	}

	orgName := req.OrgName
	if orgName == "" {
		orgName = req.Name + "'s Org"
	}

	slug := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(orgName), " ", "-"))

	user, err := h.AuthService.CreateUser(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if err.Error() == "email already exists" {
			middleware.JSONError(w, http.StatusConflict, "EMAIL_EXISTS", "Email already exists")
			return
		}
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create user")
		return
	}

	org, err := h.OrgService.CreateOrg(user.ID, orgName, slug)
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create organization")
		return
	}

	token, err := h.AuthService.GenerateToken(user.ID, org.ID, 24*time.Hour)
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate token")
		return
	}

	middleware.JSONResponse(w, http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		middleware.JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Email and password are required")
		return
	}

	user, err := h.AuthService.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		middleware.JSONError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	orgs, err := h.OrgService.ListUserOrgs(user.ID)
	if err != nil || len(orgs) == 0 {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load organizations")
		return
	}

	defaultOrg := orgs[0]

	token, err := h.AuthService.GenerateToken(user.ID, defaultOrg.ID, 24*time.Hour)
	if err != nil {
		middleware.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate token")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *APIHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		middleware.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	middleware.JSONResponse(w, http.StatusOK, user)
}

func (h *APIHandler) Logout(w http.ResponseWriter, r *http.Request) {
	middleware.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}
