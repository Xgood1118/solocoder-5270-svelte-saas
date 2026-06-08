package middleware

import (
	"context"
	"net/http"
	"strings"

	"saas-system/internal/auth"
)

type AuthService interface {
	ValidateToken(tokenString string) (*auth.Claims, error)
	GetUserByID(ctx context.Context, userID string) (*auth.User, error)
}

func AuthMiddleware(authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header is required")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization format")
				return
			}

			tokenString := parts[1]
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
				return
			}

			user, err := authService.GetUserByID(r.Context(), claims.UserID)
			if err != nil {
				JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not found")
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserKey, user)
			ctx = context.WithValue(ctx, UserIDKey, user.ID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func OptionalAuthMiddleware(authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := parts[1]
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := authService.GetUserByID(r.Context(), claims.UserID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserKey, user)
			ctx = context.WithValue(ctx, UserIDKey, user.ID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func UserFromContext(ctx context.Context) (*auth.User, bool) {
	user, ok := ctx.Value(UserKey).(*auth.User)
	return user, ok
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
