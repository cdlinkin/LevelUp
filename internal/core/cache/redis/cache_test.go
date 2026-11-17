package core_redis_cache

import (
	"testing"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
)

func TestNewConfigRejectsNonPositiveTTL(t *testing.T) {
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("REDIS_TIMEOUT", "2s")
	t.Setenv("REDIS_TTL", "0s")

	_, err := NewConfig()
	if err == nil {
		t.Fatal("NewConfig() error = nil, want error")
	}
}

func TestStatisticsKey(t *testing.T) {
	from := time.Date(2026, 1, 2, 1, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	to := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	got := statisticsKey(7, 3, &from, &to)
	want := "stats:7:3:2026-01-01:2026-02-01"
	if got != want {
		t.Errorf("statisticsKey() = %q, want %q", got, want)
	}

	got = statisticsKey(7, 0, nil, nil)
	want = "stats:7:0:all:all"
	if got != want {
		t.Errorf("statisticsKey() = %q, want %q", got, want)
	}

	if versionKey(7) != "stats:ver:7" {
		t.Errorf("versionKey() = %q, want %q", versionKey(7), "stats:ver:7")
	}
}

func TestStatisticsPayloadRoundTrip(t *testing.T) {
	average := 2 * time.Hour
	rate := 50.0
	original := domain.NewStatistics(2, 1, &rate, &average)

	payload, err := payloadFromStatistics(original)
	if err != nil {
		t.Fatalf("payloadFromStatistics() error = %v", err)
	}

	got, err := statisticsDomainFromPayload(payload)
	if err != nil {
		t.Fatalf("statisticsDomainFromPayload() error = %v", err)
	}

	if got.TasksCreated != 2 || got.TasksCompleted != 1 {
		t.Errorf("counts = %d/%d, want 2/1", got.TasksCreated, got.TasksCompleted)
	}
	if got.TasksCompletedRate == nil || *got.TasksCompletedRate != rate {
		t.Errorf("TasksCompletedRate = %v, want %v", got.TasksCompletedRate, rate)
	}
	if got.TasksAverageCompletionTime == nil || *got.TasksAverageCompletionTime != average {
		t.Errorf("TasksAverageCompletionTime = %v, want %v", got.TasksAverageCompletionTime, average)
	}
}

func TestStatisticsPayloadKeepsNilPointers(t *testing.T) {
	original := domain.NewStatistics(0, 0, nil, nil)

	payload, err := payloadFromStatistics(original)
	if err != nil {
		t.Fatalf("payloadFromStatistics() error = %v", err)
	}

	got, err := statisticsDomainFromPayload(payload)
	if err != nil {
		t.Fatalf("statisticsDomainFromPayload() error = %v", err)
	}

	if got.TasksCompletedRate != nil {
		t.Errorf("TasksCompletedRate = %v, want nil", got.TasksCompletedRate)
	}
	if got.TasksAverageCompletionTime != nil {
		t.Errorf("TasksAverageCompletionTime = %v, want nil", got.TasksAverageCompletionTime)
	}
}

func TestStatisticsPayloadRejectsGarbage(t *testing.T) {
	_, err := statisticsDomainFromPayload([]byte("{"))
	if err == nil {
		t.Fatal("statisticsDomainFromPayload() error = nil, want error")
	}
}
