package auth_service

import (
	"context"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
)

type AuthService struct {
	usersRepository UsersRepository
	tokenManager    TokenManager
}

type TokenManager interface {
	Issue(userID int) (string, error)
	TTL() time.Duration
}

type UsersRepository interface {
	CreateUser(
		ctx context.Context,
		user domain.User,
	) (domain.User, error)
	GetUserByEmail(
		ctx context.Context,
		email string,
	) (domain.User, error)
}

func NewAuthService(
	usersRepository UsersRepository,
	tokenManager TokenManager,
) *AuthService {
	return &AuthService{
		usersRepository: usersRepository,
		tokenManager:    tokenManager,
	}
}
