package agent

import (
    "errors"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "runtime"
    "strconv"
    "time"
)

type Agent struct {
    host  string
    stats *runtime.MemStats
    log   *log.Logger
    // Metrics
    pollCount int
    randValue float64
}

func New(host string, port int, log *log.Logger) *Agent {
    return &Agent{
        host:  fmt.Sprintf("%s:%d", host, port),
        stats: &runtime.MemStats{},
        log:   log,
    }
}

func (a *Agent) StartGathering() error {
    var err error

    for {
        time.Sleep(2 * time.Second)

        a.pollCount++
        a.randValue = rand.Float64()

        runtime.ReadMemStats(a.stats)
        a.log.Println("metric collected")

        if a.pollCount%5 == 0 {
            err = a.sendMetrics()
        }

        if err != nil {
            return err
        }
    }
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
    if len(a.host) == 0 {
        return errors.New("host must be configured")
    }

    resp, err := http.Post(
        fmt.Sprintf(`http://%s/update/%s/%s/%s`, a.host, t, name, value),
        "text/plain",
        nil,
    )
    if resp != nil {
        defer resp.Body.Close()
    }

    a.log.Printf(`metric sent. type: %s, name: %s, value: %s`, t, name, value)

    return err
}
