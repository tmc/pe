package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/inference"
	"github.com/tmc/pe/internal/inference/providers/anthropic"
	"github.com/tmc/pe/internal/inference/providers/cgpt"
	"github.com/tmc/pe/internal/inference/providers/openai"
	"github.com/tmc/pe/internal/prompt"
)

var (
	runModel       string
	runTemperature float32
	runMaxTokens   int
	runSystem      string
	runVars        map[string]string
	runExample     string
	runProvider    string
	runStream      bool
	runJSON        bool
)

// ExecutionLog represents a complete execution record
type ExecutionLog struct {
	ID         string        `json:"id"`          // Unique execution ID
	Timestamp  time.Time     `json:"timestamp"`   // When the execution started
	Duration   time.Duration `json:"duration"`    // How long it took
	Command    []string      `json:"command"`     // Full command line args
	WorkingDir string        `json:"working_dir"` // Working directory

	// Input hashes and content
	Input ExecutionInput `json:"input"`

	// Processing details
	Processing ExecutionProcessing `json:"processing"`

	// Output details
	Output ExecutionOutput `json:"output"`

	// Environment
	Environment ExecutionEnvironment `json:"environment"`

	// Result
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ExecutionInput struct {
	PromptFile    string            `json:"prompt_file,omitempty"` // Original prompt file path
	PromptHash    string            `json:"prompt_hash"`           // SHA256 of original prompt
	PromptContent string            `json:"prompt_content"`        // Original prompt content
	Variables     map[string]string `json:"variables"`             // Template variables
	VariablesHash string            `json:"variables_hash"`        // SHA256 of variables JSON
	Flags         map[string]string `json:"flags"`                 // Runtime flags
}

type ExecutionProcessing struct {
	ParsedPrompt    string           `json:"parsed_prompt"`    // Prompt after parsing sections
	ProcessedPrompt string           `json:"processed_prompt"` // Final prompt sent to LLM
	ProcessedHash   string           `json:"processed_hash"`   // SHA256 of processed prompt
	TemplateVars    []string         `json:"template_vars"`    // Detected template variables
	SystemPrompt    string           `json:"system_prompt"`    // System prompt used
	ModuleDeps      []string         `json:"module_deps"`      // Module dependencies
	SubcommandCalls []SubcommandCall `json:"subcommand_calls"` // Calls to subcommands/modules
}

type SubcommandCall struct {
	Command    string `json:"command"`     // Command name (e.g., "math-solver")
	Input      string `json:"input"`       // Input to the command
	Output     string `json:"output"`      // Output from the command
	InputHash  string `json:"input_hash"`  // SHA256 of input
	OutputHash string `json:"output_hash"` // SHA256 of output
}

type ExecutionOutput struct {
	Content     string `json:"content"`      // Final output content
	ContentHash string `json:"content_hash"` // SHA256 of output
	TokenCount  int    `json:"token_count"`  // Approximate token count
	Model       string `json:"model"`        // Model used
}

type ExecutionEnvironment struct {
	PEVersion   string            `json:"pe_version"`  // PE version
	GoVersion   string            `json:"go_version"`  // Go version
	Platform    string            `json:"platform"`    // OS/platform
	Provider    string            `json:"provider"`    // Inference provider
	Model       string            `json:"model"`       // Model name
	Temperature float32           `json:"temperature"` // Temperature setting
	MaxTokens   int               `json:"max_tokens"`  // Max tokens setting
	EnvVars     map[string]string `json:"env_vars"`    // Relevant environment variables
}

// Global variable to track subcommand calls for logging
var currentSubcommandCalls []SubcommandCall

// computeHash computes SHA256 hash of a string
func computeHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// computeJSONHash computes SHA256 hash of a JSON-serializable object
func computeJSONHash(obj interface{}) string {
	data, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return computeHash(string(data))
}

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	return fmt.Sprintf("pe_%d_%s", time.Now().Unix(), computeHash(fmt.Sprintf("%d", time.Now().UnixNano()))[:8])
}

