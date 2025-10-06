package handler

import (
    "bytes"
    "fmt"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "log"
    "net/http"
    "strconv"
)

const GetAllRowFormat = "%s: %s\n"

type MetricHandler struct {
    log     *log.Logger
    service service.IMetricService
}

func New(service service.IMetricService, log *log.Logger,) MetricHandler {
    return MetricHandler{log: log, service: service}
}

func (h MetricHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    res, err := h.service.GetAll()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    b := new(bytes.Buffer)
    for k, v := range res {
        fmt.Fprintf(b, GetAllRowFormat, k, v)
    }

    w.Write(b.Bytes())
}

func (h MetricHandler) GetValue(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    metricType := r.PathValue(service.MetricTypePathKey)
    name := r.PathValue(service.MetricNamePathKey)

    if len(name) == 0 || len(metricType) == 0 {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    res, err := h.service.Get(metricType, name)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    if len(res) == 0 {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    w.Write([]byte(res))
}

func (h MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
    switch r.PathValue(service.MetricTypePathKey) {
    case service.GaugeMetricTypeName:
        h.updateGauge(w, r)
    case service.CounterMetricTypeName:
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
        w.WriteHeader(http.StatusInternalServerError)
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
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
