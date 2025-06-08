package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// REPLSession manages an interactive REPL session
type REPLSession struct {
	provider    string
	temperature float64
	maxTokens   int
	history     []string
	configFile  string
	outputFile  string
	cmd         *cobra.Command
}

// NewREPLSession creates a new REPL session
func NewREPLSession(cmd *cobra.Command, provider, configFile string, temperature float64) *REPLSession {
	return &REPLSession{
		provider:    provider,
		temperature: temperature,
		maxTokens:   4096,
		history:     []string{},
		configFile:  configFile,
		cmd:         cmd,
	}
}

// Run starts the REPL session
func (r *REPLSession) Run() error {
	r.printWelcome()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(r.cmd.OutOrStdout(), "pe> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Add to history
		r.history = append(r.history, input)

		// Handle special commands
		if r.handleCommand(input) {
			continue
		}

		// Process as prompt
		if err := r.processPrompt(input); err != nil {
			fmt.Fprintf(r.cmd.OutOrStderr(), "Error: %v\n", err)
		}
		fmt.Fprintln(r.cmd.OutOrStdout())
	}

	return scanner.Err()
}

// printWelcome prints the welcome message
func (r *REPLSession) printWelcome() {
	fmt.Fprintf(r.cmd.OutOrStdout(), `
┌─────────────────────────────────────────────────────────────┐
│                    PE Interactive Mode                     │
│                Prompt Engineering Toolkit                  │
└─────────────────────────────────────────────────────────────┘

Provider:    %s
Temperature: %.1f
Max Tokens:  %d
Config:      %s

Commands:
  :help, :h                    Show help
  :quit, :q                    Exit
  :provider <name>             Switch provider
  :temp <value>                Set temperature (0.0-2.0)
  :tokens <value>              Set max tokens
  :save <filename>             Save session to file
  :load <filename>             Load configuration
  :history                     Show command history
  :clear                       Clear screen
  :multiline                   Enter multiline mode
  :benchmark <n>               Benchmark current prompt n times

Type your prompts below:

`, r.provider, r.temperature, r.maxTokens, r.getConfigDisplay())
}

// getConfigDisplay returns the config file display string
func (r *REPLSession) getConfigDisplay() string {
	if r.configFile != "" {
		return r.configFile
	}
	return "(none)"
}

// handleCommand processes special REPL commands
func (r *REPLSession) handleCommand(input string) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 || !strings.HasPrefix(parts[0], ":") {
		return false
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case ":quit", ":q", ":exit":
		fmt.Fprintln(r.cmd.OutOrStdout(), "Goodbye!")
		os.Exit(0)

	case ":help", ":h":
		r.printHelp()

	case ":provider", ":p":
		if len(args) == 0 {
			fmt.Fprintf(r.cmd.OutOrStdout(), "Current provider: %s\n", r.provider)
		} else {
			r.provider = args[0]
			fmt.Fprintf(r.cmd.OutOrStdout(), "Switched to provider: %s\n", r.provider)
		}

	case ":temp", ":temperature":
		if len(args) == 0 {
			fmt.Fprintf(r.cmd.OutOrStdout(), "Current temperature: %.1f\n", r.temperature)
		} else {
			if temp, err := strconv.ParseFloat(args[0], 64); err == nil && temp >= 0.0 && temp <= 2.0 {
				r.temperature = temp
				fmt.Fprintf(r.cmd.OutOrStdout(), "Temperature set to: %.1f\n", r.temperature)
			} else {
				fmt.Fprintf(r.cmd.OutOrStderr(), "Invalid temperature: %s (must be 0.0-2.0)\n", args[0])
			}
		}

	case ":tokens", ":max-tokens":
		if len(args) == 0 {
			fmt.Fprintf(r.cmd.OutOrStdout(), "Current max tokens: %d\n", r.maxTokens)
		} else {
			if tokens, err := strconv.Atoi(args[0]); err == nil && tokens > 0 {
				r.maxTokens = tokens
				fmt.Fprintf(r.cmd.OutOrStdout(), "Max tokens set to: %d\n", r.maxTokens)
			} else {
				fmt.Fprintf(r.cmd.OutOrStderr(), "Invalid max tokens: %s (must be positive integer)\n", args[0])
			}
		}

	case ":history":
		r.printHistory()

	case ":clear", ":cls":
		fmt.Fprint(r.cmd.OutOrStdout(), "\033[2J\033[H")

	case ":save":
		if len(args) == 0 {
			fmt.Fprintf(r.cmd.OutOrStderr(), "Usage: :save <filename>\n")
		} else {
			r.saveSession(args[0])
		}

	case ":load":
		if len(args) == 0 {
			fmt.Fprintf(r.cmd.OutOrStderr(), "Usage: :load <filename>\n")
		} else {
			r.loadConfig(args[0])
		}

	case ":multiline", ":ml":
		r.handleMultiline()

	case ":benchmark", ":bench":
		iterations := 5
		if len(args) > 0 {
			if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
				iterations = n
			}
		}
		r.runBenchmark(iterations)

	case ":status":
		r.printStatus()

	case ":models":
		r.listModels()

	default:
		fmt.Fprintf(r.cmd.OutOrStderr(), "Unknown command: %s\n", command)
		fmt.Fprintf(r.cmd.OutOrStderr(), "Type :help for available commands\n")
	}

	return true
}