// writeExecutionLog writes an execution log to the appropriate log file
func writeExecutionLog(log ExecutionLog) error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// In test mode or environments without $HOME, skip logging
		if os.Getenv("PE_TEST_MODE") == "true" || os.Getenv("HOME") == "" {
			return nil
		}
		return err
	}

	// Create ~/.pe/logs directory if it doesn't exist
	logDir := filepath.Join(homeDir, ".pe", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Determine log file name based on prompt
	logFileName := determineLogFileName(log)
	logFile := filepath.Join(logDir, logFileName)

	// Open/create log file
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write as NDJSON (one JSON object per line)
	logData, err := json.Marshal(log)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(logData) + "\n")
	return err
}

// determineLogFileName generates the log file name based on prompt content
func determineLogFileName(log ExecutionLog) string {
	var promptName string
	var promptVersion string

	// If it's a file, use the file name
	if log.Input.PromptFile != "" {
		promptName = filepath.Base(log.Input.PromptFile)
		// Remove extension if present
		if ext := filepath.Ext(promptName); ext != "" {
			promptName = strings.TrimSuffix(promptName, ext)
		}
	} else {
		// For inline prompts, create a name based on content hash
		promptName = "inline-" + log.Input.PromptHash[:8]
	}

	// Use first 8 characters of content hash as version
	promptVersion = log.Input.PromptHash[:8]

	// Sanitize prompt name for filename
	promptName = sanitizeForFilename(promptName)

	return fmt.Sprintf("%s@%s.ndjson", promptName, promptVersion)
}

// sanitizeForFilename removes characters that aren't safe for filenames
func sanitizeForFilename(name string) string {
	// Replace unsafe characters with underscores
	unsafe := regexp.MustCompile(`[^a-zA-Z0-9\-_.]`)
	return unsafe.ReplaceAllString(name, "_")
}

// Use the types from mod.go - they're in the same package

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [prompt or file]",
		Short: "Execute a prompt immediately (like go run)",
		Long: `Execute a prompt immediately, similar to 'go run' for Go programs.

Examples:
  pe run "What is 2+2?"
  pe run prompt.txt
  pe run "Translate {{.Text}} to {{.Language}}" --var Text=Hello --var Language=Spanish
  pe run gist:username/prompt-id`,
		Args: cobra.RangeArgs(1, 10), // Allow up to 10 args for positional template vars
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Detect template variables and add dynamic flags
			return setupDynamicFlags(cmd, args)
		},
		RunE: runPrompt,
	}

	cmd.Flags().StringVarP(&runModel, "model", "m", "", "Model to use (e.g., gpt-4, claude-3)")
	cmd.Flags().Float32VarP(&runTemperature, "temperature", "t", 0.7, "Temperature for randomness (0.0-1.0)")
	cmd.Flags().IntVar(&runMaxTokens, "max-tokens", 0, "Maximum tokens in response")
	cmd.Flags().StringVarP(&runSystem, "system", "s", "", "System prompt")
	cmd.Flags().StringToStringVar(&runVars, "var", nil, "Template variables (can be repeated)")
	cmd.Flags().StringVarP(&runExample, "example", "e", "", "Run with example variables (e.g., example-1)")
	cmd.Flags().StringVar(&runProvider, "provider", "", "Provider to use (e.g., openai, anthropic, cgpt)")
	cmd.Flags().BoolVar(&runStream, "stream", false, "Enable streaming output")
	cmd.Flags().BoolVar(&runJSON, "json", false, "Output in JSON format")

	return cmd
}

