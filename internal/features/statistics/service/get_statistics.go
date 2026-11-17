package statistics_service

import (
	"context"
	"fmt"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	"go.uber.org/zap"
)

func (s *StatisticsService) GetStatistics(
	ctx context.Context,
	userID *int,
	from *time.Time,
	to *time.Time,
) (domain.Statistics, error) {
	if from != nil && to != nil {
		if to.Before(*from) || to.Equal(*from) {
			return domain.Statistics{}, fmt.Errorf("'to' must be after 'from': %w", core_errors.ErrInvalidArgument)
		}
	}

	if userID != nil {
		statistics, hit, err := s.statisticsCache.GetStatistics(ctx, *userID, from, to)
		if err != nil {
			logCacheError(ctx, "get statistics from cache", err)
		} else if hit {
			return statistics, nil
		}
	}

	tasks, err := s.statisticsRepository.GetTasks(ctx, userID, from, to)
	if err != nil {
		return domain.Statistics{}, fmt.Errorf("get tasks from repository: %w", err)
	}

	statistics := calcStatistics(tasks)

	if userID != nil {
		if err := s.statisticsCache.SetStatistics(ctx, *userID, from, to, statistics); err != nil {
			logCacheError(ctx, "set statistics cache", err)
		}
	}

	return statistics, nil
}

func logCacheError(ctx context.Context, message string, err error) {
	log, ok := core_logger.TryFromContext(ctx)
	if !ok {
		return
	}

	log.Error(message, zap.Error(err))
}

func calcStatistics(tasks []domain.Task) domain.Statistics {
	if len(tasks) == 0 {
		return domain.NewStatistics(0, 0, nil, nil)
	}

	tasksCreated := len(tasks)

	tasksCompleted := 0
	var totalCompletionDuration time.Duration
	for _, task := range tasks {
		if task.Completed {
			tasksCompleted++
		}

		completionDuration := task.CompletionDuration()
		if completionDuration != nil {
			totalCompletionDuration += *completionDuration
		}
	}

	tasksCompletedRate := float64(tasksCompleted) / float64(tasksCreated) * 100

	var tasksAverageCompletionTime *time.Duration
	if tasksCompleted > 0 && totalCompletionDuration != 0 {
		avg := totalCompletionDuration / time.Duration(tasksCompleted)

		tasksAverageCompletionTime = &avg
	}
	return domain.NewStatistics(tasksCreated, tasksCompleted, &tasksCompletedRate, tasksAverageCompletionTime)
}
