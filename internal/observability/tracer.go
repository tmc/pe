package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// Span represents a traced operation
type Span struct {
	ID        string            `json:"id"`
	ParentID  string            `json:"parent_id,omitempty"`
	Operation string            `json:"operation"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time,omitempty"`
	Duration  time.Duration     `json:"duration"`
	Tags      map[string]string `json:"tags"`
	Events    []Event           `json:"events"`
	Success   bool              `json:"success"`
	Error     string            `json:"error,omitempty"`
}

// Event represents a timestamped event within a span
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	Level     string    `json:"level"` // debug, info, warn, error
}

// Tracer handles distributed tracing
type Tracer struct {
	mu        sync.RWMutex
	spans     map[string]*Span
	writer    TraceWriter
	enabled   bool
	sampler   Sampler
	idCounter int64
}

// Sampler decides whether an operation should be traced.
type Sampler interface {
	Sample(operation string) bool
}

// AlwaysSampler samples every operation.
type AlwaysSampler struct{}

// Sample implements Sampler.
func (AlwaysSampler) Sample(string) bool { return true }

// RatioSampler samples one out of N spans.
type RatioSampler struct {
	N int

	mu    sync.Mutex
	count int
}

// Sample implements Sampler.
func (s *RatioSampler) Sample(string) bool {
	if s.N <= 1 {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count++
	return s.count%s.N == 0
}

// TraceProvider configures trace creation and output.
type TraceProvider struct {
	Writer  TraceWriter
	Sampler Sampler
}

// TraceWriter interface for outputting traces
type TraceWriter interface {
	WriteSpan(span *Span) error
	Flush() error
	Close() error
}

// FileTraceWriter writes traces to a file
type FileTraceWriter struct {
	file *os.File
}

// NewFileTraceWriter creates a new file trace writer
func NewFileTraceWriter(filename string) (*FileTraceWriter, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &FileTraceWriter{file: file}, nil
}

// WriteSpan writes a span to the file
func (w *FileTraceWriter) WriteSpan(span *Span) error {
	data, err := json.Marshal(span)
	if err != nil {
		return err
	}

	_, err = w.file.Write(append(data, '\n'))
	return err
}

// Flush flushes the file buffer
func (w *FileTraceWriter) Flush() error {
	return w.file.Sync()
}

// Close closes the file
func (w *FileTraceWriter) Close() error {
	return w.file.Close()
}

// NewTracer creates a new tracer
func NewTracer(writer TraceWriter) *Tracer {
	return &Tracer{
		spans:   make(map[string]*Span),
		writer:  writer,
		enabled: true,
		sampler: AlwaysSampler{},
	}
}

// NewTracerProvider creates a tracer from provider configuration.
func NewTracerProvider(provider TraceProvider) *Tracer {
	tracer := NewTracer(provider.Writer)
	if provider.Sampler != nil {
		tracer.SetSampler(provider.Sampler)
	}
	return tracer
}

// SetSampler sets the sampler used for future spans.
func (t *Tracer) SetSampler(sampler Sampler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if sampler == nil {
		sampler = AlwaysSampler{}
	}
	t.sampler = sampler
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, operation string) (context.Context, *Span) {
	if !t.enabled {
		return ctx, nil
	}
	if t.sampler != nil && !t.sampler.Sample(operation) {
		return ctx, nil
	}

	t.mu.Lock()
	t.idCounter++
	spanID := fmt.Sprintf("span_%d", t.idCounter)
	t.mu.Unlock()

	span := &Span{
		ID:        spanID,
		Operation: operation,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
		Events:    []Event{},
		Success:   true,
	}

	// Check for parent span
	if parentSpan := SpanFromContext(ctx); parentSpan != nil {
		span.ParentID = parentSpan.ID
	}

	t.mu.Lock()
	t.spans[spanID] = span
	t.mu.Unlock()

	// Add span to context
	ctx = ContextWithSpan(ctx, span)
	if tc, ok := TraceContextFromContext(ctx); ok && tc.TraceID != "" {
		tc.SpanID = randomHex(8)
		ctx = WithTraceContext(ctx, tc)
	} else {
		ctx = WithTraceContext(ctx, TraceContext{
			TraceID: randomHex(16),
			SpanID:  randomHex(8),
			Sampled: true,
		})
	}

	return ctx, span
}

// FinishSpan finishes a span
func (t *Tracer) FinishSpan(span *Span) {
	if !t.enabled || span == nil {
		return
	}

	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)

	// Write span to output
	if t.writer != nil {
		if err := t.writer.WriteSpan(span); err != nil {
			// Log error but don't fail
			fmt.Fprintf(os.Stderr, "Failed to write span: %v\n", err)
		}
	}

	// Remove from active spans
	t.mu.Lock()
	delete(t.spans, span.ID)
	t.mu.Unlock()
}

