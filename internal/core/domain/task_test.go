package domain

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
)

func ptr[T any](v T) *T { return &v }

func TestTaskValidate(t *testing.T) {
	created := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	completedAtUndone := time.Date(2025, 5, 2, 1, 1, 1, 1, time.UTC)
	completedAtDone := time.Date(2026, 2, 3, 12, 0, 0, 0, time.UTC)

	testTasks := []struct {
		name    string
		task    Task
		wantErr error
	}{
		{
			name: "valid not completed task",
			task: Task{
				Title:     "do homework",
				CreatedAt: created,
			},
			wantErr: nil,
		},
		{
			name: "valid completed task",
			task: Task{
				Title:       "go to work",
				CreatedAt:   created,
				Completed:   true,
				CompletedAt: &completedAtDone,
			},
			wantErr: nil,
		},
		{
			name: "'title' is empty",
			task: Task{
				Title:     "",
				CreatedAt: created,
			},
			wantErr: core_errors.ErrInvalidArgument,
		},
		{
			name: "'title': len=100",
			task: Task{
				Title:     strings.Repeat("a", 100),
				CreatedAt: created,
			},
			wantErr: nil,
		},
		{
			name: "'title': len=101",
			task: Task{
				Title:     strings.Repeat("b", 101),
				CreatedAt: created,
			},
			wantErr: core_errors.ErrInvalidArgument,
		},
		{
			name: "'description' is empty",
			task: Task{
				Title:       "try",
				Description: ptr(""),
				CreatedAt:   created,
			},
			wantErr: core_errors.ErrInvalidArgument,
		},
		{
			name: "Comleted=true, CompletedAt=nil",
			task: Task{
				Title:       "try",
				Completed:   true,
				CompletedAt: nil,
				CreatedAt:   created,
			},
			wantErr: core_errors.ErrInvalidArgument,
		},
		{
			name: "CompletedAt before CreatedAt",
			task: Task{
				Title:       "try",
				Completed:   true,
				CompletedAt: &completedAtUndone,
				CreatedAt:   created,
			},
			wantErr: core_errors.ErrInvalidArgument,
		},
		{
			name: "Completed=false, CompletedAt",
			task: Task{
				Title:       "try",
				Completed:   false,
				CompletedAt: &completedAtDone,
			},
			wantErr: core_errors.ErrInvalidArgument,
		},
	}

	for _, tt := range testTasks {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("'validate()' error := %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskApplyPatch(t *testing.T) {
	created := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	baseTask := Task{
		ID:          1,
		Version:     1,
		Title:       "do homework",
		Description: ptr("golang"),
		CreatedAt:   created,
	}

	testTasks := []struct {
		name     string
		patch    TaskPatch
		wantTask Task
		wantErr  error
	}{
		{
			name:     "patch is empty",
			patch:    TaskPatch{},
			wantTask: baseTask,
			wantErr:  nil,
		},
		{
			name: "change 'title'",
			patch: TaskPatch{
				Title: Nullable[string]{
					Value: ptr("Hello"),
					Set:   true,
				},
			},
			wantTask: Task{
				ID:          1,
				Version:     1,
				Title:       "Hello",
				Description: ptr("golang"),
				CreatedAt:   created,
			},
			wantErr: nil,
		},
		{
			name: "'description' is empty",
			patch: TaskPatch{
				Description: Nullable[string]{
					Value: nil,
					Set:   true,
				},
			},
			wantTask: Task{
				ID:          1,
				Version:     1,
				Title:       "do homework",
				Description: nil,
				CreatedAt:   created,
			},
			wantErr: nil,
		},
		{
			name: "title is empty",
			patch: TaskPatch{
				Title: Nullable[string]{
					Value: ptr(""),
					Set:   true,
				},
			},
			wantTask: baseTask,
			wantErr:  core_errors.ErrInvalidArgument,
		},
		{
			name: "'title' len=101",
			patch: TaskPatch{
				Title: Nullable[string]{
					Value: ptr(strings.Repeat("a", 101)),
					Set:   true,
				},
			},
			wantTask: baseTask,
			wantErr:  core_errors.ErrInvalidArgument,
		},
		{
			name: "'completed' null",
			patch: TaskPatch{
				Completed: Nullable[bool]{
					Value: nil,
					Set:   true,
				},
			},
			wantTask: baseTask,
			wantErr:  core_errors.ErrInvalidArgument,
		},
		{
			name: "'title' is valid and 'description' is invalid",
			patch: TaskPatch{
				Title: Nullable[string]{
					Value: ptr("go to walk"),
					Set:   true,
				},
				Description: Nullable[string]{
					Value: ptr(""),
					Set:   true,
				},
			},
			wantTask: baseTask,
			wantErr:  core_errors.ErrInvalidArgument,
		},
	}

	for _, tt := range testTasks {
		t.Run(tt.name, func(t *testing.T) {
			task := baseTask

			err := task.ApplyPatch(tt.patch)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("'ApplyPatch()' error = %v: want %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(task, tt.wantTask) {
				t.Errorf("task after 'ApplyPatch()':\n got: %+v\nwant: %+v", task, tt.wantTask)
			}
		})
	}
}

func TestTaskApplyPatchComplete(t *testing.T) {
	task := Task{
		Title:     "do homework",
		CreatedAt: time.Now().Add(-time.Hour),
	}

	before := time.Now()
	err := task.ApplyPatch(TaskPatch{
		Completed: Nullable[bool]{
			Value: ptr(true),
			Set:   true,
		},
	})
	after := time.Now()

	if err != nil {
		t.Fatalf("complete: 'ApplyPatch()' error = %v, want nil", err)
	}
	if !task.Completed {
		t.Errorf("complete: 'Completed' = false, want true")
	}
	if task.CompletedAt == nil {
		t.Fatalf("complete: 'CompletedAt' = nil, want time between %v and %v",
			before,
			after,
		)
	}
	if task.CompletedAt.Before(before) || task.CompletedAt.After(after) {
		t.Errorf("complete: CompletedAt = %v, want between %v and %v",
			*task.CompletedAt,
			before,
			after,
		)
	}

	err = task.ApplyPatch(TaskPatch{
		Completed: Nullable[bool]{
			Value: ptr(false),
			Set:   true,
		},
	})
	if err != nil {
		t.Fatalf("reopen: ApplyPatch() error = %v, want nil", err)
	}
	if task.Completed {
		t.Errorf("reopen: Completed = true, want false")
	}
	if task.CompletedAt != nil {
		t.Errorf("reopen: CompletedAt = %v, want nil", *task.CompletedAt)
	}
}
