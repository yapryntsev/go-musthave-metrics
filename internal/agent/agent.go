// Package agent provides an implementation of a client for gathering and sending performance metrics to the server.
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/docker/docker/pkg/meminfo"
	"github.com/shirou/gopsutil/cpu"
	"github.com/yapryntsev/go-musthave-metrics/internal/middleware"
	"golang.org/x/sync/errgroup"

	"github.com/go-resty/resty/v2"
	models "github.com/yapryntsev/go-musthave-metrics/internal/model"
	"go.uber.org/zap"
)

// Agent handles metric collections and sending them to the server.
type Agent struct {
	mu             sync.RWMutex
	addr           string
	stats          *runtime.MemStats
	log            *zap.Logger
	client         *resty.Client
	reportInterval uint
	pollInterval   uint
	signKey        string

	// Metrics
	pollCount      int
	randValue      float64
	memTotal       int64
	memFree        int64
	cpuUtilization []float64
}

// New creates a new instance of the Agent.
func New(addr string, reportInterval uint, pollInterval uint, signKey string, log *zap.Logger) *Agent {
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	restyClient := resty.NewWithClient(&httpClient).
		SetRetryCount(3).
		SetRetryAfter(
			func(client *resty.Client, response *resty.Response) (time.Duration, error) {
				a := response.Request.Attempt - 1
				return time.Duration(1+2*a) * time.Second, nil
			},
		)

	return &Agent{
		addr:           addr,
		stats:          &runtime.MemStats{},
		log:            log,
		client:         restyClient,
		reportInterval: reportInterval,
		pollInterval:   pollInterval,
		signKey:        signKey,
	}
}

// StartGathering launches a gathering loop that collects and sends metrics based on the reportInterval and pollInterval
// flags.
//
// In order to stop the loop, cancel the context.
func (a *Agent) StartGathering(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(
		func() error {
			a.fetchMetrics(ctx)
			return nil
		},
	)

	g.Go(
		func() error {
			return a.fetchUtilMetrics(ctx)
		},
	)

	g.Go(
		func() error {
			return a.sendMetrics(ctx)
		},
	)

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

// fetchUtilMetrics collects CPU utilization metrics.
//
// Frequency of metric collection is configured by the pollInterval flag.
func (a *Agent) fetchUtilMetrics(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			cpuUtilization, err := cpu.PercentWithContext(ctx, 0, true)
			if err != nil {
				return fmt.Errorf("failed to read cpu utilization info: %w", err)
			}

			info, err := meminfo.Read()
			if err != nil {
				return fmt.Errorf("failed to read memory info: %w", err)
			}

			a.mu.Lock()
			a.cpuUtilization = cpuUtilization
			a.memTotal = info.MemTotal
			a.memFree = info.MemFree
			a.mu.Unlock()
		}
	}
}

// fetchMetrics collects memory allocation statistics.
//
// Frequency of statistic collection is configured by the pollInterval flag.
func (a *Agent) fetchMetrics(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			seed := time.Now().Unix()
			randValue := rand.New(rand.NewSource(seed)).Float64()

			a.mu.Lock()
			a.pollCount++
			a.randValue = randValue

			runtime.ReadMemStats(a.stats)
			a.mu.Unlock()

			a.log.Debug("metric collected")
		}
	}
}

// sendMetrics repeatedly sends collected metrics to the server.
//
// Send frequency is configured by the reportInterval flag.
func (a *Agent) sendMetrics(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			batch := a.makeMetricsBatch()
			err := a.sendBatch(ctx, batch)
			if err != nil {
				return fmt.Errorf("failed to send metrics batch: %w", err)
			}
		}
	}
}