// processPrompt processes a user prompt
func (r *REPLSession) processPrompt(prompt string) error {
	fmt.Fprintf(r.cmd.OutOrStdout(), "Processing with %s...\n", r.provider)

	// Create ask command and execute
	askCmd := askCmd()
	args := []string{
		"--provider", r.provider,
		"--temperature", fmt.Sprintf("%.1f", r.temperature),
		"--max-tokens", strconv.Itoa(r.maxTokens),
		"--format", "text",
		prompt,
	}

	askCmd.SetArgs(args)
	askCmd.SetOut(r.cmd.OutOrStdout())
	askCmd.SetErr(r.cmd.OutOrStderr())

	return askCmd.Execute()
}

// printHelp prints the help message
func (r *REPLSession) printHelp() {
	help := `
PE Interactive Commands:

Session Management:
  :quit, :q             Exit interactive mode
  :help, :h             Show this help
  :clear, :cls          Clear screen
  :status               Show current settings

Provider Settings:
  :provider <name>      Switch provider (e.g., openai:gpt-4)
  :temp <value>         Set temperature (0.0-2.0)
  :tokens <value>       Set max tokens
  :models               List available models

Session Features:
  :history              Show command history
  :save <file>          Save session history
  :load <file>          Load configuration
  :multiline, :ml       Enter multiline input mode
  :benchmark <n>        Benchmark current prompt

Examples:
  :provider anthropic:claude-3-haiku
  :temp 0.9
  :tokens 2048
  What is the capital of France?
  :save my-session.yaml

Tips:
  - Use up/down arrows for command history
  - Multi-word prompts don't need quotes
  - Use :multiline for complex prompts
`
	fmt.Fprint(r.cmd.OutOrStdout(), help)
}

// printHistory prints the command history
func (r *REPLSession) printHistory() {
	if len(r.history) == 0 {
		fmt.Fprintln(r.cmd.OutOrStdout(), "No history available")
		return
	}

	fmt.Fprintln(r.cmd.OutOrStdout(), "Command History:")
	for i, cmd := range r.history {
		if i >= len(r.history)-10 { // Show last 10 commands
			fmt.Fprintf(r.cmd.OutOrStdout(), "%3d: %s\n", i+1, cmd)
		}
	}
}

// printStatus prints the current session status
func (r *REPLSession) printStatus() {
	fmt.Fprintf(r.cmd.OutOrStdout(), `
Current Session Status:
  Provider:     %s
  Temperature:  %.1f
  Max Tokens:   %d
  Config File:  %s
  History Items: %d
`, r.provider, r.temperature, r.maxTokens, r.getConfigDisplay(), len(r.history))
}