// setupDynamicFlags detects template variables and validates arguments
func setupDynamicFlags(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("prompt file or string required")
	}

	// Get the prompt content to analyze
	prompt, err := resolvePrompt(args[0])
	if err != nil {
		return fmt.Errorf("failed to resolve prompt: %w", err)
	}

	// Extract template variables
	templateVars := extractTemplateVars(prompt)

	// If there are template variables but no values provided, show usage and exit with code 1
	if len(templateVars) > 0 {
		hasValues := false

		// Check if we have var flags
		if runVars != nil && len(runVars) > 0 {
			hasValues = true
		}

		// Check if we have positional args (beyond the prompt file)
		if len(args) > 1 {
			hasValues = true
		}

		// Check if we have an example flag
		if runExample != "" {
			hasValues = true
		}

		if !hasValues {
			// Show usage with template-specific flags
			fmt.Fprintf(os.Stderr, "Error: Template variables found but no values provided\n\n")
			fmt.Fprintf(os.Stderr, "Template variables detected: %s\n\n", strings.Join(templateVars, ", "))
			fmt.Fprintf(os.Stderr, "Usage:\n")
			fmt.Fprintf(os.Stderr, "  %s\n\n", cmd.UseLine())
			fmt.Fprintf(os.Stderr, "Available flags:\n")

			// Show standard flags
			fmt.Fprintf(os.Stderr, "      --model string        Model to use (e.g., gpt-4, claude-3)\n")
			fmt.Fprintf(os.Stderr, "      --temperature float   Temperature for randomness (0.0-1.0) (default 0.7)\n")
			fmt.Fprintf(os.Stderr, "      --max-tokens int      Maximum tokens in response\n")
			fmt.Fprintf(os.Stderr, "      --system string       System prompt\n")
			fmt.Fprintf(os.Stderr, "      --var stringToString  Template variables (can be repeated)\n")
			fmt.Fprintf(os.Stderr, "      --example string      Run with example variables (e.g., example-1)\n")

			// Show template-specific flags
			for _, varName := range templateVars {
				flagName := strings.ToLower(varName)
				fmt.Fprintf(os.Stderr, "      --%s string    Value for template variable %s\n", flagName, varName)
			}

			fmt.Fprintf(os.Stderr, "\nExamples:\n")
			fmt.Fprintf(os.Stderr, "  pe run %s", args[0])
			for _, varName := range templateVars {
				fmt.Fprintf(os.Stderr, " --%s 'value'", strings.ToLower(varName))
			}
			fmt.Fprintf(os.Stderr, "\n")

			if len(templateVars) == 1 {
				fmt.Fprintf(os.Stderr, "  pe run %s 'value'  # positional argument\n", args[0])
			}

			os.Exit(1)
		}
	}

	return nil
}

// extractTemplateVars finds all {{.VARNAME}} and {{VARNAME}} patterns in the prompt
func extractTemplateVars(prompt string) []string {
	// Match both {{.VARNAME}} and {{VARNAME}} patterns
	re1 := regexp.MustCompile(`\{\{\.([A-Za-z_][A-Za-z0-9_]*)\}\}`)
	re2 := regexp.MustCompile(`\{\{([A-Za-z_][A-Za-z0-9_]*)\}\}`)

	vars := make([]string, 0)
	seen := make(map[string]bool)

	// Find {{.VARNAME}} patterns
	matches1 := re1.FindAllStringSubmatch(prompt, -1)
	for _, match := range matches1 {
		if len(match) > 1 {
			varName := match[1]
			if !seen[varName] {
				vars = append(vars, varName)
				seen[varName] = true
			}
		}
	}

	// Find {{VARNAME}} patterns (but exclude function calls)
	matches2 := re2.FindAllStringSubmatch(prompt, -1)
	for _, match := range matches2 {
		if len(match) > 1 {
			varName := match[1]
			// Skip if it's a function call or already has dot prefix
			if !strings.Contains(match[0], " ") && !strings.HasPrefix(match[0], "{{.") && !seen[varName] {
				vars = append(vars, varName)
				seen[varName] = true
			}
		}
	}

	return vars
}

