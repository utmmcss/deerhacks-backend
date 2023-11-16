package helpers

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// SQL Errors

func IsUniqueViolationError(err error) bool {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return pgError.Code == "23505"
	}
	return false
}
