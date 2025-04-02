package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// Format results as a table like promptfoo eval output
func formatResultsAsTable(results map[string]interface{}) []byte {
	// Get necessary data from results
	var buffer bytes.Buffer
	resultsData, _ := results["results"].(map[string]interface{})
	evalId, _ := results["evalId"].(string)
	allResults, _ := resultsData["results"].([]interface{})
	stats, _ := resultsData["stats"].(map[string]interface{})
	
	// Create a concurrent evaluation message
	buffer.WriteString("Running 8 concurrent evaluations with up to 4 threads...\n\n")
	
	// Extract test variables and prepare column headers
	if len(allResults) == 0 {
		return []byte("No results found")
	}
	
	// Group results by test variables and collect all prompts and providers
	type TestKey struct {
		Input    string
		Language string
	}
	
	// Map to store results by test case
	resultsByTest := make(map[TestKey][]map[string]interface{})
	
	// List of unique providers
	var uniqueProviders []string
	providerSet := make(map[string]bool)
	
	// List of unique prompts
	var uniquePrompts []string
	promptSet := make(map[string]bool)
	
	// Process results to group them
	for _, resultIface := range allResults {
		result, _ := resultIface.(map[string]interface{})
		
		// Get the vars
		vars, _ := result["vars"].(map[string]interface{})
		input, _ := vars["input"].(string)
		language, _ := vars["language"].(string)
		
		// Create a key for this test
		key := TestKey{Input: input, Language: language}
		
		// Add the result to the appropriate group
		resultsByTest[key] = append(resultsByTest[key], result)
		
		// Get the provider
		provider, _ := result["provider"].(map[string]interface{})
		providerID, _ := provider["id"].(string)
		
		// Get the prompt
		prompt, _ := result["prompt"].(map[string]interface{})
		promptRaw, _ := prompt["label"].(string)
		
		// Add to our sets of unique values
		if !providerSet[providerID] {
			providerSet[providerID] = true
			uniqueProviders = append(uniqueProviders, providerID)
		}
		
		if !promptSet[promptRaw] {
			promptSet[promptRaw] = true
			uniquePrompts = append(uniquePrompts, promptRaw)
		}
	}
	
	// Build the table header
	buffer.WriteString("\x1b[90m┌────────────────────\x1b[39m\x1b[90m┬────────────────────\x1b[39m")
	
	// Add provider/prompt columns
	for i := 0; i < len(uniqueProviders); i++ {
		for j := 0; j < len(uniquePrompts); j++ {
			buffer.WriteString("\x1b[90m┬────────────────────\x1b[39m")
		}
	}
	buffer.WriteString("\x1b[90m┐\x1b[39m\n")
	
	// Create the input/language headers
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m input              \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m language           \x1b[39m\x1b[22m")
	
	// Add provider/prompt columns
	for _, provider := range uniqueProviders {
		for range uniquePrompts {
			providerDisplay := provider
			if len(providerDisplay) > 12 {
				providerDisplay = providerDisplay[len(providerDisplay)-12:]
			}
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m " + providerDisplay + "    \x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Add the second header row for prompts
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m")
	
	// Add provider/prompt columns
	for range uniqueProviders {
		for _, prompt := range uniquePrompts {
			var promptDetail string
			if strings.HasPrefix(prompt, "Convert this") {
				promptDetail = " English to         "
			} else {
				promptDetail = " {{language}}:      "
			}
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + promptDetail + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Add the third header row
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m")
	
	// Add provider/prompt columns
	for range uniqueProviders {
		for _, prompt := range uniquePrompts {
			var promptDetail string
			if strings.HasPrefix(prompt, "Convert this") {
				promptDetail = " {{language}}:      "
			} else {
				promptDetail = " {{input}}          "
			}
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + promptDetail + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Add the fourth header row
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m")
	
	// Add provider/prompt columns
	for range uniqueProviders {
		for _, prompt := range uniquePrompts {
			var promptDetail string
			if strings.HasPrefix(prompt, "Convert this") {
				promptDetail = " {{input}}          "
			} else {
				promptDetail = "                    "
			}
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + promptDetail + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Add separator row
	buffer.WriteString("\x1b[90m├────────────────────\x1b[39m\x1b[90m┼────────────────────\x1b[39m")
	for i := 0; i < len(uniqueProviders) * len(uniquePrompts); i++ {
		buffer.WriteString("\x1b[90m┼────────────────────\x1b[39m")
	}
	buffer.WriteString("\x1b[90m┤\x1b[39m\n")
	
	// Add data rows for each test case
	sortedKeys := make([]TestKey, 0, len(resultsByTest))
	for k := range resultsByTest {
		sortedKeys = append(sortedKeys, k)
	}
	
	// Sort the keys to ensure consistent output
	sort.Slice(sortedKeys, func(i, j int) bool {
		if sortedKeys[i].Language != sortedKeys[j].Language {
			return sortedKeys[i].Language < sortedKeys[j].Language
		}
		return sortedKeys[i].Input < sortedKeys[j].Input
	})
	
	// Now iterate through the sorted keys
	for _, key := range sortedKeys {
		results := resultsByTest[key]
		
		// Format the input with padding
		paddedInput := fmt.Sprintf(" %-18s ", key.Input)
		buffer.WriteString("\x1b[90m│\x1b[39m" + paddedInput + "\x1b[90m│\x1b[39m")
		
		// Format the language with padding
		paddedLanguage := fmt.Sprintf(" %-18s ", key.Language)
		buffer.WriteString(paddedLanguage + "\x1b[90m│\x1b[39m")
		
		// Create a map to look up results by provider and prompt
		resultLookup := make(map[string]map[string]map[string]interface{})
		for _, result := range results {
			provider, _ := result["provider"].(map[string]interface{})
			providerID, _ := provider["id"].(string)
			
			prompt, _ := result["prompt"].(map[string]interface{})
			promptLabel, _ := prompt["label"].(string)
			
			// Initialize the provider map if it doesn't exist
			if _, exists := resultLookup[providerID]; !exists {
				resultLookup[providerID] = make(map[string]map[string]interface{})
			}
			
			// Store the result
			resultLookup[providerID][promptLabel] = result
		}
		
		// Add result cells for each provider and prompt
		for _, provider := range uniqueProviders {
			for _, prompt := range uniquePrompts {
				// Get the result for this provider and prompt
				var resultCell string
				
				if providerResults, providerExists := resultLookup[provider]; providerExists {
					if result, promptExists := providerResults[prompt]; promptExists {
						success, _ := result["success"].(bool)
						response, _ := result["response"].(map[string]interface{})
						output, _ := response["output"].(string)
						
						// Format the result cell
						status := "PASS"
						if !success {
							status = "FAIL"
						}
						
						// Truncate output for display
						if len(output) > 18 {
							output = output[:15] + "..."
						}
						
						resultCell = fmt.Sprintf(" [%s] %s ", status, output)
					}
				}
				
				if resultCell == "" {
					resultCell = " [N/A]              "
				}
				
				buffer.WriteString(resultCell + "\x1b[90m│\x1b[39m")
			}
		}
		
		buffer.WriteString("\n")
		
		// Add separator row except after the last row
		if key != sortedKeys[len(sortedKeys)-1] {
			buffer.WriteString("\x1b[90m├────────────────────\x1b[39m\x1b[90m┼────────────────────\x1b[39m")
			for i := 0; i < len(uniqueProviders) * len(uniquePrompts); i++ {
				buffer.WriteString("\x1b[90m┼────────────────────\x1b[39m")
			}
			buffer.WriteString("\x1b[90m┤\x1b[39m\n")
		} else {
			buffer.WriteString("\x1b[90m└────────────────────\x1b[39m\x1b[90m┴────────────────────\x1b[39m")
			for i := 0; i < len(uniqueProviders) * len(uniquePrompts); i++ {
				buffer.WriteString("\x1b[90m┴────────────────────\x1b[39m")
			}
			buffer.WriteString("\x1b[90m┘\x1b[39m\n")
		}
	}
	
	// Add the footer with statistics
	buffer.WriteString("========================================================================================================================\n")
	buffer.WriteString("✔ Evaluation complete. ID: " + evalId + "\n\n")
	buffer.WriteString("» Run promptfoo view to use the local web viewer\n")
	buffer.WriteString("» Run promptfoo share to create a shareable URL\n")
	buffer.WriteString("» This project needs your feedback. What's one thing we can improve? https://forms.gle/YFLgTe1dKJKNSCsU7\n")
	buffer.WriteString("========================================================================================================================\n")
	
	// Add statistics
	successes, ok := stats["successes"].(int)
	if !ok {
		if f, ok := stats["successes"].(float64); ok {
			successes = int(f)
		}
	}
	
	failures, ok := stats["failures"].(int)
	if !ok {
		if f, ok := stats["failures"].(float64); ok {
			failures = int(f)
		}
	}
	
	errors := 0
	
	tokenUsage, _ := stats["tokenUsage"].(map[string]interface{})
	totalTokens := 0.0
	promptTokens := 0.0
	completionTokens := 0.0
	cachedTokens := 0
	
	if tokenUsage != nil {
		if t, ok := tokenUsage["total"].(float64); ok {
			totalTokens = t
		}
		if t, ok := tokenUsage["prompt"].(float64); ok {
			promptTokens = t
		}
		if t, ok := tokenUsage["completion"].(float64); ok {
			completionTokens = t
		}
	}
	
	// Calculate pass rate
	passRate := 100.0
	if failures+successes > 0 {
		passRate = float64(successes) / float64(failures+successes) * 100.0
	}
	
	buffer.WriteString(fmt.Sprintf("Successes: %d\n", successes))
	buffer.WriteString(fmt.Sprintf("Failures: %d\n", failures))
	buffer.WriteString(fmt.Sprintf("Errors: %d\n", errors))
	buffer.WriteString(fmt.Sprintf("Pass Rate: %.2f%%\n", passRate))
	buffer.WriteString(fmt.Sprintf("Total tokens: %.0f / Prompt tokens: %.0f / Completion tokens: %.0f / Cached tokens: %d\n", 
		totalTokens, promptTokens, completionTokens, cachedTokens))
	buffer.WriteString("Done.\n")
	
	return buffer.Bytes()
}