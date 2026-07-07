package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTraceWriter for testing
type MockTraceWriter struct {
	mu          sync.Mutex
	spans       []*Span
	flushCount  int
	closeCount  int
	shouldError bool
}

func NewMockTraceWriter() *MockTraceWriter {
	return &MockTraceWriter{
		spans: make([]*Span, 0),
	}
}

func (m *MockTraceWriter) WriteSpan(span *Span) error {
	if m.shouldError {
		return assert.AnError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.spans = append(m.spans, span)
	return nil
}

func (m *MockTraceWriter) Flush() error {
	if m.shouldError {
		return assert.AnError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flushCount++
	return nil
}

func (m *MockTraceWriter) Close() error {
	if m.shouldError {
		return assert.AnError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCount++
	return nil
}

func (m *MockTraceWriter) GetSpanCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.spans)
}

func TestNewTracer(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)

	assert.NotNil(t, tracer)
	assert.True(t, tracer.enabled)
	assert.Equal(t, writer, tracer.writer)
	assert.NotNil(t, tracer.spans)
	assert.Len(t, tracer.spans, 0)
	assert.Equal(t, int64(0), tracer.idCounter)
}

func TestTracer_StartFinishSpan(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)
	ctx := context.Background()

	// Start a span
	newCtx, span := tracer.StartSpan(ctx, "test_operation")
	require.NotNil(t, span)
	assert.NotEqual(t, ctx, newCtx) // Context should be different

	// Verify span properties
	assert.NotEmpty(t, span.ID)
	assert.Equal(t, "test_operation", span.Operation)
	assert.False(t, span.StartTime.IsZero())
	assert.True(t, span.EndTime.IsZero()) // Not finished yet
	assert.True(t, span.Success)
	assert.Empty(t, span.Error)
	assert.Empty(t, span.ParentID) // No parent
	assert.NotNil(t, span.Tags)
	assert.NotNil(t, span.Events)

	// Verify span is in active spans
	stats := tracer.GetStats()
	assert.Equal(t, 1, stats.ActiveSpans)

	// Verify span can be retrieved from context
	spanFromCtx := SpanFromContext(newCtx)
	assert.Equal(t, span, spanFromCtx)

	// Finish the span
	tracer.FinishSpan(span)

	// Verify span was finished
	assert.False(t, span.EndTime.IsZero())
	assert.Greater(t, span.Duration, time.Duration(0))

	// Verify span was written
	assert.Len(t, writer.spans, 1)
	assert.Equal(t, span, writer.spans[0])

	// Verify span is no longer active
	stats = tracer.GetStats()
	assert.Equal(t, 0, stats.ActiveSpans)
}

func TestTracer_NestedSpans(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)
	ctx := context.Background()

	// Start parent span
	parentCtx, parentSpan := tracer.StartSpan(ctx, "parent_operation")
	assert.Empty(t, parentSpan.ParentID)

	// Start child span
	childCtx, childSpan := tracer.StartSpan(parentCtx, "child_operation")
	assert.Equal(t, parentSpan.ID, childSpan.ParentID)

	// Start grandchild span
	_, grandchildSpan := tracer.StartSpan(childCtx, "grandchild_operation")
	assert.Equal(t, childSpan.ID, grandchildSpan.ParentID)

	// Verify all spans are active
	stats := tracer.GetStats()
	assert.Equal(t, 3, stats.ActiveSpans)

	// Finish spans in reverse order
	tracer.FinishSpan(grandchildSpan)
	tracer.FinishSpan(childSpan)
	tracer.FinishSpan(parentSpan)

	// Verify all spans were written
	assert.Len(t, writer.spans, 3)

	// Verify no active spans
	stats = tracer.GetStats()
	assert.Equal(t, 0, stats.ActiveSpans)
}

func TestTracer_AddEvent(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)
	ctx := context.Background()

	ctx, span := tracer.StartSpan(ctx, "test_operation")

	// Add events
	tracer.AddEvent(ctx, "start_processing", "Processing started", "info")
	tracer.AddEvent(ctx, "checkpoint", "Reached checkpoint", "debug")
	tracer.AddEvent(ctx, "warning", "Something might be wrong", "warn")

	// Verify events were added
	assert.Len(t, span.Events, 3)

	event1 := span.Events[0]
	assert.Equal(t, "start_processing", event1.Name)
	assert.Equal(t, "Processing started", event1.Message)
	assert.Equal(t, "info", event1.Level)
	assert.False(t, event1.Timestamp.IsZero())

	event2 := span.Events[1]
	assert.Equal(t, "checkpoint", event2.Name)
	assert.Equal(t, "debug", event2.Level)

	event3 := span.Events[2]
	assert.Equal(t, "warning", event3.Name)
	assert.Equal(t, "warn", event3.Level)

	tracer.FinishSpan(span)
}

