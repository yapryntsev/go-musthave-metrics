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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	models "github.com/yapryntsev/go-musthave-metrics/internal/model"
	"go.uber.org/zap"
)

type DatabaseMetricRepository struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func newDatabaseRepository(db *pgxpool.Pool, log *zap.Logger) *DatabaseMetricRepository {
	repo := &DatabaseMetricRepository{
		db:  db,
		log: log,
	}

	if err := repo.initAndCheckMigration(); err != nil {
		log.Fatal("failed to perform migration", zap.Error(err))
	}

	return repo
}

func (d *DatabaseMetricRepository) initAndCheckMigration() error {
	db := sql.OpenDB(stdlib.GetPoolConnector(d.db))

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to instantiate postgres driver: %w", err)
	}
	defer func() {
		_ = driver.Close()
	}()

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to instantiate migration: %w", err)
	}

	_, _, err = m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		d.log.Debug("migration version is 0. apply initial migration")

		err := m.Up()
		if err != nil {
			return fmt.Errorf("failed to migrate: %w", err)
		}
	}

	return nil
}

func (d *DatabaseMetricRepository) GetAll(ctx context.Context) ([]models.Metrics, error) {
	var result []models.Metrics

	rows, err := d.db.Query(ctx, "SELECT id, type, delta, value FROM metrics")
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
	var r models.Metrics
	query := "SELECT id, type, delta, value FROM metrics WHERE id = @id AND type = @type"

	row := d.db.QueryRow(
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
	query := `
        INSERT INTO 
            metrics (id, type, delta, value)  
        VALUES ($1, $2, $3, $4)  
        ON CONFLICT (id, type) DO UPDATE SET delta = $3, value = $4`

	_, err := d.db.Exec(
		ctx,
		query,
		metric.ID,
		metric.MType,
		metric.Delta,
		metric.Value,
	)

	return err
}

func (d *DatabaseMetricRepository) SetBatch(ctx context.Context, metrics []*models.Metrics) error {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	stmt, err := tx.Prepare(
		ctx,
		"insert_metrics",
		`INSERT INTO metrics (id, type, delta, value)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (id, type) DO UPDATE SET delta = $3, value = $4`,
	)
	if err != nil {
		return fmt.Errorf("failed to prepare sql statement: %w", err)
	}

	batch := &pgx.Batch{}

	for _, m := range metrics {
		batch.Queue(
			stmt.Name,
			m.ID,
			m.MType,
			m.Delta,
			m.Value,
		)
	}

	res := tx.SendBatch(ctx, batch)

	for range metrics {
		if _, err := res.Exec(); err != nil {
			_ = res.Close()
			return fmt.Errorf("failed to perform batched operation: %w", err)
		}
	}

	if err := res.Close(); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
