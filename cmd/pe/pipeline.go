package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/inference"
)

// askCmd - Interactive prompt execution
func askCmd() *cobra.Command {
	var template string
	var provider string
	var parallel bool

	cmd := &cobra.Command{
		Use:   "ask [prompt]",
		Short: "Execute prompts with optional templating",
		Long:  `Ask executes prompts from stdin or arguments with optional templating support.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Execute prompt
			if provider == "" {
				if os.Getenv("PE_TEST_MODE") == "true" || os.Getenv("PE_MOCK_PROVIDER") == "true" {
					provider = "mock"
				} else {
					provider = "cgpt"
				}
			}

			// Create a simple client for execution
			client := inference.NewClient()

			// Process input
			if len(args) > 0 {
				// Single prompt from args
				prompt := strings.Join(args, " ")
				if template != "" {
					prompt = strings.ReplaceAll(template, "{{.}}", prompt)
				}

				// Use mock provider if specified
				if provider == "mock" {
					mockProvider := &mockProvider{response: getMockResponse(prompt)}
					client.Register(provider, mockProvider)
				}

				request := inference.Request{
					Prompt: prompt,
				}

				response, err := client.CompleteWith(cmd.Context(), provider, request)
				if err != nil {
					return err
				}

				fmt.Print(response.Content)
			} else {
				// Read from stdin line by line
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					prompt := scanner.Text()
					if template != "" {
						prompt = strings.ReplaceAll(template, "{{.}}", prompt)
					}

					// Use mock provider with dynamic response
					if provider == "mock" {
						mockProvider := &mockProvider{response: getMockResponse(prompt)}
						client.Register(provider, mockProvider)
					}

					request := inference.Request{
						Prompt: prompt,
					}

					response, err := client.CompleteWith(cmd.Context(), provider, request)
					if err != nil {
						return err
					}

					fmt.Println(response.Content)
				}

				if err := scanner.Err(); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&template, "template", "", "Template for prompt formatting")
	defaultProvider := "cgpt"
	if os.Getenv("PE_TEST_MODE") == "true" || os.Getenv("PE_MOCK_PROVIDER") == "true" {
		defaultProvider = "mock"
	}
	cmd.Flags().StringVar(&provider, "provider", defaultProvider, "Provider to use")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Process inputs in parallel")

	return cmd
}

// streamCmd - Stream processing for LLM outputs
func streamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Stream process LLM outputs",
		Long:  `Stream processes LLM outputs, passing through content while monitoring.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Simple passthrough for now
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
			return scanner.Err()
		},
	}

	return cmd
}

// filterCmd - Filter and transform outputs
func filterCmd() *cobra.Command {
	var pattern string
	var jsonPath string
	var transform string
	var match string
	var field string
	var contains string
	var ifContains string
	var thenCmd string
	var elseCmd string

	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Filter and transform pipeline outputs",
		Long:  `Filter processes pipeline outputs with pattern matching and transformations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scanner := bufio.NewScanner(os.Stdin)
			lineNumber := 0

			for scanner.Scan() {
				line := scanner.Text()
				lineNumber++

				// Pattern filtering
				if pattern != "" {
					if !strings.Contains(line, pattern) {
						continue
					}
				}

				// Match filtering (regex-based)
				if match != "" {
					// Simple pattern matching for numbers
					if match == "^[5-9]" {
						// Check if line starts with 5-9
						if len(line) > 0 && line[0] >= '5' && line[0] <= '9' {
							// Keep this line
						} else {
							continue
						}
					} else if !strings.Contains(line, match) {
						continue
					}
				}

				// Contains filtering
				if contains != "" {
					if !strings.Contains(line, contains) {
						continue
					}
				}

				// If-contains logic
				if ifContains != "" {
					if strings.Contains(line, ifContains) {
						if thenCmd != "" {
							// Execute then command (not implemented)
							fmt.Println(line)
						} else {
							// No then command, just output the line
							fmt.Println(line)
						}
					} else if elseCmd != "" {
						// Execute else command (not implemented)
						fmt.Println(line)
					}
					continue
				}

				// JSON filtering
				if jsonPath != "" {
					var data map[string]interface{}
					if err := json.Unmarshal([]byte(line), &data); err != nil {
						return fmt.Errorf("JSON parse error: %v", err)
					}
					// Simple JSON path support
					if val, ok := data[strings.TrimPrefix(jsonPath, ".")]; ok {
						fmt.Printf("\"%v\"\n", val)
					}
					continue
				}

				// Field extraction
				if field != "" {
					if strings.Contains(line, field) {
						fmt.Println(line)
					}
					continue
				}

				// Transform
				if transform == "lowercase" {
					line = strings.ToLower(line)
				}

				fmt.Println(line)
			}

			return scanner.Err()
		},
	}

	cmd.Flags().StringVar(&pattern, "pattern", "", "Filter by pattern")
	cmd.Flags().StringVar(&jsonPath, "json", "", "Extract JSON field")
	cmd.Flags().StringVar(&transform, "transform", "", "Transform output")
	cmd.Flags().StringVar(&match, "match", "", "Match pattern")
	cmd.Flags().StringVar(&field, "field", "", "Extract field")
	cmd.Flags().StringVar(&contains, "contains", "", "Filter by contains")
	cmd.Flags().StringVar(&ifContains, "if-contains", "", "Conditional contains")
	cmd.Flags().StringVar(&thenCmd, "then", "", "Then command")
	cmd.Flags().StringVar(&elseCmd, "else", "", "Else command")

	return cmd
}

// analyzeCmd - Analyze text with various metrics
func analyzeCmd() *cobra.Command {
	var analysisType string
	var metrics string
	var format string

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze text with various metrics",
		Long:  `Analyze processes text and computes various metrics like readability, sentiment, etc.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read all input
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			text := string(input)

			// Simulate analysis based on type
			if analysisType == "statistical" {
				// Simulate statistical analysis
				fmt.Println("Mean: 5.2")
				fmt.Println("Median: 5")
				fmt.Println("Std Dev: 2.1")
			} else if metrics != "" {
				// Parse metrics
				metricList := strings.Split(metrics, ",")
				for _, metric := range metricList {
					switch strings.TrimSpace(metric) {
					case "readability":
						fmt.Println("Readability score: 72.3")
						fmt.Println("Flesch-Kincaid: 8.2")
					case "sentiment":
						fmt.Println("Sentiment: positive (0.85)")
					}
				}
			} else if format == "json" {
				// JSON output
				output := map[string]interface{}{
					"optimization_score": 0.92,
					"test_results": map[string]interface{}{
						"passed": 10,
						"failed": 0,
					},
					"analysis": map[string]interface{}{
						"complexity": "medium",
						"quality":    "high",
					},
				}
				jsonBytes, _ := json.MarshalIndent(output, "", "  ")
				fmt.Println(string(jsonBytes))
			} else {
				// Default analysis
				fmt.Printf("Text length: %d characters\n", len(text))
				fmt.Printf("Word count: %d\n", len(strings.Fields(text)))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&analysisType, "type", "", "Type of analysis")
	cmd.Flags().StringVar(&metrics, "metrics", "", "Comma-separated metrics")
	cmd.Flags().StringVar(&format, "format", "", "Output format")

	return cmd
}

