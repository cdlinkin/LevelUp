package core_http_server

import (
	"net/http"

	core_http_middleware "github.com/cdlinkin/LevelUp/internal/core/transport/http/middleware"
)

type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc

	Middleware []core_http_middleware.Middleware
}

func (r *Route) WithMiddleware() http.Handler {
	return core_http_middleware.ChainMiddleware(
		r.Handler,
		r.Middleware...,
	)
}

func WithMiddleware(routes []Route, middleware ...core_http_middleware.Middleware) []Route {
	protected := make([]Route, len(routes))

	for i, route := range routes {
		route.Middleware = append(route.Middleware, middleware...)
		protected[i] = route
	}

	return protected
}
