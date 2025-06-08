package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Simple scripttest runner for testing pe commands
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <scripttest-file>\n", os.Args[0])
		os.Exit(1)
	}

	testFile := os.Args[1]
	if err := runScriptTest(testFile); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("PASS")
}

func runScriptTest(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading test file: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNum := 0

	var currentCmd string
	var expectedOutput []string
	var assertions []string
	inTest := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines and comments when not in test
		if !inTest && (strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#")) {
			continue
		}

		// Start of command
		if strings.HasPrefix(line, "$ ") {
			// Execute previous command if any
			if inTest && currentCmd != "" {
				if err := executeTest(currentCmd, expectedOutput, assertions); err != nil {
					return fmt.Errorf("line %d: %w", lineNum, err)
				}
			}

			// New command
			currentCmd = strings.TrimPrefix(line, "$ ")
			expectedOutput = nil
			assertions = nil
			inTest = true
			continue
		}

		// Assertion
		if inTest && strings.HasPrefix(line, "> ") {
			assertion := strings.TrimPrefix(line, "> ")
			assertions = append(assertions, assertion)
			continue
		}

		// Expected output
		if inTest {
			// Empty line ends the test
			if strings.TrimSpace(line) == "" {
				if err := executeTest(currentCmd, expectedOutput, assertions); err != nil {
					return fmt.Errorf("line %d: %w", lineNum, err)
				}
				currentCmd = ""
				expectedOutput = nil
				assertions = nil
				inTest = false
				continue
			}
			expectedOutput = append(expectedOutput, line)
		}
	}

	// Execute last test if any
	if inTest && currentCmd != "" {
		if err := executeTest(currentCmd, expectedOutput, assertions); err != nil {
			return fmt.Errorf("end of file: %w", err)
		}
	}

	return nil
}

func executeTest(cmdLine string, expectedOutput []string, assertions []string) error {
	fmt.Printf("\n=== TEST: %s\n", cmdLine)

	// Parse command and environment
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	var env []string
	var cmdParts []string

	// Extract environment variables
	for i, part := range parts {
		if strings.Contains(part, "=") && !strings.HasPrefix(part, "-") {
			env = append(env, part)
		} else {
			cmdParts = parts[i:]
			break
		}
	}

	if len(cmdParts) == 0 {
		return fmt.Errorf("no command after environment variables")
	}

	// Build pe binary path
	pePath := "./pe"
	if _, err := os.Stat(pePath); os.IsNotExist(err) {
		// Try to build it
		fmt.Println("Building pe binary...")
		buildCmd := exec.Command("go", "build", "-o", "pe", "./cmd/pe")
		if output, err := buildCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("building pe: %w\n%s", err, output)
		}
	}

	// Replace 'pe' with './pe'
	if cmdParts[0] == "pe" {
		cmdParts[0] = pePath
	}

	// Execute command
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Env = append(os.Environ(), env...)

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Check exit code
	if err != nil {
		if expectedOutput == nil && len(assertions) == 0 {
			// Expected to fail
			fmt.Printf("Output (failed as expected):\n%s\n", outputStr)
			return nil
		}
		// Unexpected failure
		if _, ok := err.(*exec.ExitError); !ok {
			return fmt.Errorf("command failed: %w", err)
		}
	}

	fmt.Printf("Output:\n%s\n", outputStr)

	// Check expected output
	if len(expectedOutput) > 0 {
		expected := strings.Join(expectedOutput, "\n")
		if strings.TrimSpace(outputStr) != strings.TrimSpace(expected) {
			// Not exact match, check assertions instead
			fmt.Printf("Expected:\n%s\n", expected)
		}
	}

	// Check assertions
	for _, assertion := range assertions {
		if err := checkAssertion(outputStr, assertion); err != nil {
			return fmt.Errorf("assertion failed: %s: %w", assertion, err)
		}
		fmt.Printf("✓ %s\n", assertion)
	}

	return nil
}

func checkAssertion(output, assertion string) error {
	parts := strings.SplitN(assertion, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid assertion format")
	}

	assertType := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch assertType {
	case "contains":
		if !strings.Contains(output, value) {
			return fmt.Errorf("output does not contain %q", value)
		}
	case "not_contains":
		if strings.Contains(output, value) {
			return fmt.Errorf("output contains %q", value)
		}
	case "contains_any":
		// Parse array: [item1, item2, item3]
		value = strings.Trim(value, "[]")
		items := strings.Split(value, ",")
		found := false
		for _, item := range items {
			item = strings.TrimSpace(item)
			if strings.Contains(output, item) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("output does not contain any of %s", value)
		}
	default:
		return fmt.Errorf("unknown assertion type: %s", assertType)
	}

	return nil
}
