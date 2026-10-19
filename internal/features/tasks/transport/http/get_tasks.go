package tasks_transport_http

import (
	"fmt"
	"net/http"

	core_auth "github.com/cdlinkin/LevelUp/internal/core/auth"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_http_request "github.com/cdlinkin/LevelUp/internal/core/transport/http/request"
	core_http_response "github.com/cdlinkin/LevelUp/internal/core/transport/http/response"
)

type GetTasksResponse []TaskDTOResponse

// GetTasks		 godoc
// @Summary		 Список задач
// @Description	 Просмотр списка задач с опциональной пагинацией и/или фильтрацией по ID автора задачи
// @Tags 		tasks
// @Produce 	json
// @Param limit query int false 		"Размер страницы с задачами"
// @Param offset query int false 		"Смещение страницы с задачами"
// @Success 200 {object} GetTasksResponse 				  "Список задач"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /tasks [get]
func (h *TasksHTTPHandler) GetTasks(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	limit, offset, err := getLimitOffsetQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get userID/limit/offset query params",
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

	tasksDomains, err := h.tasksService.GetTasks(ctx, &userID, limit, offset)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get tasks",
		)
		return
	}

	response := GetTasksResponse(taskDTOsFromDomains(tasksDomains))

	responseHandler.JSONResponse(response, http.StatusOK)
}

func getLimitOffsetQueryParams(r *http.Request) (*int, *int, error) {
	const (
		limitQueryParamKey  = "limit"
		offsetQueryParamKey = "offset"
	)

	limit, err := core_http_request.GetIntQueryParam(r, limitQueryParamKey)
	if err != nil {
		return nil, nil, fmt.Errorf("get 'limit' query param: %w", err)
	}

	offset, err := core_http_request.GetIntQueryParam(r, offsetQueryParamKey)
	if err != nil {
		return nil, nil, fmt.Errorf("get 'offset' query param: %w", err)
	}

	return limit, offset, nil
}
