package domain

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	core_errors "github.com/cdlinkin/LevelUp/internal/core/errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID      int
	Version int

	FullName     string
	PhoneNumber  *string
	Email        string
	PasswordHash string
}

func NewUser(
	id int,
	version int,
	fullName string,
	phoneNumber *string,
	email string,
	passwordHash string,
) User {
	return User{
		ID:           id,
		Version:      version,
		FullName:     fullName,
		PhoneNumber:  phoneNumber,
		Email:        email,
		PasswordHash: passwordHash,
	}
}

func NewUserUninitialized(
	fullName string,
	phoneNumber *string,
	email string,
	passwordHash string,
) User {
	return NewUser(
		UninitializedID,
		UninitializedVersion,
		fullName,
		phoneNumber,
		email,
		passwordHash,
	)
}

func NormalizeEmail(email string) string {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	return email
}

func HashPassword(password string) (string, error) {
	if len(password) < 8 || len(password) > 72 {
		return "", fmt.Errorf(
			"invalid `password` len: %d: %w",
			len(password),
			core_errors.ErrInvalidArgument,
		)
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate from password: %w", err)
	}

	return string(hashPassword), nil
}

func (u *User) CheckPassword(password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return false
	}
	return true
}

func (u *User) Validate() error {
	if len(u.Email) > 250 {
		return fmt.Errorf(
			"invalid 'email' len=%d: %w",
			len(u.Email),
			core_errors.ErrInvalidArgument,
		)
	}
	parsedAddres, err := mail.ParseAddress(u.Email)
	email := NormalizeEmail(u.Email)
	if err != nil {
		return fmt.Errorf(
			"invalid `email` parse address: %w: %w",
			err,
			core_errors.ErrInvalidArgument,
		)
	}
	if parsedAddres.Address != email {
		return fmt.Errorf(
			"parsed address email: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	fullNameLen := len([]rune(u.FullName))
	if fullNameLen < 3 || fullNameLen > 100 {
		return fmt.Errorf(
			"invalid `FullName` len: %d: %w",
			fullNameLen,
			core_errors.ErrInvalidArgument,
		)
	}

	if u.PhoneNumber != nil {
		phoneNumberLen := len([]rune(*u.PhoneNumber))
		if phoneNumberLen < 10 || phoneNumberLen > 15 {
			return fmt.Errorf(
				"invalid `PhoneNumber` len: %d: %w",
				phoneNumberLen,
				core_errors.ErrInvalidArgument,
			)
		}

		re := regexp.MustCompile(`^\+[0-9]+$`)

		if !re.MatchString(*u.PhoneNumber) {
			return fmt.Errorf(
				"invalid `PhoneNumber` format: %w",
				core_errors.ErrInvalidArgument,
			)
		}
	}
	return nil
}

type UserPatch struct {
	FullName    Nullable[string]
	PhoneNumber Nullable[string]
}

func NewUserPatch(
	fullName Nullable[string],
	phoneNumber Nullable[string],
) UserPatch {
	return UserPatch{
		FullName:    fullName,
		PhoneNumber: phoneNumber,
	}
}

func (p *UserPatch) Validate() error {
	if p.FullName.Set && p.FullName.Value == nil {
		return fmt.Errorf("`FullName` can`t be patched to NULL: %w", core_errors.ErrInvalidArgument)
	}

	return nil
}

func (u *User) ApplyPatch(patch UserPatch) error {
	if err := patch.Validate(); err != nil {
		return fmt.Errorf("validate user patch: %w", err)
	}

	tmp := *u

	if patch.FullName.Set {
		tmp.FullName = *patch.FullName.Value
	}

	if patch.PhoneNumber.Set {
		tmp.PhoneNumber = patch.PhoneNumber.Value
	}

	if err := tmp.Validate(); err != nil {
		return fmt.Errorf("validate patched user: %w", err)
	}

	*u = tmp

	return nil
}
