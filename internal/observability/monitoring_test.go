package observability

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusExporter(t *testing.T) {
	collector := NewMetricsCollector(nil)
	collector.Counter("provider.request-count", 2, map[string]string{"provider": "openai"})
	collector.Gauge("provider_available", 1, nil)

	var buf bytes.Buffer
	n, err := NewPrometheusExporter(collector).WriteTo(&buf)
	require.NoError(t, err)
	assert.Positive(t, n)
	assert.Contains(t, buf.String(), `provider_request_count{provider="openai"} 2`)
	assert.Contains(t, buf.String(), "provider_available 1")
}

func TestPrometheusExporterRequiresCollector(t *testing.T) {
	var buf bytes.Buffer
	_, err := NewPrometheusExporter(nil).WriteTo(&buf)
	assert.Error(t, err)
}

func TestMonitoringDefinitions(t *testing.T) {
	alert := AlertDefinition{
		Name:       "provider-errors",
		Expression: "rate(provider_error_count[5m]) > 0.1",
		For:        5 * time.Minute,
		Severity:   "warning",
	}
	slo := SLODefinition{
		Name:      "provider-availability",
		Objective: 0.99,
		Window:    30 * 24 * time.Hour,
	}

	assert.Equal(t, "provider-errors", alert.Name)
	assert.Equal(t, 0.99, slo.Objective)
}

func TestHealthChecker(t *testing.T) {
	checker := NewHealthChecker()
	checker.Add("config", func(context.Context) error { return nil })
	checker.Add("provider", func(context.Context) error { return errors.New("down") })

	results := checker.Run(context.Background())

	assert.Equal(t, HealthOK, results["config"])
	assert.Equal(t, HealthFail, results["provider"])
}
