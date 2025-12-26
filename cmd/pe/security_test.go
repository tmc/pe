package main

import (
	"testing"
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
