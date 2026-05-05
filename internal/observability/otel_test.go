package observability

import (
	"context"
	"testing"
)

func TestTraceParentRoundTrip(t *testing.T) {
	tc := TraceContext{
		TraceID: "0123456789abcdef0123456789abcdef",
		SpanID:  "0123456789abcdef",
		Sampled: true,
	}
	header := TraceParent(tc)
	got, err := ParseTraceParent(header)
	if err != nil {
		t.Fatalf("ParseTraceParent: %v", err)
	}
	if got != tc {
		t.Fatalf("TraceContext = %+v, want %+v", got, tc)
	}
}

func TestTraceContextContext(t *testing.T) {
	tc := NewTraceContext()
	if len(tc.TraceID) != 32 {
		t.Fatalf("TraceID length = %d, want 32", len(tc.TraceID))
	}
	if len(tc.SpanID) != 16 {
		t.Fatalf("SpanID length = %d, want 16", len(tc.SpanID))
	}
	ctx := WithTraceContext(context.Background(), tc)
	got, ok := TraceContextFromContext(ctx)
	if !ok {
		t.Fatal("TraceContextFromContext ok = false")
	}
	if got != tc {
		t.Fatalf("TraceContext = %+v, want %+v", got, tc)
	}
}

func TestParseTraceParentRejectsInvalid(t *testing.T) {
	tests := []string{
		"",
		"00-short-0123456789abcdef-01",
		"01-0123456789abcdef0123456789abcdef-0123456789abcdef-01",
		"00-0123456789abcdef0123456789abcdef-0123456789abcdef-xx",
	}
	for _, test := range tests {
		if _, err := ParseTraceParent(test); err == nil {
			t.Fatalf("ParseTraceParent(%q) succeeded", test)
		}
	}
}
