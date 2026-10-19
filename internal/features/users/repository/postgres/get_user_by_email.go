package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	core_postgres_pool "github.com/cdlinkin/LevelUp/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) GetUserByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, version, full_name, phone_number, email, password_hash
	FROM todoapp.users 
	WHERE email=$1;
	`

	row := r.pool.QueryRow(ctx, query, email)

	var userModel UserModel

	if err := row.Scan(
		&userModel.ID,
		&userModel.Version,
		&userModel.FullName,
		&userModel.PhoneNumber,
		&userModel.Email,
		&userModel.PasswordHash,
	); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.User{}, fmt.Errorf(
				"user with email=%s: %w",
				email,
				core_errors.ErrNotFound,
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
