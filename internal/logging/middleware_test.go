package logging

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestRequestIDFromContext_NoValue(t *testing.T) {
    if id := RequestIDFromContext(context.Background()); id != "unknown" {
        t.Errorf("Expected 'unknown', got %s", id)
    }
}

func TestRequestIDMiddleware(t *testing.T) {
    called := false
    handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        called = true
        id := RequestIDFromContext(r.Context())
        if id == "unknown" {
            t.Error("RequestID not in context")
        }
    }))
    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    if !called {
        t.Error("Handler not called")
    }
    hdr := w.Header().Get("X-Request-Id")
    if hdr == "" {
        t.Error("X-Request-Id header not set")
    }
}
