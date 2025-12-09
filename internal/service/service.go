package service

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
    "maps"
    "slices"
)

const (
    MetricTypePathKey  = `type`
    MetricNamePathKey  = `name`
    MetricValuePathKey = `value`
)

type MetricService interface {
    GetAll(ctx context.Context) ([]models.Metrics, error)
    Get(ctx context.Context, metric *models.Metrics) (bool, error)
    UpdateCounter(ctx context.Context, mID string, value int64) error
    UpdateGauge(ctx context.Context, mID string, value float64) error
    UpdateBatch(ctx context.Context, metrics []models.Metrics) error
    Ping(ctx context.Context) error
}

type Service struct {
    repo repository.MetricRepository
    db   *pgxpool.Pool
}

func New(repo repository.MetricRepository, db *pgxpool.Pool) *Service {
    return &Service{
        repo: repo,
        db:   db,
    }
}

func (s *Service) GetAll(ctx context.Context) ([]models.Metrics, error) {
    return s.repo.GetAll(ctx)
}

func (s *Service) Get(ctx context.Context, metric *models.Metrics) (bool, error) {
    m, err := s.repo.Get(ctx, metric.ID, metric.MType)
    if err != nil {
        return false, err
    }

    if m == nil {
        return false, nil
    }

    metric.Value = m.Value
    metric.Delta = m.Delta

    return true, err
}

func (s *Service) UpdateCounter(ctx context.Context, mID string, value int64) error {
    m, err := s.getCurrentCounterMetric(ctx, mID)
    if err != nil {
        return err
    }

    *m.Delta += value

    err = s.repo.Set(ctx, m)
    if err != nil {
        return err
    }

    return nil
}

func (s *Service) UpdateGauge(ctx context.Context, mID string, value float64) error {
    return s.repo.Set(
        ctx,
        models.Metrics{
            ID:    mID,
            MType: models.Gauge,
            Value: &value,
        },
    )
}

func (s *Service) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
    if len(metrics) == 0 {
        return nil
    }

    batch := make(map[string]models.Metrics)

    for _, m := range metrics {
        switch m.MType {
        case models.Gauge:
            batch[m.ID+m.MType] = m
        case models.Counter:
            var err error

            cm, ok := batch[m.ID+m.MType]
            if !ok {
                cm, err = s.getCurrentCounterMetric(ctx, m.ID)
            }

            if err != nil {
                return err
            }

            *cm.Delta += *m.Delta
            batch[m.ID+m.MType] = cm
        }
    }

    return s.repo.SetBatch(ctx, slices.Collect(maps.Values(batch)))
}

func (s *Service) Ping(ctx context.Context) error {
    return s.db.Ping(ctx)
}

func (s *Service) getCurrentCounterMetric(ctx context.Context, mID string) (models.Metrics, error) {
    m, err := s.repo.Get(ctx, mID, models.Counter)
    if err != nil {
        return *m, err
    }

    if m == nil {
        m = &models.Metrics{
            ID:    mID,
            MType: models.Counter,
            Delta: new(int64),
        }
    }

    return *m, nil
}