// makeMetricsBatch prepare collected metrics for sending to the server.
func (a *Agent) makeMetricsBatch() []models.Metrics {
	a.mu.RLock()
	stats := a.stats
	a.mu.RUnlock()

	alloc := float64(stats.Alloc)
	bhs := float64(stats.BuckHashSys)
	frees := float64(stats.Frees)
	gccpf := stats.GCCPUFraction
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
	memTotal := float64(a.memTotal)
	memFree := float64(a.memFree)

	metrics := []models.Metrics{
		{ID: "RandomValue", MType: models.Gauge, Value: &a.randValue},
		{ID: "Alloc", MType: models.Gauge, Value: &alloc},
		{ID: "BuckHashSys", MType: models.Gauge, Value: &bhs},
		{ID: "Frees", MType: models.Gauge, Value: &frees},
		{ID: "GCCPUFraction", MType: models.Gauge, Value: &gccpf},
		{ID: "GCSys", MType: models.Gauge, Value: &gcys},
		{ID: "HeapAlloc", MType: models.Gauge, Value: &ha},
		{ID: "HeapIdle", MType: models.Gauge, Value: &hid},
		{ID: "HeapInuse", MType: models.Gauge, Value: &hin},
		{ID: "HeapObjects", MType: models.Gauge, Value: &hob},
		{ID: "HeapReleased", MType: models.Gauge, Value: &hre},
		{ID: "HeapSys", MType: models.Gauge, Value: &hsy},
		{ID: "LastGC", MType: models.Gauge, Value: &lgc},
		{ID: "Lookups", MType: models.Gauge, Value: &lup},
		{ID: "MCacheInuse", MType: models.Gauge, Value: &mci},
		{ID: "MCacheSys", MType: models.Gauge, Value: &mcs},
		{ID: "MSpanInuse", MType: models.Gauge, Value: &msi},
		{ID: "MSpanSys", MType: models.Gauge, Value: &mss},
		{ID: "Mallocs", MType: models.Gauge, Value: &mll},
		{ID: "NextGC", MType: models.Gauge, Value: &ngc},
		{ID: "NumForcedGC", MType: models.Gauge, Value: &nfg},
		{ID: "NumGC", MType: models.Gauge, Value: &nugc},
		{ID: "OtherSys", MType: models.Gauge, Value: &oss},
		{ID: "PauseTotalNs", MType: models.Gauge, Value: &ptn},
		{ID: "StackInuse", MType: models.Gauge, Value: &si},
		{ID: "StackSys", MType: models.Gauge, Value: &ss},
		{ID: "Sys", MType: models.Gauge, Value: &sys},
		{ID: "TotalAlloc", MType: models.Gauge, Value: &ta},
		{ID: "PollCount", MType: models.Counter, Delta: &pc},
		{ID: "TotalMemory", MType: models.Gauge, Value: &memTotal},
		{ID: "FreeMemory", MType: models.Gauge, Value: &memFree},
	}

	for i, v := range a.cpuUtilization {
		vCopy := v
		m := models.Metrics{ID: fmt.Sprintf("CPUutilization%d", i+1), MType: models.Gauge, Value: &vCopy}
		metrics = append(metrics, m)
	}

	return metrics
}

// sendBatch executes an HTTP request that delivers a metrics batch to the server.
func (a *Agent) sendBatch(ctx context.Context, metrics []models.Metrics) error {
	if len(a.addr) == 0 {
		return errors.New("host must be configured")
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	defer func() { _ = zw.Close() }()

	if err := json.NewEncoder(zw).Encode(metrics); err != nil {
		return err
	}

	if err := zw.Flush(); err != nil {
		return err
	}

	req := a.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	if a.signKey != "" {
		hasher := middleware.NewHasher(a.signKey)
		hasher.Write(buf.Bytes())
		hashSum := hasher.Sum(nil)
		req = req.SetHeader(middleware.SignedBodyHeader, hex.EncodeToString(hashSum))
	}

	resp, err := req.
		SetBody(buf.Bytes()).
		Post(fmt.Sprintf("http://%s/updates", a.addr))

	if err != nil {
		a.log.Error("failed to send metrics", zap.Error(err))
		return nil
	}

	a.log.Debug("metrics sent")

	if resp.StatusCode() != http.StatusOK {
		a.log.Error("unexpected status code", zap.Int("code", resp.StatusCode()))
	}

	return nil
}
