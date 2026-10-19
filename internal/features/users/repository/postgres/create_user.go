package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	core_postgres_pool "github.com/cdlinkin/LevelUp/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO todoapp.users (full_name, phone_number, email, password_hash)
	VALUES ($1, $2, $3, $4)
	RETURNING id, version, full_name, phone_number, email;
	`

	row := r.pool.QueryRow(ctx, query, user.FullName, user.PhoneNumber, user.Email, user.PasswordHash)

	var userModel UserModel

	err := row.Scan(
		&userModel.ID,
		&userModel.Version,
		&userModel.FullName,
		&userModel.PhoneNumber,
		&userModel.Email,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrUniqueViolation) {
			return domain.User{}, fmt.Errorf(
				"user with email=%s already exists: %w",
				user.Email,
				core_errors.ErrConflict,
			)
		}
		return domain.User{}, fmt.Errorf("scan error: %w", err)
	}

	userDomain := domain.NewUser(
		userModel.ID,
		userModel.Version,
		userModel.FullName,
		userModel.PhoneNumber,
		userModel.Email,
		userModel.PasswordHash,
	)

	return userDomain, nil
}
