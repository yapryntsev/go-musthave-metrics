package agent

import (
	"testing"

	"go.uber.org/zap/zaptest"
)

func BenchmarkAgent_fetchMetrics(b *testing.B) {
	agent := New("localhost:8080", 2, 1, "key", nil, nil, zaptest.NewLogger(b))

	for b.Loop() {
		_ = agent.makeMetricsBatch()
	}
}
