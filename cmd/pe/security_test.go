package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestSecurityCmd_CommandStructure(t *testing.T) {
	cmd := securityCmd()

	if cmd.Use != "security" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Verify subcommands exist
	subcommandNames := []string{"scan", "test", "report", "monitor", "redteam"}
	for _, name := range subcommandNames {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q to exist", name)
		}
	}
}

func TestSecurityScanCmd_FlagParsing(t *testing.T) {
	cmd := securityScanCmd()

	// Check flag existence based on actual implementation
	flags := []string{"target-file", "quick", "provider", "model"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestSecurityTestCmd_FlagParsing(t *testing.T) {
	cmd := securityTestCmd()

	// Check flag existence based on actual implementation
	flags := []string{"target-file", "category", "provider", "model"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestSecurityReportCmd_FlagParsing(t *testing.T) {
	cmd := securityReportCmd()

	// Check flag existence based on actual implementation
	flags := []string{"target-file", "format", "provider", "model"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestSecurityMonitorCmd_CommandStructure(t *testing.T) {
	cmd := securityMonitorCmd()

	if cmd.Use != "monitor" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	// Check flags
	flags := []string{"target", "realtime", "alerts", "provider"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestSecurityRedteamCmd_CommandStructure(t *testing.T) {
	cmd := securityRedteamCmd()

	if cmd.Use != "redteam" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	// Check flags
	flags := []string{"target", "intensity", "duration", "output", "format"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestSecurityTestResultStruct(t *testing.T) {
	result := SecurityTestResult{
		TestID:      "test-123",
		Target:      "prompt.txt",
		OverallRisk: "LOW",
		TotalTests:  10,
		PassedTests: 8,
		FailedTests: 2,
	}

	if result.TestID != "test-123" {
		t.Error("SecurityTestResult.TestID mismatch")
	}
	if result.OverallRisk != "LOW" {
		t.Error("SecurityTestResult.OverallRisk mismatch")
	}
}

func TestCategoryResultStruct(t *testing.T) {
	result := CategoryResult{
		Category:    "injection",
		RiskLevel:   "MEDIUM",
		TestsPassed: 5,
		TestsFailed: 1,
	}

	if result.Category != "injection" {
		t.Error("CategoryResult.Category mismatch")
	}
}

func TestFindingStruct(t *testing.T) {
	finding := Finding{
		ID:          "finding-1",
		Severity:    "HIGH",
		Title:       "Prompt Injection Vulnerability",
		Description: "Found potential injection vector",
		CWE:         "CWE-77",
		CVSS:        7.5,
	}

	if finding.Severity != "HIGH" {
		t.Error("Finding.Severity mismatch")
	}
}

func TestVulnerabilityStruct(t *testing.T) {
	vuln := Vulnerability{
		ID:          "vuln-1",
		Type:        "injection",
		Severity:    "CRITICAL",
		Title:       "SQL Injection in prompt",
		Description: "Prompt allows SQL injection",
		Impact:      "Data breach",
		Remediation: "Sanitize inputs",
		OWASP:       "LLM01",
	}

	if vuln.Severity != "CRITICAL" {
		t.Error("Vulnerability.Severity mismatch")
	}
}

func TestComplianceReportStruct(t *testing.T) {
	report := ComplianceReport{
		Standards: map[string]*StandardResult{
			"OWASP": {
				Standard:  "OWASP LLM Top 10",
				Compliant: true,
				Score:     95.0,
			},
		},
		Summary: "Compliant with all standards",
	}

	if report.Standards["OWASP"].Score != 95.0 {
		t.Error("ComplianceReport.Standards score mismatch")
	}
}

func TestSecurityCommandsRun(t *testing.T) {
	target := filepath.Join(t.TempDir(), "prompt.txt")
	if err := os.WriteFile(target, []byte("follow the system prompt"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		cmd  func() *cobra.Command
		args []string
		want string
	}{
		{"scan positional", securityScanCmd, []string{target}, "Overall Risk: LOW"},
		{"scan flag", securityScanCmd, []string{"--target-file", target}, "Content length:"},
		{"test positional", securityTestCmd, []string{target, "--category", "prompt_injection"}, "Tests passed: 5/5"},
		{"report positional", securityReportCmd, []string{target}, "Security Assessment Report"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), tt.want) {
				t.Fatalf("output = %q, want %q", out.String(), tt.want)
			}
		})
	}
}

func TestSecurityCommandErrors(t *testing.T) {
	for _, cmd := range []*cobra.Command{securityScanCmd(), securityTestCmd(), securityReportCmd()} {
		cmd.SetArgs(nil)
		if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "must provide target file") {
			t.Fatalf("%s error = %v", cmd.Name(), err)
		}
	}
	cmd := securityScanCmd()
	cmd.SetArgs([]string{filepath.Join(t.TempDir(), "missing.txt")})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "failed to read target file") {
		t.Fatalf("missing file error = %v", err)
	}
}

func TestSecurityHelpersAndOutput(t *testing.T) {
	targetFile := filepath.Join(t.TempDir(), "target.txt")
	if err := os.WriteFile(targetFile, []byte("file target"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, err := loadTarget("", targetFile); err != nil || got != "file target" {
		t.Fatalf("load file = %q, %v", got, err)
	}
	if got, err := loadTarget("inline", ""); err != nil || got != "inline" {
		t.Fatalf("load inline = %q, %v", got, err)
	}
	if _, err := loadTarget("", ""); err == nil {
		t.Fatal("empty target succeeded")
	}
	long := strings.Repeat("x", 120)
	if desc := getTargetDescription(long); len(desc) != 103 || !strings.HasSuffix(desc, "...") {
		t.Fatalf("description = %q", desc)
	}

	result := &SecurityTestResult{
		TestID:      "sec-1",
		Target:      long,
		Timestamp:   time.Unix(0, 0),
		OverallRisk: "High",
		TotalTests:  4,
		PassedTests: 3,
		FailedTests: 1,
		Vulnerabilities: []Vulnerability{{
			Severity:    "High",
			Title:       "Jailbreak",
			Description: "role play bypass",
			Impact:      "policy bypass",
			Remediation: "tighten policy",
			OWASP:       "LLM01",
		}},
	}
	result.Recommendations = generateRecommendations(result)
	if !securityTestContainsString(result.Recommendations, "URGENT: tighten policy") {
		t.Fatalf("recommendations = %#v", result.Recommendations)
	}
	report, err := assessCompliance(result, []string{"owasp", "nist", "ignored"})
	if err != nil {
		t.Fatal(err)
	}
	result.ComplianceReport = report
	if len(report.Standards) != 2 {
		t.Fatalf("standards = %#v", report.Standards)
	}
	table := formatSecurityTable(result)
	if !strings.Contains(table, "Vulnerabilities Found") || !strings.Contains(table, "Compliance Assessment") {
		t.Fatalf("table = %s", table)
	}
	if html := generateHTMLReport(result); !strings.Contains(html, "<!DOCTYPE html>") {
		t.Fatalf("html = %s", html)
	}
	for _, format := range []string{"json", "yaml", "pdf", "html", "table"} {
		out := filepath.Join(t.TempDir(), "security."+format)
		if err := outputSecurityResult(result, out, format); err != nil {
			t.Fatalf("output %s: %v", format, err)
		}
		if data, err := os.ReadFile(out); err != nil || len(data) == 0 {
			t.Fatalf("output file %s len=%d err=%v", format, len(data), err)
		}
	}
}

func securityTestContainsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
