package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pev1 "github.com/tmc/pe/proto/pe/v1"
)

// createCmd creates a new prompt or prompt revision.
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new prompt or prompt revision",
	Long:  `Create a new prompt or prompt revision with the specified parameters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		promptName, _ := cmd.Flags().GetString("name")
		systemPrompt, _ := cmd.Flags().GetString("system-prompt")
		modelName, _ := cmd.Flags().GetString("model")
		maxTokens, _ := cmd.Flags().GetInt32("max-tokens")
		temperature, _ := cmd.Flags().GetFloat32("temperature")
		outputPath, _ := cmd.Flags().GetString("output")

		// Create a new prompt
		prompt := &pev1.Prompt{
			Id:        generateID(),
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Name:      promptName,
			Revisions: []*pev1.PromptRevision{},
			Metadata:  map[string]string{},
		}

		// Create the initial revision
		revision := &pev1.PromptRevision{
			Id:                 generateID(),
			CreatedAt:          timestamppb.Now(),
			SystemPrompt:       systemPrompt,
			ModelName:          modelName,
			MaxTokensToSample:  maxTokens,
			Temperature:        float32(temperature),
			Tools:              []*structpb.Struct{},
			Messages:           []*pev1.Message{},
			Examples:           []*pev1.Example{},
			PromptId:           prompt.Id,
			Variables:          []*pev1.Variable{},
			TestCases:          []*pev1.TestCase{},
			Metadata:           map[string]string{},
		}

		// Add the revision to the prompt
		prompt.Revisions = append(prompt.Revisions, revision)
		prompt.LatestRevision = revision

		// Write to file
		if outputPath == "" {
			outputPath = promptName + ".pb"
		}
		return writePromptToFile(prompt, outputPath)
	},
}

// showCmd displays a prompt or prompt revision.
var showCmd = &cobra.Command{
	Use:   "show [file]",
	Short: "Show prompt details",
	Long:  `Display details of a prompt or prompt revision in various formats.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		format, _ := cmd.Flags().GetString("format")
		revisionID, _ := cmd.Flags().GetString("revision")
		
		// Read the prompt from file
		prompt, err := readPromptFromFile(filePath)
		if err != nil {
			return err
		}

		// If revision ID is specified, show only that revision
		if revisionID != "" {
			for _, rev := range prompt.Revisions {
				if rev.Id == revisionID {
					return displayRevision(rev, format, cmd.OutOrStdout())
				}
			}
			return fmt.Errorf("revision with ID %s not found", revisionID)
		}

		// Otherwise show the whole prompt
		return displayPrompt(prompt, format, cmd.OutOrStdout())
	},
}

// analyzeCmd analyzes a prompt or prompt revision.
var analyzeCmd = &cobra.Command{
	Use:   "analyze [file]",
	Short: "Analyze a prompt",
	Long:  `Analyze a prompt to identify potential issues, calculate statistics, etc.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		revisionID, _ := cmd.Flags().GetString("revision")
		
		// Read the prompt from file
		prompt, err := readPromptFromFile(filePath)
		if err != nil {
			return err
		}

		// Get the revision to analyze
		var revision *pev1.PromptRevision
		if revisionID == "" {
			revision = prompt.LatestRevision
		} else {
			for _, rev := range prompt.Revisions {
				if rev.Id == revisionID {
					revision = rev
					break
				}
			}
			if revision == nil {
				return fmt.Errorf("revision with ID %s not found", revisionID)
			}
		}

		// Perform analysis
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		
		fmt.Fprintf(w, "Analysis for Prompt: %s (Revision: %s)\n\n", prompt.Name, revision.Id)
		
		// System prompt analysis
		wordCount := len(strings.Fields(revision.SystemPrompt))
		fmt.Fprintf(w, "System Prompt:\n")
		fmt.Fprintf(w, "  Word Count:\t%d\n", wordCount)
		fmt.Fprintf(w, "  Character Count:\t%d\n", len(revision.SystemPrompt))
		
		// Variables analysis
		fmt.Fprintf(w, "\nVariables:\t%d\n", len(revision.Variables))
		for _, v := range revision.Variables {
			fmt.Fprintf(w, "  %s:\t%s\n", v.Name, v.Description)
		}
		
		// Test cases analysis
		fmt.Fprintf(w, "\nTest Cases:\t%d\n", len(revision.TestCases))
		passed := 0
		for _, tc := range revision.TestCases {
			if tc.IsSuccess {
				passed++
			}
		}
		fmt.Fprintf(w, "  Passed:\t%d\n", passed)
		fmt.Fprintf(w, "  Failed:\t%d\n", len(revision.TestCases)-passed)
		
		// Model settings
		fmt.Fprintf(w, "\nModel Settings:\n")
		fmt.Fprintf(w, "  Model:\t%s\n", revision.ModelName)
		fmt.Fprintf(w, "  Max Tokens:\t%d\n", revision.MaxTokensToSample)
		fmt.Fprintf(w, "  Temperature:\t%.2f\n", revision.Temperature)
		
		return w.Flush()
	},
}

// validateCmd validates a prompt or prompt revision.
var validateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate a prompt",
	Long:  `Validate a prompt or prompt revision against schema and semantic rules.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		
		// Read the prompt from file
		prompt, err := readPromptFromFile(filePath)
		if err != nil {
			return err
		}

		// Perform validation
		issues := validatePrompt(prompt)
		
		if len(issues) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Validation passed: No issues found.\n")
			return nil
		}
		
		fmt.Fprintf(cmd.OutOrStdout(), "Validation found %d issues:\n\n", len(issues))
		for i, issue := range issues {
			fmt.Fprintf(cmd.OutOrStdout(), "%d. %s\n", i+1, issue)
		}
		
		return fmt.Errorf("validation failed with %d issues", len(issues))
	},
}

