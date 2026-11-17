package web_transport_http

import (
	"net/http"
	"path"

	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_http_response "github.com/cdlinkin/LevelUp/internal/core/transport/http/response"
)

func (h *WebHTTPHandler) GetPublicFile(rw http.ResponseWriter, r *http.Request) {
	h.writePublicFile(rw, r, r.PathValue("file"))
}

func (h *WebHTTPHandler) GetBrandFile(rw http.ResponseWriter, r *http.Request) {
	h.writePublicFile(rw, r, path.Join("brand", r.PathValue("file")))
}

func (h *WebHTTPHandler) writePublicFile(rw http.ResponseWriter, r *http.Request, rel string) {
	log := core_logger.FromContext(r.Context())
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	file, err := h.webService.GetPublicFile(rel)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get public file")
		return
	}

	rw.Header().Set("Content-Type", publicContentType(rel))
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(file)
}

func publicContentType(rel string) string {
	switch path.Ext(rel) {
	case ".css":
		return "text/css; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}
