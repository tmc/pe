package observability

import (
	"context"
	"testing"
)

func TestTracerProviderSampler(t *testing.T) {
	writer := NewMockTraceWriter()
	tracer := NewTracerProvider(TraceProvider{
		Writer:  writer,
		Sampler: &RatioSampler{N: 2},
	})

	_, first := tracer.StartSpan(context.Background(), "first")
	if first != nil {
		t.Fatal("first span was sampled, want skipped")
	}
	_, second := tracer.StartSpan(context.Background(), "second")
	if second == nil {
		t.Fatal("second span was skipped, want sampled")
	}
	tracer.FinishSpan(second)
	if writer.GetSpanCount() != 1 {
		t.Fatalf("written spans = %d, want 1", writer.GetSpanCount())
	}
}

func TestTracerSetSamplerNilUsesAlways(t *testing.T) {
	tracer := NewTracer(nil)
	tracer.SetSampler(nil)
	_, span := tracer.StartSpan(context.Background(), "operation")
	if span == nil {
		t.Fatal("span was skipped after nil sampler")
	}
}
