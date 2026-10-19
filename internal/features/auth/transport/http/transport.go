package auth_transport_http

import (
	"context"
	"net/http"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_http_server "github.com/cdlinkin/LevelUp/internal/core/transport/http/server"
)

type AuthHTTPHandler struct {
	authService AuthService
}

type AuthService interface {
	Register(
		ctx context.Context,
		fullName string,
		phoneNumber *string,
		email string,
		password string,
	) (domain.User, error)
}

func NewAuthHTTPHandler(
	authService AuthService,
) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		authService: authService,
	}
}

func (h *AuthHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/auth/register",
			Handler: h.Register,
		},
	}
}
