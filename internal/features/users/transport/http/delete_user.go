package users_transport_http

import (
	"net/http"

	core_auth "github.com/cdlinkin/LevelUp/internal/core/auth"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_http_request "github.com/cdlinkin/LevelUp/internal/core/transport/http/request"
	core_http_response "github.com/cdlinkin/LevelUp/internal/core/transport/http/response"
)

// DeleteUser 		godoc
// @Summary 		Удалить пользователя
// @Description 	Удаление своего аккаунта. Чужой ID возвращает 403, пользователь с задачами — 409.
// @Tags 			users
// @Security 		BearerAuth
// @Param 			id path int true 						      "ID удаляемого пользователя"
// @Success 		204 										  "Успешное удаление пользователя"
// @Failure 		400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 		401 {object} core_http_response.ErrorResponse "Unauthorized"
// @Failure 		403 {object} core_http_response.ErrorResponse "Forbidden"
// @Failure 		404 {object} core_http_response.ErrorResponse "User not found"
// @Failure 		409 {object} core_http_response.ErrorResponse "User has tasks"
// @Failure 		500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router 			/users/{id} [delete]
func (h *UsersHTTPHandler) DeleteUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	currentUserID, ok := core_auth.UserIDFromContext(ctx)
	if !ok {
		responseHandler.ErrorResponse(
			core_errors.ErrUnauthorized,
			"failed to get user from context",
		)
		return
	}

	userID, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get userID path value",
		)
		return
	}

	if err := h.usersService.DeleteUser(ctx, currentUserID, userID); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to delete user",
		)
		return
	}
	responseHandler.NoContentResponse()
}
