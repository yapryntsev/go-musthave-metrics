package agent

import (
    "bytes"
    "compress/gzip"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/go-resty/resty/v2"
    log "github.com/sirupsen/logrus"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "math/rand"
    "net/http"
    "runtime"
    "time"
)

var errTypeCast = errors.New("failed to cast type")

type Agent struct {
    addr           string
    stats          *runtime.MemStats
    log            *log.Entry
    client         *resty.Client
    reportInterval uint
    pollInterval   uint

    // Metrics
    pollCount int
    randValue float64
}

func New(addr string, reportInterval uint, pollInterval uint, log *log.Entry) *Agent {
    client := http.Client{
        Timeout: 5 * time.Second,
    }

    return &Agent{
        addr:           addr,
        stats:          &runtime.MemStats{},
        log:            log,
        client:         resty.NewWithClient(&client),
        reportInterval: reportInterval,
        pollInterval:   pollInterval,
    }
}

func (a *Agent) StartGathering(ctx context.Context) error {
    lastReportTime := time.Now()
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            err := a.scheduleMetricsFetchAndUpload(&lastReportTime)
            if err != nil {
                return err
            }
        }
    }
}

func (a *Agent) scheduleMetricsFetchAndUpload(lastReportTime *time.Time) error {
    time.Sleep(time.Duration(a.pollInterval) * time.Second)

    a.pollCount++
    seed := time.Now().Unix()
    a.randValue = rand.New(rand.NewSource(seed)).Float64()

    runtime.ReadMemStats(a.stats)
    a.log.Println("metric collected")

    if time.Since(*lastReportTime).Seconds() < float64(a.reportInterval) {
        return nil
    }
    *lastReportTime = time.Now()

    err := a.sendMetrics()
    return err
}

func (a *Agent) sendMetrics() error {
    stats := a.stats

    gaugeMetrics := []struct {
        name  string
        value float64
    }{
        {"RandomValue", a.randValue},
        {"Alloc", float64(stats.Alloc)},
        {"BuckHashSys", float64(stats.BuckHashSys)},
        {"Frees", float64(stats.Frees)},
        {"GCCPUFraction", float64(stats.GCCPUFraction)},
        {"GCSys", float64(stats.GCSys)},
        {"HeapAlloc", float64(stats.HeapAlloc)},
        {"HeapIdle", float64(stats.HeapIdle)},
        {"HeapInuse", float64(stats.HeapInuse)},
        {"HeapObjects", float64(stats.HeapObjects)},
        {"HeapReleased", float64(stats.HeapReleased)},
        {"HeapSys", float64(stats.HeapSys)},
        {"LastGC", float64(stats.LastGC)},
        {"Lookups", float64(stats.Lookups)},
        {"MCacheInuse", float64(stats.MCacheInuse)},
        {"MCacheSys", float64(stats.MCacheSys)},
        {"MSpanInuse", float64(stats.MSpanInuse)},
        {"MSpanSys", float64(stats.MSpanSys)},
        {"Mallocs", float64(stats.Mallocs)},
        {"NextGC", float64(stats.NextGC)},
        {"NumForcedGC", float64(stats.NumForcedGC)},
        {"NumGC", float64(stats.NumGC)},
        {"OtherSys", float64(stats.OtherSys)},
        {"PauseTotalNs", float64(stats.PauseTotalNs)},
        {"StackInuse", float64(stats.StackInuse)},
        {"StackSys", float64(stats.StackSys)},
        {"Sys", float64(stats.Sys)},
        {"TotalAlloc", float64(stats.TotalAlloc)},
    }

    for _, m := range gaugeMetrics {
        err := a.sendMetric(models.Gauge, m.name, m.value)
        if err != nil {
            return err
        }
    }

    return a.sendMetric(models.Counter, "PollCount", int64(a.pollCount))
}

func (a *Agent) sendMetric(t string, name string, value interface{}) error {
    if len(a.addr) == 0 {
        return errors.New("host must be configured")
    }

    metric := models.Metrics{
        ID:    name,
        MType: t,
    }

    switch t {
    case models.Counter:
        v, ok := value.(int64)
        if !ok {
            return errTypeCast
        }
        metric.Delta = &v
    case models.Gauge:
        v, ok := value.(float64)
        if !ok {
            return errTypeCast
        }
        metric.Value = &v
    }

    var buf bytes.Buffer
    zw := gzip.NewWriter(&buf)
    defer zw.Close()

    if err := json.NewEncoder(zw).Encode(metric); err != nil {
        return err
    }

    if err := zw.Flush(); err != nil {
        return err
    }

    resp, err := a.client.R().
        SetHeader("Content-Type", "application/json").
        SetHeader("Content-Encoding", "gzip").
        SetBody(buf.Bytes()).
        Post(fmt.Sprintf("http://%s/update", a.addr))

    if err != nil {
        a.log.Println(err)
    }

    a.log.Printf("metric sent. type: %s, name: %s, value: %v", t, name, value)

    if resp.StatusCode() != http.StatusOK {
        a.log.Printf("unexpected status code: %d", resp.StatusCode())
    }

    return nil
}
