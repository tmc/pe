package errors

import (
	"fmt"
	"io"
	"sort"
)

// Dashboard summarizes reporter metrics for operators.
type Dashboard struct {
	Total      int               `json:"total"`
	Retryable  int               `json:"retryable"`
	ByCode     map[ErrorCode]int `json:"by_code"`
	BySeverity map[Severity]int  `json:"by_severity"`
}

// Dashboard returns the current reporter dashboard.
func (r *ErrorReporter) Dashboard() Dashboard {
	metrics := r.Metrics()
	return Dashboard{
		Total:      metrics.Total,
		Retryable:  metrics.Retryable,
		ByCode:     metrics.ByCode,
		BySeverity: metrics.BySeverity,
	}
}

// WriteDashboard writes a text dashboard.
func (r *ErrorReporter) WriteDashboard(w io.Writer) error {
	dashboard := r.Dashboard()
	if _, err := fmt.Fprintf(w, "Errors: %d\nRetryable: %d\n", dashboard.Total, dashboard.Retryable); err != nil {
		return err
	}
	codes := make([]string, 0, len(dashboard.ByCode))
	for code := range dashboard.ByCode {
		codes = append(codes, string(code))
	}
	sort.Strings(codes)
	for _, code := range codes {
		if _, err := fmt.Fprintf(w, "%s: %d\n", code, dashboard.ByCode[ErrorCode(code)]); err != nil {
			return err
		}
	}
	return nil
}

// Notifier sends error notifications.
type Notifier interface {
	Notify(ReportEntry) error
}

// Notify sends every current entry to notifier.
func (r *ErrorReporter) Notify(notifier Notifier) error {
	if notifier == nil {
		return fmt.Errorf("notifier is required")
	}
	for _, entry := range r.Entries() {
		if err := notifier.Notify(entry); err != nil {
			return err
		}
	}
	return nil
}

// Translation returns a user-facing error string for err in language.
func Translation(err error, language string) string {
	if language == "" || language == "en" {
		return Suggestion(err)
	}
	code := GetCode(err)
	switch language {
	case "es":
		switch code {
		case ErrCodeProviderAuth, ErrCodeAuthentication:
			return "revise las credenciales configuradas"
		case ErrCodeFileNotFound:
			return "compruebe que la ruta existe"
		}
	}
	return Suggestion(err)
}
