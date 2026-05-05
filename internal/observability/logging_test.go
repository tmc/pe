package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructuredLoggerWritesJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger()
	logger.SetOutput(&buf)
	logger.SetIncludeCaller(false)

	logger.Info("started", String("command", "run"))

	var entry LogEntry
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry))
	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "started", entry.Message)
	assert.Equal(t, "run", entry.Fields["command"])
}

func TestStructuredLoggerFiltersLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger()
	logger.SetOutput(&buf)
	logger.SetIncludeCaller(false)
	logger.SetLevel(WarnLevel)

	logger.Info("hidden")
	logger.Warn("shown")

	assert.NotContains(t, buf.String(), "hidden")
	assert.Contains(t, buf.String(), "shown")
}

func TestTextLogFormatter(t *testing.T) {
	entry := LogEntry{
		Level:   "INFO",
		Message: "ready",
		Fields:  map[string]interface{}{"b": 2, "a": 1},
		TraceID: "trace-1",
	}

	out, err := TextLogFormatter{}.Format(entry)
	require.NoError(t, err)
	text := string(out)
	assert.Contains(t, text, "[INFO] ready")
	assert.Contains(t, text, "trace_id=trace-1")
	assert.True(t, strings.Index(text, "a=1") < strings.Index(text, "b=2"))
}

func TestStructuredLoggerAggregatesEntries(t *testing.T) {
	var buf bytes.Buffer
	aggregator := NewLogAggregator()
	logger := NewStructuredLogger()
	logger.SetOutput(&buf)
	logger.SetIncludeCaller(false)
	logger.SetAggregator(aggregator)

	logger.Info("one")
	logger.Error("two")

	assert.Equal(t, 1, aggregator.Count(InfoLevel))
	assert.Equal(t, 1, aggregator.Count(ErrorLevel))
	require.Len(t, aggregator.Entries(), 2)
}

func TestStructuredLoggerCorrelationID(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger()
	logger.SetOutput(&buf)
	logger.SetIncludeCaller(false)

	ctx := WithCorrelationID(context.Background(), "corr-1")
	logger.WithContext(ctx).Info("correlated")

	var entry LogEntry
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry))
	assert.Equal(t, "corr-1", entry.TraceID)
}

func TestRotatingLogWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pe.log")
	writer, err := NewRotatingLogWriter(path, 16)
	require.NoError(t, err)
	defer writer.Close()

	_, err = writer.Write([]byte("first line\n"))
	require.NoError(t, err)
	_, err = writer.Write([]byte("second line\n"))
	require.NoError(t, err)

	_, err = os.Stat(path)
	require.NoError(t, err)
	_, err = os.Stat(path + ".1")
	require.NoError(t, err)
}

type failingFormatter struct{}

func (failingFormatter) Format(LogEntry) ([]byte, error) { return nil, errors.New("format failed") }

type closeBuffer struct {
	bytes.Buffer
	closed bool
}

func (b *closeBuffer) Close() error {
	b.closed = true
	return nil
}

func TestLogLevelFieldsFormatterFallbackAndClose(t *testing.T) {
	if LogLevel(99).String() != "UNKNOWN" {
		t.Fatal("unknown level string mismatch")
	}
	when := time.Unix(1, 2).UTC()
	fields := []Field{
		String("s", "v"),
		Int("i", 1),
		Int64("i64", 2),
		Float64("f", 1.5),
		Bool("b", true),
		Duration("d", time.Second),
		Time("t", when),
		Error(errors.New("boom")),
		Error(nil),
		Any("a", []string{"x"}),
	}
	got := map[string]interface{}{}
	for _, field := range fields {
		got[field.Key] = field.Value
	}
	if got["s"] != "v" || got["i"] != 1 || got["d"] != "1s" || got["error"] != nil {
		t.Fatalf("fields = %#v", got)
	}

	var buf bytes.Buffer
	logger := NewStructuredLogger()
	logger.SetOutput(&buf)
	logger.SetIncludeCaller(false)
	logger.SetFormatter(nil)
	logger.Info("json")
	if !strings.Contains(buf.String(), `"message":"json"`) {
		t.Fatalf("json output = %s", buf.String())
	}
	buf.Reset()
	logger.SetFormatter(failingFormatter{})
	logger.Info("fallback", String("trace_id", "trace-1"))
	if !strings.Contains(buf.String(), "INFO trace-1: fallback") {
		t.Fatalf("fallback output = %q", buf.String())
	}

	closer := &closeBuffer{}
	logger.SetOutput(closer)
	if err := logger.Close(); err != nil || !closer.closed {
		t.Fatalf("close err=%v closed=%v", err, closer.closed)
	}
	logger.SetOutput(io.Discard)
	if err := logger.Close(); err != nil {
		t.Fatalf("discard close = %v", err)
	}
}

