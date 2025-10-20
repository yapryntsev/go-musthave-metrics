package repository

import (
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "go.uber.org/zap"
    "time"
)

type MetricRepository interface {
    GetAll() ([]*models.Metrics, error)
    Get(mID string, mType string) (*models.Metrics, error)
    Set(metric *models.Metrics) error
}

func New(storeInt time.Duration, filePath string, restore bool, l *zap.Logger) MetricRepository {
    repo := &FileMetricRepository{
        l:          l,
        inMemory:   NewInMemoryRepo(),
        filePath:   filePath,
        storeInt:   storeInt,
        lastUpdate: time.Now(),
    }

    if restore {
        repo.ReadFromFile()
    }

    return repo
}