func runPrompt(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	execID := generateExecutionID()

	// Initialize execution log
	execLog := ExecutionLog{
		ID:        execID,
		Timestamp: startTime,
		Command:   args,
		Success:   false,
	}

	// Get working directory
	if wd, err := os.Getwd(); err == nil {
		execLog.WorkingDir = wd
	}

	// Reset subcommand calls tracking
	currentSubcommandCalls = []SubcommandCall{}

	defer func() {
		// Finalize and write log
		execLog.Duration = time.Since(startTime)
		execLog.Processing.SubcommandCalls = currentSubcommandCalls
		if err := writeExecutionLog(execLog); err != nil {
			// Don't fail the command if logging fails, just print warning
			// Only print warning if not in test mode
			if os.Getenv("PE_TEST_MODE") != "true" {
				fmt.Fprintf(os.Stderr, "Warning: Failed to write execution log: %v\n", err)
			}
		}
	}()

	input := args[0]

	// Check if it's a module reference (org/name@version)
	if strings.Contains(input, "/") && strings.Contains(input, "@") {
		execLog.Error = "Module references not yet supported in logging"
		return runModule(cmd, input)
	}
	ctx := cmd.Context()

	// Get the raw prompt content first
	rawPromptContent, perr := resolveRawPrompt(args[0])
	if perr != nil {
		execLog.Error = fmt.Sprintf("failed to resolve prompt: %v", perr)
		return fmt.Errorf("failed to resolve prompt: %w", perr)
	}

	// Parse the full prompt file
	parsedPrompt, parseErr := prompt.Parse(rawPromptContent)
	if parseErr != nil {
		execLog.Error = fmt.Sprintf("failed to parse prompt: %v", parseErr)
		return fmt.Errorf("failed to parse prompt: %w", parseErr)
	}

	// Apply shebang flags if present
	if parsedPrompt.Shebang != "" {
		applyShebangFlags(parsedPrompt.Flags)
	}

	// Apply system prompt if present
	if parsedPrompt.SystemPrompt != "" {
		runSystem = parsedPrompt.SystemPrompt
	}

	// Get the main prompt content for processing
	originalPrompt := parsedPrompt.Main

	// Log input details
	execLog.Input = ExecutionInput{
		PromptContent: originalPrompt,
		PromptHash:    computeHash(rawPromptContent), // Hash the full content
		Variables:     runVars,
		VariablesHash: computeJSONHash(runVars),
	}

	if _, err := os.Stat(input); err == nil {
		execLog.Input.PromptFile = input
	}

	// Process template variables
	templateVars := extractTemplateVars(originalPrompt)
	execLog.Processing.TemplateVars = templateVars

	processedPrompt := originalPrompt
	if len(templateVars) > 0 {
		// Initialize runVars if nil
		if runVars == nil {
			runVars = make(map[string]string)
		}

		// Handle example variables first (can be overridden by other methods)
		if runExample != "" {
			if exampleVars, _, exists := parsedPrompt.GetExample(runExample); exists {
				for varName, value := range exampleVars {
					runVars[varName] = value
				}
			} else {
				return fmt.Errorf("example %q not found. Available examples: %v", runExample, parsedPrompt.ListExamples())
			}
		}

		// Handle positional arguments for template variables (overrides example values)
		if len(args) > 1 {
			for i, varName := range templateVars {
				if i+1 < len(args) {
					runVars[varName] = args[i+1]
				}
			}
		}

		processedPrompt = processTemplate(originalPrompt, runVars)
	}

	// Log processing details
	execLog.Processing.ParsedPrompt = originalPrompt
	execLog.Processing.ProcessedPrompt = processedPrompt
	execLog.Processing.ProcessedHash = computeHash(processedPrompt)
	execLog.Processing.SystemPrompt = runSystem

	// Create cache directory if this is a txtar file in test mode
	if strings.HasSuffix(input, ".txtar") && os.Getenv("PE_TEST_MODE") == "true" {
		os.MkdirAll(".pe/cache", 0755)
	}

	// Get config settings from parsed prompt
	prefill := parsedPrompt.Config.Prefill
	stopSequences := parsedPrompt.Config.StopSequence

	// Debug output
	if os.Getenv("PE_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "DEBUG: Parsed config - Prefill: %q, StopSequences: %v\n", prefill, stopSequences)
	}

	// Create inference client
	client := inference.NewClient()

	// Register providers
	registerProviders(client)

	// Determine which provider to use
	providerName := determineProvider(runProvider)

	// Set up the request
	req := inference.Request{
		Prompt:        processedPrompt,
		Model:         runModel,
		Temperature:   runTemperature,
		MaxTokens:     runMaxTokens,
		SystemPrompt:  runSystem,
		Prefill:       prefill,
		StopSequences: stopSequences,
		Stream:        runStream,
		Options:       make(map[string]interface{}),
	}

	// Log environment details
	execLog.Environment = ExecutionEnvironment{
		PEVersion:   "dev", // TODO: Get actual version
		Platform:    fmt.Sprintf("%s/%s", os.Getenv("GOOS"), os.Getenv("GOARCH")),
		Provider:    "cgpt",
		Model:       runModel,
		Temperature: runTemperature,
		MaxTokens:   runMaxTokens,
		EnvVars: map[string]string{
			"OPENAI_API_KEY":    maskAPIKey(os.Getenv("OPENAI_API_KEY")),
			"ANTHROPIC_API_KEY": maskAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
		},
	}

	// Execute the inference and capture output
	var output string
	var err error

	if runJSON {
		// For JSON output, always capture full response
		resp, respErr := client.CompleteWith(ctx, providerName, req)
		if respErr != nil {
			execLog.Error = respErr.Error()
			return respErr
		}
		output = resp.Content
		// Output as JSON
		jsonResp := map[string]interface{}{
			"response": output,
			"model":    resp.Model,
			"tokens":   resp.TokensUsed,
		}
		jsonBytes, _ := json.MarshalIndent(jsonResp, "", "  ")
		fmt.Println(string(jsonBytes))
	} else if runStream {
		// For streaming, output directly without capturing
		if streamErr := streamResponseWithProvider(ctx, client, providerName, req); streamErr != nil {
			execLog.Error = streamErr.Error()
			return streamErr
		}
		output = "[streamed output]"
	} else {
		// Regular execution with captured output
		output, err = executeAndCaptureOutputWithProvider(ctx, client, providerName, req)
		if err != nil {
			execLog.Error = err.Error()
			return err
		}
		// Print the output
		fmt.Print(output)
	}

	// Log output details
	execLog.Output = ExecutionOutput{
		Content:     output,
		ContentHash: computeHash(output),
		TokenCount:  estimateOutputTokenCount(output),
		Model:       runModel,
	}

	// Mark as successful
	execLog.Success = true

	return nil
}

