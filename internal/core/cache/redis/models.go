package core_redis_cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
)

type statisticsModel struct {
	TasksCreated               int      `json:"tasks_created"`
	TasksCompleted             int      `json:"tasks_completed"`
	TasksCompletedRate         *float64 `json:"tasks_completed_rate"`
	TasksAverageCompletionTime *int64   `json:"tasks_average_completion_time_ns"`
}

func statisticsModelFromDomain(statistics domain.Statistics) statisticsModel {
	model := statisticsModel{
		TasksCreated:       statistics.TasksCreated,
		TasksCompleted:     statistics.TasksCompleted,
		TasksCompletedRate: statistics.TasksCompletedRate,
	}

	if statistics.TasksAverageCompletionTime != nil {
		nanoseconds := statistics.TasksAverageCompletionTime.Nanoseconds()
		model.TasksAverageCompletionTime = &nanoseconds
	}

	return model
}

func statisticsDomainFromModel(model statisticsModel) domain.Statistics {
	var average *time.Duration
	if model.TasksAverageCompletionTime != nil {
		duration := time.Duration(*model.TasksAverageCompletionTime)
		average = &duration
	}

	return domain.NewStatistics(
		model.TasksCreated,
		model.TasksCompleted,
		model.TasksCompletedRate,
		average,
	)
}

func payloadFromStatistics(statistics domain.Statistics) ([]byte, error) {
	payload, err := json.Marshal(statisticsModelFromDomain(statistics))
	if err != nil {
		return nil, fmt.Errorf("marshal statistics: %w", err)
	}

	return payload, nil
}

func statisticsDomainFromPayload(payload []byte) (domain.Statistics, error) {
	var model statisticsModel
	if err := json.Unmarshal(payload, &model); err != nil {
		return domain.Statistics{}, fmt.Errorf("unmarshal statistics: %w", err)
	}

	return statisticsDomainFromModel(model), nil
}
