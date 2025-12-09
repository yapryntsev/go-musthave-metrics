package handler

import (
    "bytes"
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "go.uber.org/zap"
    "net/http"
    "strconv"
    "time"
)

const GetAllRowFormat = "%s: %s\n"

type MetricHandler struct {
    l       *zap.Logger
    db      *sql.DB
    service service.MetricService
}

func New(service service.MetricService, db *sql.DB, l *zap.Logger) MetricHandler {
    return MetricHandler{l: l, db: db, service: service}
}

func (h MetricHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    res, err := h.service.GetAll()
    if err != nil {
        h.l.Error("failed to get all metrics", zap.Error(err))

        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    b := new(bytes.Buffer)
    for _, m := range res {
        switch m.MType {
        case models.Counter:
            v := strconv.Itoa(int(*m.Delta))
            fmt.Fprintf(b, GetAllRowFormat, m.ID, v)
        case models.Gauge:
            v := fmt.Sprintf(`%.f`, *m.Value)
            fmt.Fprintf(b, GetAllRowFormat, m.ID, v)
        }
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
        h.l.Error(
            "failed to get metric",
            zap.Error(err),
            zap.String("mID", mID),
            zap.String("mType", mType),
        )

        w.WriteHeader(http.StatusInternalServerError)
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
        h.l.Error("failed to format metric value", zap.Error(err))

        w.WriteHeader(http.StatusInternalServerError)
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
        h.l.Error("failed to decode request body", zap.Error(err))

        w.WriteHeader(http.StatusBadRequest)
        return
    }

    ok, err := h.service.Get(metric)
    if err != nil {
        h.l.Error(
            "failed to get metric",
            zap.Error(err),
            zap.String("mID", metric.ID),
            zap.String("mType", metric.MType),
        )

        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    if !ok {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    if err := json.NewEncoder(w).Encode(metric); err != nil {
        h.l.Error("failed to encode response", zap.Error(err))

        w.WriteHeader(http.StatusInternalServerError)
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
        h.l.Error("failed to parse metric value", zap.Error(err), zap.String("value", rawValue))

        w.WriteHeader(http.StatusBadRequest)
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
        h.l.Error("failed to decode request body", zap.Error(err))

        w.WriteHeader(http.StatusBadRequest)
        return
    }

    h.updateMetric(w, metric)
}

func (h MetricHandler) Ping(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
    defer cancel()

    if err := h.db.PingContext(ctx); err != nil {
        h.l.Error("failed to ping db", zap.Error(err))
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
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
