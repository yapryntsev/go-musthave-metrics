package service

import (
    "errors"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
    "log"
)

const (
    MetricTypePathKey     = `type`
    MetricNamePathKey     = `name`
    MetricValuePathKey    = `value`
    GaugeMetricTypeName   = `gauge`
    CounterMetricTypeName = `counter`
)

var ErrMetricTypeMismatch = errors.New(`metric with the same name but a different type already exists`)

type IMetricService interface {
    UpdateCounter(name string, value int64) error
    UpdateGauge(name string, value float64) error
}

type MetricService struct {
    log  *log.Logger
    repo repository.IMetricRepository
}

func New(repo repository.IMetricRepository, log *log.Logger) *MetricService {
    return &MetricService{
        log:  log,
        repo: repo,
    }
}

func (s *MetricService) UpdateCounter(name string, value int64) error {
    metric, err := s.repo.Get(name)
    if err != nil {
        s.log.Printf("failed to fetch counter metric: %s", err.Error())
        return err
    }

    if metric == nil {
        metric = &repository.Metric{
            MetricName: name,
            TypeName:   CounterMetricTypeName,
        }
    }

    if metric.TypeName != CounterMetricTypeName {
        return ErrMetricTypeMismatch
    }

    metric.Value += float64(value)
    err = s.repo.Set(metric)
    if err != nil {
        s.log.Printf("failed to save counter metric: %s", err.Error())
        return err
    }

    return nil
}

func (s *MetricService) UpdateGauge(name string, value float64) error {
    oldMetric, err := s.repo.Get(name)
    newMetric := &repository.Metric{
        MetricName: name,
        TypeName:   GaugeMetricTypeName,
        Value:      value,
    }

    if oldMetric != nil && oldMetric.TypeName != GaugeMetricTypeName {
        return ErrMetricTypeMismatch
    }

    err = s.repo.Set(newMetric)
    if err != nil {
        s.log.Printf("failed to save gauge metric: %s", err.Error())
        return err
    }

    return nil
}
