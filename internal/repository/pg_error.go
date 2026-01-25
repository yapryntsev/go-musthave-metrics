package repository

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return isRetriablePgError(pgErr)
	}

	return false
}

func isRetriablePgError(pgErr *pgconn.PgError) bool {
	switch pgErr.Code {
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure,
		pgerrcode.TransactionRollback,
		pgerrcode.TransactionIntegrityConstraintViolation,
		pgerrcode.SerializationFailure,
		pgerrcode.StatementCompletionUnknown:
		return true
	default:
		return false
	}
}
