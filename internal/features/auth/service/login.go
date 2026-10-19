package auth_service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
)

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (string, time.Duration, error) {
	email = domain.NormalizeEmail(email)

	user, err := s.usersRepository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			return "", 0, fmt.Errorf(
				"invalid email or password: %w",
				core_errors.ErrUnauthorized,
			)
		}

		return "", 0, fmt.Errorf(
			"get user by email: %w",
			err,
		)
	}

	if !user.CheckPassword(password) {
		return "", 0, fmt.Errorf(
			"invalid email or password: %w",
			core_errors.ErrUnauthorized,
		)
	}

	token, err := s.tokenManager.Issue(user.ID)
	if err != nil {
		return "", 0, fmt.Errorf(
			"issue token: %w",
			err,
		)
	}

	return token, s.tokenManager.TTL(), nil
}
