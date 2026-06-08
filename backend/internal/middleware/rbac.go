package middleware

import (
	"net/http"
)

var roleHierarchy = map[string]int{
	"owner":  4,
	"admin":  3,
	"member": 2,
	"guest":  1,
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := OrgRoleFromContext(r.Context())
			if !ok {
				JSONError(w, http.StatusForbidden, "FORBIDDEN", "Role information not available")
				return
			}

			userLevel := roleHierarchy[userRole]

			for _, role := range roles {
				requiredLevel := roleHierarchy[role]
				if userLevel >= requiredLevel {
					next.ServeHTTP(w, r)
					return
				}
			}

			JSONError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
		})
	}
}
