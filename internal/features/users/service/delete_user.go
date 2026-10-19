package users_service

import (
	"context"
	"fmt"

	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
)

func (s *UsersService) DeleteUser(
	ctx context.Context,
	currentUserID int,
	id int,
) error {
	if currentUserID != id {
		return fmt.Errorf(
			"user %d cannot modify user %d: %w",
			currentUserID,
			id,
			core_errors.ErrForbidden,
		)
	}

	if err := s.usersRepository.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
