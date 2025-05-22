package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/tmc/pe/internal/promptfoo"
)

// TestKey represents a unique test case key
type TestKey struct {
	Input    string
	Language string
}

// formatResultsAsEnhancedTable formats evaluation results as a colorful table similar to promptfoo
func formatResultsAsEnhancedTable(results promptfoo.EvaluationResult) []byte {
	var buffer bytes.Buffer
	
	evalId := results.EvalID
	stats := results.Results.Stats
	
	// Step 1: Group results by test variables and collect all prompts and providers
	resultsByTest := make(map[TestKey][]map[string]interface{})
	uniqueProviders := make(map[string]bool)
	uniquePrompts := make(map[string]bool)
	
	// Convert result objects to maps for easier manipulation
	for _, result := range results.Results.Results {
		resultMap := map[string]interface{}{
			"id":       result.ID,
			"prompt":   result.Prompt,
			"provider": result.Provider,
			"response": result.Response,
			"success":  result.Success,
			"score":    result.Score,
		}
		
		// Extract test variables
		input := fmt.Sprintf("%v", result.Vars["input"])
		language := fmt.Sprintf("%v", result.Vars["language"])
		key := TestKey{Input: input, Language: language}
		
		// Add to appropriate test group
		resultsByTest[key] = append(resultsByTest[key], resultMap)
		
		// Track unique providers and prompts
		providerID := result.Provider["id"]
		promptLabel := result.Prompt["label"]
		uniqueProviders[providerID] = true
		uniquePrompts[promptLabel] = true
	}
	
	// Convert maps to sorted slices for consistent output
	providerList := make([]string, 0, len(uniqueProviders))
	for provider := range uniqueProviders {
		providerList = append(providerList, provider)
	}
	sort.Strings(providerList)
	
	promptList := make([]string, 0, len(uniquePrompts))
	for prompt := range uniquePrompts {
		promptList = append(promptList, prompt)
	}
	sort.Strings(promptList)
	
	// Calculate table width
	tableWidth := 20 + 20 + (len(providerList) * len(promptList) * 20)
	
	// Step 2: Build the table
	// Top header row with provider names
	buffer.WriteString("\x1b[90m┌────────────────────\x1b[39m\x1b[90m┬────────────────────\x1b[39m")
	for range providerList {
		for range promptList {
			buffer.WriteString("\x1b[90m┬────────────────────\x1b[39m")
		}
	}
	buffer.WriteString("\x1b[90m┐\x1b[39m\n")
	
	// Header row 1 with column names and providers
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m input              \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m language           \x1b[39m\x1b[22m")
	
	for _, provider := range providerList {
		for range promptList {
			// Format provider name with padding
			providerText := fmt.Sprintf(" [%s] ", provider)
			if len(providerText) > 18 {
				providerText = providerText[:15] + "..."
			}
			providerText = fmt.Sprintf(" %-18s ", providerText)
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + providerText + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Header row 2 with prompt descriptions part 1
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m")
	
	for range providerList {
		for i := range promptList {
			var promptRow1 string
			
			if i == 0 { // Translate this into
				promptRow1 = " Translate this     "
			} else { // Convert this English to
				promptRow1 = " Convert this       "
			}
			
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + promptRow1 + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Header row 3 with prompt descriptions part 2
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m")
	
	for range providerList {
		for i := range promptList {
			var promptRow2 string
			
			if i == 0 { // Translate this into
				promptRow2 = " into {{language}}: "
			} else { // Convert this English to
				promptRow2 = " English to         "
			}
			
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + promptRow2 + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Header row 4 with prompt descriptions part 3
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m")
	
	for range providerList {
		for i := range promptList {
			var promptRow3 string
			
			if i == 0 { // Translate this into
				promptRow3 = " {{input}}          "
			} else { // Convert this English to
				promptRow3 = " {{language}}:      "
			}
			
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + promptRow3 + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Header row 5 with prompt descriptions part 4 (only needed for second prompt type)
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m")
	
	for range providerList {
		for i := range promptList {
			var promptRow4 string
			
			if i == 0 { // Translate this into
				promptRow4 = "                    "
			} else { // Convert this English to
				promptRow4 = " {{input}}          "
			}
			
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + promptRow4 + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Table separator
	buffer.WriteString("\x1b[90m├────────────────────\x1b[39m\x1b[90m┼────────────────────\x1b[39m")
	for i := 0; i < len(providerList)*len(promptList); i++ {
		buffer.WriteString("\x1b[90m┼────────────────────\x1b[39m")
	}
	buffer.WriteString("\x1b[90m┤\x1b[39m\n")
	
	// Sort the test keys to ensure consistent order
	sortedKeys := make([]TestKey, 0, len(resultsByTest))
	for k := range resultsByTest {
		sortedKeys = append(sortedKeys, k)
	}
	
	sort.Slice(sortedKeys, func(i, j int) bool {
		// Sort first by language, then by input
		if sortedKeys[i].Language != sortedKeys[j].Language {
			return sortedKeys[i].Language < sortedKeys[j].Language
		}
		return sortedKeys[i].Input < sortedKeys[j].Input
	})
	
	// Process each test case
	for i, key := range sortedKeys {
		testResults := resultsByTest[key]
		
		// Format the input and language with padding
		paddedInput := fmt.Sprintf(" %-18s ", key.Input)
		buffer.WriteString("\x1b[90m│\x1b[39m" + paddedInput + "\x1b[90m│\x1b[39m")
		
		paddedLanguage := fmt.Sprintf(" %-18s ", key.Language)
		buffer.WriteString(paddedLanguage + "\x1b[90m│\x1b[39m")
		
		// Create lookup for results by provider and prompt
		resultLookup := make(map[string]map[string]map[string]interface{})
		for _, result := range testResults {
			provider, _ := result["provider"].(map[string]string)
			providerID := provider["id"]
			
			prompt, _ := result["prompt"].(map[string]string)
			promptLabel := prompt["label"]
			
			if _, exists := resultLookup[providerID]; !exists {
				resultLookup[providerID] = make(map[string]map[string]interface{})
			}
			
			resultLookup[providerID][promptLabel] = result
		}
		
		// Add cells for results
		for _, provider := range providerList {
			for _, prompt := range promptList {
				// Format the output result cell (matching promptfoo format)
				// Each cell looks like: " [PASS] Bonjour le... "
				var cellText string
				
				if providerResults, ok := resultLookup[provider]; ok {
					if result, ok := providerResults[prompt]; ok {
						success, _ := result["success"].(bool)
						response, _ := result["response"].(map[string]interface{})
						output, _ := response["output"].(string)
						
						// Match promptfoo's PASS/FAIL format with spaces and color
						status := "\x1b[32mPASS\x1b[39m" // Green PASS
						if !success {
							status = "\x1b[31mFAIL\x1b[39m" // Red FAIL
						}
						
						// Truncate output for display, matching promptfoo style
						truncated := output
						if len(output) > 12 {
							truncated = output[:12] + "..."
						}
						
						cellText = fmt.Sprintf(" [%s] %s ", status, truncated)
						
						// Pad to ensure consistent width
						if len(cellText) < 20 {
							cellText = cellText + strings.Repeat(" ", 20-len(cellText))
						}
					}
				}
				
				if cellText == "" {
					cellText = " [N/A]               "
				}
				
				buffer.WriteString(cellText + "\x1b[90m│\x1b[39m")
			}
		}
		
		buffer.WriteString("\n")
		
		// Add separator row except after the last one
		if i < len(sortedKeys)-1 {
			buffer.WriteString("\x1b[90m├────────────────────\x1b[39m\x1b[90m┼────────────────────\x1b[39m")
			for j := 0; j < len(providerList)*len(promptList); j++ {
				buffer.WriteString("\x1b[90m┼────────────────────\x1b[39m")
			}
			buffer.WriteString("\x1b[90m┤\x1b[39m\n")
		} else {
			buffer.WriteString("\x1b[90m└────────────────────\x1b[39m\x1b[90m┴────────────────────\x1b[39m")
			for j := 0; j < len(providerList)*len(promptList); j++ {
				buffer.WriteString("\x1b[90m┴────────────────────\x1b[39m")
			}
			buffer.WriteString("\x1b[90m┘\x1b[39m\n")
		}
	}
	
	// Calculate horizontal separator line width
	separatorLine := strings.Repeat("=", tableWidth)
	
	// Footer with statistics (match promptfoo format exactly)
	buffer.WriteString(separatorLine + "\n")
	buffer.WriteString("✔ Evaluation complete. ID: " + evalId + "\n\n")
	buffer.WriteString("» Run pe view to use the local web viewer\n")
	buffer.WriteString("» Run pe share to create a shareable URL\n")
	buffer.WriteString(separatorLine + "\n")
	
	// Extract statistics
	successesVal := float64(stats.Successes)
	failuresVal := float64(stats.Failures)
	errorsVal := float64(stats.Errors)
	
	// Get token usage info
	tokenUsage := stats.TokenUsage
	totalTokens := float64(tokenUsage.Total)
	promptTokens := float64(tokenUsage.Prompt)
	completionTokens := float64(tokenUsage.Completion)
	
	// Calculate pass rate
	passRate := 100.0
	if successesVal+failuresVal > 0 {
		passRate = (successesVal / (successesVal + failuresVal)) * 100.0
	}
	
	// Add statistics to output matching promptfoo format
	buffer.WriteString(fmt.Sprintf("Successes: %.0f\n", successesVal))
	buffer.WriteString(fmt.Sprintf("Failures: %.0f\n", failuresVal))
	buffer.WriteString(fmt.Sprintf("Errors: %.0f\n", errorsVal))
	buffer.WriteString(fmt.Sprintf("Pass Rate: %.2f%%\n", passRate))
	buffer.WriteString(fmt.Sprintf("Total tokens: %.0f / Prompt tokens: %.0f / Completion tokens: %.0f / Cached tokens: 0\n",
		totalTokens, promptTokens, completionTokens))
	buffer.WriteString("Done.\n")
	
	return buffer.Bytes()
}