package agent

import (
    "bytes"
    "compress/gzip"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/go-resty/resty/v2"
    models "github.com/yapryntsev/go-musthave-metrics/internal/model"
    "go.uber.org/zap"
    "math/rand"
    "net/http"
    "runtime"
    "time"
)

type Agent struct {
    addr           string
    stats          *runtime.MemStats
    l              *zap.Logger
    client         *resty.Client
    reportInterval uint
    pollInterval   uint

    // Metrics
    pollCount int
    randValue float64
}

func New(addr string, reportInterval uint, pollInterval uint, l *zap.Logger) *Agent {
    client := http.Client{
        Timeout: 5 * time.Second,
    }

    return &Agent{
        addr:           addr,
        stats:          &runtime.MemStats{},
        l:              l,
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
    a.l.Debug("metric collected")

    if time.Since(*lastReportTime).Seconds() < float64(a.reportInterval) {
        return nil
    }
    *lastReportTime = time.Now()

    err := a.sendMetrics()
    return err
}

func (a *Agent) sendMetrics() error {
    stats := a.stats

    alloc := float64(stats.Alloc)
    bhs := float64(stats.BuckHashSys)
    frees := float64(stats.Frees)
    gccpf := float64(stats.GCCPUFraction)
    gcys := float64(stats.GCSys)
    ha := float64(stats.HeapAlloc)
    hid := float64(stats.HeapIdle)
    hin := float64(stats.HeapInuse)
    hob := float64(stats.HeapObjects)
    hre := float64(stats.HeapReleased)
    hsy := float64(stats.HeapSys)
    lgc := float64(stats.LastGC)
    lup := float64(stats.Lookups)
    mci := float64(stats.MCacheInuse)
    mcs := float64(stats.MCacheSys)
    msi := float64(stats.MSpanInuse)
    mss := float64(stats.MSpanSys)
    mll := float64(stats.Mallocs)
    ngc := float64(stats.NextGC)
    nfg := float64(stats.NumForcedGC)
    nugc := float64(stats.NumGC)
    oss := float64(stats.OtherSys)
    ptn := float64(stats.PauseTotalNs)
    si := float64(stats.StackInuse)
    ss := float64(stats.StackSys)
    sys := float64(stats.Sys)
    ta := float64(stats.TotalAlloc)
    pc := int64(a.pollCount)

    metrics := []models.Metrics{
        {ID: "RandomValue", Value: &a.randValue},
        {ID: "Alloc", Value: &alloc},
        {ID: "BuckHashSys", Value: &bhs},
        {ID: "Frees", Value: &frees},
        {ID: "GCCPUFraction", Value: &gccpf},
        {ID: "GCSys", Value: &gcys},
        {ID: "HeapAlloc", Value: &ha},
        {ID: "HeapIdle", Value: &hid},
        {ID: "HeapInuse", Value: &hin},
        {ID: "HeapObjects", Value: &hob},
        {ID: "HeapReleased", Value: &hre},
        {ID: "HeapSys", Value: &hsy},
        {ID: "LastGC", Value: &lgc},
        {ID: "Lookups", Value: &lup},
        {ID: "MCacheInuse", Value: &mci},
        {ID: "MCacheSys", Value: &mcs},
        {ID: "MSpanInuse", Value: &msi},
        {ID: "MSpanSys", Value: &mss},
        {ID: "Mallocs", Value: &mll},
        {ID: "NextGC", Value: &ngc},
        {ID: "NumForcedGC", Value: &nfg},
        {ID: "NumGC", Value: &nugc},
        {ID: "OtherSys", Value: &oss},
        {ID: "PauseTotalNs", Value: &ptn},
        {ID: "StackInuse", Value: &si},
        {ID: "StackSys", Value: &ss},
        {ID: "Sys", Value: &sys},
        {ID: "TotalAlloc", Value: &ta},
        {ID: "PollCount", Delta: &pc},
    }

    err := a.sendBatch(metrics)
    if err != nil {
        return fmt.Errorf("failed to send metrics batch: %w", err)
    }

    return nil
}

func (a *Agent) sendBatch(metrics []models.Metrics) error {
    if len(a.addr) == 0 {
        return errors.New("host must be configured")
    }

    var buf bytes.Buffer
    zw := gzip.NewWriter(&buf)
    defer zw.Close()

    if err := json.NewEncoder(zw).Encode(metrics); err != nil {
        return err
    }

    if err := zw.Flush(); err != nil {
        return err
    }

    resp, err := a.client.R().
        SetHeader("Content-Type", "application/json").
        SetHeader("Content-Encoding", "gzip").
        SetBody(buf.Bytes()).
        Post(fmt.Sprintf("http://%s/updates", a.addr))

    if err != nil {
        a.l.Error("failed to send metrics", zap.Error(err))
    }

    a.l.Debug("metrics sent")

    if resp.StatusCode() != http.StatusOK {
        a.l.Error("unexpected status code", zap.Int("code", resp.StatusCode()))
    }

    return nil
}
