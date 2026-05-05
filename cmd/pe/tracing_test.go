package main

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/observability"
)

type commandTraceWriter struct {
	spans []*observability.Span
}

func (w *commandTraceWriter) WriteSpan(span *observability.Span) error {
	w.spans = append(w.spans, span)
	return nil
}

func (w *commandTraceWriter) Flush() error { return nil }
func (w *commandTraceWriter) Close() error { return nil }

func TestInstrumentCommandTracingRunE(t *testing.T) {
	writer := &commandTraceWriter{}
	observability.InitGlobalTracer(writer)

	var sawSpan bool
	cmd := &cobra.Command{
		Use: "run",
		Annotations: map[string]string{
			commandGroupAnnotation: "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			sawSpan = observability.SpanFromContext(cmd.Context()) != nil
			return nil
		},
	}

	instrumentCommand(cmd)
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("ExecuteContext: %v", err)
	}
	if !sawSpan {
		t.Fatal("RunE did not receive traced context")
	}
	if len(writer.spans) != 1 {
		t.Fatalf("spans = %d, want 1", len(writer.spans))
	}
	span := writer.spans[0]
	if got, want := span.Operation, "command.run"; got != want {
		t.Fatalf("operation = %q, want %q", got, want)
	}
	if got, want := span.Tags["group"], "core"; got != want {
		t.Fatalf("group tag = %q, want %q", got, want)
	}
}

func TestInstrumentCommandTracingError(t *testing.T) {
	writer := &commandTraceWriter{}
	observability.InitGlobalTracer(writer)
	wantErr := errors.New("boom")
	cmd := &cobra.Command{
		Use: "eval",
		RunE: func(cmd *cobra.Command, args []string) error {
			return wantErr
		},
	}

	instrumentCommand(cmd)
	err := cmd.ExecuteContext(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("ExecuteContext error = %v, want %v", err, wantErr)
	}
	if len(writer.spans) != 1 {
		t.Fatalf("spans = %d, want 1", len(writer.spans))
	}
	if writer.spans[0].Success {
		t.Fatal("error span was marked successful")
	}
	if writer.spans[0].Error != "boom" {
		t.Fatalf("span error = %q, want boom", writer.spans[0].Error)
	}
}
