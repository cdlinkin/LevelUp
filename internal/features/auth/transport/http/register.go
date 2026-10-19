package auth_transport_http

import (
	"net/http"

	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_http_request "github.com/cdlinkin/LevelUp/internal/core/transport/http/request"
	core_http_response "github.com/cdlinkin/LevelUp/internal/core/transport/http/response"
)

func (h *AuthHTTPHandler) Register(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request RegisterRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)
		return
	}

	user, err := h.authService.Register(
		ctx,
		request.FullName,
		request.PhoneNumber,
		request.Email,
		request.Password,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to register user",
		)
		return
	}

	response := RegisterResponse{
		ID:          user.ID,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}

	responseHandler.JSONResponse(response, http.StatusCreated)
}
