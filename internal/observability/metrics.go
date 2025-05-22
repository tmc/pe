package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

// MetricType represents different types of metrics
type MetricType string

const (
	CounterMetric   MetricType = "counter"
	GaugeMetric     MetricType = "gauge"
	HistogramMetric MetricType = "histogram"
	SummaryMetric   MetricType = "summary"
)

// Metric represents a single metric
type Metric struct {
	Name        string            `json:"name"`
	Type        MetricType        `json:"type"`
	Value       float64           `json:"value"`
	Tags        map[string]string `json:"tags"`
	Timestamp   time.Time         `json:"timestamp"`
	Description string            `json:"description"`
}

// HistogramData contains histogram-specific data
type HistogramData struct {
	Buckets []HistogramBucket `json:"buckets"`
	Count   uint64            `json:"count"`
	Sum     float64           `json:"sum"`
}

// HistogramBucket represents a histogram bucket
type HistogramBucket struct {
	UpperBound float64 `json:"upper_bound"`
	Count      uint64  `json:"count"`
}

// SummaryData contains summary-specific data
type SummaryData struct {
	Count      uint64              `json:"count"`
	Sum        float64             `json:"sum"`
	Quantiles  map[float64]float64 `json:"quantiles"`
}

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	mu          sync.RWMutex
	metrics     map[string]*Metric
	histograms  map[string]*HistogramData
	summaries   map[string]*SummaryData
	enabled     bool
	writer      MetricsWriter
}

// MetricsWriter interface for outputting metrics
type MetricsWriter interface {
	WriteMetric(metric *Metric) error
	WriteHistogram(name string, data *HistogramData, tags map[string]string) error
	WriteSummary(name string, data *SummaryData, tags map[string]string) error
	Flush() error
	Close() error
}

// JSONMetricsWriter writes metrics to JSON files
type JSONMetricsWriter struct {
	file *os.File
}

// NewJSONMetricsWriter creates a new JSON metrics writer
func NewJSONMetricsWriter(filename string) (*JSONMetricsWriter, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	
	return &JSONMetricsWriter{file: file}, nil
}

// WriteMetric writes a metric to the file
func (w *JSONMetricsWriter) WriteMetric(metric *Metric) error {
	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}
	
	_, err = w.file.Write(append(data, '\n'))
	return err
}

// WriteHistogram writes a histogram to the file
func (w *JSONMetricsWriter) WriteHistogram(name string, data *HistogramData, tags map[string]string) error {
	entry := map[string]interface{}{
		"name":      name,
		"type":      "histogram",
		"data":      data,
		"tags":      tags,
		"timestamp": time.Now(),
	}
	
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	
	_, err = w.file.Write(append(jsonData, '\n'))
	return err
}

// WriteSummary writes a summary to the file
func (w *JSONMetricsWriter) WriteSummary(name string, data *SummaryData, tags map[string]string) error {
	entry := map[string]interface{}{
		"name":      name,
		"type":      "summary",
		"data":      data,
		"tags":      tags,
		"timestamp": time.Now(),
	}
	
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	
	_, err = w.file.Write(append(jsonData, '\n'))
	return err
}

// Flush flushes the file buffer
func (w *JSONMetricsWriter) Flush() error {
	return w.file.Sync()
}

// Close closes the file
func (w *JSONMetricsWriter) Close() error {
	return w.file.Close()
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(writer MetricsWriter) *MetricsCollector {
	return &MetricsCollector{
		metrics:    make(map[string]*Metric),
		histograms: make(map[string]*HistogramData),
		summaries:  make(map[string]*SummaryData),
		enabled:    true,
		writer:     writer,
	}
}

// Counter increments a counter metric
func (mc *MetricsCollector) Counter(name string, value float64, tags map[string]string) {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	key := mc.metricKey(name, tags)
	if metric, exists := mc.metrics[key]; exists {
		metric.Value += value
		metric.Timestamp = time.Now()
	} else {
		metric = &Metric{
			Name:      name,
			Type:      CounterMetric,
			Value:     value,
			Tags:      tags,
			Timestamp: time.Now(),
		}
		mc.metrics[key] = metric
	}
	
	if mc.writer != nil {
		mc.writer.WriteMetric(mc.metrics[key])
	}
}

// Gauge sets a gauge metric
func (mc *MetricsCollector) Gauge(name string, value float64, tags map[string]string) {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	key := mc.metricKey(name, tags)
	metric := &Metric{
		Name:      name,
		Type:      GaugeMetric,
		Value:     value,
		Tags:      tags,
		Timestamp: time.Now(),
	}
	mc.metrics[key] = metric
	
	if mc.writer != nil {
		mc.writer.WriteMetric(metric)
	}
}

// Histogram observes a value in a histogram
func (mc *MetricsCollector) Histogram(name string, value float64, buckets []float64, tags map[string]string) {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	key := mc.metricKey(name, tags)
	histogram, exists := mc.histograms[key]
	if !exists {
		histogram = &HistogramData{
			Buckets: make([]HistogramBucket, len(buckets)),
		}
		for i, bound := range buckets {
			histogram.Buckets[i] = HistogramBucket{UpperBound: bound}
		}
		mc.histograms[key] = histogram
	}
	
	// Update histogram
	histogram.Count++
	histogram.Sum += value
	
	for i := range histogram.Buckets {
		if value <= histogram.Buckets[i].UpperBound {
			histogram.Buckets[i].Count++
		}
	}
	
	if mc.writer != nil {
		mc.writer.WriteHistogram(name, histogram, tags)
	}
}

// Summary observes a value in a summary
func (mc *MetricsCollector) Summary(name string, value float64, quantiles []float64, tags map[string]string) {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	key := mc.metricKey(name, tags)
	summary, exists := mc.summaries[key]
	if !exists {
		summary = &SummaryData{
			Quantiles: make(map[float64]float64),
		}
		mc.summaries[key] = summary
	}
	
	// Simple implementation - in production, you'd use a more sophisticated algorithm
	summary.Count++
	summary.Sum += value
	
	// Calculate quantiles (simplified)
	for _, q := range quantiles {
		summary.Quantiles[q] = value // Placeholder - would need proper quantile calculation
	}
	
	if mc.writer != nil {
		mc.writer.WriteSummary(name, summary, tags)
	}
}

// Timer measures execution time
func (mc *MetricsCollector) Timer(name string, tags map[string]string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		mc.Histogram(name+"_duration_ms", duration.Seconds()*1000, 
			[]float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000}, tags)
	}
}

