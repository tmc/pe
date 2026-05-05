package observability

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// TraceReport summarizes completed spans.
type TraceReport struct {
	SpanCount       int                       `json:"span_count"`
	FailedSpanCount int                       `json:"failed_span_count"`
	TotalDuration   time.Duration             `json:"total_duration"`
	ByOperation     map[string]OperationTrace `json:"by_operation"`
	Roots           []*TraceNode              `json:"roots,omitempty"`
}

// OperationTrace summarizes spans for one operation.
type OperationTrace struct {
	Count         int           `json:"count"`
	Failed        int           `json:"failed"`
	TotalDuration time.Duration `json:"total_duration"`
	MaxDuration   time.Duration `json:"max_duration"`
}

// TraceNode is a span tree node.
type TraceNode struct {
	Span     Span         `json:"span"`
	Children []*TraceNode `json:"children,omitempty"`
}

// ReadTraceSpans reads newline-delimited JSON spans from r.
func ReadTraceSpans(r io.Reader) ([]Span, error) {
	scanner := bufio.NewScanner(r)
	var spans []Span
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var span Span
		if err := json.Unmarshal([]byte(line), &span); err != nil {
			return nil, fmt.Errorf("decode trace span: %w", err)
		}
		spans = append(spans, span)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read trace spans: %w", err)
	}
	return spans, nil
}

// AnalyzeTraceSpans builds summary and tree data for spans.
func AnalyzeTraceSpans(spans []Span) TraceReport {
	report := TraceReport{
		ByOperation: make(map[string]OperationTrace),
	}
	nodes := make(map[string]*TraceNode)
	for i := range spans {
		span := spans[i]
		report.SpanCount++
		if !span.Success {
			report.FailedSpanCount++
		}
		report.TotalDuration += span.Duration
		op := report.ByOperation[span.Operation]
		op.Count++
		if !span.Success {
			op.Failed++
		}
		op.TotalDuration += span.Duration
		if span.Duration > op.MaxDuration {
			op.MaxDuration = span.Duration
		}
		report.ByOperation[span.Operation] = op
		nodes[span.ID] = &TraceNode{Span: span}
	}
	for _, node := range nodes {
		parent := nodes[node.Span.ParentID]
		if parent == nil {
			report.Roots = append(report.Roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}
	sortTraceNodes(report.Roots)
	return report
}

// WriteTextTraceReport writes a human-readable trace report.
func WriteTextTraceReport(w io.Writer, report TraceReport) error {
	if _, err := fmt.Fprintf(w, "Trace Report\nSpans: %d\nFailed: %d\nTotal duration: %s\n\nOperations:\n",
		report.SpanCount, report.FailedSpanCount, report.TotalDuration); err != nil {
		return err
	}
	ops := make([]string, 0, len(report.ByOperation))
	for op := range report.ByOperation {
		ops = append(ops, op)
	}
	sort.Strings(ops)
	for _, op := range ops {
		summary := report.ByOperation[op]
		if _, err := fmt.Fprintf(w, "  %s count=%d failed=%d total=%s max=%s\n",
			op, summary.Count, summary.Failed, summary.TotalDuration, summary.MaxDuration); err != nil {
			return err
		}
	}
	if len(report.Roots) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "\nTrace tree:"); err != nil {
		return err
	}
	for _, root := range report.Roots {
		if err := writeTraceNode(w, root, ""); err != nil {
			return err
		}
	}
	return nil
}

// WriteDOTTraceReport writes a Graphviz DOT trace graph.
func WriteDOTTraceReport(w io.Writer, report TraceReport) error {
	if _, err := fmt.Fprintln(w, "digraph trace {"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, `  node [shape=box];`); err != nil {
		return err
	}
	for _, root := range report.Roots {
		if err := writeDOTTraceNode(w, root); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, "}")
	return err
}

func writeTraceNode(w io.Writer, node *TraceNode, indent string) error {
	status := "ok"
	if !node.Span.Success {
		status = "failed"
	}
	if _, err := fmt.Fprintf(w, "%s- %s %s %s\n", indent, node.Span.ID, node.Span.Operation, status); err != nil {
		return err
	}
	for _, child := range node.Children {
		if err := writeTraceNode(w, child, indent+"  "); err != nil {
			return err
		}
	}
	return nil
}

func writeDOTTraceNode(w io.Writer, node *TraceNode) error {
	label := fmt.Sprintf("%s\\n%s", node.Span.Operation, node.Span.Duration)
	if _, err := fmt.Fprintf(w, "  %q [label=%q];\n", node.Span.ID, label); err != nil {
		return err
	}
	for _, child := range node.Children {
		if _, err := fmt.Fprintf(w, "  %q -> %q;\n", node.Span.ID, child.Span.ID); err != nil {
			return err
		}
		if err := writeDOTTraceNode(w, child); err != nil {
			return err
		}
	}
	return nil
}

func sortTraceNodes(nodes []*TraceNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Span.StartTime.Before(nodes[j].Span.StartTime)
	})
	for _, node := range nodes {
		sortTraceNodes(node.Children)
	}
}
