package db

import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
)

var dbConnection *sql.DB

func NewConnection(dsn string) (*sql.DB, error) {
    db, err := sql.Open("pgx", dsn)
    dbConnection = db

    return db, err
}

func CloseConnection() error {
    if dbConnection != nil {
        return dbConnection.Close()
    }

    return nil
}
