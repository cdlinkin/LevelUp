package tasks_service

import (
	"context"
	"fmt"
)

func (s *TasksService) DeleteTask(
	ctx context.Context,
	userID int,
	id int,
) error {
	if _, err := s.GetTask(ctx, userID, id); err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	if err := s.tasksRepository.DeleteTask(ctx, id); err != nil {
		return fmt.Errorf("delete task from repository: %w", err)
	}

	return nil
}
