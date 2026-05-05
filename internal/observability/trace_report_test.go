package observability

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestAnalyzeTraceSpans(t *testing.T) {
	spans := []Span{
		{
			ID:        "root",
			Operation: "run",
			StartTime: time.Unix(1, 0),
			Duration:  10 * time.Millisecond,
			Success:   true,
		},
		{
			ID:        "child",
			ParentID:  "root",
			Operation: "provider",
			StartTime: time.Unix(2, 0),
			Duration:  20 * time.Millisecond,
			Success:   false,
			Error:     "boom",
		},
	}

	report := AnalyzeTraceSpans(spans)
	if report.SpanCount != 2 {
		t.Fatalf("SpanCount = %d, want 2", report.SpanCount)
	}
	if report.FailedSpanCount != 1 {
		t.Fatalf("FailedSpanCount = %d, want 1", report.FailedSpanCount)
	}
	if got := report.ByOperation["provider"].Failed; got != 1 {
		t.Fatalf("provider failed = %d, want 1", got)
	}
	if len(report.Roots) != 1 || len(report.Roots[0].Children) != 1 {
		t.Fatalf("trace tree = %+v", report.Roots)
	}
}

func TestReadTraceSpansAndWriteReports(t *testing.T) {
	input := `{"id":"root","operation":"run","start_time":"2026-05-05T00:00:00Z","duration":1000000,"success":true}` + "\n" +
		`{"id":"child","parent_id":"root","operation":"eval","start_time":"2026-05-05T00:00:01Z","duration":2000000,"success":true}` + "\n"

	spans, err := ReadTraceSpans(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadTraceSpans: %v", err)
	}
	report := AnalyzeTraceSpans(spans)

	var text bytes.Buffer
	if err := WriteTextTraceReport(&text, report); err != nil {
		t.Fatalf("WriteTextTraceReport: %v", err)
	}
	for _, want := range []string{"Trace Report", "run", "eval", "Trace tree"} {
		if !strings.Contains(text.String(), want) {
			t.Fatalf("text report missing %q:\n%s", want, text.String())
		}
	}

	var dot bytes.Buffer
	if err := WriteDOTTraceReport(&dot, report); err != nil {
		t.Fatalf("WriteDOTTraceReport: %v", err)
	}
	for _, want := range []string{"digraph trace", `"root" -> "child"`} {
		if !strings.Contains(dot.String(), want) {
			t.Fatalf("dot report missing %q:\n%s", want, dot.String())
		}
	}
}
