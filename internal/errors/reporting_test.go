package errors

import (
	"bytes"
	"strings"
	"testing"
)

func TestErrorReporterMetricsAndJSON(t *testing.T) {
	reporter := NewReporter()
	err := New(ErrCodeInvalidInput, "bad input").WithComponent("cmd")
	err.Retryable = true
	reporter.Report(err)

	metrics := reporter.Metrics()
	if metrics.Total != 1 {
		t.Fatalf("Total = %d, want 1", metrics.Total)
	}
	if metrics.Retryable != 1 {
		t.Fatalf("Retryable = %d, want 1", metrics.Retryable)
	}
	if metrics.ByCode[ErrCodeInvalidInput] != 1 {
		t.Fatalf("ByCode invalid input = %d, want 1", metrics.ByCode[ErrCodeInvalidInput])
	}

	var buf bytes.Buffer
	if err := reporter.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	if !strings.Contains(buf.String(), `"code":"INVALID_INPUT"`) {
		t.Fatalf("json output missing code: %s", buf.String())
	}
}

func TestErrorReporterAggregate(t *testing.T) {
	reporter := NewReporter()
	err := reporter.Aggregate(
		New(ErrCodeInvalidInput, "bad input"),
		New(ErrCodeInternal, "internal"),
	)
	if err == nil {
		t.Fatal("Aggregate returned nil")
	}
	if len(reporter.Entries()) != 2 {
		t.Fatalf("entries = %d, want 2", len(reporter.Entries()))
	}
}
