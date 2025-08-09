package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewMetricsHandler returns a handler for Prometheus metrics
func NewMetricsHandler() http.Handler {
	return promhttp.Handler()
}
