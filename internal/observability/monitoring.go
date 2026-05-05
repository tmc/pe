package observability

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// PrometheusExporter writes metrics in Prometheus text format.
type PrometheusExporter struct {
	collector *MetricsCollector
}

// NewPrometheusExporter creates an exporter for collector.
func NewPrometheusExporter(collector *MetricsCollector) *PrometheusExporter {
	return &PrometheusExporter{collector: collector}
}

// WriteTo writes current metrics to w.
func (e *PrometheusExporter) WriteTo(w io.Writer) (int64, error) {
	if e.collector == nil {
		return 0, fmt.Errorf("no metrics collector configured")
	}
	var total int64
	metrics := e.collector.GetMetrics()
	keys := make([]string, 0, len(metrics))
	for key := range metrics {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		metric := metrics[key]
		line := fmt.Sprintf("%s%s %g\n", sanitizeMetricName(metric.Name), prometheusTags(metric.Tags), metric.Value)
		n, err := io.WriteString(w, line)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func prometheusTags(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%q", sanitizeMetricName(key), tags[key]))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func sanitizeMetricName(name string) string {
	var b strings.Builder
	for i, r := range name {
		if r == '_' || r == ':' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || i > 0 && r >= '0' && r <= '9' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	return b.String()
}

// AlertDefinition describes one alert rule.
type AlertDefinition struct {
	Name        string
	Expression  string
	For         time.Duration
	Severity    string
	Description string
}

// SLODefinition describes a service-level objective.
type SLODefinition struct {
	Name        string
	Objective   float64
	Window      time.Duration
	Description string
}

// HealthStatus is a health check result.
type HealthStatus string

const (
	HealthOK   HealthStatus = "ok"
	HealthFail HealthStatus = "fail"
)

// HealthCheck checks one subsystem.
type HealthCheck func(context.Context) error

// HealthChecker runs named health checks.
type HealthChecker struct {
	checks map[string]HealthCheck
}

// NewHealthChecker creates an empty health checker.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{checks: make(map[string]HealthCheck)}
}

// Add registers a named check.
func (h *HealthChecker) Add(name string, check HealthCheck) {
	h.checks[name] = check
}

// Run executes all checks and returns their status.
func (h *HealthChecker) Run(ctx context.Context) map[string]HealthStatus {
	results := make(map[string]HealthStatus, len(h.checks))
	for name, check := range h.checks {
		if err := check(ctx); err != nil {
			results[name] = HealthFail
			continue
		}
		results[name] = HealthOK
	}
	return results
}
