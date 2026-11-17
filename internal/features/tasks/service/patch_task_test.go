package tasks_service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	tasks_service "github.com/cdlinkin/LevelUp/internal/features/tasks/service"
)

type fakeTasksRepository struct {
	getTaskResult domain.Task
	getTaskErr    error

	patchTaskErr    error
	patchTaskCalled bool
	patchTaskArg    domain.Task

	getTasksCalled bool

	deleteTaskCalled bool
}

func (f *fakeTasksRepository) CreateTask(ctx context.Context, task domain.Task) (domain.Task, error) {
	return task, nil
}

func (f *fakeTasksRepository) GetTasks(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error) {
	f.getTasksCalled = true
	return nil, nil
}

func (f *fakeTasksRepository) GetTask(ctx context.Context, id int) (domain.Task, error) {
	return f.getTaskResult, f.getTaskErr
}

func (f *fakeTasksRepository) DeleteTask(ctx context.Context, id int) error {
	f.deleteTaskCalled = true
	return nil
}

type fakeStatisticsInvalidator struct {
	called bool
	userID int
	err    error
}

func (f *fakeStatisticsInvalidator) InvalidateStatistics(ctx context.Context, userID int) error {
	f.called = true
	f.userID = userID
	return f.err
}

func newTasksService(repo *fakeTasksRepository) (*tasks_service.TasksService, *fakeStatisticsInvalidator) {
	invalidator := &fakeStatisticsInvalidator{}
	return tasks_service.NewTasksService(repo, invalidator), invalidator
}

func (f *fakeTasksRepository) PatchTask(ctx context.Context, id int, task domain.Task) (domain.Task, error) {
	f.patchTaskCalled = true
	f.patchTaskArg = task

	if f.patchTaskErr != nil {
		return domain.Task{}, f.patchTaskErr
	}
	return task, nil
}

func ptr[T any](v T) *T { return &v }

func existingTask() domain.Task {
	return domain.Task{
		ID:           1,
		Version:      1,
		Title:        "do homework",
		CreatedAt:    time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		AuthorUserID: 1,
	}
}

func TestPatchTask_NotFound(t *testing.T) {
	repo := &fakeTasksRepository{
		getTaskErr: fmt.Errorf("task with id=1: %w", core_errors.ErrNotFound),
	}
	service, invalidator := newTasksService(repo)

	patch := domain.TaskPatch{
		Title: domain.Nullable[string]{Value: ptr("new title"), Set: true},
	}

	_, err := service.PatchTask(context.Background(), 1, 1, patch)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Errorf("PatchTask() error = %v, want %v", err, core_errors.ErrNotFound)
	}
	if repo.patchTaskCalled {
		t.Errorf("repository PatchTask was called, want not called")
	}
	if invalidator.called {
		t.Errorf("statistics were invalidated, want not invalidated")
	}
}

func TestPatchTask_InvalidPatch(t *testing.T) {
	repo := &fakeTasksRepository{
		getTaskResult: existingTask(),
	}
	service, invalidator := newTasksService(repo)

	patch := domain.TaskPatch{
		Title: domain.Nullable[string]{Value: nil, Set: true},
	}

	_, err := service.PatchTask(context.Background(), 1, 1, patch)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Errorf("PatchTask() error = %v, want %v", err, core_errors.ErrInvalidArgument)
	}
	if repo.patchTaskCalled {
		t.Errorf("repository PatchTask was called with invalid data, want not called")
	}
	if invalidator.called {
		t.Errorf("statistics were invalidated, want not invalidated")
	}
}

func TestPatchTask_Success(t *testing.T) {
	repo := &fakeTasksRepository{
		getTaskResult: existingTask(),
	}
	service, invalidator := newTasksService(repo)

	patch := domain.TaskPatch{
		Title: domain.Nullable[string]{Value: ptr("new title"), Set: true},
	}

	got, err := service.PatchTask(context.Background(), 1, 1, patch)
	if err != nil {
		t.Fatalf("PatchTask() error = %v, want nil", err)
	}
	if !repo.patchTaskCalled {
		t.Fatalf("repository PatchTask was not called, want called")
	}
	if repo.patchTaskArg.Title != "new title" {
		t.Errorf("repository PatchTask got Title = %q, want %q", repo.patchTaskArg.Title, "new title")
	}
	if got.Title != "new title" {
		t.Errorf("PatchTask() Title = %q, want %q", got.Title, "new title")
	}
	if !invalidator.called || invalidator.userID != 1 {
		t.Errorf("InvalidateStatistics called = %v, userID = %d, want true, 1", invalidator.called, invalidator.userID)
	}
}

