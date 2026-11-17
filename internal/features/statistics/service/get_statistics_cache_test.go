package statistics_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
)

type fakeStatisticsRepository struct {
	tasks  []domain.Task
	err    error
	called int
}

func (f *fakeStatisticsRepository) GetTasks(
	ctx context.Context,
	userID *int,
	from *time.Time,
	to *time.Time,
) ([]domain.Task, error) {
	f.called++
	return f.tasks, f.err
}

type fakeStatisticsCache struct {
	statistics domain.Statistics
	hit        bool
	getErr     error
	setErr     error
	getCalled  int
	setCalled  int
	setValue   domain.Statistics
}

func (f *fakeStatisticsCache) GetStatistics(
	ctx context.Context,
	userID int,
	from *time.Time,
	to *time.Time,
) (domain.Statistics, bool, error) {
	f.getCalled++
	return f.statistics, f.hit, f.getErr
}

func (f *fakeStatisticsCache) SetStatistics(
	ctx context.Context,
	userID int,
	from *time.Time,
	to *time.Time,
	statistics domain.Statistics,
) error {
	f.setCalled++
	f.setValue = statistics
	return f.setErr
}

func TestGetStatistics_CacheHitSkipsRepository(t *testing.T) {
	userID := 4
	cached := domain.NewStatistics(9, 3, ptr(33.0), nil)
	repo := &fakeStatisticsRepository{
		tasks: []domain.Task{notCompleted()},
	}
	cache := &fakeStatisticsCache{
		statistics: cached,
		hit:        true,
	}
	service := NewStatisticsService(repo, cache)

	got, err := service.GetStatistics(context.Background(), &userID, nil, nil)
	if err != nil {
		t.Fatalf("GetStatistics() error = %v, want nil", err)
	}
	if repo.called != 0 {
		t.Errorf("repository called %d times, want 0", repo.called)
	}
	if cache.setCalled != 0 {
		t.Errorf("cache set called %d times, want 0", cache.setCalled)
	}
	if got.TasksCreated != 9 || got.TasksCompleted != 3 {
		t.Errorf("statistics = %d/%d, want 9/3", got.TasksCreated, got.TasksCompleted)
	}
}

func TestGetStatistics_CacheMissStoresCalculation(t *testing.T) {
	userID := 4
	repo := &fakeStatisticsRepository{
		tasks: []domain.Task{
			completedIn(2 * time.Hour),
			notCompleted(),
		},
	}
	cache := &fakeStatisticsCache{}
	service := NewStatisticsService(repo, cache)

	got, err := service.GetStatistics(context.Background(), &userID, nil, nil)
	if err != nil {
		t.Fatalf("GetStatistics() error = %v, want nil", err)
	}
	if repo.called != 1 {
		t.Errorf("repository called %d times, want 1", repo.called)
	}
	if cache.setCalled != 1 {
		t.Fatalf("cache set called %d times, want 1", cache.setCalled)
	}
	if got.TasksCreated != 2 || cache.setValue.TasksCreated != 2 {
		t.Errorf("stored TasksCreated = %d, want 2", cache.setValue.TasksCreated)
	}
	if cache.setValue.TasksCompleted != 1 {
		t.Errorf("stored TasksCompleted = %d, want 1", cache.setValue.TasksCompleted)
	}
}

func TestGetStatistics_EmptyResultIsStored(t *testing.T) {
	userID := 4
	repo := &fakeStatisticsRepository{}
	cache := &fakeStatisticsCache{}
	service := NewStatisticsService(repo, cache)

	got, err := service.GetStatistics(context.Background(), &userID, nil, nil)
	if err != nil {
		t.Fatalf("GetStatistics() error = %v, want nil", err)
	}
	if got.TasksCreated != 0 || cache.setCalled != 1 {
		t.Errorf("empty statistics setCalled = %d, created = %d, want 1, 0", cache.setCalled, got.TasksCreated)
	}
}

func TestGetStatistics_CacheGetErrorFallsThrough(t *testing.T) {
	userID := 4
	repo := &fakeStatisticsRepository{
		tasks: []domain.Task{notCompleted()},
	}
	cache := &fakeStatisticsCache{
		getErr: errors.New("redis down"),
	}
	service := NewStatisticsService(repo, cache)

	got, err := service.GetStatistics(context.Background(), &userID, nil, nil)
	if err != nil {
		t.Fatalf("GetStatistics() error = %v, want nil", err)
	}
	if repo.called != 1 {
		t.Errorf("repository called %d times, want 1", repo.called)
	}
	if got.TasksCreated != 1 {
		t.Errorf("TasksCreated = %d, want 1", got.TasksCreated)
	}
	if cache.setCalled != 1 {
		t.Errorf("cache set called %d times, want 1", cache.setCalled)
	}
}

func TestGetStatistics_CacheSetErrorStillReturnsStatistics(t *testing.T) {
	userID := 4
	repo := &fakeStatisticsRepository{
		tasks: []domain.Task{notCompleted()},
	}
	cache := &fakeStatisticsCache{
		setErr: errors.New("redis down"),
	}
	service := NewStatisticsService(repo, cache)

	got, err := service.GetStatistics(context.Background(), &userID, nil, nil)
	if err != nil {
		t.Fatalf("GetStatistics() error = %v, want nil", err)
	}
	if got.TasksCreated != 1 {
		t.Errorf("TasksCreated = %d, want 1", got.TasksCreated)
	}
}

func TestGetStatistics_NilUserSkipsCache(t *testing.T) {
	repo := &fakeStatisticsRepository{
		tasks: []domain.Task{notCompleted()},
	}
	cache := &fakeStatisticsCache{hit: true}
	service := NewStatisticsService(repo, cache)

	got, err := service.GetStatistics(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("GetStatistics() error = %v, want nil", err)
	}
	if cache.getCalled != 0 || cache.setCalled != 0 {
		t.Errorf("cache get/set = %d/%d, want 0/0", cache.getCalled, cache.setCalled)
	}
	if repo.called != 1 || got.TasksCreated != 1 {
		t.Errorf("repository called %d, created %d, want 1, 1", repo.called, got.TasksCreated)
	}
}

func TestGetStatistics_InvalidPeriodSkipsCache(t *testing.T) {
	userID := 4
	from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	to := from
	repo := &fakeStatisticsRepository{}
	cache := &fakeStatisticsCache{}
	service := NewStatisticsService(repo, cache)

	_, err := service.GetStatistics(context.Background(), &userID, &from, &to)
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Errorf("GetStatistics() error = %v, want %v", err, core_errors.ErrInvalidArgument)
	}
	if cache.getCalled != 0 || repo.called != 0 {
		t.Errorf("cache/repository called %d/%d, want 0/0", cache.getCalled, repo.called)
	}
}

func TestGetStatistics_RepositoryError(t *testing.T) {
	userID := 4
	repo := &fakeStatisticsRepository{err: errors.New("db down")}
	cache := &fakeStatisticsCache{}
	service := NewStatisticsService(repo, cache)

	_, err := service.GetStatistics(context.Background(), &userID, nil, nil)
	if err == nil {
		t.Fatal("GetStatistics() error = nil, want error")
	}
	if cache.setCalled != 0 {
		t.Errorf("cache set called %d times, want 0", cache.setCalled)
	}
}
