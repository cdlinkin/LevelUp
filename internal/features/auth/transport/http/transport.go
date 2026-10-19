package auth_transport_http

import (
	"context"
	"net/http"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_http_middleware "github.com/cdlinkin/LevelUp/internal/core/transport/http/middleware"
	core_http_server "github.com/cdlinkin/LevelUp/internal/core/transport/http/server"
)

type AuthHTTPHandler struct {
	authService    AuthService
	authMiddleware core_http_middleware.Middleware
}

type AuthService interface {
	Register(
		ctx context.Context,
		fullName string,
		phoneNumber *string,
		email string,
		password string,
	) (domain.User, error)
	Login(
		ctx context.Context,
		email string,
		password string,
	) (string, time.Duration, error)
	Me(
		ctx context.Context,
		userID int,
	) (domain.User, error)
}

func NewAuthHTTPHandler(
	authService AuthService,
	authMiddleware core_http_middleware.Middleware,
) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		authService:    authService,
		authMiddleware: authMiddleware,
	}
}

func (h *AuthHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/auth/register",
			Handler: h.Register,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/login",
			Handler: h.Login,
		},
		{
			Method:     http.MethodGet,
			Path:       "/auth/me",
			Handler:    h.Me,
			Middleware: []core_http_middleware.Middleware{h.authMiddleware},
		},
	}
}
