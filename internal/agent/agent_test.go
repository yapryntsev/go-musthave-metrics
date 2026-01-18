package agent

import (
    "testing"

    "go.uber.org/zap/zaptest"
)

func BenchmarkAgent_fetchMetrics(b *testing.B) {
    agent := New("localhost:8080", 2, 1, "key", zaptest.NewLogger(b))
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _ = agent.makeMetricsBatch()
    }
}