func TestTracer_SetTag(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)
	ctx := context.Background()

	ctx, span := tracer.StartSpan(ctx, "test_operation")

	// Set tags
	tracer.SetTag(ctx, "service", "test-service")
	tracer.SetTag(ctx, "version", "1.0.0")
	tracer.SetTag(ctx, "user_id", "user123")

	// Verify tags were set
	assert.Len(t, span.Tags, 3)
	assert.Equal(t, "test-service", span.Tags["service"])
	assert.Equal(t, "1.0.0", span.Tags["version"])
	assert.Equal(t, "user123", span.Tags["user_id"])

	// Update existing tag
	tracer.SetTag(ctx, "version", "1.0.1")
	assert.Equal(t, "1.0.1", span.Tags["version"])

	tracer.FinishSpan(span)
}

func TestTracer_SetError(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)
	ctx := context.Background()

	ctx, span := tracer.StartSpan(ctx, "test_operation")

	// Initially successful
	assert.True(t, span.Success)
	assert.Empty(t, span.Error)

	// Set error
	testError := assert.AnError
	tracer.SetError(ctx, testError)

	// Verify error was set
	assert.False(t, span.Success)
	assert.Equal(t, testError.Error(), span.Error)

	tracer.FinishSpan(span)
}

func TestTracer_EnableDisable(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)
	ctx := context.Background()

	// Initially enabled
	assert.True(t, tracer.enabled)

	// Disable tracing
	tracer.Disable()
	assert.False(t, tracer.enabled)

	// Operations should be no-ops when disabled
	newCtx, span := tracer.StartSpan(ctx, "disabled_operation")
	assert.Equal(t, ctx, newCtx) // Context unchanged
	assert.Nil(t, span)

	// Re-enable
	tracer.Enable()
	assert.True(t, tracer.enabled)

	// Should work again
	newCtx, span = tracer.StartSpan(ctx, "enabled_operation")
	assert.NotEqual(t, ctx, newCtx)
	assert.NotNil(t, span)

	tracer.FinishSpan(span)
	assert.Len(t, writer.spans, 1)
}

func TestTracer_FlushClose(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)

	err := tracer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, 1, writer.flushCount)

	err = tracer.Close()
	assert.NoError(t, err)
	assert.Equal(t, 1, writer.closeCount)

	// Test with nil writer
	tracer2 := NewTracer(nil)
	err = tracer2.Flush()
	assert.NoError(t, err)
	err = tracer2.Close()
	assert.NoError(t, err)
}

func TestFileTraceWriter(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/traces.jsonl"

	writer, err := NewFileTraceWriter(filename)
	require.NoError(t, err)
	defer writer.Close()

	// Create test span
	span := &Span{
		ID:        "test-span-1",
		Operation: "test_operation",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(100 * time.Millisecond),
		Duration:  100 * time.Millisecond,
		Tags:      map[string]string{"service": "test"},
		Events: []Event{
			{
				Timestamp: time.Now(),
				Name:      "test_event",
				Message:   "Test message",
				Level:     "info",
			},
		},
		Success: true,
	}

	// Write span
	err = writer.WriteSpan(span)
	require.NoError(t, err)

	// Flush and close
	err = writer.Flush()
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	// Verify file contents
	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 1)

	// Parse span
	var parsedSpan Span
	err = json.Unmarshal([]byte(lines[0]), &parsedSpan)
	require.NoError(t, err)

	assert.Equal(t, span.ID, parsedSpan.ID)
	assert.Equal(t, span.Operation, parsedSpan.Operation)
	assert.Equal(t, span.Success, parsedSpan.Success)
	assert.Len(t, parsedSpan.Events, 1)
}

func TestSpanContext(t *testing.T) {
	ctx := context.Background()

	// Test with no span
	span := SpanFromContext(ctx)
	assert.Nil(t, span)

	// Test with span
	testSpan := &Span{ID: "test-span", Operation: "test"}
	ctx = ContextWithSpan(ctx, testSpan)

	retrievedSpan := SpanFromContext(ctx)
	assert.Equal(t, testSpan, retrievedSpan)

	// Test context replacement
	newSpan := &Span{ID: "new-span", Operation: "new"}
	ctx = ContextWithSpan(ctx, newSpan)

	retrievedSpan = SpanFromContext(ctx)
	assert.Equal(t, newSpan, retrievedSpan)
}

