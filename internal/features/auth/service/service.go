package auth_service

import (
	"context"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
)

type AuthService struct {
	usersRepository UsersRepository
}

type UsersRepository interface {
	CreateUser(
		ctx context.Context,
		user domain.User,
	) (domain.User, error)
}

func NewAuthService(
	usersRepository UsersRepository,
) *AuthService {
	return &AuthService{
		usersRepository: usersRepository,
	}
}
