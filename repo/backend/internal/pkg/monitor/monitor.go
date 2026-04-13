package monitor

import (
	"math"
	"sort"
	"sync"
	"time"
)

type MetricCollector struct {
	mu          sync.RWMutex
	latencies   []timedValue
	errorCounts map[string]*errorCounter
	maxAge      time.Duration
}

type timedValue struct {
	value float64
	at    time.Time
}

type errorCounter struct {
	count4xx int64
	count5xx int64
}

type PerformanceSnapshot struct {
	P50Latency    float64            `json:"p50_latency_ms"`
	P95Latency    float64            `json:"p95_latency_ms"`
	P99Latency    float64            `json:"p99_latency_ms"`
	RequestCount  int                `json:"request_count"`
	ErrorRates    map[string]int64   `json:"error_rates"`
}

func NewMetricCollector() *MetricCollector {
	return &MetricCollector{
		latencies:   make([]timedValue, 0, 10000),
		errorCounts: make(map[string]*errorCounter),
		maxAge:      5 * time.Minute,
	}
}

func (m *MetricCollector) RecordLatency(latencyMs float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latencies = append(m.latencies, timedValue{value: latencyMs, at: time.Now()})
}

func (m *MetricCollector) RecordError(endpoint string, statusCode int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ec, ok := m.errorCounts[endpoint]
	if !ok {
		ec = &errorCounter{}
		m.errorCounts[endpoint] = ec
	}
	if statusCode >= 400 && statusCode < 500 {
		ec.count4xx++
	} else if statusCode >= 500 {
		ec.count5xx++
	}
}

func (m *MetricCollector) GetSnapshot() PerformanceSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-m.maxAge)

	// Prune old entries
	valid := make([]float64, 0, len(m.latencies))
	kept := make([]timedValue, 0, len(m.latencies))
	for _, tv := range m.latencies {
		if tv.at.After(cutoff) {
			valid = append(valid, tv.value)
			kept = append(kept, tv)
		}
	}
	m.latencies = kept

	sort.Float64s(valid)

	snapshot := PerformanceSnapshot{
		RequestCount: len(valid),
		ErrorRates:   make(map[string]int64),
	}

	if len(valid) > 0 {
		snapshot.P50Latency = percentile(valid, 0.50)
		snapshot.P95Latency = percentile(valid, 0.95)
		snapshot.P99Latency = percentile(valid, 0.99)
	}

	for endpoint, ec := range m.errorCounts {
		snapshot.ErrorRates[endpoint+"_4xx"] = ec.count4xx
		snapshot.ErrorRates[endpoint+"_5xx"] = ec.count5xx
	}

	return snapshot
}

func (m *MetricCollector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latencies = m.latencies[:0]
	m.errorCounts = make(map[string]*errorCounter)
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := p * float64(len(sorted)-1)
	lower := int(math.Floor(idx))
	upper := int(math.Ceil(idx))
	if lower == upper {
		return sorted[lower]
	}
	frac := idx - float64(lower)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
}
