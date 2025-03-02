package tests

import (
	"testing"

	"github.com/tmc/pe/internal/assertutil"
)

func TestAssertions(t *testing.T) {
	// Test data
	sampleOutput := "The capital of France is Paris. It is known as the City of Light."
	
	// Test contains assertion
	t.Run("Contains", func(t *testing.T) {
		assertion := assertutil.Assertion{
			Type:  assertutil.Contains,
			Value: "Paris",
		}
		
		result := assertion.Check(sampleOutput)
		if !result.Success {
			t.Errorf("Expected 'contains' assertion to pass, got failure: %s", result.Reason)
		}
		
		// Test negative case
		assertion.Value = "Tokyo"
		result = assertion.Check(sampleOutput)
		if result.Success {
			t.Error("Expected 'contains' assertion to fail for non-contained string, but it passed")
		}
	})
	
	// Test not-contains assertion
	t.Run("NotContains", func(t *testing.T) {
		assertion := assertutil.Assertion{
			Type:  assertutil.NotContains,
			Value: "Tokyo",
		}
		
		result := assertion.Check(sampleOutput)
		if !result.Success {
			t.Errorf("Expected 'not-contains' assertion to pass, got failure: %s", result.Reason)
		}
		
		// Test negative case
		assertion.Value = "Paris"
		result = assertion.Check(sampleOutput)
		if result.Success {
			t.Error("Expected 'not-contains' assertion to fail for contained string, but it passed")
		}
	})
	
	// Test equals assertion
	t.Run("Equals", func(t *testing.T) {
		assertion := assertutil.Assertion{
			Type:  assertutil.Equals,
			Value: sampleOutput,
		}
		
		result := assertion.Check(sampleOutput)
		if !result.Success {
			t.Errorf("Expected 'equals' assertion to pass, got failure: %s", result.Reason)
		}
		
		// Test negative case
		assertion.Value = "Different text"
		result = assertion.Check(sampleOutput)
		if result.Success {
			t.Error("Expected 'equals' assertion to fail for different string, but it passed")
		}
	})
	
	// Test regex assertion
	t.Run("Regex", func(t *testing.T) {
		assertion := assertutil.Assertion{
			Type:  assertutil.Regex,
			Value: "capital of [A-Za-z]+ is [A-Za-z]+",
		}
		
		result := assertion.Check(sampleOutput)
		if !result.Success {
			t.Errorf("Expected 'regex' assertion to pass, got failure: %s", result.Reason)
		}
		
		// Test negative case
		assertion.Value = "capital of [0-9]+"
		result = assertion.Check(sampleOutput)
		if result.Success {
			t.Error("Expected 'regex' assertion to fail for non-matching pattern, but it passed")
		}
		
		// Test invalid regex
		assertion.Value = "["  // Invalid regex
		result = assertion.Check(sampleOutput)
		if result.Success {
			t.Error("Expected 'regex' assertion to fail for invalid regex, but it passed")
		}
	})
	
	// Test starts-with assertion
	t.Run("StartsWith", func(t *testing.T) {
		assertion := assertutil.Assertion{
			Type:  assertutil.StartsWith,
			Value: "The capital",
		}
		
		result := assertion.Check(sampleOutput)
		if !result.Success {
			t.Errorf("Expected 'starts-with' assertion to pass, got failure: %s", result.Reason)
		}
		
		// Test negative case
		assertion.Value = "Paris"
		result = assertion.Check(sampleOutput)
		if result.Success {
			t.Error("Expected 'starts-with' assertion to fail for non-prefix string, but it passed")
		}
	})
	
	// Test ends-with assertion
	t.Run("EndsWith", func(t *testing.T) {
		assertion := assertutil.Assertion{
			Type:  assertutil.EndsWith,
			Value: "City of Light.",
		}
		
		result := assertion.Check(sampleOutput)
		if !result.Success {
			t.Errorf("Expected 'ends-with' assertion to pass, got failure: %s", result.Reason)
		}
		
		// Test negative case
		assertion.Value = "Paris"
		result = assertion.Check(sampleOutput)
		if result.Success {
			t.Error("Expected 'ends-with' assertion to fail for non-suffix string, but it passed")
		}
	})
	
	// Test length assertion
	t.Run("Length", func(t *testing.T) {
		// Test exact length
		assertion := assertutil.Assertion{
			Type:  assertutil.Length,
			Value: float64(len(sampleOutput)),
		}
		
		result := assertion.Check(sampleOutput)
		if !result.Success {
			t.Errorf("Expected 'length' assertion to pass, got failure: %s", result.Reason)
		}
		
		// Test length with comparison
		assertion = assertutil.Assertion{
			Type:  assertutil.Length,
			Value: ">10",
		}
		
		result = assertion.Check(sampleOutput)
		if !result.Success {
			t.Errorf("Expected 'length >10' assertion to pass, got failure: %s", result.Reason)
		}
		
		// Test negative case
		assertion.Value = "<10"
		result = assertion.Check(sampleOutput)
		if result.Success {
			t.Error("Expected 'length <10' assertion to fail for longer string, but it passed")
		}
	})
	
	// Test CheckAll and AllPassed
	t.Run("CheckAllAndAllPassed", func(t *testing.T) {
		assertions := []assertutil.Assertion{
			{Type: assertutil.Contains, Value: "Paris"},
			{Type: assertutil.NotContains, Value: "Tokyo"},
			{Type: assertutil.StartsWith, Value: "The"},
		}
		
		results := assertutil.CheckAll(sampleOutput, assertions)
		
		if len(results) != len(assertions) {
			t.Fatalf("Expected %d results, got %d", len(assertions), len(results))
		}
		
		if !assertutil.AllPassed(results) {
			t.Error("Expected all assertions to pass, but AllPassed returned false")
		}
		
		// Add a failing assertion
		assertions = append(assertions, assertutil.Assertion{
			Type:  assertutil.Contains,
			Value: "London",
		})
		
		results = assertutil.CheckAll(sampleOutput, assertions)
		
		if assertutil.AllPassed(results) {
			t.Error("Expected AllPassed to return false when one assertion fails, but it returned true")
		}
	})
	
	// Test ParseFromMap
	t.Run("ParseFromMap", func(t *testing.T) {
		assertMap := map[string]interface{}{
			"type":  "contains",
			"value": "Paris",
		}
		
		assertion, err := assertutil.ParseFromMap(assertMap)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if assertion.Type != assertutil.Contains {
			t.Errorf("Expected type 'contains', got: %s", assertion.Type)
		}
		
		if assertion.Value != "Paris" {
			t.Errorf("Expected value 'Paris', got: %v", assertion.Value)
		}
		
		// Test with missing fields
		badMap := map[string]interface{}{
			"type": "contains",
			// Missing "value" field
		}
		
		_, err = assertutil.ParseFromMap(badMap)
		if err == nil {
			t.Error("Expected error for missing 'value' field, got nil")
		}
	})
	
	// Test ParseAll
	t.Run("ParseAll", func(t *testing.T) {
		assertionsList := []interface{}{
			map[string]interface{}{
				"type":  "contains",
				"value": "Paris",
			},
			map[string]interface{}{
				"type":  "not-contains",
				"value": "Tokyo",
			},
		}
		
		assertions, err := assertutil.ParseAll(assertionsList)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if len(assertions) != 2 {
			t.Fatalf("Expected 2 assertions, got %d", len(assertions))
		}
		
		if assertions[0].Type != assertutil.Contains {
			t.Errorf("Expected first assertion type 'contains', got: %s", assertions[0].Type)
		}
		
		if assertions[1].Type != assertutil.NotContains {
			t.Errorf("Expected second assertion type 'not-contains', got: %s", assertions[1].Type)
		}
		
		// Test with invalid assertion
		badAssertionsList := []interface{}{
			map[string]interface{}{
				"type": "contains",
				// Missing "value" field
			},
		}
		
		_, err = assertutil.ParseAll(badAssertionsList)
		if err == nil {
			t.Error("Expected error for invalid assertion, got nil")
		}
		
		// Test with non-map item
		nonMapAssertionsList := []interface{}{
			"not a map",
		}
		
		_, err = assertutil.ParseAll(nonMapAssertionsList)
		if err == nil {
			t.Error("Expected error for non-map assertion, got nil")
		}
	})
}