package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
