package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolationCode = "23505"

func ViolatedUniqueConstraint(err error) (string, bool) {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == uniqueViolationCode {
		return postgresError.ConstraintName, true
	}
	return "", false
}
