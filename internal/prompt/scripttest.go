package prompt

import (
	"bufio"
	"fmt"
	"strings"
)

// ScriptTest represents a single test case in scripttest format
type ScriptTest struct {
	LineNum    int                   // Line number where test starts
	Variables  map[string]string     // Variable assignments
	Expected   string                // Expected output (exact match)
	Assertions []ScriptTestAssertion // Additional assertions
}

// ScriptTestAssertion represents an assertion in a test
type ScriptTestAssertion struct {
	Type      string
	Value     string
	Threshold float64
}

// ParseScriptTests parses scripttest format from the evals section
func ParseScriptTests(content string) ([]ScriptTest, error) {
	var tests []ScriptTest

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	var currentTest *ScriptTest
	var expectedLines []string
	inTest := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments when not in a test
		if !inTest && (trimmed == "" || strings.HasPrefix(trimmed, "#")) {
			continue
		}

		// Start of a new test
		if strings.HasPrefix(trimmed, "$") {
			// Save previous test if exists
			if currentTest != nil && inTest {
				currentTest.Expected = strings.TrimSpace(strings.Join(expectedLines, "\n"))
				tests = append(tests, *currentTest)
			}

			// Parse new test
			currentTest = &ScriptTest{
				LineNum:   lineNum,
				Variables: make(map[string]string),
			}
			expectedLines = nil
			inTest = true

			// Parse variable assignments
			varPart := strings.TrimPrefix(trimmed, "$")
			varPart = strings.TrimSpace(varPart)

			if err := parseVariables(varPart, currentTest.Variables); err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNum, err)
			}
			continue
		}

		// Assertion line
		if inTest && strings.HasPrefix(trimmed, ">") {
			assertionStr := strings.TrimPrefix(trimmed, ">")
			assertionStr = strings.TrimSpace(assertionStr)

			assertion, err := parseAssertion(assertionStr)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNum, err)
			}

			if currentTest != nil {
				currentTest.Assertions = append(currentTest.Assertions, assertion)
			}
			continue
		}

		// Expected output line
		if inTest {
			// Empty line ends the expected output if we don't have assertions
			if trimmed == "" && len(currentTest.Assertions) == 0 {
				currentTest.Expected = strings.TrimSpace(strings.Join(expectedLines, "\n"))
				tests = append(tests, *currentTest)
				currentTest = nil
				expectedLines = nil
				inTest = false
				continue
			}

			// Otherwise it's part of expected output
			expectedLines = append(expectedLines, line)
		}
	}

	// Save last test if exists
	if currentTest != nil && inTest {
		currentTest.Expected = strings.TrimSpace(strings.Join(expectedLines, "\n"))
		tests = append(tests, *currentTest)
	}

	return tests, nil
}

// parseVariables parses variable assignments like: var1=value1 var2="quoted value"
func parseVariables(input string, vars map[string]string) error {
	// Simple parser for var=value pairs
	// Handles quoted values and escapes

	runes := []rune(input)
	i := 0

	for i < len(runes) {
		// Skip whitespace
		for i < len(runes) && runes[i] == ' ' {
			i++
		}

		if i >= len(runes) {
			break
		}

		// Parse variable name
		nameStart := i
		for i < len(runes) && runes[i] != '=' && runes[i] != ' ' {
			i++
		}

		if i >= len(runes) || runes[i] != '=' {
			return fmt.Errorf("invalid variable assignment at position %d", i)
		}

		name := string(runes[nameStart:i])
		i++ // skip '='

		// Parse value
		var value string
		if i < len(runes) && runes[i] == '"' {
			// Quoted value
			i++ // skip opening quote
			valueStart := i
			escaped := false

			for i < len(runes) {
				if escaped {
					escaped = false
					i++
					continue
				}

				if runes[i] == '\\' {
					escaped = true
					i++
					continue
				}

				if runes[i] == '"' {
					value = string(runes[valueStart:i])
					i++ // skip closing quote
					break
				}

				i++
			}
		} else {
			// Unquoted value - read until space
			valueStart := i
			for i < len(runes) && runes[i] != ' ' {
				i++
			}
			value = string(runes[valueStart:i])
		}

		vars[name] = value
	}

	return nil
}

// parseAssertion parses assertion lines like: contains: value
func parseAssertion(input string) (ScriptTestAssertion, error) {
	parts := strings.SplitN(input, ":", 2)
	if len(parts) != 2 {
		return ScriptTestAssertion{}, fmt.Errorf("invalid assertion format (expected 'type: value')")
	}

	assertion := ScriptTestAssertion{
		Type: strings.TrimSpace(parts[0]),
	}

	valueStr := strings.TrimSpace(parts[1])

	// Handle special assertions with additional parameters
	switch assertion.Type {
	case "similar":
		// Parse: "text" threshold=0.9
		if idx := strings.Index(valueStr, "threshold="); idx > 0 {
			assertion.Value = strings.TrimSpace(valueStr[:idx])
			assertion.Value = strings.Trim(assertion.Value, `"`)

			thresholdStr := strings.TrimSpace(valueStr[idx+10:])
			var threshold float64
			fmt.Sscanf(thresholdStr, "%f", &threshold)
			assertion.Threshold = threshold
		} else {
			assertion.Value = strings.Trim(valueStr, `"`)
			assertion.Threshold = 0.8 // default
		}

	case "contains_any", "not_contains_any":
		// Parse: [item1, item2, item3]
		assertion.Value = valueStr

	default:
		// Simple value
		assertion.Value = strings.Trim(valueStr, `"`)
	}

	return assertion, nil
}

// ConvertToPromptFooTests converts script tests to promptfoo test format
func ConvertScriptTestsToYAML(tests []ScriptTest) string {
	var b strings.Builder

	b.WriteString("tests:\n")

	for i, test := range tests {
		b.WriteString(fmt.Sprintf("  - description: test_%d\n", i+1))

		// Variables
		if len(test.Variables) > 0 {
			b.WriteString("    vars:\n")
			for k, v := range test.Variables {
				// Quote values that might need it
				if strings.Contains(v, ":") || strings.Contains(v, "\n") {
					b.WriteString(fmt.Sprintf("      %s: |\n", k))
					for _, line := range strings.Split(v, "\n") {
						b.WriteString(fmt.Sprintf("        %s\n", line))
					}
				} else {
					b.WriteString(fmt.Sprintf("      %s: %q\n", k, v))
				}
			}
		}

		// Assertions
		var assertions []string

		// Add exact match assertion if we have expected output
		if test.Expected != "" {
			assertions = append(assertions,
				fmt.Sprintf("      - type: equals\n        value: %q", test.Expected))
		}

		// Add other assertions
		for _, assert := range test.Assertions {
			switch assert.Type {
			case "similar":
				assertions = append(assertions,
					fmt.Sprintf("      - type: similar\n        value: %q\n        threshold: %v",
						assert.Value, assert.Threshold))
			case "contains_any", "not_contains_any":
				assertions = append(assertions,
					fmt.Sprintf("      - type: %s\n        value: %s", assert.Type, assert.Value))
			default:
				assertions = append(assertions,
					fmt.Sprintf("      - type: %s\n        value: %q", assert.Type, assert.Value))
			}
		}

		if len(assertions) > 0 {
			b.WriteString("    assert:\n")
			b.WriteString(strings.Join(assertions, "\n"))
			b.WriteString("\n")
		}
	}

	return b.String()
}