// convertCmd converts a prompt between different formats.
var convertCmd = &cobra.Command{
	Use:   "convert [input-file] [output-file]",
	Short: "Convert prompt between formats",
	Long:  `Convert a prompt between different formats (protobuf, JSON, YAML, etc.).`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		outputPath := args[1]
		
		inputFormat := inferFormat(inputPath)
		outputFormat := inferFormat(outputPath)
		
		if inputFormat == outputFormat {
			return fmt.Errorf("input and output formats are the same: %s", inputFormat)
		}
		
		// Read the prompt from the input file
		prompt, err := readPromptFromFile(inputPath)
		if err != nil {
			return err
		}
		
		// Write the prompt to the output file
		return writePromptToFile(prompt, outputPath)
	},
}

// evalCmd evaluates a prompt against test cases.
var evalCmd = &cobra.Command{
	Use:   "eval [file]",
	Short: "Evaluate a prompt",
	Long:  `Evaluate a prompt against its test cases using an LLM provider.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		provider, _ := cmd.Flags().GetString("provider")
		
		// Read the prompt from file
		prompt, err := readPromptFromFile(filePath)
		if err != nil {
			return err
		}
		
		// Get the revision to evaluate
		revision := prompt.LatestRevision
		if revision == nil && len(prompt.Revisions) > 0 {
			revision = prompt.Revisions[len(prompt.Revisions)-1]
		}
		
		if revision == nil {
			return fmt.Errorf("no revision found to evaluate")
		}
		
		// Mock evaluation results for now
		fmt.Fprintf(cmd.OutOrStdout(), "Evaluating prompt '%s' with provider '%s'...\n\n", prompt.Name, provider)
		
		for i, testCase := range revision.TestCases {
			// In a real implementation, we would call the LLM provider here
			time.Sleep(500 * time.Millisecond)
			fmt.Fprintf(cmd.OutOrStdout(), "Test case %d: %s\n", i+1, testCase.Description)
			fmt.Fprintf(cmd.OutOrStdout(), "  Variable values: %v\n", testCase.VariableValues)
			fmt.Fprintf(cmd.OutOrStdout(), "  Expected output: %s\n", testCase.ExpectedOutput)
			fmt.Fprintf(cmd.OutOrStdout(), "  Completion: %s\n", testCase.CompletionText)
			
			// Mock success/failure
			success := i%2 == 0
			fmt.Fprintf(cmd.OutOrStdout(), "  Result: %s\n\n", map[bool]string{true: "PASS", false: "FAIL"}[success])
			
			// Update test case with result
			testCase.IsSuccess = success
			testCase.CompletedAt = timestamppb.Now()
		}
		
		// Save the updated prompt
		return writePromptToFile(prompt, filePath)
	},
}

// formatCmd formats a prompt file.
var formatCmd = &cobra.Command{
	Use:   "format [file]",
	Short: "Format a prompt file",
	Long:  `Format a prompt file according to style guidelines.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		
		// Read the prompt from file
		prompt, err := readPromptFromFile(filePath)
		if err != nil {
			return err
		}
		
		// No actual formatting for now, just write it back
		return writePromptToFile(prompt, filePath)
	},
}

