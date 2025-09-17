package handler

import (
    "errors"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "log"
    "net/http"
    "strconv"
)

type MetricHandler struct {
    log     *log.Logger
    service service.IMetricService
}

func New(service service.IMetricService, log *log.Logger,) MetricHandler {
    return MetricHandler{log: log, service: service}
}

func (h MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
    switch r.PathValue(service.MetricTypePathKey) {
    case service.GaugeMetricName:
        h.updateGauge(w, r)
    case service.CounterMetricName:
        h.updateCounter(w, r)
    default:
        w.WriteHeader(http.StatusBadRequest)
    }
}

func (h MetricHandler) updateGauge(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    name := r.PathValue(service.MetricNamePathKey)
    rawValue := r.PathValue(service.MetricValuePathKey)

    if len(name) == 0 || len(rawValue) == 0 {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    value, err := strconv.ParseFloat(rawValue, 64)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    err = h.service.UpdateGauge(name, value)
    if err != nil {
        if errors.Is(err, service.MetricTypeMismatch) {
            w.WriteHeader(http.StatusBadRequest)
        } else {
            w.WriteHeader(http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (h MetricHandler) updateCounter(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    name := r.PathValue(service.MetricNamePathKey)
    rawValue := r.PathValue(service.MetricValuePathKey)

    if len(name) == 0 || len(rawValue) == 0 {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    value, err := strconv.Atoi(rawValue)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    err = h.service.UpdateCounter(name, int64(value))
    if err != nil {
        if errors.Is(err, service.MetricTypeMismatch) {
            w.WriteHeader(http.StatusBadRequest)
        } else {
            w.WriteHeader(http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusOK)
}
