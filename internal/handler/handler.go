package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"

    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "github.com/yapryntsev/go-musthave-metrics/internal/service"
    "github.com/yapryntsev/go-musthave-metrics/internal/service/audit"
    "go.uber.org/zap"
)

const GetAllRowFormat = "%s: %s\n"

type MetricHandler struct {
    log      *zap.Logger
    service  service.MetricService
    auditors map[audit.AuditorID]audit.Auditor
}

func New(service service.MetricService, log *zap.Logger) MetricHandler {
    return MetricHandler{log: log, service: service, auditors: make(map[audit.AuditorID]audit.Auditor)}
}

func (h *MetricHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    res, err := h.service.GetAll(r.Context())
    if err != nil {
        h.log.Error("failed to get all metrics", zap.Error(err))

        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    b := new(bytes.Buffer)
    for _, m := range res {
        switch m.MType {
        case models.Counter:
            v := strconv.Itoa(int(*m.Delta))
            _, _ = fmt.Fprintf(b, GetAllRowFormat, m.ID, v)
        case models.Gauge:
            v := fmt.Sprintf(`%.f`, *m.Value)
            _, _ = fmt.Fprintf(b, GetAllRowFormat, m.ID, v)
        }
    }

    w.Header().Set("Content-Type", "text/html")
    _, _ = w.Write(b.Bytes())
}

func (h *MetricHandler) GetValue(w http.ResponseWriter, r *http.Request) {
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

    ok, err := h.service.Get(r.Context(), metric)
    if err != nil {
        h.log.Error(
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
        h.log.Error("failed to format metric value", zap.Error(err))

        w.WriteHeader(http.StatusInternalServerError)
        return
    }
}

func (h *MetricHandler) GetObject(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    metric := &models.Metrics{}
    if err := json.NewDecoder(r.Body).Decode(metric); err != nil {
        h.log.Error("failed to decode request body", zap.Error(err))

        w.WriteHeader(http.StatusBadRequest)
        return
    }

    ok, err := h.service.Get(r.Context(), metric)
    if err != nil {
        h.log.Error(
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
        h.log.Error("failed to encode response", zap.Error(err))

        w.WriteHeader(http.StatusInternalServerError)
        return
    }
}

func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
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
        h.log.Error("failed to parse metric value", zap.Error(err), zap.String("value", rawValue))

        w.WriteHeader(http.StatusBadRequest)
        return
    }

    h.updateMetric(r.Context(), w, metric)
}

func (h *MetricHandler) UpdateObject(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    metric := &models.Metrics{}
    if err := json.NewDecoder(r.Body).Decode(metric); err != nil {
        h.log.Error("failed to decode request body", zap.Error(err))

        w.WriteHeader(http.StatusBadRequest)
        return
    }

    h.updateMetric(r.Context(), w, metric)
}

func (h *MetricHandler) Ping(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    if err := h.service.Ping(r.Context()); err != nil {
        h.log.Error("failed to ping db", zap.Error(err))
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
}

func (h *MetricHandler) UpdateBatch(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var metrics []models.Metrics
    if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
        h.log.Error("failed to decode request body", zap.Error(err))

        w.WriteHeader(http.StatusBadRequest)
        return
    }

    if err := h.service.UpdateBatch(r.Context(), metrics); err != nil {
        h.log.Error("failed to save batch", zap.Error(err))

        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    metricIDs := make([]string, len(metrics))
    for _, metric := range metrics {
        metricIDs = append(metricIDs, metric.ID)
    }

    auditPayload := audit.Payload{
        Timestamp: strconv.Itoa(int(time.Now().Unix())),
        Metrics:   metricIDs,
        IPAddress: r.RemoteAddr,
    }

    for _, auditor := range h.auditors {
        auditor.Process(auditPayload)
    }
}

func (h *MetricHandler) updateMetric(ctx context.Context, w http.ResponseWriter, metric *models.Metrics) {
    var err error

    switch metric.MType {
    case models.Gauge:
        if metric.Value == nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        err = h.service.UpdateGauge(ctx, metric.ID, *metric.Value)
    case models.Counter:
        if metric.Delta == nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        err = h.service.UpdateCounter(ctx, metric.ID, *metric.Delta)
    }

    if err != nil {
        h.log.Error("failed to update metric", zap.Error(err))
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
}

func (h *MetricHandler) Audit(auditor audit.Auditor) {
    h.auditors[auditor.ID()] = auditor
}

func (h *MetricHandler) Neglect(auditorID audit.AuditorID) {
    delete(h.auditors, auditorID)
}
