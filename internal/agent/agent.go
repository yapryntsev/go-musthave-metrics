package agent

import (
    "context"
    "errors"
    "fmt"
    "github.com/go-resty/resty/v2"
    log "github.com/sirupsen/logrus"
    "math/rand"
    "net/http"
    "runtime"
    "strconv"
    "time"
)

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
        {"random-value", a.randValue},
        {"alloc", float64(stats.Alloc)},
        {"buck-hash-sys", float64(stats.BuckHashSys)},
        {"frees", float64(stats.Frees)},
        {"gc-cpu-fraction", float64(stats.GCCPUFraction)},
        {"gc-sys", float64(stats.GCSys)},
        {"heap-alloc", float64(stats.HeapAlloc)},
        {"heap-idle", float64(stats.HeapIdle)},
        {"heap-inuse", float64(stats.HeapInuse)},
        {"heap-objects", float64(stats.HeapObjects)},
        {"heap-released", float64(stats.HeapReleased)},
        {"heap-sys", float64(stats.HeapSys)},
        {"last-gc", float64(stats.LastGC)},
        {"lookups", float64(stats.Lookups)},
        {"m-cache-inuse", float64(stats.MCacheInuse)},
        {"m-cache-sys", float64(stats.MCacheSys)},
        {"m-span-inuse", float64(stats.MSpanInuse)},
        {"m-span-sys", float64(stats.MSpanSys)},
        {"mallocs", float64(stats.Mallocs)},
        {"next-gc", float64(stats.NextGC)},
        {"num-forced-gc", float64(stats.NumForcedGC)},
        {"num-gc", float64(stats.NumGC)},
        {"other-sys", float64(stats.OtherSys)},
        {"pause-total-ns", float64(stats.PauseTotalNs)},
        {"stack-inuse", float64(stats.StackInuse)},
        {"stack-sys", float64(stats.StackSys)},
        {"sys", float64(stats.Sys)},
        {"total-alloc", float64(stats.TotalAlloc)},
    }

    for _, m := range gaugeMetrics {
        err := a.sendMetric("gauge", m.name, fmt.Sprintf("%.2f", m.value))
        if err != nil {
            return err
        }
    }

    return a.sendMetric("counter", "poll-count", strconv.Itoa(a.pollCount))
}

func (a *Agent) sendMetric(t string, name string, value string) error {
    if len(a.addr) == 0 {
        return errors.New("host must be configured")
    }

    resp, err := a.client.R().
        SetHeader(`Content-Type`, `text/plain`).
        Post(fmt.Sprintf(`http://%s/update/%s/%s/%s`, a.addr, t, name, value))

    if err != nil {
        return err
    }

    if resp.StatusCode() != http.StatusOK {
        return fmt.Errorf(`unexpected status code d: %d`, resp.StatusCode())
    }

    a.log.Printf(`metric sent. type: %s, name: %s, value: %s`, t, name, value)
    return nil
}
