package service

import (
    "errors"
    "fmt"
    "github.com/yapryntsev/go-musthave-metrics/internal/repository"
    "log"
    "strconv"
)

const (
    MetricTypePathKey     = `type`
    MetricNamePathKey     = `name`
    MetricValuePathKey    = `value`
    GaugeMetricTypeName   = `gauge`
    CounterMetricTypeName = `counter`
)

type IMetricService interface {
    GetAll() (map[string]string, error)
    Get(metricType string, name string) (string, error)
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

func (s *MetricService) GetAll() (map[string]string, error) {
    var err error
    res := make(map[string]string)

    intMetrics, err := s.repo.GetAllInt()
    if err != nil {
        return nil, err
    }

    for n, v := range intMetrics {
        res[n] = strconv.Itoa(int(v))
    }

    floatMetrics, err := s.repo.GetAllFloat()
    if err != nil {
        return nil, err
    }

    for n, v := range floatMetrics {
        res[n] = fmt.Sprintf(`%.3f`, v)
    }

    return res, nil
}

func (s *MetricService) Get(metricType string, name string) (string, error) {
    var res string
    var err error

    switch metricType {
    case CounterMetricTypeName:
        var v int64
        v, err = s.repo.GetInt(name)
        if err == nil {
            res = strconv.Itoa(int(v))
        }
    case GaugeMetricTypeName:
        var v float64
        v, err = s.repo.GetFloat(name)
        if err == nil {
            res = fmt.Sprintf(`%.3f`, v)
        }
    }

    if err != nil && !errors.Is(err, repository.ErrValueNotFound) {
        s.log.Printf("failed to fetch counter metric: %s", err.Error())
        return ``, err
    }

    return res, nil
}

func (s *MetricService) UpdateCounter(name string, value int64) error {
    metric, err := s.repo.GetInt(name)
    if err != nil && !errors.Is(err, repository.ErrValueNotFound) {
        s.log.Printf("failed to fetch counter metric: %s", err.Error())
        return err
    }

    metric += value
    err = s.repo.SetInt(name, metric)
    if err != nil {
        s.log.Printf("failed to save counter metric: %s", err.Error())
        return err
    }

    return nil
}

func (s *MetricService) UpdateGauge(name string, value float64) error {
    err := s.repo.SetFloat(name, value)
    if err != nil {
        s.log.Printf("failed to save gauge metric: %s", err.Error())
        return err
    }

    return nil
}
