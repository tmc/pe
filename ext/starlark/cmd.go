package starlark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.starlark.net/starlark"
)

// RunStarlarkTest executes a Starlark test file against a response
func RunStarlarkTest(testFile, response string, testFunction string) error {
	// Read the Starlark test file
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read test file %s: %w", testFile, err)
	}

	// Create evaluator and execute
	evaluator := NewEvaluator()
	globals, err := evaluator.EvalFile(testFile, content)
	if err != nil {
		return fmt.Errorf("failed to execute Starlark file: %w", err)
	}

	// If no specific test function specified, try to find one
	if testFunction == "" {
		testFunction = findTestFunction(globals)
		if testFunction == "" {
			return fmt.Errorf("no test function found in %s", testFile)
		}
	}

	// Run the test
	result, err := evaluator.EvaluateTest(globals, testFunction, response)
	if err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// Output result as JSON
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// ListTestFunctions lists all available test functions in a Starlark file
func ListTestFunctions(testFile string) error {
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read test file %s: %w", testFile, err)
	}

	evaluator := NewEvaluator()
	globals, err := evaluator.EvalFile(testFile, content)
	if err != nil {
		return fmt.Errorf("failed to execute Starlark file: %w", err)
	}

	// Find all functions that start with "test_"
	testFunctions := []string{}
	for name := range globals {
		if strings.HasPrefix(name, "test_") {
			testFunctions = append(testFunctions, name)
		}
	}

	if len(testFunctions) == 0 {
		fmt.Printf("No test functions found in %s\n", testFile)
		return nil
	}

	fmt.Printf("Test functions in %s:\n", testFile)
	for _, fn := range testFunctions {
		fmt.Printf("  - %s\n", fn)
	}

	return nil
}

// RunStarlarkTestSuite runs all test functions in a Starlark file
func RunStarlarkTestSuite(testFile, response string) error {
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read test file %s: %w", testFile, err)
	}

	evaluator := NewEvaluator()
	globals, err := evaluator.EvalFile(testFile, content)
	if err != nil {
		return fmt.Errorf("failed to execute Starlark file: %w", err)
	}

	// Find all test functions
	testFunctions := []string{}
	for name := range globals {
		if strings.HasPrefix(name, "test_") {
			testFunctions = append(testFunctions, name)
		}
	}

	if len(testFunctions) == 0 {
		return fmt.Errorf("no test functions found in %s", testFile)
	}

	// Run all tests and collect results
	suiteResults := make(map[string]*TestResult)
	var totalScore float64
	passedTests := 0

	for _, testName := range testFunctions {
		result, err := evaluator.EvaluateTest(globals, testName, response)
		if err != nil {
			result = &TestResult{
				Pass:   false,
				Reason: fmt.Sprintf("Execution error: %v", err),
			}
		}

		suiteResults[testName] = result
		if result.Pass {
			passedTests++
		}
		totalScore += result.Score
	}

	// Calculate suite summary
	passRate := float64(passedTests) / float64(len(testFunctions))
	avgScore := totalScore / float64(len(testFunctions))

	summary := map[string]interface{}{
		"file":          testFile,
		"total_tests":   len(testFunctions),
		"passed_tests":  passedTests,
		"failed_tests":  len(testFunctions) - passedTests,
		"pass_rate":     passRate,
		"average_score": avgScore,
		"suite_passed":  passRate >= 0.7, // Configurable threshold
		"results":       suiteResults,
	}

	// Output results as JSON
	output, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// ValidateStarlarkFile checks if a Starlark file is syntactically valid
func ValidateStarlarkFile(testFile string) error {
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read test file %s: %w", testFile, err)
	}

	evaluator := NewEvaluator()
	_, err = evaluator.EvalFile(testFile, content)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Printf("✓ %s is valid\n", testFile)
	return nil
}

// FindStarlarkFiles discovers .star files in a directory
func FindStarlarkFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".star") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// findTestFunction finds the first function that starts with "test_"
func findTestFunction(globals starlark.StringDict) string {
	for name := range globals {
		if strings.HasPrefix(name, "test_") {
			return name
		}
	}
	return ""
}

// Example CLI-style functions that could be integrated into PE

// StarlarkEvalCommand implements `pe starlark eval` functionality
func StarlarkEvalCommand(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: pe starlark eval <file.star> <response>")
	}

	testFile := args[0]
	response := args[1]
	testFunction := ""

	if len(args) > 2 {
		testFunction = args[2]
	}

	return RunStarlarkTest(testFile, response, testFunction)
}

// StarlarkListCommand implements `pe starlark list` functionality
func StarlarkListCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pe starlark list <file.star>")
	}

	return ListTestFunctions(args[0])
}

// StarlarkSuiteCommand implements `pe starlark suite` functionality
func StarlarkSuiteCommand(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: pe starlark suite <file.star> <response>")
	}

	return RunStarlarkTestSuite(args[0], args[1])
}

// StarlarkValidateCommand implements `pe starlark validate` functionality
func StarlarkValidateCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pe starlark validate <file.star>")
	}

	return ValidateStarlarkFile(args[0])
}

// StarlarkDiscoverCommand implements `pe starlark discover` functionality
func StarlarkDiscoverCommand(args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	files, err := FindStarlarkFiles(dir)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Printf("No .star files found in %s\n", dir)
		return nil
	}

	fmt.Printf("Starlark test files found:\n")
	for _, file := range files {
		fmt.Printf("  %s\n", file)
	}

	return nil
}
