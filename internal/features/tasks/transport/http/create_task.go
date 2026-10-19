package tasks_transport_http

import (
	"net/http"

	core_auth "github.com/cdlinkin/LevelUp/internal/core/auth"
	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_http_request "github.com/cdlinkin/LevelUp/internal/core/transport/http/request"
	core_http_response "github.com/cdlinkin/LevelUp/internal/core/transport/http/response"
)

type CreateTaskRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=100" example:"to go outside"`
	Description *string `json:"description" validate:"omitempty,min=1,max=1000" example:"to walk 100 km"`
}

type CreateTaskResponse TaskDTOResponse

// CreateTask godoc
// @Summary      Создать задачу
// @Description  Создать новую задачу. Автором становится пользователь из access token.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateTaskRequest true "CreateTask тело запроса"
// @Success      201 {object} CreateTaskResponse "Успешно созданная задача"
// @Failure      400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure      401 {object} core_http_response.ErrorResponse "Unauthorized"
// @Failure      500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router       /tasks [post]
func (h *TasksHTTPHandler) CreateTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request CreateTaskRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)

		return
	}

	userID, ok := core_auth.UserIDFromContext(ctx)
	if !ok {
		responseHandler.ErrorResponse(
			core_errors.ErrUnauthorized,
			"failed to get user from context",
		)
		return
	}

	taskDomain := domain.NewTaskUninitialized(
		request.Title,
		request.Description,
		userID,
	)

	taskDomain, err := h.tasksService.CreateTask(ctx, taskDomain)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to create task",
		)

		return
	}

	response := CreateTaskResponse(taskDTOFromDomain(taskDomain))

	responseHandler.JSONResponse(response, http.StatusCreated)
}
