package users_service

import (
	"context"
	"fmt"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
)

func (s *UsersService) GetUser(
	ctx context.Context,
	currentUserID int,
	id int,
) (domain.User, error) {
	if currentUserID != id {
		return domain.User{}, fmt.Errorf(
			"user %d cannot view user %d: %w",
			currentUserID,
			id,
			core_errors.ErrForbidden,
		)
	}

	user, err := s.usersRepository.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user from repository: %w", err)
	}

	return user, nil
}
