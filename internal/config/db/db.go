package db

import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
)

var dbConnection *sql.DB

func NewConnection() (*sql.DB, error) {
    db, err := sql.Open("pgx", "host=localhost user=ya dbname=metrics sslmode=disable")
    dbConnection = db

    return db, err
}

func CloseConnection() error {
    return dbConnection.Close()
}
