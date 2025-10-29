package repository

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    "github.com/jackc/pgx/v5"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "go.uber.org/zap"
    "time"
)

type DatabaseMetricRepository struct {
    db *sql.DB
    l  *zap.Logger
}

func newDatabaseRepository(db *sql.DB, l *zap.Logger) *DatabaseMetricRepository {
    repo := &DatabaseMetricRepository{
        db: db,
        l:  l,
    }

    if err := repo.initAndCheckMigration(); err != nil {
        l.Fatal("failed to perform migration", zap.Error(err))
    }

    return repo
}

func (d *DatabaseMetricRepository) initAndCheckMigration() error {
    driver, err := postgres.WithInstance(d.db, &postgres.Config{})
    if err != nil {
        return fmt.Errorf("failed to instantiate postgres driver: %w", err)
    }

    m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
    if err != nil {
        return fmt.Errorf("failed to instantiate migration: %w", err)
    }

    _, _, err = m.Version()
    if errors.Is(err, migrate.ErrNilVersion) {
        d.l.Debug("migration version is 0. apply initial migration")

        err := m.Up()
        if err != nil {
            return fmt.Errorf("failed to migrate: %w", err)
        }
    }

    return nil
}

func (d *DatabaseMetricRepository) GetAll(ctx context.Context) ([]models.Metrics, error) {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    var result []models.Metrics

    rows, err := d.db.QueryContext(ctx, "SELECT id, type, delta, value FROM metrics")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var r models.Metrics

        err = rows.Scan(&r.ID, &r.MType, &r.Delta, &r.Value)
        if err != nil {
            return nil, err
        }

        result = append(result, r)
    }

    err = rows.Err()
    if err != nil {
        return nil, err
    }

    return result, nil
}

func (d *DatabaseMetricRepository) Get(ctx context.Context, mID string, mType string) (*models.Metrics, error) {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    var r models.Metrics
    query := "SELECT id, type, delta, value FROM metrics WHERE id = @id AND type = @type"

    row := d.db.QueryRowContext(
        ctx, query,
        pgx.NamedArgs{
            "id":   mID,
            "type": mType,
        },
    )

    err := row.Scan(&r.ID, &r.MType, &r.Delta, &r.Value)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }

    return &r, err
}

func (d *DatabaseMetricRepository) Set(ctx context.Context, metric models.Metrics) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    query := "INSERT INTO metrics (id, type, delta, value) VALUES (@id, @type, @delta, @value)"

    _, err := d.db.ExecContext(
        ctx, query,
        pgx.NamedArgs{
            "id":    metric.ID,
            "type":  metric.MType,
            "delta": metric.Delta,
            "value": metric.Value,
        },
    )

    return err
}

func (d *DatabaseMetricRepository) SetBatch(ctx context.Context, metrics []models.Metrics) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    tx, err := d.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to create transaction: %w", err)
    }

    stmt, err := tx.PrepareContext(
        ctx,
        "INSERT INTO metrics (id, type, delta, value) "+
            "VALUES ($1, $2, $3, $4) "+
            "ON CONFLICT (id, type) DO UPDATE SET delta = $3, value = $4",
    )
    if err != nil {
        return fmt.Errorf("failed to prepare sql statement: %w", err)
    }

    for _, m := range metrics {
        _, err := stmt.ExecContext(
            ctx,
            m.ID,
            m.MType,
            m.Delta,
            m.Value,
        )
        if err != nil {
            tx.Rollback()
            return fmt.Errorf("failed to perform insert: %w", err)
        }
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