// saveSession saves the current session to a file
func (r *REPLSession) saveSession(filename string) {
	session := map[string]interface{}{
		"provider":    r.provider,
		"temperature": r.temperature,
		"max_tokens":  r.maxTokens,
		"history":     r.history,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	data, err := yaml.Marshal(session)
	if err != nil {
		fmt.Fprintf(r.cmd.OutOrStderr(), "Error marshaling session: %v\n", err)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Fprintf(r.cmd.OutOrStderr(), "Error saving session: %v\n", err)
		return
	}

	fmt.Fprintf(r.cmd.OutOrStdout(), "Session saved to: %s\n", filename)
}

// loadConfig loads a configuration file
func (r *REPLSession) loadConfig(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Fprintf(r.cmd.OutOrStderr(), "File not found: %s\n", filename)
		return
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(r.cmd.OutOrStderr(), "Error reading file: %v\n", err)
		return
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Fprintf(r.cmd.OutOrStderr(), "Error parsing config: %v\n", err)
		return
	}

	// Apply config
	if provider, ok := config["provider"].(string); ok {
		r.provider = provider
	}
	if temp, ok := config["temperature"].(float64); ok {
		r.temperature = temp
	}
	if tokens, ok := config["max_tokens"].(int); ok {
		r.maxTokens = tokens
	}

	r.configFile = filename
	fmt.Fprintf(r.cmd.OutOrStdout(), "Configuration loaded from: %s\n", filename)
}

// handleMultiline handles multiline input mode
func (r *REPLSession) handleMultiline() {
	fmt.Fprintln(r.cmd.OutOrStdout(), "Entering multiline mode. Type 'END' on a new line to finish:")

	var lines []string
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Fprint(r.cmd.OutOrStdout(), "... ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if strings.TrimSpace(line) == "END" {
			break
		}

		lines = append(lines, line)
	}

	if len(lines) > 0 {
		prompt := strings.Join(lines, "\n")
		fmt.Fprintln(r.cmd.OutOrStdout(), "Processing multiline prompt...")
		if err := r.processPrompt(prompt); err != nil {
			fmt.Fprintf(r.cmd.OutOrStderr(), "Error: %v\n", err)
		}
	}
}

// runBenchmark runs a benchmark of the current prompt
func (r *REPLSession) runBenchmark(iterations int) {
	fmt.Fprintln(r.cmd.OutOrStdout(), "Enter the prompt to benchmark:")
	fmt.Fprint(r.cmd.OutOrStdout(), "pe> ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}

	prompt := strings.TrimSpace(scanner.Text())
	if prompt == "" {
		fmt.Fprintln(r.cmd.OutOrStderr(), "No prompt provided")
		return
	}

	fmt.Fprintf(r.cmd.OutOrStdout(), "Running benchmark with %d iterations...\n", iterations)

	var latencies []time.Duration
	var errors int

	for i := 0; i < iterations; i++ {
		fmt.Fprintf(r.cmd.OutOrStdout(), "Iteration %d/%d... ", i+1, iterations)

		start := time.Now()
		if err := r.processPrompt(prompt); err != nil {
			errors++
			fmt.Fprintf(r.cmd.OutOrStderr(), "Error: %v\n", err)
		} else {
			latency := time.Since(start)
			latencies = append(latencies, latency)
			fmt.Fprintf(r.cmd.OutOrStdout(), "%.2fs\n", latency.Seconds())
		}
	}

	// Calculate statistics
	if len(latencies) > 0 {
		var total time.Duration
		for _, lat := range latencies {
			total += lat
		}
		avg := total / time.Duration(len(latencies))

		fmt.Fprintf(r.cmd.OutOrStdout(), `
Benchmark Results:
  Iterations:    %d
  Successful:    %d
  Errors:        %d
  Average Time:  %.2fs
  Success Rate:  %.1f%%
`, iterations, len(latencies), errors, avg.Seconds(), float64(len(latencies))/float64(iterations)*100)
	}
}

// listModels lists available models for the current provider
func (r *REPLSession) listModels() {
	fmt.Fprintln(r.cmd.OutOrStdout(), "Commonly supported models:")

	models := map[string][]string{
		"openai": {
			"gpt-4", "gpt-4-turbo", "gpt-4-turbo-preview",
			"gpt-3.5-turbo", "gpt-3.5-turbo-0125",
		},
		"anthropic": {
			"claude-3-opus-20240229", "claude-3-sonnet-20240229",
			"claude-3-haiku-20240307", "claude-2.1", "claude-instant-1.2",
		},
	}

	providerName := strings.Split(r.provider, ":")[0]
	if providerModels, exists := models[providerName]; exists {
		for _, model := range providerModels {
			fmt.Fprintf(r.cmd.OutOrStdout(), "  %s:%s\n", providerName, model)
		}
	} else {
		fmt.Fprintf(r.cmd.OutOrStdout(), "  No predefined models for provider: %s\n", providerName)
	}
}
