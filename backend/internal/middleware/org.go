package middleware

import (
	"context"
	"net/http"

	"saas-system/internal/billing"
	"saas-system/internal/db"
)

type OrgService interface {
	GetMemberRole(orgID, userID string) (string, error)
}

func OrgMiddleware(orgService OrgService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}

			orgID := r.Header.Get("X-Org-Id")
			if orgID == "" {
				JSONError(w, http.StatusBadRequest, "ORG_ID_REQUIRED", "X-Org-Id header is required")
				return
			}

			role, err := orgService.GetMemberRole(orgID, userID)
			if err != nil {
				if err.Error() == "member not found" || err.Error() == "permission denied" {
					JSONError(w, http.StatusForbidden, "FORBIDDEN", "User does not belong to this organization")
					return
				}
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify organization membership")
				return
			}

			ctx := r.Context()
			ctx = db.WithOrgID(ctx, orgID)
			ctx = billing.WithOrgID(ctx, orgID)
			ctx = context.WithValue(ctx, OrgRoleKey, role)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func OrgRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(OrgRoleKey).(string)
	return role, ok
}
