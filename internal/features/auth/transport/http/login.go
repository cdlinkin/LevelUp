package auth_transport_http

import (
	"net/http"

	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_http_request "github.com/cdlinkin/LevelUp/internal/core/transport/http/request"
	core_http_response "github.com/cdlinkin/LevelUp/internal/core/transport/http/response"
)

// Login 			godoc
// @Summary 		Вход в систему
// @Description 	Проверяет email и пароль, возвращает JWT access token.
// @Description 	Токен передаётся в остальные запросы заголовком `Authorization: Bearer <access_token>`.
// @Description 	На неизвестный email и неверный пароль ответ одинаковый: 401.
// @Tags 			auth
// @Accept 			json
// @Produce 		json
// @Param 			request body LoginRequest true "Login тело запроса"
// @Success 		200 {object} LoginResponse "Успешный вход"
// @Failure 		400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 		401 {object} core_http_response.ErrorResponse "Invalid email or password"
// @Failure 		500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router 			/auth/login [post]
func (h *AuthHTTPHandler) Login(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request LoginRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)
		return
	}

	token, ttl, err := h.authService.Login(
		ctx,
		request.Email,
		request.Password,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)
		return
	}

	response := LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(ttl.Seconds()),
	}

	responseHandler.JSONResponse(response, http.StatusOK)
}
