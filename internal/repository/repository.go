package repository

import (
    "database/sql"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "go.uber.org/zap"
    "time"
)

type MetricRepository interface {
    GetAll() ([]*models.Metrics, error)
    Get(mID string, mType string) (*models.Metrics, error)
    Set(metric *models.Metrics) error
}

func New(db *sql.DB, storeInt time.Duration, filePath string, restore bool, l *zap.Logger) MetricRepository {
    if db != nil {
        return newDatabaseRepository(db, l)
    }

    return newFileRepository(storeInt, filePath, restore, l)
}
