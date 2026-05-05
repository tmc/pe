package errors

import (
	"bytes"
	"testing"
)

type captureNotifier struct {
	entries []ReportEntry
}

func (n *captureNotifier) Notify(entry ReportEntry) error {
	n.entries = append(n.entries, entry)
	return nil
}

func TestErrorReporterDashboard(t *testing.T) {
	reporter := NewReporter()
	reporter.Report(New(ErrCodeFileNotFound, "missing"))
	reporter.Report(New(ErrCodeProviderAuth, "auth"))

	var buf bytes.Buffer
	if err := reporter.WriteDashboard(&buf); err != nil {
		t.Fatalf("WriteDashboard: %v", err)
	}
	if got := reporter.Dashboard().Total; got != 2 {
		t.Fatalf("dashboard total = %d, want 2", got)
	}
	if !bytes.Contains(buf.Bytes(), []byte("FILE_NOT_FOUND: 1")) {
		t.Fatalf("dashboard output = %s", buf.String())
	}
}

func TestErrorReporterNotify(t *testing.T) {
	reporter := NewReporter()
	reporter.Report(New(ErrCodeFileNotFound, "missing"))
	notifier := &captureNotifier{}

	if err := reporter.Notify(notifier); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if len(notifier.entries) != 1 {
		t.Fatalf("notified entries = %d, want 1", len(notifier.entries))
	}
}

func TestTranslation(t *testing.T) {
	err := New(ErrCodeProviderAuth, "auth failed")
	if got := Translation(err, "es"); got != "revise las credenciales configuradas" {
		t.Fatalf("Translation = %q", got)
	}
	if got := Translation(err, "en"); got == "" {
		t.Fatal("english translation is empty")
	}
}
