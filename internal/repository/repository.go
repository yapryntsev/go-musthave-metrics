package repository

import (
    "context"
    "database/sql"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "go.uber.org/zap"
    "time"
)

type MetricRepository interface {
    GetAll(ctx context.Context) ([]models.Metrics, error)
    Get(ctx context.Context, mID string, mType string) (*models.Metrics, error)
    Set(ctx context.Context, metric models.Metrics) error
    SetBatch(ctx context.Context, metrics []models.Metrics) error
}

func New(db *sql.DB, storeInt time.Duration, filePath string, restore bool, l *zap.Logger) MetricRepository {
    if db != nil {
        return newDatabaseRepository(db, l)
    }

    return newFileRepository(storeInt, filePath, restore, l)
}
