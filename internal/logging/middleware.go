package logging

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIDKey = "request_id"

// Middleware injects a request ID into the context and header
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-Id")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		w.Header().Set("X-Request-Id", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FromContext extracts the request ID from context
func RequestIDFromContext(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "unknown"
}

// S returns a sugared logger with request_id field if present in context
func S(ctx context.Context) *zap.SugaredLogger {
	reqID := RequestIDFromContext(ctx)
	return zap.S().With("request_id", reqID)
}
