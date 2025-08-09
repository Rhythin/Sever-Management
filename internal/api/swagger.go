package api

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

// NewSwaggerHandler returns a handler for Swagger UI
func NewSwaggerHandler() http.Handler {
	return httpSwagger.WrapHandler
}