// AddEvent adds an event to the current span
func (t *Tracer) AddEvent(ctx context.Context, name, message, level string) {
	if !t.enabled {
		return
	}

	span := SpanFromContext(ctx)
	if span == nil {
		return
	}

	event := Event{
		Timestamp: time.Now(),
		Name:      name,
		Message:   message,
		Level:     level,
	}

	span.Events = append(span.Events, event)
}

// SetTag sets a tag on the current span
func (t *Tracer) SetTag(ctx context.Context, key, value string) {
	if !t.enabled {
		return
	}

	span := SpanFromContext(ctx)
	if span == nil {
		return
	}

	span.Tags[key] = value
}

// SetError marks the current span as failed
func (t *Tracer) SetError(ctx context.Context, err error) {
	if !t.enabled {
		return
	}

	span := SpanFromContext(ctx)
	if span == nil {
		return
	}

	span.Success = false
	if err != nil {
		span.Error = err.Error()
	}
}

// Enable enables tracing
func (t *Tracer) Enable() {
	t.enabled = true
}

// Disable disables tracing
func (t *Tracer) Disable() {
	t.enabled = false
}

// Flush flushes all pending traces
func (t *Tracer) Flush() error {
	if t.writer != nil {
		return t.writer.Flush()
	}
	return nil
}

// Close closes the tracer
func (t *Tracer) Close() error {
	if t.writer != nil {
		return t.writer.Close()
	}
	return nil
}

// GetStats returns tracing statistics
func (t *Tracer) GetStats() TracingStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return TracingStats{
		ActiveSpans: len(t.spans),
		Enabled:     t.enabled,
	}
}

// TracingStats contains tracing statistics
type TracingStats struct {
	ActiveSpans int  `json:"active_spans"`
	Enabled     bool `json:"enabled"`
}

// Context key for spans
type spanKey struct{}

// ContextWithSpan adds a span to context
func ContextWithSpan(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanKey{}, span)
}

// SpanFromContext retrieves a span from context
func SpanFromContext(ctx context.Context) *Span {
	span, _ := ctx.Value(spanKey{}).(*Span)
	return span
}

// Global tracer instance
var globalTracer *Tracer
var tracerOnce sync.Once

// GetGlobalTracer returns the global tracer instance
func GetGlobalTracer() *Tracer {
	if globalTracer != nil {
		return globalTracer
	}
	tracerOnce.Do(func() {
		// Default to no-op tracer
		globalTracer = NewTracer(nil)
	})
	return globalTracer
}

// InitGlobalTracer initializes the global tracer
func InitGlobalTracer(writer TraceWriter) {
	globalTracer = NewTracer(writer)
	// Reset the sync.Once so GetGlobalTracer will use the new instance
	tracerOnce = sync.Once{}
}

// Convenience functions using global tracer

// StartSpan starts a span using the global tracer
func StartSpan(ctx context.Context, operation string) (context.Context, func()) {
	tracer := GetGlobalTracer()
	newCtx, span := tracer.StartSpan(ctx, operation)

	// Return a finish function
	finish := func() {
		tracer.FinishSpan(span)
	}

	return newCtx, finish
}

// AddEvent adds an event using the global tracer
func AddEvent(ctx context.Context, name, message, level string) {
	GetGlobalTracer().AddEvent(ctx, name, message, level)
}

// SetTag sets a tag using the global tracer
func SetTag(ctx context.Context, key, value string) {
	GetGlobalTracer().SetTag(ctx, key, value)
}

// SetError sets an error using the global tracer
func SetError(ctx context.Context, err error) {
	GetGlobalTracer().SetError(ctx, err)
}

// TraceFunction is a helper to trace function execution
func TraceFunction(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, finish := StartSpan(ctx, name)
	defer finish()

	// Get function caller info
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			SetTag(ctx, "function", fn.Name())
		}
		SetTag(ctx, "file", file)
		SetTag(ctx, "line", fmt.Sprintf("%d", line))
	}

	err := fn(ctx)
	if err != nil {
		SetError(ctx, err)
	}

	return err
}

// TraceMeasure is a helper to trace and measure execution time
func TraceMeasure(ctx context.Context, name string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	ctx, finish := StartSpan(ctx, name)
	defer finish()

	start := time.Now()
	result, err := fn(ctx)
	duration := time.Since(start)

	SetTag(ctx, "duration_ms", fmt.Sprintf("%.2f", duration.Seconds()*1000))

	if err != nil {
		SetError(ctx, err)
	}

	return result, err
}
