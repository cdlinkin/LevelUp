package statistics_service

import (
	"math"
	"testing"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
)

var testCreatedAt = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func notCompleted() domain.Task {
	return domain.Task{
		Title:     "task",
		CreatedAt: testCreatedAt,
	}
}

func completedIn(d time.Duration) domain.Task {
	completedAt := testCreatedAt.Add(d)

	return domain.Task{
		Title:       "task",
		Completed:   true,
		CreatedAt:   testCreatedAt,
		CompletedAt: &completedAt,
	}
}

func ptr[T any](v T) *T { return &v }

func assertRate(t *testing.T, got, want *float64) {
	t.Helper()

	if got == nil || want == nil {
		if got != want {
			t.Errorf("TasksCompletedRate = %v, want %v", got, want)
		}
		return
	}

	if math.Abs(*got-*want) > 1e-9 {
		t.Errorf("TasksCompletedRate = %v, want %v", *got, *want)
	}
}

func assertDuration(t *testing.T, got, want *time.Duration) {
	t.Helper()

	if got == nil || want == nil {
		if got != want {
			t.Errorf("TasksAverageCompletionTime = %v, want %v", got, want)
		}
		return
	}

	if *got != *want {
		t.Errorf("TasksAverageCompletionTime = %v, want %v", *got, *want)
	}
}

func TestCalcStatistics(t *testing.T) {
	tests := []struct {
		name          string
		tasks         []domain.Task
		wantCreated   int
		wantCompleted int
		wantRate      *float64
		wantAvg       *time.Duration
	}{
		{
			name:          "empty list",
			tasks:         nil,
			wantCreated:   0,
			wantCompleted: 0,
			wantRate:      nil,
			wantAvg:       nil,
		},
		{
			name: "one not completed",
			tasks: []domain.Task{
				notCompleted(),
			},
			wantCreated:   1,
			wantCompleted: 0,
			wantRate:      ptr(0.0),
			wantAvg:       nil,
		},
		{
			name: "one completed in 2h and one not completed",
			tasks: []domain.Task{
				completedIn(2 * time.Hour),
				notCompleted(),
			},
			wantCreated:   2,
			wantCompleted: 1,
			wantRate:      ptr(50.0),
			wantAvg:       ptr(2 * time.Hour),
		},
		{
			name: "two completed in 1h and 3h",
			tasks: []domain.Task{
				completedIn(time.Hour),
				completedIn(3 * time.Hour),
			},
			wantCreated:   2,
			wantCompleted: 2,
			wantRate:      ptr(100.0),
			wantAvg:       ptr(2 * time.Hour),
		},
		{
			name: "one of three completed",
			tasks: []domain.Task{
				completedIn(30 * time.Minute),
				notCompleted(),
				notCompleted(),
			},
			wantCreated:   3,
			wantCompleted: 1,
			wantRate:      ptr(100.0 / 3),
			wantAvg:       ptr(30 * time.Minute),
		},
		{
			name: "completed instantly",
			tasks: []domain.Task{
				completedIn(0),
			},
			wantCreated:   1,
			wantCompleted: 1,
			wantRate:      ptr(100.0),
			wantAvg:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcStatistics(tt.tasks)

			if got.TasksCreated != tt.wantCreated {
				t.Errorf("TasksCreated = %d, want %d", got.TasksCreated, tt.wantCreated)
			}
			if got.TasksCompleted != tt.wantCompleted {
				t.Errorf("TasksCompleted = %d, want %d", got.TasksCompleted, tt.wantCompleted)
			}
			assertRate(t, got.TasksCompletedRate, tt.wantRate)
			assertDuration(t, got.TasksAverageCompletionTime, tt.wantAvg)
		})
	}
}
