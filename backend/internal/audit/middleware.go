package audit

import (
	"context"
	"net"
	"net/http"
	"strings"

	"saas-system/internal/db"
	"saas-system/internal/middleware"
)

func (s *Service) AuditMiddleware(action, entityType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			orgID, _ := db.OrgIDFromContext(ctx)

			userID, _ := ctx.Value(middleware.UserIDKey).(string)

			ip := extractIP(r)
			userAgent := r.UserAgent()

			ctx = context.WithValue(ctx, auditContextKey, &auditContext{
				orgID:      orgID,
				userID:     userID,
				action:     action,
				entityType: entityType,
				ip:         ip,
				userAgent:  userAgent,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type contextKey string

const auditContextKey contextKey = "audit_context"

type auditContext struct {
	orgID      string
	userID     string
	action     string
	entityType string
	entityID   string
	beforeData string
	afterData  string
	ip         string
	userAgent  string
}

func GetAuditContext(ctx context.Context) *auditContext {
	ac, _ := ctx.Value(auditContextKey).(*auditContext)
	return ac
}

func SetEntityID(ctx context.Context, entityID string) {
	if ac := GetAuditContext(ctx); ac != nil {
		ac.entityID = entityID
	}
}

func SetBeforeData(ctx context.Context, beforeData string) {
	if ac := GetAuditContext(ctx); ac != nil {
		ac.beforeData = beforeData
	}
}

func SetAfterData(ctx context.Context, afterData string) {
	if ac := GetAuditContext(ctx); ac != nil {
		ac.afterData = afterData
	}
}

func (s *Service) RecordAuditLog(ctx context.Context) error {
	ac := GetAuditContext(ctx)
	if ac == nil {
		return nil
	}

	if ac.orgID == "" {
		ac.orgID, _ = db.OrgIDFromContext(ctx)
	}

	if ac.userID == "" {
		ac.userID, _ = ctx.Value(middleware.UserIDKey).(string)
	}

	return s.Log(ctx, ac.orgID, ac.userID, ac.action, ac.entityType, ac.entityID, ac.beforeData, ac.afterData, ac.ip, ac.userAgent)
}

func extractIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
