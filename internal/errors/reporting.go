package errors

import (
	"encoding/json"
	"io"
	"sync"
)

// ReportEntry is a structured error report row.
type ReportEntry struct {
	Code      ErrorCode              `json:"code"`
	Message   string                 `json:"message"`
	Severity  Severity               `json:"severity"`
	Retryable bool                   `json:"retryable"`
	Component string                 `json:"component,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ErrorReporter records errors and can emit structured logs.
type ErrorReporter struct {
	mu      sync.Mutex
	entries []ReportEntry
}

// NewReporter creates an empty reporter.
func NewReporter() *ErrorReporter {
	return &ErrorReporter{}
}

// Report records err when non-nil.
func (r *ErrorReporter) Report(err error) {
	if err == nil {
		return
	}
	entry := ReportEntry{
		Code:      GetCode(err),
		Message:   err.Error(),
		Severity:  GetSeverity(err),
		Retryable: IsRetryable(err),
		Component: GetComponent(err),
		Context:   GetContext(err),
	}
	r.mu.Lock()
	r.entries = append(r.entries, entry)
	r.mu.Unlock()
}

// Entries returns a snapshot of reported errors.
func (r *ErrorReporter) Entries() []ReportEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]ReportEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

// Metrics returns aggregate counts by code and severity.
func (r *ErrorReporter) Metrics() ErrorMetrics {
	metrics := ErrorMetrics{
		ByCode:     make(map[ErrorCode]int),
		BySeverity: make(map[Severity]int),
	}
	for _, entry := range r.Entries() {
		metrics.Total++
		metrics.ByCode[entry.Code]++
		metrics.BySeverity[entry.Severity]++
		if entry.Retryable {
			metrics.Retryable++
		}
	}
	return metrics
}

// WriteJSON writes entries as newline-delimited JSON.
func (r *ErrorReporter) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	for _, entry := range r.Entries() {
		if err := enc.Encode(entry); err != nil {
			return err
		}
	}
	return nil
}

// ErrorMetrics summarizes reported errors.
type ErrorMetrics struct {
	Total      int               `json:"total"`
	Retryable  int               `json:"retryable"`
	ByCode     map[ErrorCode]int `json:"by_code"`
	BySeverity map[Severity]int  `json:"by_severity"`
}

// Aggregate combines multiple errors into one chain and reports each input.
func (r *ErrorReporter) Aggregate(errs ...error) error {
	for _, err := range errs {
		r.Report(err)
	}
	return Chain(errs...)
}
