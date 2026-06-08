package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey    contextKey = "request_id"
	UserKey         contextKey = "user"
	UserIDKey       contextKey = "user_id"
	OrgRoleKey      contextKey = "org_role"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func JSONError(w http.ResponseWriter, status int, errorCode, message string) {
	JSONResponse(w, status, ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Org-Id, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		r = r.WithContext(context.WithValue(r.Context(), RequestIDKey, requestID))
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID, _ := r.Context().Value(RequestIDKey).(string)

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s %d %s %s",
			requestID,
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			duration,
			r.RemoteAddr,
		)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v\n%s", err, debug.Stack())
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
