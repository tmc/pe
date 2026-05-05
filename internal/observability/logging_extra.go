package observability

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type correlationIDKey struct{}

// WithCorrelationID returns a context carrying id for log correlation.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey{}, id)
}

// CorrelationIDFromContext returns the correlation ID in ctx, if any.
func CorrelationIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(correlationIDKey{}).(string)
	return id
}

// LogAggregator stores log summaries in memory.
type LogAggregator struct {
	mu      sync.RWMutex
	levels  map[string]int
	entries []LogEntry
}

// NewLogAggregator creates an empty log aggregator.
func NewLogAggregator() *LogAggregator {
	return &LogAggregator{levels: make(map[string]int)}
}

// Add records one log entry.
func (a *LogAggregator) Add(entry LogEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.levels[entry.Level]++
	a.entries = append(a.entries, entry)
}

// Count returns the number of entries at level.
func (a *LogAggregator) Count(level LogLevel) int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.levels[level.String()]
}

// Entries returns a copy of aggregated entries.
func (a *LogAggregator) Entries() []LogEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	entries := make([]LogEntry, len(a.entries))
	copy(entries, a.entries)
	return entries
}

// RotatingLogWriter rotates a log file after it reaches MaxBytes.
type RotatingLogWriter struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	file     *os.File
}

// NewRotatingLogWriter creates a log writer that rotates path at maxBytes.
func NewRotatingLogWriter(path string, maxBytes int64) (*RotatingLogWriter, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("max bytes must be positive")
	}
	w := &RotatingLogWriter{path: path, maxBytes: maxBytes}
	if err := w.open(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *RotatingLogWriter) open() error {
	if err := os.MkdirAll(filepath.Dir(w.path), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file = file
	return nil
}

// Write writes p, rotating before the write if needed.
func (w *RotatingLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	info, err := w.file.Stat()
	if err != nil {
		return 0, err
	}
	if info.Size()+int64(len(p)) > w.maxBytes {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

func (w *RotatingLogWriter) rotate() error {
	if err := w.file.Close(); err != nil {
		return err
	}
	rotated := w.path + ".1"
	_ = os.Remove(rotated)
	if err := os.Rename(w.path, rotated); err != nil && !os.IsNotExist(err) {
		return err
	}
	return w.open()
}

// Close closes the active log file.
func (w *RotatingLogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}
