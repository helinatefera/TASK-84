package unit_tests_test

import (
	"math"
	"testing"

	"github.com/localinsights/portal/internal/pkg/monitor"
)

func TestRecordLatency(t *testing.T) {
	mc := monitor.NewMetricCollector()
	mc.RecordLatency(10.0)
	mc.RecordLatency(20.0)
	mc.RecordLatency(30.0)

	snap := mc.GetSnapshot()
	if snap.RequestCount != 3 {
		t.Errorf("expected RequestCount 3, got %d", snap.RequestCount)
	}
	if snap.P50Latency <= 0 {
		t.Errorf("expected positive P50Latency, got %f", snap.P50Latency)
	}
	if snap.P95Latency <= 0 {
		t.Errorf("expected positive P95Latency, got %f", snap.P95Latency)
	}
	if snap.P99Latency <= 0 {
		t.Errorf("expected positive P99Latency, got %f", snap.P99Latency)
	}
}

func TestPercentileCalculation(t *testing.T) {
	mc := monitor.NewMetricCollector()
	values := []float64{1, 2, 3, 4, 5}
	for _, v := range values {
		mc.RecordLatency(v)
	}

	snap := mc.GetSnapshot()
	// For [1,2,3,4,5], p50 index = 0.5 * 4 = 2.0, so p50 = sorted[2] = 3
	if math.Abs(snap.P50Latency-3.0) > 0.01 {
		t.Errorf("expected P50 = 3.0, got %f", snap.P50Latency)
	}
}

func TestErrorCounting(t *testing.T) {
	mc := monitor.NewMetricCollector()
	mc.RecordError("/api/users", 404)
	mc.RecordError("/api/users", 404)
	mc.RecordError("/api/users", 500)
	mc.RecordError("/api/posts", 503)

	snap := mc.GetSnapshot()

	if snap.ErrorRates["/api/users_4xx"] != 2 {
		t.Errorf("expected /api/users_4xx = 2, got %d", snap.ErrorRates["/api/users_4xx"])
	}
	if snap.ErrorRates["/api/users_5xx"] != 1 {
		t.Errorf("expected /api/users_5xx = 1, got %d", snap.ErrorRates["/api/users_5xx"])
	}
	if snap.ErrorRates["/api/posts_5xx"] != 1 {
		t.Errorf("expected /api/posts_5xx = 1, got %d", snap.ErrorRates["/api/posts_5xx"])
	}
}

func TestReset(t *testing.T) {
	mc := monitor.NewMetricCollector()
	mc.RecordLatency(100.0)
	mc.RecordLatency(200.0)
	mc.RecordError("/api/test", 500)

	mc.Reset()

	snap := mc.GetSnapshot()
	if snap.RequestCount != 0 {
		t.Errorf("expected RequestCount 0 after reset, got %d", snap.RequestCount)
	}
	if snap.P50Latency != 0 {
		t.Errorf("expected P50Latency 0 after reset, got %f", snap.P50Latency)
	}
	if len(snap.ErrorRates) != 0 {
		t.Errorf("expected empty ErrorRates after reset, got %v", snap.ErrorRates)
	}
}
