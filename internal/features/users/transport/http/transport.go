package users_transport_http

import (
	"context"
	"net/http"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_http_server "github.com/cdlinkin/LevelUp/internal/core/transport/http/server"
)

type UsersHTTPHandler struct {
	usersService UsersService
}

type UsersService interface {
	GetUser(ctx context.Context, userID int, id int) (domain.User, error)
	DeleteUser(ctx context.Context, userID int, id int) error
	PatchUser(ctx context.Context, userID int, id int, patch domain.UserPatch) (domain.User, error)
}

func NewUsersHTTPHandler(
	usersService UsersService,
) *UsersHTTPHandler {
	return &UsersHTTPHandler{
		usersService: usersService,
	}
}

func (h *UsersHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/users/{id}",
			Handler: h.GetUser,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/users/{id}",
			Handler: h.DeleteUser,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/users/{id}",
			Handler: h.PatchUser,
		},
	}
}