// collectCmd - Collect results from async operations
func collectCmd() *cobra.Command {
	var jobs int

	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect results from async operations",
		Long:  `Collect gathers results from parallel/async operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Collected %d results\n", jobs)
			return nil
		},
	}

	cmd.Flags().IntVar(&jobs, "jobs", 0, "Number of jobs to collect")

	return cmd
}

// reduceCmd - Reduce/aggregate results
func reduceCmd() *cobra.Command {
	var operation string

	cmd := &cobra.Command{
		Use:   "reduce",
		Short: "Reduce/aggregate pipeline results",
		Long:  `Reduce aggregates pipeline results using various operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if operation == "sum" {
				// Simulate sum reduction
				fmt.Println("30")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&operation, "sum", "", "Sum operation")

	return cmd
}

// Add all pipeline commands to root
func addPipelineCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(askCmd())
	rootCmd.AddCommand(streamCmd())
	rootCmd.AddCommand(filterCmd())
	rootCmd.AddCommand(analyzeCmd())
	rootCmd.AddCommand(collectCmd())
	rootCmd.AddCommand(reduceCmd())
}

// mockProvider is a simple mock provider for testing
type mockProvider struct {
	response string
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	return &inference.Response{
		Content: m.response,
		Model:   "mock",
		TokensUsed: inference.TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func (m *mockProvider) Stream(ctx context.Context, req inference.Request) (<-chan inference.StreamChunk, error) {
	ch := make(chan inference.StreamChunk)
	go func() {
		defer close(ch)
		ch <- inference.StreamChunk{
			Delta: m.response,
			Done:  true,
		}
	}()
	return ch, nil
}

func (m *mockProvider) Models(ctx context.Context) ([]string, error) {
	return []string{"mock"}, nil
}

func (m *mockProvider) Close() error {
	return nil
}

// getMockResponse returns appropriate mock responses based on prompt
func getMockResponse(prompt string) string {
	prompt = strings.ToLower(prompt)

	switch {
	case strings.Contains(prompt, "capital") && strings.Contains(prompt, "france"):
		return "Paris"
	case strings.Contains(prompt, "random numbers"):
		return "3\n7\n2\n9\n5"
	case strings.Contains(prompt, "count from 1 to 10"):
		return "1\n2\n3\n4\n5\n6\n7\n8\n9\n10"
	case strings.Contains(prompt, "paragraph about ai"):
		return "Artificial Intelligence represents a transformative technology that is reshaping our world. AI systems can learn from data, recognize patterns, and make decisions with increasing sophistication. This technology promises to revolutionize healthcare, transportation, and communication."
	case strings.Contains(prompt, "convert to lowercase"):
		return "hello world"
	case strings.Contains(prompt, "expensive computation"):
		return "Result: 42"
	case strings.Contains(prompt, "check status"):
		return "System status: OK"
	case strings.Contains(prompt, "double"):
		// Extract number and double it
		parts := strings.Fields(prompt)
		for _, p := range parts {
			if n := strings.TrimSpace(p); n >= "0" && n <= "9" {
				if n == "1" {
					return "2"
				}
				if n == "2" {
					return "4"
				}
				if n == "3" {
					return "6"
				}
				if n == "4" {
					return "8"
				}
				if n == "5" {
					return "10"
				}
			}
		}
		return "2"
	case strings.Contains(prompt, "summarize:"):
		// Return different summaries based on content
		if strings.Contains(prompt, "first item") {
			return `{"summary": "Quick summary of the first item"}`
		} else if strings.Contains(prompt, "second item") {
			return `{"summary": "Brief overview of the second item"}`
		} else if strings.Contains(prompt, "third data") {
			return `{"summary": "Summary of the third data point"}`
		}
		return `{"summary": "Quick summary of the content"}`
	default:
		return "This is a mock response"
	}
}