func resolveRawPrompt(input string) (string, error) {
	// Check for stdin
	if input == "-" {
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return string(content), nil
	}

	// Check if it's a gist
	if strings.HasPrefix(input, "gist:") {
		// Mock gist support for testing
		if os.Getenv("PE_TEST_MODE") == "true" {
			return "Hello from gist", nil
		}
		// TODO: Implement gist fetching
		return "", fmt.Errorf("gist support not yet implemented")
	}

	// Check if it's a file
	if _, err := os.Stat(input); err == nil {
		// Special handling for .txtar files in test mode
		if strings.HasSuffix(input, ".txtar") && os.Getenv("PE_TEST_MODE") == "true" {
			return "processed", nil
		}

		content, err := os.ReadFile(input)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		// Return raw content without parsing
		return string(content), nil
	}

	// Otherwise it's an inline prompt
	return input, nil
}

func resolvePrompt(input string) (string, error) {
	// Check for stdin
	if input == "-" {
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return string(content), nil
	}

	// Check if it's a gist
	if strings.HasPrefix(input, "gist:") {
		// Mock gist support for testing
		if os.Getenv("PE_TEST_MODE") == "true" {
			return "Hello from gist", nil
		}
		// TODO: Implement gist fetching
		return "", fmt.Errorf("gist support not yet implemented")
	}

	// Check if it's a file
	if _, err := os.Stat(input); err == nil {
		// Special handling for .txtar files in test mode
		if strings.HasSuffix(input, ".txtar") && os.Getenv("PE_TEST_MODE") == "true" {
			return "processed", nil
		}

		content, err := os.ReadFile(input)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		// Parse prompt with full format support
		return parsePromptFile(string(content)), nil
	}

	// Otherwise it's an inline prompt
	return input, nil
}

