package handlers

import (
	"net/http"

	"github.com/rhythin/sever-management/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewSwaggerHandler returns a handler for Swagger UI
func NewSwaggerHandler() http.Handler {
	docs.SwaggerInfo.BasePath = "/"
	return httpSwagger.WrapHandler
}
