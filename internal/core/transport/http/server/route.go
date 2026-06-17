package core_http_server

// Описывает HTTP маршруты (какой метод путь или хандлер выбрать)

import (
	"net/http"

	core_http_middleware "github.com/d1mas2k3/url_shortener/internal/core/transport/http/middleware"
)

type Route struct {
	Method     string
	Path       string
	Handler    http.HandlerFunc
	Middleware []core_http_middleware.Middleware
}

func (r *Route) WithMiddleware() http.Handler {
	return core_http_middleware.ChainMiddleware(
		r.Handler,
		r.Middleware...,
	)
}
