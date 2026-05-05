package errors

import (
	"bytes"
	stderrors "errors"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/observability"
)

func TestIntegrationProviderReportsAndLogs(t *testing.T) {
	var buf bytes.Buffer
	logger := observability.NewStructuredLogger()
	logger.SetOutput(&buf)
	logger.SetIncludeCaller(false)

	reporter := NewReporter()
	integration := NewIntegration(reporter, logger)

	err := integration.Provider("local", "test-model", stderrors.New("request failed"))
	if err == nil {
		t.Fatal("Provider returned nil")
	}
	if got, want := GetCode(err), ErrCodeProviderAPI; got != want {
		t.Fatalf("code = %s, want %s", got, want)
	}
	if got, want := GetComponent(err), "provider"; got != want {
		t.Fatalf("component = %q, want %q", got, want)
	}
	entries := reporter.Entries()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if got, want := entries[0].Context["provider"], "local"; got != want {
		t.Fatalf("provider context = %v, want %q", got, want)
	}
	log := buf.String()
	for _, want := range []string{"error reported", "PROVIDER_API", "provider", "request failed"} {
		if !strings.Contains(log, want) {
			t.Fatalf("log does not contain %q: %s", want, log)
		}
	}
}

func TestIntegrationPreservesStructuredErrors(t *testing.T) {
	var buf bytes.Buffer
	logger := observability.NewStructuredLogger()
	logger.SetOutput(&buf)
	logger.SetIncludeCaller(false)

	reporter := NewReporter()
	integration := NewIntegration(reporter, logger)
	base := NewEvaluationTimeoutError(30)

	err := integration.Evaluation("smoke", base)
	if err != base {
		t.Fatalf("Evaluation did not preserve structured error")
	}
	if got := len(reporter.Entries()); got != 1 {
		t.Fatalf("entries = %d, want 1", got)
	}
	if got, want := GetContext(err)["suite"], "smoke"; got != want {
		t.Fatalf("suite context = %v, want %q", got, want)
	}
}

func TestIntegrationComponentWrappers(t *testing.T) {
	tests := []struct {
		name      string
		run       func(*Integration, error) error
		wantCode  ErrorCode
		wantComp  string
		wantKey   string
		wantValue string
	}{
		{
			name:      "command",
			run:       func(i *Integration, err error) error { return i.Command("run", err) },
			wantCode:  ErrCodeInternal,
			wantComp:  "command",
			wantKey:   "command",
			wantValue: "run",
		},
		{
			name:      "optimization",
			run:       func(i *Integration, err error) error { return i.Optimization("semantic", err) },
			wantCode:  ErrCodeOptimizationFailed,
			wantComp:  "optimization",
			wantKey:   "method",
			wantValue: "semantic",
		},
		{
			name:      "evaluation",
			run:       func(i *Integration, err error) error { return i.Evaluation("regression", err) },
			wantCode:  ErrCodeEvaluationFailed,
			wantComp:  "evaluation",
			wantKey:   "suite",
			wantValue: "regression",
		},
		{
			name:      "module",
			run:       func(i *Integration, err error) error { return i.Module("example.com/mod", err) },
			wantCode:  ErrCodeModuleInvalid,
			wantComp:  "module",
			wantKey:   "module",
			wantValue: "example.com/mod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := observability.NewStructuredLogger()
			logger.SetOutput(&bytes.Buffer{})
			logger.SetIncludeCaller(false)
			reporter := NewReporter()
			integration := NewIntegration(reporter, logger)

			err := tt.run(integration, stderrors.New("boom"))
			if got := GetCode(err); got != tt.wantCode {
				t.Fatalf("code = %s, want %s", got, tt.wantCode)
			}
			if got := GetComponent(err); got != tt.wantComp {
				t.Fatalf("component = %q, want %q", got, tt.wantComp)
			}
			if got := GetContext(err)[tt.wantKey]; got != tt.wantValue {
				t.Fatalf("context[%s] = %v, want %q", tt.wantKey, got, tt.wantValue)
			}
			if got := len(reporter.Entries()); got != 1 {
				t.Fatalf("entries = %d, want 1", got)
			}
		})
	}
}
