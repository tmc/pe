package config

import "testing"

func TestValidateConfigReportValid(t *testing.T) {
	report := ValidateConfigReport(DefaultConfig())
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	if len(report.Issues) != 0 {
		t.Fatalf("issues = %+v, want none", report.Issues)
	}
}

func TestValidateConfigReportCollectsIssues(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Providers.Default = ""
	cfg.Eval.MaxConcurrency = 0
	cfg.Optimization.LearningRate = 0

	report := ValidateConfigReport(cfg)
	if report.Valid {
		t.Fatal("report is valid, want invalid")
	}
	if len(report.Issues) < 3 {
		t.Fatalf("issues = %+v, want at least 3", report.Issues)
	}
	assertIssue(t, report, "providers.default")
	assertIssue(t, report, "eval.max_concurrency")
	assertIssue(t, report, "optimization.learning_rate")
}

func TestValidatorCustomValidator(t *testing.T) {
	validator := NewValidator(func(*Config) []ValidationIssue {
		return []ValidationIssue{{Path: "custom", Message: "custom validation failed"}}
	})
	report := validator.Validate(DefaultConfig())
	if report.Valid {
		t.Fatal("report is valid, want invalid")
	}
	assertIssue(t, report, "custom")
}

func assertIssue(t *testing.T, report ValidationReport, path string) {
	t.Helper()
	for _, issue := range report.Issues {
		if issue.Path == path {
			return
		}
	}
	t.Fatalf("missing issue path %q in %+v", path, report.Issues)
}
