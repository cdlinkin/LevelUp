package web_transport_http

import (
	"net/http"

	core_http_server "github.com/cdlinkin/LevelUp/internal/core/transport/http/server"
)

type WebHTTPHandler struct {
	webService WebService
}

type WebService interface {
	GetMainPage() ([]byte, error)
	GetPublicFile(rel string) ([]byte, error)
}

func NewWebHTTPHandler(
	webService WebService,
) *WebHTTPHandler {
	return &WebHTTPHandler{
		webService: webService,
	}
}

func (h *WebHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/{$}",
			Handler: h.GetMainPage,
		},
		{
			Method:  http.MethodGet,
			Path:    "/{file}",
			Handler: h.GetPublicFile,
		},
		{
			Method:  http.MethodGet,
			Path:    "/brand/{file}",
			Handler: h.GetBrandFile,
		},
	}
}
