package metrics

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestCounters(t *testing.T) {
	atomic.StoreInt64(&totalServers, 0)
	atomic.StoreInt64(&runningServers, 0)

	IncTotalServers()
	if got := GetTotalServers(); got != 1 {
		t.Errorf("TotalServers = %d; want 1", got)
	}
	DecTotalServers()
	if got := GetTotalServers(); got != 0 {
		t.Errorf("TotalServers = %d; want 0", got)
	}

	atomic.StoreInt64(&runningServers, 5)
	IncRunningServers()
	if got := GetRunningServers(); got != 6 {
		t.Errorf("RunningServers = %d; want 6", got)
	}
	DecRunningServers()
	if got := GetRunningServers(); got != 5 {
		t.Errorf("RunningServers = %d; want 5", got)
	}
}

func TestNewMetricsHandler(t *testing.T) {
	handler := NewMetricsHandler()
	if handler == nil {
		t.Fatal("NewMetricsHandler returned nil")
	}
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Metrics handler status = %d; want %d", w.Code, http.StatusOK)
	}
}
