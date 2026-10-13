package core_postgres_pool

import "errors"

var (
	ErrNoRows             = errors.New("No rows")
	ErrViolatesForeignKey = errors.New("violates foreign key")
	ErrUnknown            = errors.New("unknown")
)
