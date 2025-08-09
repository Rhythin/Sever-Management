package metrics

import (
	"net/http"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var totalServers int64
var runningServers int64

func IncTotalServers()         { atomic.AddInt64(&totalServers, 1) }
func DecTotalServers()         { atomic.AddInt64(&totalServers, -1) }
func IncRunningServers()       { atomic.AddInt64(&runningServers, 1) }
func DecRunningServers()       { atomic.AddInt64(&runningServers, -1) }
func GetTotalServers() int64   { return atomic.LoadInt64(&totalServers) }
func GetRunningServers() int64 { return atomic.LoadInt64(&runningServers) }

// NewMetricsHandler returns a handler for Prometheus metrics
func NewMetricsHandler() http.Handler {
	return promhttp.Handler()
}
