package auth_transport_http

import (
	"net/http"

	core_auth "github.com/cdlinkin/LevelUp/internal/core/auth"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_http_response "github.com/cdlinkin/LevelUp/internal/core/transport/http/response"
)

func (h *AuthHTTPHandler) Me(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, ok := core_auth.UserIDFromContext(ctx)
	if !ok {
		responseHandler.ErrorResponse(
			core_errors.ErrUnauthorized,
			"no user in context",
		)
		return
	}
	user, err := h.authService.Me(ctx, userID)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get user",
		)
	}

	response := RegisterResponse{
		ID:          userID,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}

	responseHandler.JSONResponse(response, http.StatusOK)
}
