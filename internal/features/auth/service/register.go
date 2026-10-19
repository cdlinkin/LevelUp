package auth_service

import (
	"context"
	"fmt"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
)

func (s *AuthService) Register(
	ctx context.Context,
	fullName string,
	phoneNumber *string,
	email string,
	password string,
) (domain.User, error) {
	email = domain.NormalizeEmail(email)
	passwordHash, err := domain.HashPassword(password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := domain.NewUserUninitialized(fullName, phoneNumber, email, passwordHash)
	if err := user.Validate(); err != nil {
		return domain.User{}, fmt.Errorf("user validate: %w", err)
	}

	createdUser, err := s.usersRepository.CreateUser(
		ctx,
		user,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("user register service: %w", err)
	}
	return createdUser, nil
}
