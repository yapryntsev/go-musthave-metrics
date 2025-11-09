package pgerror

import (
    "errors"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgerrcode"
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
        pgerrcode.ConnectionFailure:
        return true
    default:
        return false
    }
}