func TestGlobalTracer(t *testing.T) {
	// Reset global state
	globalTracer = nil
	tracerOnce = sync.Once{}

	// Test getting global tracer (should initialize)
	tracer := GetGlobalTracer()
	assert.NotNil(t, tracer)
	assert.Nil(t, tracer.writer) // Default is nil writer

	// Test that subsequent calls return the same instance
	tracer2 := GetGlobalTracer()
	assert.Same(t, tracer, tracer2)

	// Test initializing with custom writer
	writer := NewMockTraceWriter()
	InitGlobalTracer(writer)

	// Should use the new writer
	assert.Same(t, writer, globalTracer.writer)
}

func TestConvenienceFunctions(t *testing.T) {
	// Reset global state
	globalTracer = nil
	tracerOnce = sync.Once{}

	writer := NewMockTraceWriter()
	InitGlobalTracer(writer)

	ctx := context.Background()

	// Test global StartSpan
	newCtx, finish := StartSpan(ctx, "global_operation")
	assert.NotEqual(t, ctx, newCtx)

	span := SpanFromContext(newCtx)
	assert.NotNil(t, span)
	assert.Equal(t, "global_operation", span.Operation)

	// Test global AddEvent
	AddEvent(newCtx, "test_event", "Test message", "info")
	assert.Len(t, span.Events, 1)
	assert.Equal(t, "test_event", span.Events[0].Name)

	// Test global SetTag
	SetTag(newCtx, "test_tag", "test_value")
	assert.Equal(t, "test_value", span.Tags["test_tag"])

	// Test global SetError
	testError := assert.AnError
	SetError(newCtx, testError)
	assert.False(t, span.Success)
	assert.Equal(t, testError.Error(), span.Error)

	// Finish span
	finish()

	// Verify span was written
	assert.Len(t, writer.spans, 1)
}

func TestTraceFunction(t *testing.T) {
	writer := NewMockTraceWriter()
	InitGlobalTracer(writer)

	ctx := context.Background()
	executed := false

	// Test successful function
	err := TraceFunction(ctx, "test_function", func(ctx context.Context) error {
		executed = true

		// Verify span is available in context
		span := SpanFromContext(ctx)
		assert.NotNil(t, span)
		assert.Equal(t, "test_function", span.Operation)

		// Verify function info was tagged
		assert.Contains(t, span.Tags, "function")
		assert.Contains(t, span.Tags, "file")
		assert.Contains(t, span.Tags, "line")

		return nil
	})

	assert.NoError(t, err)
	assert.True(t, executed)
	assert.Len(t, writer.spans, 1)
	assert.True(t, writer.spans[0].Success)

	// Test function with error
	testError := assert.AnError
	err = TraceFunction(ctx, "error_function", func(ctx context.Context) error {
		return testError
	})

	assert.Equal(t, testError, err)
	assert.Len(t, writer.spans, 2)
	assert.False(t, writer.spans[1].Success)
	assert.Equal(t, testError.Error(), writer.spans[1].Error)
}

func TestTraceMeasure(t *testing.T) {
	writer := NewMockTraceWriter()
	InitGlobalTracer(writer)

	ctx := context.Background()

	// Test successful function
	result, err := TraceMeasure(ctx, "measure_function", func(ctx context.Context) (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Len(t, writer.spans, 1)

	span := writer.spans[0]
	assert.True(t, span.Success)
	assert.Contains(t, span.Tags, "duration_ms")

	// Parse duration and verify it's reasonable
	durationStr := span.Tags["duration_ms"]
	assert.NotEmpty(t, durationStr)

	// Test function with error
	testError := assert.AnError
	result, err = TraceMeasure(ctx, "error_measure", func(ctx context.Context) (interface{}, error) {
		return nil, testError
	})

	assert.Equal(t, testError, err)
	assert.Nil(t, result)
	assert.Len(t, writer.spans, 2)
	assert.False(t, writer.spans[1].Success)
}

func TestTracer_AddEventNoSpan(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)
	ctx := context.Background()

	// Adding event without span should be no-op
	tracer.AddEvent(ctx, "orphan_event", "message", "info")

	// No spans should be created or written
	assert.Len(t, writer.spans, 0)
	stats := tracer.GetStats()
	assert.Equal(t, 0, stats.ActiveSpans)
}

func TestTracer_SetTagNoSpan(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)
	ctx := context.Background()

	// Setting tag without span should be no-op
	tracer.SetTag(ctx, "orphan_tag", "value")

	// No spans should be created or written
	assert.Len(t, writer.spans, 0)
	stats := tracer.GetStats()
	assert.Equal(t, 0, stats.ActiveSpans)
}