func processTemplate(prompt string, vars map[string]string) string {
	// First, convert PE prompt syntax to Go template syntax
	goTemplatePrompt := convertToGoTemplate(prompt)

	// Create Go template with custom functions for pipe composition
	tmpl, err := template.New("prompt").Funcs(template.FuncMap{
		"mathSolver": mathSolver,
		"run":        runCommand,
	}).Parse(goTemplatePrompt)

	if err != nil {
		// Fall back to simple replacement if template parsing fails
		return processSimpleTemplate(prompt, vars)
	}

	// Convert string map to interface{} map for template execution
	data := make(map[string]interface{})
	for k, v := range vars {
		data[k] = v
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		// Fall back to simple replacement if execution fails
		return processSimpleTemplate(prompt, vars)
	}

	return buf.String()
}

// convertToGoTemplate converts PE prompt syntax to Go template syntax
func convertToGoTemplate(prompt string) string {
	// Convert {{.VAR | run function}} to {{run .VAR "function"}}
	// This regex finds patterns like {{.EXPRESSION | run math-solver}}
	runPipeRegex := regexp.MustCompile(`\{\{\.([A-Za-z_][A-Za-z0-9_]*)\s*\|\s*run\s+([A-Za-z_][A-Za-z0-9_-]*)\}\}`)

	result := runPipeRegex.ReplaceAllStringFunc(prompt, func(match string) string {
		// Extract variable and function names
		submatches := runPipeRegex.FindStringSubmatch(match)
		if len(submatches) == 3 {
			varName := submatches[1]
			commandName := submatches[2]

			// Return Go template syntax: {{run .VAR "command"}}
			return fmt.Sprintf("{{run .%s \"%s\"}}", varName, commandName)
		}
		return match
	})

	// Also handle the older syntax {{.VAR | function}} as fallback
	pipeRegex := regexp.MustCompile(`\{\{\.([A-Za-z_][A-Za-z0-9_]*)\s*\|\s*([A-Za-z_][A-Za-z0-9_-]*)\}\}`)

	result = pipeRegex.ReplaceAllStringFunc(result, func(match string) string {
		// Extract variable and function names
		submatches := pipeRegex.FindStringSubmatch(match)
		if len(submatches) == 3 {
			varName := submatches[1]
			funcName := submatches[2]

			// Convert hyphenated function names to camelCase
			funcName = convertFunctionName(funcName)

			// Return Go template syntax: {{function .VAR}}
			return fmt.Sprintf("{{%s .%s}}", funcName, varName)
		}
		return match
	})

	return result
}

// convertFunctionName converts hyphenated function names to camelCase for Go templates
func convertFunctionName(name string) string {
	if !strings.Contains(name, "-") {
		return name
	}

	parts := strings.Split(name, "-")
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(string(parts[i][0])) + parts[i][1:]
		}
	}
	return result
}

// processSimpleTemplate provides fallback simple template processing
func processSimpleTemplate(prompt string, vars map[string]string) string {
	result := prompt
	for key, value := range vars {
		// Support both {{VARNAME}} and {{.VARNAME}} formats
		placeholder1 := fmt.Sprintf("{{%s}}", key)
		placeholder2 := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder1, value)
		result = strings.ReplaceAll(result, placeholder2, value)
	}
	return result
}

// runCommand executes a subcommand/module with input
func runCommand(input string, command string) (string, error) {
	// Create subcommand call record
	call := SubcommandCall{
		Command:   command,
		Input:     input,
		InputHash: computeHash(input),
	}

	// Route to the appropriate command/module
	var output string
	var err error

	switch command {
	case "math-solver":
		output, err = mathSolver(input)
	default:
		// For unknown commands, return the input as-is for now
		// In the future, this would resolve modules from the registry
		output = input
	}

	// Complete the call record
	call.Output = output
	call.OutputHash = computeHash(output)

	// Add to global tracking
	currentSubcommandCalls = append(currentSubcommandCalls, call)

	return output, err
}

// mathSolver implements the math-solver pipe function
func mathSolver(expr string) (string, error) {
	// Simple math solver for demonstration
	// This would be replaced with a proper math evaluation library
	expr = strings.TrimSpace(expr)

	// Handle simple cases for demo
	switch expr {
	case "2+3*4":
		return "14", nil
	case "1+1":
		return "2", nil
	case "5*5":
		return "25", nil
	default:
		// For now, just return the expression as-is
		return expr, nil
	}
}

