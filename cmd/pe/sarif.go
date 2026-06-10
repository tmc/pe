package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// SARIF 2.1.0 output for security findings, so results can be uploaded to
// GitHub code scanning and other SARIF-aware tools. Only the subset of the
// schema needed to represent pe's vulnerabilities is modeled.
//
// Reference: https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html

const sarifSchema = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri,omitempty"`
	Rules          []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name,omitempty"`
	ShortDescription sarifText              `json:"shortDescription"`
	FullDescription  *sarifText             `json:"fullDescription,omitempty"`
	HelpURI          string                 `json:"helpUri,omitempty"`
	Properties       map[string]interface{} `json:"properties,omitempty"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifText       `json:"message"`
	Locations []sarifLocation `json:"locations,omitempty"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifText struct {
	Text string `json:"text"`
}

// formatSecuritySARIF renders the security result as a SARIF 2.1.0 log. Each
// distinct vulnerability type becomes a rule; each vulnerability becomes a
// result referencing that rule, with severity mapped to a SARIF level.
func formatSecuritySARIF(result *SecurityTestResult) ([]byte, error) {
	rules := make([]sarifRule, 0)
	seenRule := make(map[string]bool)
	results := make([]sarifResult, 0, len(result.Vulnerabilities))

	for _, v := range result.Vulnerabilities {
		ruleID := v.Type
		if ruleID == "" {
			ruleID = v.ID
		}
		if !seenRule[ruleID] {
			seenRule[ruleID] = true
			rule := sarifRule{
				ID:               ruleID,
				Name:             v.Title,
				ShortDescription: sarifText{Text: v.Title},
			}
			if v.Description != "" {
				rule.FullDescription = &sarifText{Text: v.Description}
			}
			if len(v.References) > 0 {
				rule.HelpURI = v.References[0]
			}
			if v.OWASP != "" {
				rule.Properties = map[string]interface{}{"owasp": v.OWASP}
			}
			rules = append(rules, rule)
		}

		msg := v.Description
		if v.Remediation != "" {
			msg = strings.TrimSpace(msg + "\nRemediation: " + v.Remediation)
		}
		results = append(results, sarifResult{
			RuleID:  ruleID,
			Level:   sarifLevel(v.Severity),
			Message: sarifText{Text: msg},
			Locations: []sarifLocation{{
				PhysicalLocation: sarifPhysicalLocation{
					ArtifactLocation: sarifArtifactLocation{URI: sarifTargetURI(result.Target)},
				},
			}},
		})
	}

	// Stable rule order for deterministic output.
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })

	log := sarifLog{
		Schema:  sarifSchema,
		Version: "2.1.0",
		Runs: []sarifRun{{
			Tool: sarifTool{Driver: sarifDriver{
				Name:           "pe security",
				InformationURI: "https://github.com/tmc/pe",
				Rules:          rules,
			}},
			Results: results,
		}},
	}
	return json.MarshalIndent(log, "", "  ")
}

// sarifLevel maps pe severity strings to SARIF result levels.
func sarifLevel(severity string) string {
	switch strings.ToLower(severity) {
	case "critical", "high":
		return "error"
	case "medium", "moderate":
		return "warning"
	case "low", "info", "informational":
		return "note"
	default:
		return "warning"
	}
}

// sarifTargetURI yields a non-empty artifact URI; SARIF requires a location and
// GitHub code scanning expects a path-like URI even for non-file targets.
func sarifTargetURI(target string) string {
	if target == "" {
		return "pe-security-target"
	}
	return fmt.Sprintf("%v", target)
}
