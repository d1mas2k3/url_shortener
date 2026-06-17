package core_http_server

// Обертка над обычным http.ServeMux
// Позволяет хранить версию API, удобно регистрировать routes, структурировать backend

import (
	"fmt"
	"net/http"

	core_http_middleware "github.com/d1mas2k3/url_shortener/internal/core/transport/http/middleware"
)

type ApiVersion string

var ApiVersion1 = ApiVersion("v1")
var ApiVersion2 = ApiVersion("v2")
var ApiVersion3 = ApiVersion("v3")

type APIVersionRouter struct {
	*http.ServeMux
	apiVersion ApiVersion
	middleware []core_http_middleware.Middleware
}

func NewAPIVersionRouter(
	apiVersion ApiVersion,
	middleware ...core_http_middleware.Middleware,
) *APIVersionRouter {
	return &APIVersionRouter{
		ServeMux:   http.NewServeMux(),
		apiVersion: apiVersion,
		middleware: middleware,
	}
}

func (r *APIVersionRouter) RegisterRoutes(routes ...Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)

		r.Handle(pattern, route.WithMiddleware()) // Регистрирует endpoint в ServeMux
	}
}

func (r *APIVersionRouter) WithMiddleware() http.Handler {
    return core_http_middleware.ChainMiddleware(
        r,
        r.middleware...,
    )
}