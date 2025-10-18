package repository

import (
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "maps"
    "slices"
    "sync"
)

type MemoryMetricRepository struct {
    mu      sync.RWMutex
    storage map[string]*models.Metrics
}

func NewInMemoryRepo() *MemoryMetricRepository {
    return &MemoryMetricRepository{
        storage: make(map[string]*models.Metrics),
    }
}

func (r *MemoryMetricRepository) GetAll() ([]*models.Metrics, error) {
    return slices.Collect(maps.Values(r.storage)), nil
}

func (r *MemoryMetricRepository) Get(mID string, mType string) (*models.Metrics, error) {
    key := r.constructKey(mID, mType)
    return r.storage[key], nil
}

func (r *MemoryMetricRepository) Set(metric *models.Metrics) error {
    key := r.keyForMetric(metric)
    r.storage[key] = metric

    return nil
}

func (r *MemoryMetricRepository) constructKey(mID string, mType string) string {
    return mID + "-" + mType
}

func (r *MemoryMetricRepository) keyForMetric(metrics *models.Metrics) string {
    return r.constructKey(metrics.ID, metrics.MType)
}
