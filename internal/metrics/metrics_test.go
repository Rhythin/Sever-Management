package metrics

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNewMetricsHandler(t *testing.T) {
	tests := []struct {
		name string
		want http.Handler
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMetricsHandler(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMetricsHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
