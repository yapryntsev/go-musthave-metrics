package handler

import (
    "bytes"
    "encoding/json"
    "fmt"
    log "github.com/sirupsen/logrus"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "net/http"
    "strconv"
)

const GetAllRowFormat = "%s: %s\n"

type MetricHandler struct {
    log     *log.Entry
    service service.MetricService
}

func New(service service.MetricService, log *log.Entry) MetricHandler {
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

    w.Header().Set("Content-Type", "text/html")
    w.Write(b.Bytes())
}

func (h MetricHandler) GetValue(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    mType := r.PathValue(service.MetricTypePathKey)
    mID := r.PathValue(service.MetricNamePathKey)

    if len(mID) == 0 || len(mType) == 0 {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    metric := &models.Metrics{
        ID:    mID,
        MType: mType,
    }

    ok, err := h.service.Get(metric)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !ok {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    switch metric.MType {
    case models.Counter:
        v := strconv.Itoa(int(*metric.Delta))
        _, err = w.Write([]byte(v))
    case models.Gauge:
        v := strconv.FormatFloat(*metric.Value, 'f', -1, 64)
        _, err = w.Write([]byte(v))
    }

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func (h MetricHandler) GetObject(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    metric := &models.Metrics{}
    if err := json.NewDecoder(r.Body).Decode(metric); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    ok, err := h.service.Get(metric)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !ok {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    if err := json.NewEncoder(w).Encode(metric); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func (h MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
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

    metric := &models.Metrics{
        ID: name,
    }

    var value float64
    var delta int64
    var err error

    switch r.PathValue(service.MetricTypePathKey) {
    case models.Gauge:
        value, err = strconv.ParseFloat(rawValue, 64)

        metric.MType = models.Gauge
        metric.Value = &value
    case models.Counter:
        var v int

        v, err = strconv.Atoi(rawValue)
        delta = int64(v)

        metric.MType = models.Counter
        metric.Delta = &delta
    default:
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    h.updateMetric(w, metric)
}

func (h MetricHandler) UpdateObject(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    metric := &models.Metrics{}
    if err := json.NewDecoder(r.Body).Decode(metric); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    h.updateMetric(w, metric)
}

func (h MetricHandler) updateMetric(w http.ResponseWriter, metric *models.Metrics) {
    var err error

    switch metric.MType {
    case models.Gauge:
        if metric.Value == nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        err = h.service.UpdateGauge(metric.ID, *metric.Value)
    case models.Counter:
        if metric.Delta == nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        err = h.service.UpdateCounter(metric.ID, *metric.Delta)
    }

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