// mergeCmd merges multiple prompts or prompt revisions.
var mergeCmd = &cobra.Command{
	Use:   "merge [file1] [file2] [output-file]",
	Short: "Merge prompts",
	Long:  `Merge multiple prompts or prompt revisions into a single prompt.`,
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath := args[len(args)-1]
		inputPaths := args[:len(args)-1]
		
		var prompts []*pev1.Prompt
		for _, path := range inputPaths {
			prompt, err := readPromptFromFile(path)
			if err != nil {
				return err
			}
			prompts = append(prompts, prompt)
		}
		
		// Create a new prompt with all revisions from the input prompts
		result := &pev1.Prompt{
			Id:        generateID(),
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Name:      "Merged Prompt",
			Revisions: []*pev1.PromptRevision{},
			Metadata:  map[string]string{"merged_from": strings.Join(inputPaths, ",")},
		}
		
		for _, prompt := range prompts {
			for _, rev := range prompt.Revisions {
				// Create a copy of the revision with a new ID
				newRev := proto.Clone(rev).(*pev1.PromptRevision)
				newRev.Id = generateID()
				newRev.PromptId = result.Id
				
				result.Revisions = append(result.Revisions, newRev)
			}
		}
		
		// Set the latest revision
		if len(result.Revisions) > 0 {
			result.LatestRevision = result.Revisions[len(result.Revisions)-1]
		}
		
		return writePromptToFile(result, outputPath)
	},
}

// diffCmd shows differences between prompts or prompt revisions.
var diffCmd = &cobra.Command{
	Use:   "diff [file1] [file2]",
	Short: "Show differences",
	Long:  `Show differences between two prompts or prompt revisions.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		file1 := args[0]
		file2 := args[1]
		
		prompt1, err := readPromptFromFile(file1)
		if err != nil {
			return err
		}
		
		prompt2, err := readPromptFromFile(file2)
		if err != nil {
			return err
		}
		
		// Simple comparison for now
		fmt.Fprintf(cmd.OutOrStdout(), "Comparing %s and %s:\n\n", file1, file2)
		
		fmt.Fprintf(cmd.OutOrStdout(), "Prompt Name:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", file1, prompt1.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n\n", file2, prompt2.Name)
		
		fmt.Fprintf(cmd.OutOrStdout(), "Number of Revisions:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  %s: %d\n", file1, len(prompt1.Revisions))
		fmt.Fprintf(cmd.OutOrStdout(), "  %s: %d\n\n", file2, len(prompt2.Revisions))
		
		rev1 := prompt1.LatestRevision
		rev2 := prompt2.LatestRevision
		
		if rev1 != nil && rev2 != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Latest Revision:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  System Prompt Diff:\n")
			
			if rev1.SystemPrompt == rev2.SystemPrompt {
				fmt.Fprintf(cmd.OutOrStdout(), "    (identical)\n")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "    %s: %d chars\n", file1, len(rev1.SystemPrompt))
				fmt.Fprintf(cmd.OutOrStdout(), "    %s: %d chars\n", file2, len(rev2.SystemPrompt))
			}
			
			fmt.Fprintf(cmd.OutOrStdout(), "\n  Model:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "    %s: %s\n", file1, rev1.ModelName)
			fmt.Fprintf(cmd.OutOrStdout(), "    %s: %s\n", file2, rev2.ModelName)
			
			fmt.Fprintf(cmd.OutOrStdout(), "\n  Temperature:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "    %s: %.2f\n", file1, rev1.Temperature)
			fmt.Fprintf(cmd.OutOrStdout(), "    %s: %.2f\n", file2, rev2.Temperature)
		}
		
		return nil
	},
}

// exportCmd exports a prompt to a different format.
var exportCmd = &cobra.Command{
	Use:   "export [file] [output-file]",
	Short: "Export a prompt",
	Long:  `Export a prompt to a different format or platform.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		outputPath := args[1]
		format, _ := cmd.Flags().GetString("format")
		
		if format == "" {
			format = inferFormat(outputPath)
		}
		
		prompt, err := readPromptFromFile(inputPath)
		if err != nil {
			return err
		}
		
		switch format {
		case "json":
			return exportPromptToJSON(prompt, outputPath)
		case "yaml":
			return exportPromptToYAML(prompt, outputPath)
		case "proto":
			return writePromptToFile(prompt, outputPath)
		case "text":
			return exportPromptToText(prompt, outputPath)
		default:
			return fmt.Errorf("unsupported export format: %s", format)
		}
	},
}

