package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    *CallerInfo            `json:"caller,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
}

// LogFormatter formats log entries for output.
type LogFormatter interface {
	Format(entry LogEntry) ([]byte, error)
}

// JSONLogFormatter formats entries as newline-delimited JSON.
type JSONLogFormatter struct{}

// Format formats entry as JSON.
func (JSONLogFormatter) Format(entry LogEntry) ([]byte, error) {
	return json.Marshal(entry)
}

// TextLogFormatter formats entries for human-readable logs.
type TextLogFormatter struct{}

// Format formats entry as text.
func (TextLogFormatter) Format(entry LogEntry) ([]byte, error) {
	line := fmt.Sprintf("%s [%s] %s", entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message)
	if entry.TraceID != "" {
		line += " trace_id=" + entry.TraceID
	}
	if entry.SpanID != "" {
		line += " span_id=" + entry.SpanID
	}
	if len(entry.Fields) > 0 {
		keys := make([]string, 0, len(entry.Fields))
		for k := range entry.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			line += fmt.Sprintf(" %s=%v", k, entry.Fields[k])
		}
	}
	return []byte(line), nil
}

// CallerInfo contains information about the code location that generated the log
type CallerInfo struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// Logger interface defines the logging contract
type Logger interface {
	Debug(message string, fields ...Field)
	Info(message string, fields ...Field)
	Warn(message string, fields ...Field)
	Error(message string, fields ...Field)
	Fatal(message string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
	SetLevel(level LogLevel)
	SetOutput(w io.Writer)
	Close() error
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// Helper functions to create fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

func Error(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// StructuredLogger is the main implementation of the Logger interface
type StructuredLogger struct {
	mu            sync.RWMutex
	level         LogLevel
	output        io.Writer
	formatter     LogFormatter
	fields        map[string]interface{}
	includeCaller bool
	traceEnabled  bool
	aggregator    *LogAggregator
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger() *StructuredLogger {
	return &StructuredLogger{
		level:         InfoLevel,
		output:        os.Stdout,
		formatter:     JSONLogFormatter{},
		fields:        make(map[string]interface{}),
		includeCaller: true,
		traceEnabled:  true,
	}
}

// SetLevel sets the minimum log level
func (l *StructuredLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput sets the output writer
func (l *StructuredLogger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// SetFormatter sets the formatter used for future log entries.
func (l *StructuredLogger) SetFormatter(formatter LogFormatter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if formatter == nil {
		formatter = JSONLogFormatter{}
	}
	l.formatter = formatter
}

// SetIncludeCaller enables or disables caller information in logs
func (l *StructuredLogger) SetIncludeCaller(include bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.includeCaller = include
}

// SetTraceEnabled enables or disables trace context extraction
func (l *StructuredLogger) SetTraceEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.traceEnabled = enabled
}

// SetAggregator records future log entries in aggregator.
func (l *StructuredLogger) SetAggregator(aggregator *LogAggregator) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.aggregator = aggregator
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(message string, fields ...Field) {
	l.log(DebugLevel, message, fields...)
}

// Info logs an info message
func (l *StructuredLogger) Info(message string, fields ...Field) {
	l.log(InfoLevel, message, fields...)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(message string, fields ...Field) {
	l.log(WarnLevel, message, fields...)
}

// Error logs an error message
func (l *StructuredLogger) Error(message string, fields ...Field) {
	l.log(ErrorLevel, message, fields...)
}

// Fatal logs a fatal message and exits
func (l *StructuredLogger) Fatal(message string, fields ...Field) {
	l.log(FatalLevel, message, fields...)
	os.Exit(1)
}

// With returns a new logger with additional fields
func (l *StructuredLogger) With(fields ...Field) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}

	for _, field := range fields {
		newFields[field.Key] = field.Value
	}

	return &StructuredLogger{
		level:         l.level,
		output:        l.output,
		formatter:     l.formatter,
		fields:        newFields,
		includeCaller: l.includeCaller,
		traceEnabled:  l.traceEnabled,
		aggregator:    l.aggregator,
	}
}

// WithContext returns a logger with trace context if available
func (l *StructuredLogger) WithContext(ctx context.Context) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}

	// Extract trace information if available
	if l.traceEnabled {
		if span := SpanFromContext(ctx); span != nil {
			newFields["trace_id"] = span.ID
			if span.ParentID != "" {
				newFields["parent_span_id"] = span.ParentID
			}
		}
		if correlationID := CorrelationIDFromContext(ctx); correlationID != "" {
			newFields["correlation_id"] = correlationID
		}
	}

	return &StructuredLogger{
		level:         l.level,
		output:        l.output,
		formatter:     l.formatter,
		fields:        newFields,
		includeCaller: l.includeCaller,
		traceEnabled:  l.traceEnabled,
		aggregator:    l.aggregator,
	}
}

// log is the internal logging method
func (l *StructuredLogger) log(level LogLevel, message string, fields ...Field) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level.String(),
		Message:   message,
		Fields:    make(map[string]interface{}),
	}

	// Add base fields
	for k, v := range l.fields {
		entry.Fields[k] = v
	}

	// Add provided fields
	for _, field := range fields {
		entry.Fields[field.Key] = field.Value
	}

	// Extract trace IDs if present in fields
	if traceID, ok := entry.Fields["trace_id"].(string); ok && traceID != "" {
		entry.TraceID = traceID
		delete(entry.Fields, "trace_id")
	}
	if correlationID, ok := entry.Fields["correlation_id"].(string); ok && correlationID != "" {
		entry.TraceID = correlationID
		delete(entry.Fields, "correlation_id")
	}
	if spanID, ok := entry.Fields["span_id"].(string); ok && spanID != "" {
		entry.SpanID = spanID
		delete(entry.Fields, "span_id")
	}

	// Add caller information if enabled
	if l.includeCaller {
		if pc, file, line, ok := runtime.Caller(2); ok {
			entry.Caller = &CallerInfo{
				File: file,
				Line: line,
			}
			if fn := runtime.FuncForPC(pc); fn != nil {
				entry.Caller.Function = fn.Name()
			}
		}
	}

	// Remove empty fields map if no additional fields
	if len(entry.Fields) == 0 {
		entry.Fields = nil
	}

	if l.aggregator != nil {
		l.aggregator.Add(entry)
	}

	data, err := l.formatter.Format(entry)
	if err != nil {
		// Fallback to simple text format
		fmt.Fprintf(l.output, "[%s] %s %s: %s\n",
			entry.Timestamp.Format(time.RFC3339),
			entry.Level,
			entry.TraceID,
			entry.Message)
		return
	}

	l.output.Write(data)
	l.output.Write([]byte("\n"))
}

// Close closes any resources used by the logger
func (l *StructuredLogger) Close() error {
	// If output implements io.Closer, close it
	if closer, ok := l.output.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Console logger for human-readable output
type ConsoleLogger struct {
	*StructuredLogger
}

// NewConsoleLogger creates a logger optimized for console output
func NewConsoleLogger() *ConsoleLogger {
	logger := NewStructuredLogger()
	logger.SetFormatter(TextLogFormatter{})
	return &ConsoleLogger{StructuredLogger: logger}
}

// log overrides the structured logger to provide console-friendly output
func (l *ConsoleLogger) log(level LogLevel, message string, fields ...Field) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if level < l.level {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	levelColor := l.levelColor(level)
	resetColor := "\033[0m"

	// Build the log line
	logLine := fmt.Sprintf("%s %s[%s]%s %s", timestamp, levelColor, level.String(), resetColor, message)

	// Add fields if any
	allFields := make(map[string]interface{})
	for k, v := range l.fields {
		allFields[k] = v
	}
	for _, field := range fields {
		allFields[field.Key] = field.Value
	}

	if len(allFields) > 0 {
		logLine += " "
		first := true
		for k, v := range allFields {
			if !first {
				logLine += " "
			}
			logLine += fmt.Sprintf("%s=%v", k, v)
			first = false
		}
	}

	logLine += "\n"
	l.output.Write([]byte(logLine))
}

// levelColor returns ANSI color codes for different log levels
func (l *ConsoleLogger) levelColor(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "\033[36m" // Cyan
	case InfoLevel:
		return "\033[32m" // Green
	case WarnLevel:
		return "\033[33m" // Yellow
	case ErrorLevel:
		return "\033[31m" // Red
	case FatalLevel:
		return "\033[35m" // Magenta
	default:
		return "\033[0m" // Reset
	}
}

// Global logger instance
var (
	globalLogger Logger = NewStructuredLogger()
	loggerMu     sync.RWMutex
)

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger Logger) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return globalLogger
}

// Global logging functions for convenience

// Debug logs a debug message using the global logger
func Debug(message string, fields ...Field) {
	GetGlobalLogger().Debug(message, fields...)
}

// Info logs an info message using the global logger
func Info(message string, fields ...Field) {
	GetGlobalLogger().Info(message, fields...)
}

// Warn logs a warning message using the global logger
func Warn(message string, fields ...Field) {
	GetGlobalLogger().Warn(message, fields...)
}

// LogError logs an error message using the global logger
func LogError(message string, fields ...Field) {
	GetGlobalLogger().Error(message, fields...)
}

// Fatal logs a fatal message using the global logger
func Fatal(message string, fields ...Field) {
	GetGlobalLogger().Fatal(message, fields...)
}

// WithFields returns a logger with additional fields
func WithFields(fields ...Field) Logger {
	return GetGlobalLogger().With(fields...)
}

// WithContext returns a logger with trace context
func WithContext(ctx context.Context) Logger {
	return GetGlobalLogger().WithContext(ctx)
}

// Predefined logging helpers for PE operations

// LogProviderRequest logs a provider request
func LogProviderRequest(ctx context.Context, provider, model string, prompt string) {
	WithContext(ctx).Info("Provider request initiated",
		String("provider", provider),
		String("model", model),
		String("prompt_preview", truncateString(prompt, 100)),
		Int("prompt_length", len(prompt)),
	)
}

// LogProviderResponse logs a provider response
func LogProviderResponse(ctx context.Context, provider, model string, response string, duration time.Duration, tokenUsage map[string]int) {
	fields := []Field{
		String("provider", provider),
		String("model", model),
		String("response_preview", truncateString(response, 100)),
		Int("response_length", len(response)),
		Duration("duration", duration),
	}

	for key, value := range tokenUsage {
		fields = append(fields, Int("tokens_"+key, value))
	}

	WithContext(ctx).Info("Provider request completed", fields...)
}

// LogProviderError logs a provider error
func LogProviderError(ctx context.Context, provider, model string, err error, duration time.Duration) {
	WithContext(ctx).Error("Provider request failed",
		String("provider", provider),
		String("model", model),
		Error(err),
		Duration("duration", duration),
	)
}

// LogCommandStart logs the start of a command
func LogCommandStart(ctx context.Context, command string, args []string) {
	WithContext(ctx).Info("Command started",
		String("command", command),
		Any("args", args),
	)
}

// LogCommandEnd logs the end of a command
func LogCommandEnd(ctx context.Context, command string, duration time.Duration, exitCode int) {
	level := InfoLevel
	if exitCode != 0 {
		level = ErrorLevel
	}

	logger := WithContext(ctx)
	fields := []Field{
		String("command", command),
		Duration("duration", duration),
		Int("exit_code", exitCode),
	}

	if level == ErrorLevel {
		logger.Error("Command failed", fields...)
	} else {
		logger.Info("Command completed", fields...)
	}
}

// LogEvaluationStart logs the start of an evaluation
func LogEvaluationStart(ctx context.Context, configName string, testCount int) {
	WithContext(ctx).Info("Evaluation started",
		String("config", configName),
		Int("test_count", testCount),
	)
}

// LogEvaluationEnd logs the end of an evaluation
func LogEvaluationEnd(ctx context.Context, configName string, passed, failed int, duration time.Duration) {
	successRate := float64(passed) / float64(passed+failed) * 100

	WithContext(ctx).Info("Evaluation completed",
		String("config", configName),
		Int("passed", passed),
		Int("failed", failed),
		Float64("success_rate", successRate),
		Duration("duration", duration),
	)
}

// LogOptimizationIteration logs an optimization iteration
func LogOptimizationIteration(ctx context.Context, method string, iteration int, score float64) {
	WithContext(ctx).Debug("Optimization iteration",
		String("method", method),
		Int("iteration", iteration),
		Float64("score", score),
	)
}

// LogOptimizationComplete logs optimization completion
func LogOptimizationComplete(ctx context.Context, method string, finalScore float64, iterations int, duration time.Duration) {
	WithContext(ctx).Info("Optimization completed",
		String("method", method),
		Float64("final_score", finalScore),
		Int("iterations", iterations),
		Duration("duration", duration),
	)
}

// Helper function to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
