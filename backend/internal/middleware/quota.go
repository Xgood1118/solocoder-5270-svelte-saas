package middleware

import (
	"context"
	"net/http"

	"saas-system/internal/db"
)

type QuotaChecker interface {
	CheckAndConsumeQuota(ctx context.Context, metric string, amount int) (bool, error)
}

func QuotaMiddleware(checker QuotaChecker, metric string, amount int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := db.OrgIDFromContext(r.Context())
			if !ok {
				JSONError(w, http.StatusBadRequest, "ORG_CONTEXT_REQUIRED", "Organization context is required")
				return
			}

			ok, err := checker.CheckAndConsumeQuota(r.Context(), metric, amount)
			if err != nil {
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check quota")
				return
			}

			if !ok {
				JSONError(w, http.StatusTooManyRequests, "QUOTA_EXCEEDED", "Quota exceeded for metric: "+metric)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
