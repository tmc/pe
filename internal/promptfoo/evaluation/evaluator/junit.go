package evaluator

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/tmc/pe/internal/promptfoo"
)

// JUnit XML report types. The shape matches the de-facto JUnit schema consumed
// by CI systems (GitLab, Jenkins, GitHub test reporters): a <testsuites> root
// holding one <testsuite>, each result a <testcase> with an optional <failure>.

type junitTestSuites struct {
	XMLName  xml.Name         `xml:"testsuites"`
	Name     string           `xml:"name,attr"`
	Tests    int              `xml:"tests,attr"`
	Failures int              `xml:"failures,attr"`
	Errors   int              `xml:"errors,attr"`
	Time     float64          `xml:"time,attr"`
	Suites   []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Time      float64         `xml:"time,attr"`
	Timestamp string          `xml:"timestamp,attr,omitempty"`
	Cases     []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      float64       `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
	SystemOut string        `xml:"system-out,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr,omitempty"`
	Body    string `xml:",chardata"`
}

// formatResultsAsJUnit renders evaluation results as a JUnit XML report so CI
// systems can consume pass/fail per test case. Each TestResult becomes a
// <testcase>; a failed result carries a <failure> whose body lists the failing
// assertion reasons.
func formatResultsAsJUnit(results promptfoo.EvaluationResult) ([]byte, error) {
	suite := junitTestSuite{
		Name:      results.EvalID,
		Timestamp: results.Results.Timestamp,
	}

	var totalTime float64
	for _, r := range results.Results.Results {
		seconds := float64(r.LatencyMs) / 1000.0
		totalTime += seconds

		name := r.ID
		if p := r.Prompt["raw"]; p != "" {
			name = fmt.Sprintf("%s: %s", r.ID, truncate(p, 80))
		}
		tc := junitTestCase{
			Name:      name,
			Classname: r.Provider["id"],
			Time:      seconds,
		}
		if !r.Success {
			suite.Failures++
			tc.Failure = &junitFailure{
				Message: firstFailureReason(r),
				Type:    "assertion",
				Body:    failureBody(r),
			}
		}
		suite.Cases = append(suite.Cases, tc)
		suite.Tests++
	}
	suite.Time = totalTime

	report := junitTestSuites{
		Name:     results.EvalID,
		Tests:    suite.Tests,
		Failures: suite.Failures,
		Errors:   suite.Errors,
		Time:     totalTime,
		Suites:   []junitTestSuite{suite},
	}

	body, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal junit: %w", err)
	}
	return append([]byte(xml.Header), append(body, '\n')...), nil
}

// firstFailureReason returns a short message for the testcase failure attribute.
func firstFailureReason(r promptfoo.TestResult) string {
	for _, c := range r.GradingResult.ComponentResults {
		if !c.Pass {
			return fmt.Sprintf("%s: %s", c.Assertion.Type, c.Reason)
		}
	}
	if r.GradingResult.Reason != "" {
		return r.GradingResult.Reason
	}
	return "assertion failed"
}

// failureBody lists every failing assertion plus the produced output for
// debugging from CI logs.
func failureBody(r promptfoo.TestResult) string {
	var b strings.Builder
	for _, c := range r.GradingResult.ComponentResults {
		if !c.Pass {
			fmt.Fprintf(&b, "[%s] %s\n", c.Assertion.Type, c.Reason)
		}
	}
	fmt.Fprintf(&b, "\noutput:\n%s", r.Response.Output)
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