func TestPatchTask_Conflict(t *testing.T) {
	repo := &fakeTasksRepository{
		getTaskResult: existingTask(),
		patchTaskErr:  fmt.Errorf("task with id=1 concurrently accessed: %w", core_errors.ErrConflict),
	}
	service, invalidator := newTasksService(repo)

	patch := domain.TaskPatch{
		Title: domain.Nullable[string]{Value: ptr("new title"), Set: true},
	}

	_, err := service.PatchTask(context.Background(), 1, 1, patch)

	if !errors.Is(err, core_errors.ErrConflict) {
		t.Errorf("PatchTask() error = %v, want %v", err, core_errors.ErrConflict)
	}
	if invalidator.called {
		t.Errorf("statistics were invalidated, want not invalidated")
	}
}

func TestGetTasks_NegativeLimit(t *testing.T) {
	repo := &fakeTasksRepository{}
	service, invalidator := newTasksService(repo)

	_, err := service.GetTasks(context.Background(), nil, ptr(-1), nil)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Errorf("GetTasks() error = %v, want %v", err, core_errors.ErrInvalidArgument)
	}
	if repo.getTasksCalled {
		t.Errorf("repository GetTasks was called, want not called")
	}
	if invalidator.called {
		t.Errorf("statistics were invalidated, want not invalidated")
	}
}

func TestPatchTask_OtherUsersTask(t *testing.T) {
	repo := &fakeTasksRepository{
		getTaskResult: existingTask(),
	}
	service, invalidator := newTasksService(repo)

	patch := domain.TaskPatch{
		Title: domain.Nullable[string]{Value: ptr("new title"), Set: true},
	}

	_, err := service.PatchTask(context.Background(), 2, 1, patch)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Errorf("PatchTask() error = %v, want %v", err, core_errors.ErrNotFound)
	}
	if repo.patchTaskCalled {
		t.Errorf("repository PatchTask was called for other user's task, want not called")
	}
	if invalidator.called {
		t.Errorf("statistics were invalidated, want not invalidated")
	}
}

func TestDeleteTask_OtherUsersTask(t *testing.T) {
	repo := &fakeTasksRepository{
		getTaskResult: existingTask(),
	}
	service, invalidator := newTasksService(repo)

	err := service.DeleteTask(context.Background(), 2, 1)

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Errorf("DeleteTask() error = %v, want %v", err, core_errors.ErrNotFound)
	}
	if repo.deleteTaskCalled {
		t.Errorf("repository DeleteTask was called for other user's task, want not called")
	}
	if invalidator.called {
		t.Errorf("statistics were invalidated, want not invalidated")
	}
}

func TestCreateTask_InvalidatesStatistics(t *testing.T) {
	repo := &fakeTasksRepository{}
	service, invalidator := newTasksService(repo)

	_, err := service.CreateTask(context.Background(), domain.NewTaskUninitialized("title", nil, 7))
	if err != nil {
		t.Fatalf("CreateTask() error = %v, want nil", err)
	}
	if !invalidator.called || invalidator.userID != 7 {
		t.Errorf("InvalidateStatistics called = %v, userID = %d, want true, 7", invalidator.called, invalidator.userID)
	}
}

func TestCreateTask_InvalidDoesNotInvalidate(t *testing.T) {
	repo := &fakeTasksRepository{}
	service, invalidator := newTasksService(repo)

	_, err := service.CreateTask(context.Background(), domain.NewTaskUninitialized("", nil, 7))
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Errorf("CreateTask() error = %v, want %v", err, core_errors.ErrInvalidArgument)
	}
	if invalidator.called {
		t.Errorf("statistics were invalidated, want not invalidated")
	}
}

func TestDeleteTask_InvalidatesStatistics(t *testing.T) {
	repo := &fakeTasksRepository{
		getTaskResult: existingTask(),
	}
	service, invalidator := newTasksService(repo)

	err := service.DeleteTask(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v, want nil", err)
	}
	if !repo.deleteTaskCalled {
		t.Fatalf("repository DeleteTask was not called, want called")
	}
	if !invalidator.called || invalidator.userID != 1 {
		t.Errorf("InvalidateStatistics called = %v, userID = %d, want true, 1", invalidator.called, invalidator.userID)
	}
}

func TestPatchTask_InvalidateErrorDoesNotFail(t *testing.T) {
	repo := &fakeTasksRepository{
		getTaskResult: existingTask(),
	}
	service, invalidator := newTasksService(repo)
	invalidator.err = errors.New("redis down")

	patch := domain.TaskPatch{
		Title: domain.Nullable[string]{Value: ptr("new title"), Set: true},
	}

	got, err := service.PatchTask(context.Background(), 1, 1, patch)
	if err != nil {
		t.Fatalf("PatchTask() error = %v, want nil", err)
	}
	if got.Title != "new title" {
		t.Errorf("PatchTask() Title = %q, want %q", got.Title, "new title")
	}
}
