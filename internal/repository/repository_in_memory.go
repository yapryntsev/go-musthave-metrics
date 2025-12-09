package repository

import (
    "context"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "maps"
    "slices"
    "sync"
)

type MemoryMetricRepository struct {
    mu      sync.RWMutex
    storage map[string]models.Metrics
}

func newInMemoryRepo() *MemoryMetricRepository {
    return &MemoryMetricRepository{
        storage: make(map[string]models.Metrics),
    }
}

func (r *MemoryMetricRepository) GetAll(ctx context.Context) ([]models.Metrics, error) {
    r.mu.RLock()
    metrics := slices.Collect(maps.Values(r.storage))
    r.mu.RUnlock()

    return metrics, nil
}

func (r *MemoryMetricRepository) Get(ctx context.Context, mID string, mType string) (*models.Metrics, error) {
    key := r.constructKey(mID, mType)

    r.mu.RLock()
    metric, ok := r.storage[key]
    r.mu.RUnlock()

    if !ok {
        return nil, nil
    }

    return &metric, nil
}

func (r *MemoryMetricRepository) Set(ctx context.Context, metric models.Metrics) error {
    key := r.keyForMetric(metric)

    r.mu.Lock()
    r.storage[key] = metric
    r.mu.Unlock()

    return nil
}

func (r *MemoryMetricRepository) constructKey(mID string, mType string) string {
    return mID + "-" + mType
}

func (r *MemoryMetricRepository) keyForMetric(metrics models.Metrics) string {
    return r.constructKey(metrics.ID, metrics.MType)
}
