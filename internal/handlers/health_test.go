package handlers

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHealthzHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
    w := httptest.NewRecorder()
    HealthzHandler(w, req)
    if w.Code != http.StatusOK {
        t.Errorf("Healthz status = %d; want %d", w.Code, http.StatusOK)
    }
    if body := w.Body.String(); body != "ok" {
        t.Errorf("Healthz body = %s; want ok", body)
    }
}

func TestReadyzHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
    w := httptest.NewRecorder()
    ReadyzHandler(w, req)
    if w.Code != http.StatusOK {
        t.Errorf("Readyz status = %d; want %d", w.Code, http.StatusOK)
    }
    if body := w.Body.String(); body != "ready" {
        t.Errorf("Readyz body = %s; want ready", body)
    }
}
