package api

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestNewServerRouter_NotFound(t *testing.T) {
    r := NewServerRouter(nil)
    req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusNotFound {
        t.Errorf("expected 404 for unknown route, got %d", w.Code)
    }
}