// importCmd imports a prompt from a different format.
var importCmd = &cobra.Command{
	Use:   "import [file] [output-file]",
	Short: "Import a prompt",
	Long:  `Import a prompt from a different format or platform.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		outputPath := args[1]
		format, _ := cmd.Flags().GetString("format")
		
		if format == "" {
			format = inferFormat(inputPath)
		}
		
		var prompt *pev1.Prompt
		var err error
		
		switch format {
		case "json":
			prompt, err = importPromptFromJSON(inputPath)
		case "yaml":
			prompt, err = importPromptFromYAML(inputPath)
		case "proto":
			prompt, err = readPromptFromFile(inputPath)
		default:
			return fmt.Errorf("unsupported import format: %s", format)
		}
		
		if err != nil {
			return err
		}
		
		return writePromptToFile(prompt, outputPath)
	},
}

// testCmd runs test cases for a prompt.
var testCmd = &cobra.Command{
	Use:   "test [file]",
	Short: "Test a prompt",
	Long:  `Run test cases for a prompt and report results.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		
		prompt, err := readPromptFromFile(filePath)
		if err != nil {
			return err
		}
		
		revision := prompt.LatestRevision
		if revision == nil && len(prompt.Revisions) > 0 {
			revision = prompt.Revisions[len(prompt.Revisions)-1]
		}
		
		if revision == nil {
			return fmt.Errorf("no revision found to test")
		}
		
		if len(revision.TestCases) == 0 {
			return fmt.Errorf("no test cases found")
		}
		
		fmt.Fprintf(cmd.OutOrStdout(), "Running tests for prompt '%s'...\n\n", prompt.Name)
		
		passed := 0
		for i, tc := range revision.TestCases {
			fmt.Fprintf(cmd.OutOrStdout(), "Test %d: %s\n", i+1, tc.Description)
			
			// In a real implementation, we would evaluate the test case against the LLM
			success := i%2 == 0 // Mock result
			statusText := map[bool]string{true: "PASS", false: "FAIL"}[success]
			
			fmt.Fprintf(cmd.OutOrStdout(), "  Status: %s\n", statusText)
			fmt.Fprintf(cmd.OutOrStdout(), "  Expected: %s\n", tc.ExpectedOutput)
			fmt.Fprintf(cmd.OutOrStdout(), "  Actual: %s\n\n", tc.CompletionText)
			
			if success {
				passed++
			}
		}
		
		fmt.Fprintf(cmd.OutOrStdout(), "Results: %d/%d tests passed (%.1f%%)\n", 
			passed, len(revision.TestCases), float64(passed)*100/float64(len(revision.TestCases)))
		
		return nil
	},
}

