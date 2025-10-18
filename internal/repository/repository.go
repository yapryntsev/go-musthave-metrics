package repository

import (
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
)

type MetricRepository interface {
    GetAll() ([]*models.Metrics, error)
    Get(mID string, mType string) (*models.Metrics, error)
    Set(metric *models.Metrics) error
}
