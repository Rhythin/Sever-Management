package handlers

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNewSwaggerHandler(t *testing.T) {

	tests := []struct {
		name string
		want http.Handler
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSwaggerHandler(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSwaggerHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