// statCmd shows statistics for a prompt.
var statCmd = &cobra.Command{
	Use:   "stat [file]",
	Short: "Show statistics",
	Long:  `Show statistics for a prompt or collection of prompts.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		
		prompt, err := readPromptFromFile(filePath)
		if err != nil {
			return err
		}
		
		fmt.Fprintf(cmd.OutOrStdout(), "Statistics for prompt '%s':\n\n", prompt.Name)
		
		fmt.Fprintf(cmd.OutOrStdout(), "General:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Created: %s\n", prompt.CreatedAt.AsTime().Format(time.RFC3339))
		fmt.Fprintf(cmd.OutOrStdout(), "  Updated: %s\n", prompt.UpdatedAt.AsTime().Format(time.RFC3339))
		fmt.Fprintf(cmd.OutOrStdout(), "  Revisions: %d\n", len(prompt.Revisions))
		
		if prompt.LatestRevision != nil {
			rev := prompt.LatestRevision
			fmt.Fprintf(cmd.OutOrStdout(), "\nLatest Revision:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  ID: %s\n", rev.Id)
			fmt.Fprintf(cmd.OutOrStdout(), "  Created: %s\n", rev.CreatedAt.AsTime().Format(time.RFC3339))
			fmt.Fprintf(cmd.OutOrStdout(), "  Model: %s\n", rev.ModelName)
			fmt.Fprintf(cmd.OutOrStdout(), "  Variables: %d\n", len(rev.Variables))
			fmt.Fprintf(cmd.OutOrStdout(), "  Test Cases: %d\n", len(rev.TestCases))
			fmt.Fprintf(cmd.OutOrStdout(), "  Messages: %d\n", len(rev.Messages))
			fmt.Fprintf(cmd.OutOrStdout(), "  Examples: %d\n", len(rev.Examples))
			
			// Test results
			passed := 0
			for _, tc := range rev.TestCases {
				if tc.IsSuccess {
					passed++
				}
			}
			if len(rev.TestCases) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  Test Success Rate: %.1f%% (%d/%d)\n", 
					float64(passed)*100/float64(len(rev.TestCases)), passed, len(rev.TestCases))
			}
			
			// System prompt stats
			words := len(strings.Fields(rev.SystemPrompt))
			fmt.Fprintf(cmd.OutOrStdout(), "  System Prompt: %d words, %d chars\n", words, len(rev.SystemPrompt))
		}
		
		return nil
	},
}

// searchCmd searches for prompts matching criteria.
var searchCmd = &cobra.Command{
	Use:   "search [pattern]",
	Short: "Search for prompts",
	Long:  `Search for prompts matching specified criteria.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		dir, _ := cmd.Flags().GetString("dir")
		
		if dir == "" {
			dir = "."
		}
		
		matches, err := searchPrompts(dir, pattern)
		if err != nil {
			return err
		}
		
		if len(matches) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No prompts found matching '%s'\n", pattern)
			return nil
		}
		
		fmt.Fprintf(cmd.OutOrStdout(), "Found %d prompts matching '%s':\n\n", len(matches), pattern)
		
		for i, match := range matches {
			prompt, err := readPromptFromFile(match)
			if err != nil {
				continue
			}
			
			fmt.Fprintf(cmd.OutOrStdout(), "%d. %s\n", i+1, match)
			fmt.Fprintf(cmd.OutOrStdout(), "   Name: %s\n", prompt.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "   Revisions: %d\n", len(prompt.Revisions))
			if prompt.LatestRevision != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "   Model: %s\n", prompt.LatestRevision.ModelName)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n")
		}
		
		return nil
	},
}

// Helper functions

// inferFormat infers the format from a file path.
func inferFormat(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	
	switch ext {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".pb", ".proto":
		return "proto"
	case ".txt":
		return "text"
	default:
		return "proto"
	}
}

// generateID generates a simple ID for new prompts and revisions.
func generateID() string {
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}

// readPromptFromFile reads a prompt from a file in the appropriate format.
func readPromptFromFile(path string) (*pev1.Prompt, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	format := inferFormat(path)
	
	var prompt pev1.Prompt
	
	switch format {
	case "json":
		err = protojson.Unmarshal(data, &prompt)
	case "proto":
		err = proto.Unmarshal(data, &prompt)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal prompt: %w", err)
	}
	
	return &prompt, nil
}

