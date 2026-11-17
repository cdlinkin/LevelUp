package tasks_service

import (
	"context"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	"go.uber.org/zap"
)

type TasksService struct {
	tasksRepository       TasksRepository
	statisticsInvalidator StatisticsInvalidator
}

type TasksRepository interface {
	CreateTask(
		ctx context.Context,
		task domain.Task,
	) (domain.Task, error)

	GetTasks(
		ctx context.Context,
		userID *int,
		limit *int,
		offset *int,
	) ([]domain.Task, error)

	GetTask(
		ctx context.Context,
		id int,
	) (domain.Task, error)

	DeleteTask(
		ctx context.Context,
		id int,
	) error

	PatchTask(
		ctx context.Context,
		id int,
		patch domain.Task,
	) (domain.Task, error)
}

type StatisticsInvalidator interface {
	InvalidateStatistics(
		ctx context.Context,
		userID int,
	) error
}

func NewTasksService(
	tasksRepository TasksRepository,
	statisticsInvalidator StatisticsInvalidator,
) *TasksService {
	return &TasksService{
		tasksRepository:       tasksRepository,
		statisticsInvalidator: statisticsInvalidator,
	}
}

func (s *TasksService) invalidateStatistics(ctx context.Context, userID int) {
	if err := s.statisticsInvalidator.InvalidateStatistics(ctx, userID); err != nil {
		log, ok := core_logger.TryFromContext(ctx)
		if !ok {
			return
		}

		log.Error("invalidate statistics", zap.Error(err))
	}
}
