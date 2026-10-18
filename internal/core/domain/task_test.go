package domain

import (
	"errors"
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