// writePromptToFile writes a prompt to a file in the appropriate format.
func writePromptToFile(prompt *pev1.Prompt, path string) error {
	format := inferFormat(path)
	
	var data []byte
	var err error
	
	switch format {
	case "json":
		data, err = protojson.Marshal(prompt)
	case "proto":
		data, err = proto.Marshal(prompt)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	
	if err != nil {
		return fmt.Errorf("failed to marshal prompt: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// displayPrompt displays a prompt in the specified format.
func displayPrompt(prompt *pev1.Prompt, format string, out io.Writer) error {
	switch format {
	case "json":
		data, err := protojson.Marshal(prompt)
		if err != nil {
			return err
		}
		
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
			return err
		}
		
		_, err = prettyJSON.WriteTo(out)
		return err
		
	case "summary":
		fmt.Fprintf(out, "Prompt: %s\n", prompt.Name)
		fmt.Fprintf(out, "ID: %s\n", prompt.Id)
		fmt.Fprintf(out, "Created: %s\n", prompt.CreatedAt.AsTime().Format(time.RFC3339))
		fmt.Fprintf(out, "Updated: %s\n", prompt.UpdatedAt.AsTime().Format(time.RFC3339))
		fmt.Fprintf(out, "Revisions: %d\n", len(prompt.Revisions))
		
		if len(prompt.Metadata) > 0 {
			fmt.Fprintf(out, "\nMetadata:\n")
			for k, v := range prompt.Metadata {
				fmt.Fprintf(out, "  %s: %s\n", k, v)
			}
		}
		
		if prompt.LatestRevision != nil {
			fmt.Fprintf(out, "\nLatest Revision:\n")
			fmt.Fprintf(out, "  ID: %s\n", prompt.LatestRevision.Id)
			fmt.Fprintf(out, "  Model: %s\n", prompt.LatestRevision.ModelName)
			fmt.Fprintf(out, "  Temperature: %.2f\n", prompt.LatestRevision.Temperature)
			
			if prompt.LatestRevision.SystemPrompt != "" {
				fmt.Fprintf(out, "\nSystem Prompt:\n")
				fmt.Fprintf(out, "%s\n", prompt.LatestRevision.SystemPrompt)
			}
		}
		
		return nil
		
	default:
		return fmt.Errorf("unsupported display format: %s", format)
	}
}

// displayRevision displays a prompt revision in the specified format.
func displayRevision(revision *pev1.PromptRevision, format string, out io.Writer) error {
	switch format {
	case "json":
		data, err := protojson.Marshal(revision)
		if err != nil {
			return err
		}
		
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
			return err
		}
		
		_, err = prettyJSON.WriteTo(out)
		return err
		
	case "summary":
		fmt.Fprintf(out, "Revision: %s\n", revision.Id)
		fmt.Fprintf(out, "Created: %s\n", revision.CreatedAt.AsTime().Format(time.RFC3339))
		fmt.Fprintf(out, "Model: %s\n", revision.ModelName)
		fmt.Fprintf(out, "Max Tokens: %d\n", revision.MaxTokensToSample)
		fmt.Fprintf(out, "Temperature: %.2f\n", revision.Temperature)
		
		if revision.AverageRating != nil {
			fmt.Fprintf(out, "Average Rating: %.2f\n", revision.AverageRating.Value)
		}
		
		fmt.Fprintf(out, "Variables: %d\n", len(revision.Variables))
		fmt.Fprintf(out, "Test Cases: %d\n", len(revision.TestCases))
		fmt.Fprintf(out, "Messages: %d\n", len(revision.Messages))
		fmt.Fprintf(out, "Examples: %d\n", len(revision.Examples))
		
		if len(revision.Metadata) > 0 {
			fmt.Fprintf(out, "\nMetadata:\n")
			for k, v := range revision.Metadata {
				fmt.Fprintf(out, "  %s: %s\n", k, v)
			}
		}
		
		if revision.SystemPrompt != "" {
			fmt.Fprintf(out, "\nSystem Prompt:\n")
			fmt.Fprintf(out, "%s\n", revision.SystemPrompt)
		}
		
		return nil
		
	default:
		return fmt.Errorf("unsupported display format: %s", format)
	}
}

// validatePrompt validates a prompt against schema and semantic rules.
func validatePrompt(prompt *pev1.Prompt) []string {
	var issues []string
	
	// Check required fields
	if prompt.Id == "" {
		issues = append(issues, "Prompt is missing an ID")
	}
	
	if prompt.Name == "" {
		issues = append(issues, "Prompt is missing a name")
	}
	
	if prompt.CreatedAt == nil {
		issues = append(issues, "Prompt is missing creation timestamp")
	}
	
	if prompt.UpdatedAt == nil {
		issues = append(issues, "Prompt is missing update timestamp")
	}
	
	// Check revisions
	if len(prompt.Revisions) == 0 {
		issues = append(issues, "Prompt has no revisions")
	}
	
	if prompt.LatestRevision == nil && len(prompt.Revisions) > 0 {
		issues = append(issues, "Prompt has revisions but no latest_revision is set")
	}
	
	// Check each revision
	for i, rev := range prompt.Revisions {
		prefix := fmt.Sprintf("Revision %d (%s)", i, rev.Id)
		
		if rev.Id == "" {
			issues = append(issues, fmt.Sprintf("%s: Missing ID", prefix))
		}
		
		if rev.CreatedAt == nil {
			issues = append(issues, fmt.Sprintf("%s: Missing creation timestamp", prefix))
		}
		
		if rev.PromptId == "" {
			issues = append(issues, fmt.Sprintf("%s: Missing prompt_id reference", prefix))
		}
		
		if rev.PromptId != prompt.Id {
			issues = append(issues, fmt.Sprintf("%s: prompt_id doesn't match parent prompt ID", prefix))
		}
		
		if rev.ModelName == "" {
			issues = append(issues, fmt.Sprintf("%s: Missing model_name", prefix))
		}
	}
	
	return issues
}

// exportPromptToJSON exports a prompt to a JSON file.
func exportPromptToJSON(prompt *pev1.Prompt, outputPath string) error {
	data, err := protojson.Marshal(prompt)
	if err != nil {
		return fmt.Errorf("failed to marshal prompt to JSON: %w", err)
	}
	
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	
	if err := os.WriteFile(outputPath, prettyJSON.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// exportPromptToYAML exports a prompt to a YAML file.
func exportPromptToYAML(prompt *pev1.Prompt, outputPath string) error {
	// For now, just export as JSON (would use a proper YAML marshaler in real implementation)
	return exportPromptToJSON(prompt, outputPath)
}

// exportPromptToText exports a prompt to a text file.
func exportPromptToText(prompt *pev1.Prompt, outputPath string) error {
	var buf bytes.Buffer
	
	if err := displayPrompt(prompt, "summary", &buf); err != nil {
		return err
	}
	
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// importPromptFromJSON imports a prompt from a JSON file.
func importPromptFromJSON(inputPath string) (*pev1.Prompt, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	var prompt pev1.Prompt
	if err := protojson.Unmarshal(data, &prompt); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	return &prompt, nil
}

// importPromptFromYAML imports a prompt from a YAML file.
func importPromptFromYAML(inputPath string) (*pev1.Prompt, error) {
	// For now, just assume it's JSON (would use a proper YAML parser in real implementation)
	return importPromptFromJSON(inputPath)
}

// searchPrompts searches for prompt files matching a pattern.
func searchPrompts(dir, pattern string) ([]string, error) {
	var matches []string
	
	// Get a list of potential prompt files
	files, err := filepath.Glob(filepath.Join(dir, "*.pb"))
	if err != nil {
		return nil, err
	}
	
	jsonFiles, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	
	files = append(files, jsonFiles...)
	
	// Check each file
	for _, file := range files {
		// Try to read as a prompt
		prompt, err := readPromptFromFile(file)
		if err != nil {
			continue
		}
		
		// Check if the prompt matches the pattern
		if strings.Contains(prompt.Name, pattern) || 
		   strings.Contains(file, pattern) {
			matches = append(matches, file)
			continue
		}
		
		// Check system prompts and models
		for _, rev := range prompt.Revisions {
			if strings.Contains(rev.SystemPrompt, pattern) || 
			   strings.Contains(rev.ModelName, pattern) {
				matches = append(matches, file)
				break
			}
		}
	}
	
	return matches, nil
}

// Init function to set up command flags
func init() {
	// createCmd flags
	createCmd.Flags().StringP("name", "n", "New Prompt", "Name of the prompt")
	createCmd.Flags().StringP("system-prompt", "s", "", "System prompt text")
	createCmd.Flags().StringP("model", "m", "gpt-4", "Model name")
	createCmd.Flags().Int32P("max-tokens", "t", 1024, "Maximum tokens to sample")
	createCmd.Flags().Float32P("temperature", "p", 0.7, "Temperature")
	createCmd.Flags().StringP("output", "o", "", "Output file path")
	
	// showCmd flags
	showCmd.Flags().StringP("format", "f", "summary", "Output format (json, summary)")
	showCmd.Flags().StringP("revision", "r", "", "Show specific revision by ID")
	
	// analyzeCmd flags
	analyzeCmd.Flags().StringP("revision", "r", "", "Analyze specific revision by ID")
	
	// evalCmd flags
	evalCmd.Flags().StringP("provider", "p", "openai", "LLM provider to use for evaluation")
	
	// exportCmd flags
	exportCmd.Flags().StringP("format", "f", "", "Export format (json, yaml, text, proto)")
	
	// importCmd flags
	importCmd.Flags().StringP("format", "f", "", "Import format (json, yaml, proto)")
	
	// searchCmd flags
	searchCmd.Flags().StringP("dir", "d", ".", "Directory to search in")
}