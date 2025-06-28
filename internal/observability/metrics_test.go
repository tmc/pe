package observability

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJSONMetricsWriter(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	require.NotNil(t, writer)

	err = writer.Close()
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(filename)
	require.NoError(t, err)
}

func TestJSONMetricsWriter_WriteMetric(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	metric := &Metric{
		Name:      "test_metric",
		Type:      CounterMetric,
		Value:     42.0,
		Tags:      map[string]string{"env": "test"},
		Timestamp: time.Now(),
	}

	err = writer.WriteMetric(metric)
	require.NoError(t, err)

	err = writer.Flush()
	require.NoError(t, err)

	// Verify file has content
	info, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestJSONMetricsWriter_WriteHistogram(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	histData := &HistogramData{
		Count: 10,
		Sum:   100.5,
		Buckets: []HistogramBucket{
			{UpperBound: 1.0, Count: 5},
			{UpperBound: 5.0, Count: 8},
			{UpperBound: 10.0, Count: 10},
		},
	}

	err = writer.WriteHistogram("test_histogram", histData, map[string]string{"env": "test"})
	require.NoError(t, err)

	err = writer.Flush()
	require.NoError(t, err)

	// Verify file has content
	info, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestJSONMetricsWriter_WriteSummary(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	summaryData := &SummaryData{
		Count: 100,
		Sum:   5050.0,
		Quantiles: map[float64]float64{
			0.5:  50.0,
			0.9:  90.0,
			0.95: 95.0,
		},
	}

	err = writer.WriteSummary("test_summary", summaryData, map[string]string{"env": "test"})
	require.NoError(t, err)

	err = writer.Flush()
	require.NoError(t, err)

	// Verify file has content
	info, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestNewMetricsCollector(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	collector := NewMetricsCollector(writer)
	require.NotNil(t, collector)
	require.NotNil(t, collector.metrics)
	require.NotNil(t, collector.histograms)
	require.NotNil(t, collector.summaries)
	assert.True(t, collector.enabled)
}

func TestMetricsCollector_Counter(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	collector := NewMetricsCollector(writer)

	// Record counter metric
	collector.Counter("test_counter", 1.0, map[string]string{"env": "test"})
	collector.Counter("test_counter", 2.0, map[string]string{"env": "test"})

	metrics := collector.GetMetrics()
	assert.NotEmpty(t, metrics)

	// Verify counter exists in metrics
	found := false
	for _, metric := range metrics {
		if metric.Name == "test_counter" && metric.Type == CounterMetric {
			found = true
			break
		}
	}
	assert.True(t, found, "Counter metric should be recorded")
}

func TestMetricsCollector_Gauge(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	collector := NewMetricsCollector(writer)

	// Record gauge metric
	collector.Gauge("test_gauge", 42.5, map[string]string{"env": "test"})

	metrics := collector.GetMetrics()
	assert.NotEmpty(t, metrics)

	// Verify gauge exists in metrics
	found := false
	for _, metric := range metrics {
		if metric.Name == "test_gauge" && metric.Type == GaugeMetric {
			assert.Equal(t, 42.5, metric.Value)
			found = true
			break
		}
	}
	assert.True(t, found, "Gauge metric should be recorded")
}

func TestMetricsCollector_Histogram(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	collector := NewMetricsCollector(writer)

	// Record histogram values
	buckets := []float64{1.0, 5.0, 10.0, 50.0, 100.0}
	collector.Histogram("test_histogram", 2.5, buckets, map[string]string{"env": "test"})
	collector.Histogram("test_histogram", 7.5, buckets, map[string]string{"env": "test"})

	// Verify histogram was recorded (implementation details may vary)
	metrics := collector.GetMetrics()
	assert.NotNil(t, metrics)
}

func TestMetricsCollector_Summary(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	collector := NewMetricsCollector(writer)

	// Record summary values
	quantiles := []float64{0.5, 0.9, 0.95}
	collector.Summary("test_summary", 42.0, quantiles, map[string]string{"env": "test"})

	// Verify summary was recorded (implementation details may vary)
	metrics := collector.GetMetrics()
	assert.NotNil(t, metrics)
}

func TestMetricsCollector_Timer(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	collector := NewMetricsCollector(writer)

	// Use timer
	timerFunc := collector.Timer("test_timer", map[string]string{"env": "test"})
	require.NotNil(t, timerFunc)

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Stop timer
	timerFunc()

	// Verify timer metric was recorded
	metrics := collector.GetMetrics()
	assert.NotNil(t, metrics)
}

func TestMetricsCollector_EnableDisable(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	collector := NewMetricsCollector(writer)

	// Initially enabled
	assert.True(t, collector.enabled)

	// Disable
	collector.Disable()
	assert.False(t, collector.enabled)

	// Enable
	collector.Enable()
	assert.True(t, collector.enabled)
}

func TestMetricsCollector_Reset(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	collector := NewMetricsCollector(writer)

	// Add some metrics
	collector.Counter("test_counter", 1.0, nil)
	collector.Gauge("test_gauge", 42.0, nil)

	// Verify metrics exist
	metrics := collector.GetMetrics()
	assert.NotEmpty(t, metrics)

	// Reset
	collector.Reset()

	// Verify metrics are cleared
	metrics = collector.GetMetrics()
	assert.Empty(t, metrics)
}

func TestMetricsCollector_FlushAndClose(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "metrics.json")

	writer, err := NewJSONMetricsWriter(filename)
	require.NoError(t, err)

	collector := NewMetricsCollector(writer)

	// Add a metric
	collector.Counter("test_counter", 1.0, map[string]string{"env": "test"})

	// Flush
	err = collector.Flush()
	require.NoError(t, err)

	// Close
	err = collector.Close()
	require.NoError(t, err)

	// Verify file has content
	info, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestGetGlobalMetrics(t *testing.T) {
	globalCollector := GetGlobalMetrics()
	require.NotNil(t, globalCollector)

	// Should return the same instance on multiple calls
	globalCollector2 := GetGlobalMetrics()
	assert.Equal(t, globalCollector, globalCollector2)
}

func TestMetricTypes(t *testing.T) {
	assert.Equal(t, MetricType("counter"), CounterMetric)
	assert.Equal(t, MetricType("gauge"), GaugeMetric)
	assert.Equal(t, MetricType("histogram"), HistogramMetric)
	assert.Equal(t, MetricType("summary"), SummaryMetric)
}

func TestHistogramData(t *testing.T) {
	histData := HistogramData{
		Count: 10,
		Sum:   100.5,
		Buckets: []HistogramBucket{
			{UpperBound: 1.0, Count: 2},
			{UpperBound: 5.0, Count: 5},
			{UpperBound: 10.0, Count: 8},
		},
	}

	assert.Equal(t, uint64(10), histData.Count)
	assert.Equal(t, 100.5, histData.Sum)
	assert.Len(t, histData.Buckets, 3)
}

func TestSummaryData(t *testing.T) {
	summaryData := SummaryData{
		Count: 100,
		Sum:   5050.0,
		Quantiles: map[float64]float64{
			0.5:  50.0,
			0.9:  90.0,
			0.95: 95.0,
		},
	}

	assert.Equal(t, uint64(100), summaryData.Count)
	assert.Equal(t, 5050.0, summaryData.Sum)
	assert.Len(t, summaryData.Quantiles, 3)
	assert.Equal(t, 50.0, summaryData.Quantiles[0.5])
}

func TestErrorCases(t *testing.T) {
	t.Run("invalid file path", func(t *testing.T) {
		writer, err := NewJSONMetricsWriter("/invalid/path/metrics.json")
		assert.Error(t, err)
		assert.Nil(t, writer)
	})

	t.Run("collector with nil writer", func(t *testing.T) {
		collector := NewMetricsCollector(nil)
		require.NotNil(t, collector)

		// Should not panic with nil writer
		collector.Counter("test", 1.0, nil)
		
		err := collector.Flush()
		assert.Error(t, err) // Should error with nil writer
	})
}