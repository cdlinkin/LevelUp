package users_service

import (
	"context"
	"fmt"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
)

func (s *UsersService) PatchUser(ctx context.Context, userID int, id int, patch domain.UserPatch) (domain.User, error) {
	if userID != id {
		return domain.User{}, fmt.Errorf(
			"user %d cannot modify user %d: %w",
			userID,
			id,
			core_errors.ErrForbidden,
		)
	}
	user, err := s.usersRepository.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}

	if err := user.ApplyPatch(patch); err != nil {
		return domain.User{}, fmt.Errorf("apply user patch: %w", err)
	}

	patchedUser, err := s.usersRepository.PatchUser(ctx, id, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("patched user: %w", err)
	}

	return patchedUser, nil
}
