package tasks_service

import (
	"context"
	"fmt"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
)

func (s *TasksService) PatchTask(
	ctx context.Context,
	userID int,
	id int,
	patch domain.TaskPatch,
) (domain.Task, error) {
	task, err := s.tasksRepository.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get tasks: %w", err)
	}

	if task.AuthorUserID != userID {
		return domain.Task{}, fmt.Errorf(
			"task with id=%d: %w",
			id,
			core_errors.ErrNotFound,
		)
	}

	if err := task.ApplyPatch(patch); err != nil {
		return domain.Task{}, fmt.Errorf("apply task patch: %w", err)
	}

	patchedTask, err := s.tasksRepository.PatchTask(ctx, id, task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("patch task: %w", err)
	}

	return patchedTask, nil
}