// GetMetrics returns all current metrics
func (mc *MetricsCollector) GetMetrics() map[string]*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	metrics := make(map[string]*Metric)
	for k, v := range mc.metrics {
		metrics[k] = v
	}
	return metrics
}

// Reset resets all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics = make(map[string]*Metric)
	mc.histograms = make(map[string]*HistogramData)
	mc.summaries = make(map[string]*SummaryData)
}

// Enable enables metrics collection
func (mc *MetricsCollector) Enable() {
	mc.enabled = true
}

// Disable disables metrics collection
func (mc *MetricsCollector) Disable() {
	mc.enabled = false
}

// Flush flushes all pending metrics
func (mc *MetricsCollector) Flush() error {
	if mc.writer != nil {
		return mc.writer.Flush()
	}
	return nil
}

// Close closes the metrics collector
func (mc *MetricsCollector) Close() error {
	if mc.writer != nil {
		return mc.writer.Close()
	}
	return nil
}

// metricKey generates a unique key for a metric
func (mc *MetricsCollector) metricKey(name string, tags map[string]string) string {
	if len(tags) == 0 {
		return name
	}
	
	// Sort tags for consistent key generation
	var keys []string
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	key := name
	for _, k := range keys {
		key += fmt.Sprintf(",%s=%s", k, tags[k])
	}
	
	return key
}

// MetricsReport contains a summary of collected metrics
type MetricsReport struct {
	Timestamp time.Time            `json:"timestamp"`
	Counters  map[string]float64   `json:"counters"`
	Gauges    map[string]float64   `json:"gauges"`
	Summary   MetricsSummary       `json:"summary"`
}

// MetricsSummary contains high-level metrics summary
type MetricsSummary struct {
	TotalMetrics    int `json:"total_metrics"`
	TotalCounters   int `json:"total_counters"`
	TotalGauges     int `json:"total_gauges"`
	TotalHistograms int `json:"total_histograms"`
	TotalSummaries  int `json:"total_summaries"`
}

// GenerateReport generates a metrics report
func (mc *MetricsCollector) GenerateReport() MetricsReport {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	counters := make(map[string]float64)
	gauges := make(map[string]float64)
	
	for _, metric := range mc.metrics {
		switch metric.Type {
		case CounterMetric:
			counters[metric.Name] = metric.Value
		case GaugeMetric:
			gauges[metric.Name] = metric.Value
		}
	}
	
	return MetricsReport{
		Timestamp: time.Now(),
		Counters:  counters,
		Gauges:    gauges,
		Summary: MetricsSummary{
			TotalMetrics:    len(mc.metrics),
			TotalCounters:   len(counters),
			TotalGauges:     len(gauges),
			TotalHistograms: len(mc.histograms),
			TotalSummaries:  len(mc.summaries),
		},
	}
}

// Global metrics collector
var globalMetrics *MetricsCollector
var metricsOnce sync.Once

// GetGlobalMetrics returns the global metrics collector
func GetGlobalMetrics() *MetricsCollector {
	metricsOnce.Do(func() {
		globalMetrics = NewMetricsCollector(nil)
	})
	return globalMetrics
}

// InitGlobalMetrics initializes the global metrics collector
func InitGlobalMetrics(writer MetricsWriter) {
	globalMetrics = NewMetricsCollector(writer)
}

// Convenience functions using global metrics

// Counter increments a counter using global metrics
func Counter(name string, value float64, tags map[string]string) {
	GetGlobalMetrics().Counter(name, value, tags)
}

// Gauge sets a gauge using global metrics
func Gauge(name string, value float64, tags map[string]string) {
	GetGlobalMetrics().Gauge(name, value, tags)
}

// Timer measures execution time using global metrics
func Timer(name string, tags map[string]string) func() {
	return GetGlobalMetrics().Timer(name, tags)
}

// MeasureLatency is a helper to measure function execution time
func MeasureLatency(ctx context.Context, name string, tags map[string]string, fn func(context.Context) error) error {
	stopTimer := Timer(name, tags)
	defer stopTimer()
	
	return fn(ctx)
}

// MeasureLatencyWithResult is a helper to measure function execution time with result
func MeasureLatencyWithResult(ctx context.Context, name string, tags map[string]string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	stopTimer := Timer(name, tags)
	defer stopTimer()
	
	return fn(ctx)
}