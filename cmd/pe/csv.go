package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
)

// formatResultsAsCSV formats evaluation results as a CSV file
func formatResultsAsCSV(results map[string]interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	
	// Write the header
	header := []string{
		"Test ID", 
		"Provider", 
		"Prompt",
		"Status",
		"Output",
		"Variables",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("error writing header: %v", err)
	}
	
	// Access the results data
	resultsData, ok := results["results"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid results format: results.results is not a map")
	}
	
	// Get the detailed results (printing the type for debugging)
	detailedResults, ok := resultsData["results"].([]interface{})
	if !ok {
		// Try to dump what we have for debugging
		fmt.Printf("Results data: %T\n", resultsData["results"])
		
		// Try as map[string]interface{} then get each result
		if resultsArr, ok := resultsData["results"].([]map[string]interface{}); ok {
			// Convert to []interface{}
			for _, r := range resultsArr {
				detailedResults = append(detailedResults, r)
			}
		} else {
			return nil, fmt.Errorf("invalid results format: expected array of results")
		}
	}
	
	// Process each result
	for _, r := range detailedResults {
		result, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		
		// Extract the data we need
		id, _ := result["id"].(string)
		
		provider := ""
		if providerData, ok := result["provider"].(map[string]interface{}); ok {
			provider, _ = providerData["id"].(string)
		}
		
		promptText := ""
		if promptData, ok := result["prompt"].(map[string]interface{}); ok {
			promptText, _ = promptData["raw"].(string)
		}
		
		success, _ := result["success"].(bool)
		status := "PASS"
		if !success {
			status = "FAIL"
		}
		
		output := ""
		if responseData, ok := result["response"].(map[string]interface{}); ok {
			output, _ = responseData["output"].(string)
		}
		
		// Format variables
		varsStr := ""
		if vars, ok := result["vars"].(map[string]interface{}); ok {
			for k, v := range vars {
				varsStr += fmt.Sprintf("%s=%v; ", k, v)
			}
		}
		
		// Write the row
		row := []string{
			id,
			provider,
			promptText,
			status,
			output,
			varsStr,
		}
		
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("error writing row: %v", err)
		}
	}
	
	writer.Flush()
	
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("error flushing writer: %v", err)
	}
	
	return buffer.Bytes(), nil
}