func TestStructuredLoggerContextCallerAndGlobals(t *testing.T) {
	var buf bytes.Buffer
	aggregator := NewLogAggregator()
	logger := NewStructuredLogger()
	logger.SetOutput(&buf)
	logger.SetIncludeCaller(true)
	logger.SetAggregator(aggregator)
	logger.SetTraceEnabled(false)
	logger.With(String("base", "field")).Info("with field", String("trace_id", "trace"), String("span_id", "span"))

	var entry LogEntry
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry))
	assert.Equal(t, "trace", entry.TraceID)
	assert.Equal(t, "span", entry.SpanID)
	assert.Equal(t, "field", entry.Fields["base"])
	assert.NotNil(t, entry.Caller)

	buf.Reset()
	logger.SetTraceEnabled(true)
	ctx := WithCorrelationID(context.Background(), "corr-2")
	logger.WithContext(ctx).Warn("ctx")
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry))
	assert.Equal(t, "corr-2", entry.TraceID)

	old := GetGlobalLogger()
	defer SetGlobalLogger(old)
	global := NewStructuredLogger()
	global.SetOutput(&buf)
	global.SetIncludeCaller(false)
	SetGlobalLogger(global)
	Debug("hidden")
	Info("info")
	Warn("warn")
	LogError("err")
	assert.Contains(t, buf.String(), "info")
	assert.Contains(t, buf.String(), "warn")
	assert.Contains(t, buf.String(), "err")
	WithFields(String("k", "v")).Info("fields")
	WithContext(context.Background()).Info("context")
}

func TestConsoleLoggerAndOperationHelpers(t *testing.T) {
	var buf bytes.Buffer
	logger := NewConsoleLogger()
	logger.SetOutput(&buf)
	logger.SetLevel(DebugLevel)
	for _, level := range []LogLevel{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, LogLevel(99)} {
		_ = logger.levelColor(level)
	}
	logger.Debug("debug", String("k", "v"))
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
	out := buf.String()
	for _, want := range []string{"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]", "k=v"} {
		assert.Contains(t, out, want)
	}

	buf.Reset()
	structured := NewStructuredLogger()
	structured.SetOutput(&buf)
	structured.SetIncludeCaller(false)
	old := GetGlobalLogger()
	defer SetGlobalLogger(old)
	SetGlobalLogger(structured)
	ctx := context.Background()
	LogProviderRequest(ctx, "openai", "gpt", strings.Repeat("p", 120))
	LogProviderResponse(ctx, "openai", "gpt", strings.Repeat("r", 120), time.Second, map[string]int{"total": 3})
	LogProviderError(ctx, "openai", "gpt", errors.New("bad"), time.Second)
	LogCommandStart(ctx, "run", []string{"a"})
	LogCommandEnd(ctx, "run", time.Second, 0)
	LogCommandEnd(ctx, "run", time.Second, 1)
	LogEvaluationStart(ctx, "cfg", 2)
	LogEvaluationEnd(ctx, "cfg", 1, 1, time.Second)
	LogOptimizationIteration(ctx, "gaso", 1, 0.5)
	LogOptimizationComplete(ctx, "gaso", 0.9, 2, time.Second)
	logs := buf.String()
	for _, want := range []string{"Provider request initiated", "Provider request completed", "Provider request failed", "Command started", "Command completed", "Command failed", "Evaluation started", "Evaluation completed", "Optimization completed", "..."} {
		assert.Contains(t, logs, want)
	}
	if truncateString("short", 10) != "short" || truncateString("1234567890", 6) != "123..." {
		t.Fatal("truncate mismatch")
	}
}