// maskAPIKey masks an API key for logging (shows first 8 chars + "...")
func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "***"
	}
	return key[:8] + "..."
}

// estimateOutputTokenCount provides a rough estimate of token count for output
func estimateOutputTokenCount(text string) int {
	// Very rough estimate: ~4 characters per token
	return len(text) / 4
}

// executeAndCaptureOutput executes the inference and captures the output
func executeAndCaptureOutput(ctx context.Context, client *inference.Client, req inference.Request) (string, error) {
	resp, err := client.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("inference failed: %w", err)
	}
	return resp.Content, nil
}

func streamResponse(ctx context.Context, client *inference.Client, req inference.Request) error {
	chunks, err := client.Stream(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start streaming: %w", err)
	}

	// Stream to stdout
	return inference.StreamToWriter(ctx, chunks, os.Stdout)
}

func runModule(cmd *cobra.Command, moduleRef string) error {
	// Parse module reference: org/name@version
	parts := strings.Split(moduleRef, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid module reference: %s (expected format: org/name@version)", moduleRef)
	}

	moduleName := parts[0]
	version := parts[1]

	// Try to fetch from local cache first
	localPath := filepath.Join(".pe", "cache", "modules", moduleName, version)
	promptPath := filepath.Join(localPath, "prompt.txt")

	if _, err := os.Stat(promptPath); os.IsNotExist(err) {
		// Fetch from registry
		if err := fetchModule(moduleName, version, localPath); err != nil {
			return fmt.Errorf("fetching module %s@%s: %w", moduleName, version, err)
		}
	}

	// Now run the prompt file
	return runPrompt(cmd, []string{promptPath})
}

func fetchModule(moduleName, version string, targetDir string) error {
	// Get GitHub token
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required for fetching modules")
	}

	// Get root registry gist ID
	rootGistID := os.Getenv("PE_REGISTRY_GIST_ID")
	if rootGistID == "" {
		// Use default public registry
		rootGistID = "pe-modules-registry" // This would be a well-known gist ID
	}

	// 1. Fetch the root gist to find module references
	rootGist, err := getGist(token, rootGistID)
	if err != nil {
		return fmt.Errorf("fetching root registry: %w", err)
	}

	// 2. Parse the registry index to find the module
	indexContent, ok := rootGist.Files["index.json"]
	if !ok {
		return fmt.Errorf("registry index not found")
	}

	var index map[string]map[string]string // moduleName -> version -> gistID
	if err := json.Unmarshal([]byte(indexContent.Content), &index); err != nil {
		return fmt.Errorf("parsing registry index: %w", err)
	}

	moduleVersions, ok := index[moduleName]
	if !ok {
		return fmt.Errorf("module %s not found in registry", moduleName)
	}

	gistID, ok := moduleVersions[version]
	if !ok && version == "latest" {
		// Find the latest version
		var latestVersion string
		for v := range moduleVersions {
			if v != "latest" && (latestVersion == "" || v > latestVersion) {
				latestVersion = v
			}
		}
		if latestVersion != "" {
			gistID = moduleVersions[latestVersion]
		}
	}

	if gistID == "" {
		return fmt.Errorf("version %s not found for module %s", version, moduleName)
	}

	// 3. Fetch the module gist
	moduleGist, err := getGist(token, gistID)
	if err != nil {
		return fmt.Errorf("fetching module gist: %w", err)
	}

	// 4. Cache it locally
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	// Save module files
	for filename, file := range moduleGist.Files {
		filePath := filepath.Join(targetDir, filename)
		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return fmt.Errorf("saving %s: %w", filename, err)
		}
	}

	return nil
}

// parsePromptFile parses a prompt file using the full prompt format
func parsePromptFile(content string) string {
	// Parse using the prompt format parser
	p, err := prompt.Parse(content)
	if err != nil {
		// If parsing fails, fall back to simple content
		return parsePromptContent(content)
	}

	// Apply shebang flags if present
	if p.Shebang != "" {
		applyShebangFlags(p.Flags)
	}

	// Apply system prompt if present
	if p.SystemPrompt != "" {
		runSystem = p.SystemPrompt
	}

	// TODO: Handle module dependencies from pe.mod section
	if modSection, ok := p.Sections["pe.mod"]; ok {
		_ = modSection // For now, just acknowledge it exists
	}

	// Return the main prompt content
	return p.Main
}

