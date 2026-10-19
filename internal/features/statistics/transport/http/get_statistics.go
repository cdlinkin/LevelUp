package statistics_transport_http

import (
	"fmt"
	"net/http"
	"time"

	core_auth "github.com/cdlinkin/LevelUp/internal/core/auth"
	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_http_request "github.com/cdlinkin/LevelUp/internal/core/transport/http/request"
	core_http_response "github.com/cdlinkin/LevelUp/internal/core/transport/http/response"
)

type GetStatisticsResponse struct {
	TasksCreated               int      `json:"tasks_created" example:"50"`
	TasksCompleted             int      `json:"tasks_completed" example:"10"`
	TasksCompletedRate         *float64 `json:"tasks_completed_rate" example:"20"`
	TasksAverageCompletionTime *string  `json:"tasks_average_completion_time" example:"2m30s"`
}

// GetStatistics godoc
// @Summary Получение статистики
// @Description Статистика по задачам текущего пользователя с опциональной фильтрацией по временному промежутку
// @Tags statistics
// @Produce json
// @Security BearerAuth
// @Param from query string false "Начало промежутка рассмотрения статистики (включительно), формат: YYYY-MM-DD"
// @Param to query string false "Конец промежутка рассмотрения статистики (не включительно), формат: YYYY-MM-DD"
// @Success 200 {object} GetStatisticsResponse "Успешное получение статистики"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 401 {object} core_http_response.ErrorResponse "Unauthorized"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /statistics [get]
func (h *StatisticsHTTPHandler) GetStatistics(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	from, to, err := getFromToQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get 'user_id', 'from', 'to' query param",
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

	statistics, err := h.statisticsService.GetStatistics(ctx, &userID, from, to)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get statistics",
		)
		return
	}

	response := toDTOFromDomain(statistics)

	responseHandler.JSONResponse(response, http.StatusOK)
}

func toDTOFromDomain(statistics domain.Statistics) GetStatisticsResponse {
	var avgTime *string
	if statistics.TasksAverageCompletionTime != nil {
		duration := statistics.TasksAverageCompletionTime.String()
		avgTime = &duration
	}

	return GetStatisticsResponse{
		TasksCreated:               statistics.TasksCreated,
		TasksCompleted:             statistics.TasksCompleted,
		TasksCompletedRate:         statistics.TasksCompletedRate,
		TasksAverageCompletionTime: avgTime,
	}
}

func getFromToQueryParams(r *http.Request) (*time.Time, *time.Time, error) {
	const (
		fromQueryParamKey = "from"
		toQueryParamKey   = "to"
	)

	from, err := core_http_request.GetDateQueryParam(r, fromQueryParamKey)
	if err != nil {
		return nil, nil, fmt.Errorf("get 'from' query param: %w", err)
	}

	to, err := core_http_request.GetDateQueryParam(r, toQueryParamKey)
	if err != nil {
		return nil, nil, fmt.Errorf("get 'to' query param: %w", err)
	}

	return from, to, nil
}
