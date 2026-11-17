package statistics_service

import (
	"context"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
)

type StatisticsService struct {
	statisticsRepository StatisticsRepository
	statisticsCache      StatisticsCache
}

type StatisticsRepository interface {
	GetTasks(
		ctx context.Context,
		userID *int,
		from *time.Time,
		to *time.Time,
	) ([]domain.Task, error)
}

type StatisticsCache interface {
	GetStatistics(
		ctx context.Context,
		userID int,
		from *time.Time,
		to *time.Time,
	) (domain.Statistics, bool, error)

	SetStatistics(
		ctx context.Context,
		userID int,
		from *time.Time,
		to *time.Time,
		statistics domain.Statistics,
	) error
}

func NewStatisticsService(
	statisticsRepository StatisticsRepository,
	statisticsCache StatisticsCache,
) *StatisticsService {
	return &StatisticsService{
		statisticsRepository: statisticsRepository,
		statisticsCache:      statisticsCache,
	}
}
