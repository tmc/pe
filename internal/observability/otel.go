package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

type traceContextKey struct{}

// TraceContext carries W3C trace context fields.
type TraceContext struct {
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
	Sampled bool   `json:"sampled"`
}

// NewTraceContext returns a sampled W3C trace context.
func NewTraceContext() TraceContext {
	return TraceContext{
		TraceID: randomHex(16),
		SpanID:  randomHex(8),
		Sampled: true,
	}
}

// WithTraceContext returns a context carrying tc.
func WithTraceContext(ctx context.Context, tc TraceContext) context.Context {
	return context.WithValue(ctx, traceContextKey{}, tc)
}

// TraceContextFromContext returns W3C trace context from ctx.
func TraceContextFromContext(ctx context.Context) (TraceContext, bool) {
	tc, ok := ctx.Value(traceContextKey{}).(TraceContext)
	return tc, ok
}

// TraceParent formats tc as a W3C traceparent header.
func TraceParent(tc TraceContext) string {
	flags := "00"
	if tc.Sampled {
		flags = "01"
	}
	return fmt.Sprintf("00-%s-%s-%s", tc.TraceID, tc.SpanID, flags)
}

// ParseTraceParent parses a W3C traceparent header.
func ParseTraceParent(header string) (TraceContext, error) {
	parts := strings.Split(header, "-")
	if len(parts) != 4 {
		return TraceContext{}, fmt.Errorf("invalid traceparent")
	}
	if parts[0] != "00" {
		return TraceContext{}, fmt.Errorf("unsupported traceparent version")
	}
	if len(parts[1]) != 32 || len(parts[2]) != 16 || len(parts[3]) != 2 {
		return TraceContext{}, fmt.Errorf("invalid traceparent length")
	}
	if _, err := hex.DecodeString(parts[1]); err != nil {
		return TraceContext{}, fmt.Errorf("invalid trace id: %w", err)
	}
	if _, err := hex.DecodeString(parts[2]); err != nil {
		return TraceContext{}, fmt.Errorf("invalid span id: %w", err)
	}
	if _, err := hex.DecodeString(parts[3]); err != nil {
		return TraceContext{}, fmt.Errorf("invalid trace flags: %w", err)
	}
	return TraceContext{
		TraceID: parts[1],
		SpanID:  parts[2],
		Sampled: parts[3] == "01",
	}, nil
}

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		for i := range buf {
			buf[i] = byte(i + 1)
		}
	}
	return hex.EncodeToString(buf)
}