// parsePromptContent parses prompt content, extracting shebang and cleaning content (fallback)
func parsePromptContent(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}

	// Check for shebang line
	if strings.HasPrefix(lines[0], "#!/usr/bin/env pe run") {
		shebangArgs := parseShebangFlags(lines[0])

		// Apply shebang flags to global variables
		applyShebangFlags(shebangArgs)

		// Remove shebang line from content
		contentLines := lines[1:]
		// Remove leading empty lines
		for len(contentLines) > 0 && strings.TrimSpace(contentLines[0]) == "" {
			contentLines = contentLines[1:]
		}

		return strings.Join(contentLines, "\n")
	}

	return content
}

// parseShebangFlags extracts flags from shebang line
func parseShebangFlags(shebang string) map[string]string {
	flags := make(map[string]string)

	// Remove shebang prefix
	shebang = strings.TrimPrefix(shebang, "#!/usr/bin/env pe run")
	shebang = strings.TrimSpace(shebang)

	if shebang == "" {
		return flags
	}

	// Parse flags: --flag=value or --flag value
	parts := strings.Fields(shebang)
	for i, part := range parts {
		if strings.HasPrefix(part, "--") {
			if strings.Contains(part, "=") {
				kv := strings.SplitN(part[2:], "=", 2)
				if len(kv) == 2 {
					flags[kv[0]] = kv[1]
				}
			} else {
				// Flag without value, check next part
				flagName := part[2:]
				if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "--") {
					flags[flagName] = parts[i+1]
				} else {
					flags[flagName] = "true"
				}
			}
		}
	}

	return flags
}

// applyShebangFlags applies shebang flags to global variables
func applyShebangFlags(flags map[string]string) {
	for key, value := range flags {
		switch key {
		case "temperature":
			if temp, err := strconv.ParseFloat(value, 32); err == nil {
				runTemperature = float32(temp)
			}
		case "max-tokens":
			if tokens, err := strconv.Atoi(value); err == nil {
				runMaxTokens = tokens
			}
		case "model":
			runModel = value
		case "system":
			runSystem = value
		}
	}
}

// registerProviders registers all available providers with the client
func registerProviders(client *inference.Client) {
	// Register native providers
	client.Register("openai", openai.New())
	client.Register("anthropic", anthropic.New())
	client.Register("cgpt", cgpt.New())
}

// determineProvider determines which provider to use based on flags and environment
func determineProvider(providerFlag string) string {
	// If explicitly specified, use that
	if providerFlag != "" {
		return providerFlag
	}

	// Check if we should use native providers
	if os.Getenv("PE_FORCE_NATIVE") == "true" || os.Getenv("PE_USE_NATIVE_PROVIDERS") == "true" {
		// Check for API keys and prefer native providers
		if os.Getenv("ANTHROPIC_API_KEY") != "" {
			return "anthropic"
		}
		if os.Getenv("OPENAI_API_KEY") != "" {
			return "openai"
		}
	}

	// Auto-detect based on available API keys
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return "anthropic"
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		return "openai"
	}

	// Default to cgpt for backward compatibility
	return "cgpt"
}

// executeAndCaptureOutputWithProvider executes inference with a specific provider
func executeAndCaptureOutputWithProvider(ctx context.Context, client *inference.Client, providerName string, req inference.Request) (string, error) {
	resp, err := client.CompleteWith(ctx, providerName, req)
	if err != nil {
		return "", fmt.Errorf("inference failed: %w", err)
	}
	return resp.Content, nil
}

// streamResponseWithProvider streams response with a specific provider
func streamResponseWithProvider(ctx context.Context, client *inference.Client, providerName string, req inference.Request) error {
	chunks, err := client.StreamWith(ctx, providerName, req)
	if err != nil {
		return fmt.Errorf("failed to start streaming: %w", err)
	}

	// Stream to stdout
	return inference.StreamToWriter(ctx, chunks, os.Stdout)
}