func TestTracer_SetErrorNoSpan(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)
	ctx := context.Background()

	// Setting error without span should be no-op
	tracer.SetError(ctx, assert.AnError)

	// No spans should be created or written
	assert.Len(t, writer.spans, 0)
	stats := tracer.GetStats()
	assert.Equal(t, 0, stats.ActiveSpans)
}

func TestTracer_WriterError(t *testing.T) {
	writer := NewMockTraceWriter()
	writer.shouldError = true
	tracer := NewTracer(writer)
	ctx := context.Background()

	// Start and finish span - should not panic even if writer fails
	_, span := tracer.StartSpan(ctx, "error_test")
	assert.NotNil(t, span)

	// This should not panic
	tracer.FinishSpan(span)

	// Writer error should not prevent span from being removed from active spans
	stats := tracer.GetStats()
	assert.Equal(t, 0, stats.ActiveSpans)
}

func TestSpan_JSONSerialization(t *testing.T) {
	span := &Span{
		ID:        "test-span-123",
		ParentID:  "parent-span-456",
		Operation: "test_operation",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(100 * time.Millisecond),
		Duration:  100 * time.Millisecond,
		Tags:      map[string]string{"service": "test", "version": "1.0"},
		Events: []Event{
			{
				Timestamp: time.Now(),
				Name:      "event1",
				Message:   "First event",
				Level:     "info",
			},
		},
		Success: true,
		Error:   "",
	}

	// Test JSON marshaling
	data, err := json.Marshal(span)
	require.NoError(t, err)

	var unmarshaled Span
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, span.ID, unmarshaled.ID)
	assert.Equal(t, span.ParentID, unmarshaled.ParentID)
	assert.Equal(t, span.Operation, unmarshaled.Operation)
	assert.Equal(t, span.Success, unmarshaled.Success)
	assert.Equal(t, span.Error, unmarshaled.Error)
	assert.Len(t, unmarshaled.Events, 1)
	assert.Equal(t, span.Events[0].Name, unmarshaled.Events[0].Name)
}

func TestEvent_JSONSerialization(t *testing.T) {
	event := Event{
		Timestamp: time.Now(),
		Name:      "test_event",
		Message:   "Test message",
		Level:     "error",
	}

	// Test JSON marshaling
	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled Event
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.Name, unmarshaled.Name)
	assert.Equal(t, event.Message, unmarshaled.Message)
	assert.Equal(t, event.Level, unmarshaled.Level)
}

// Benchmark tests
func BenchmarkTracer_StartFinishSpan(b *testing.B) {
	tracer := NewTracer(nil) // No writer for benchmarking
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, span := tracer.StartSpan(ctx, "benchmark_operation")
		tracer.FinishSpan(span)
	}
}

func BenchmarkTracer_AddEvent(b *testing.B) {
	tracer := NewTracer(nil)
	ctx := context.Background()
	ctx, span := tracer.StartSpan(ctx, "benchmark_span")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracer.AddEvent(ctx, "benchmark_event", "message", "info")
	}

	tracer.FinishSpan(span)
}

func BenchmarkTracer_SetTag(b *testing.B) {
	tracer := NewTracer(nil)
	ctx := context.Background()
	ctx, span := tracer.StartSpan(ctx, "benchmark_span")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracer.SetTag(ctx, "benchmark_tag", "value")
	}

	tracer.FinishSpan(span)
}

func BenchmarkSpanFromContext(b *testing.B) {
	span := &Span{ID: "benchmark-span"}
	ctx := ContextWithSpan(context.Background(), span)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SpanFromContext(ctx)
	}
}

// Test concurrent access
func TestTracer_ConcurrentAccess(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracer(writer)
	ctx := context.Background()

	const numGoroutines = 50
	const spansPerGoroutine = 10

	done := make(chan bool, numGoroutines)

	// Start multiple goroutines that create spans
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < spansPerGoroutine; j++ {
				spanCtx, span := tracer.StartSpan(ctx, fmt.Sprintf("operation_%d_%d", id, j))

				tracer.AddEvent(spanCtx, "event", "message", "info")
				tracer.SetTag(spanCtx, "goroutine", fmt.Sprintf("%d", id))
				tracer.SetTag(spanCtx, "iteration", fmt.Sprintf("%d", j))

				time.Sleep(time.Microsecond) // Small delay to encourage concurrency

				tracer.FinishSpan(span)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all spans were written
	expectedSpans := numGoroutines * spansPerGoroutine
	assert.Equal(t, expectedSpans, writer.GetSpanCount())

	// Verify no active spans
	stats := tracer.GetStats()
	assert.Equal(t, 0, stats.ActiveSpans)
}
