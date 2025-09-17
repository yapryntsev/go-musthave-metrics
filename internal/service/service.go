package service

import (
    "errors"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
    "log"
)

const (
    MetricTypePathKey  = `type`
    MetricNamePathKey  = `name`
    MetricValuePathKey = `value`
    GaugeMetricName    = `gauge`
    CounterMetricName  = `counter`
)

var ErrMetricTypeMismatch = errors.New(`metric with the same name but a different type already exists`)

type IMetricService interface {
    UpdateCounter(name string, value int64) error
    UpdateGauge(name string, value float64) error
}

func New(repo repository.IMetricRepository, log *log.Logger) IMetricService {
    return &metricService{
        log:  log,
        repo: repo,
    }
}

type metricService struct {
    log  *log.Logger
    repo repository.IMetricRepository
}

func (s *metricService) UpdateCounter(name string, value int64) error {
    metric, err := s.repo.Get(name)
    if err != nil {
        s.log.Printf("failed to fetch counter metric: %s", err.Error())
        return err
    }

    if metric == nil {
        metric = &repository.Metric{
            MetricName: name,
            TypeName:   CounterMetricName,
            Value:      float64(value),
        }
    }

    if metric.TypeName != CounterMetricName {
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

func (s *metricService) UpdateGauge(name string, value float64) error {
    metric := &repository.Metric{
        MetricName: name,
        TypeName:   GaugeMetricName,
        Value:      value,
    }

    err := s.repo.Set(metric)
    if err != nil {
        s.log.Printf("failed to save gauge metric: %s", err.Error())
        return err
    }

    return nil
}
