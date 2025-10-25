package service

import (
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
)

const (
    MetricTypePathKey  = `type`
    MetricNamePathKey  = `name`
    MetricValuePathKey = `value`
)

type MetricService interface {
    GetAll() ([]*models.Metrics, error)
    Get(metric *models.Metrics) (bool, error)
    UpdateCounter(mID string, value int64) error
    UpdateGauge(mID string, value float64) error
}

type Service struct {
    repo repository.MetricRepository
}

func New(repo repository.MetricRepository) *Service {
    return &Service{
        repo: repo,
    }
}

func (s *Service) GetAll() ([]*models.Metrics, error) {
    return s.repo.GetAll()
}

func (s *Service) Get(metric *models.Metrics) (bool, error) {
    m, err := s.repo.Get(metric.ID, metric.MType)
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

func (s *Service) UpdateCounter(mID string, value int64) error {
    m, err := s.repo.Get(mID, models.Counter)
    if err != nil {
        return err
    }

    if m == nil {
        m = &models.Metrics{
            ID:    mID,
            MType: models.Counter,
            Delta: new(int64),
        }
    }

    *m.Delta += value

    err = s.repo.Set(m)
    if err != nil {
        return err
    }

    return nil
}

func (s *Service) UpdateGauge(mID string, value float64) error {
    return s.repo.Set(
        &models.Metrics{
            ID:    mID,
            MType: models.Gauge,
            Value: &value,
        },
    )
}
