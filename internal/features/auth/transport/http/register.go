package auth_transport_http

import (
	"net/http"

	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_http_request "github.com/cdlinkin/LevelUp/internal/core/transport/http/request"
	core_http_response "github.com/cdlinkin/LevelUp/internal/core/transport/http/response"
)

// Register 		godoc
// @Summary 		Регистрация пользователя
// @Description 	Создаёт пользователя с email и паролем.
// @Description 	Email приводится к нижнему регистру, пароль хранится только в виде bcrypt-хеша.
// @Description 	Ограничения: пароль от 8 до 72 байт, `full_name` от 3 до 100 символов.
// @Tags 			auth
// @Accept 			json
// @Produce 		json
// @Param 			request body RegisterRequest true "Register тело запроса"
// @Success 		201 {object} RegisterResponse "Пользователь зарегистрирован"
// @Failure 		400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 		409 {object} core_http_response.ErrorResponse "Email already exists"
// @Failure 		500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router 			/auth/register [post]
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
