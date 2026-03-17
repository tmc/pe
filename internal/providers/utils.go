package providers

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Type string
	Data string
	ID   string
}

// SSEScanner handles parsing of Server-Sent Events
type SSEScanner struct {
	scanner *bufio.Scanner
	event   SSEEvent
	err     error
}

// NewSSEScanner creates a new SSE scanner
func NewSSEScanner(r io.Reader) *SSEScanner {
	return &SSEScanner{
		scanner: bufio.NewScanner(r),
	}
}

// Scan reads the next SSE event
func (s *SSEScanner) Scan() bool {
	s.event = SSEEvent{}

	var lines []string
	for s.scanner.Scan() {
		line := s.scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			break
		}

		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return false
	}

	// Parse the event
	for _, line := range lines {
		if strings.HasPrefix(line, "event: ") {
			s.event.Type = strings.TrimSpace(strings.TrimPrefix(line, "event: "))
		} else if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if s.event.Data != "" {
				s.event.Data += "\n"
			}
			s.event.Data += data
		} else if strings.HasPrefix(line, "id: ") {
			s.event.ID = strings.TrimSpace(strings.TrimPrefix(line, "id: "))
		}
	}

	return true
}

// Event returns the current event
func (s *SSEScanner) Event() SSEEvent {
	return s.event
}

// Err returns any scanning error
func (s *SSEScanner) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.scanner.Err()
}

// Helper functions for parsing options

func getStringOption(options map[string]interface{}, key, defaultValue string) string {
	if value, exists := options[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getIntOption(options map[string]interface{}, key string, defaultValue int) int {
	if value, exists := options[key]; exists {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

func getFloat64Option(options map[string]interface{}, key string, defaultValue float64) float64 {
	if value, exists := options[key]; exists {
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return defaultValue
}

func getBoolOption(options map[string]interface{}, key string, defaultValue bool) bool {
	if value, exists := options[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
		if str, ok := value.(string); ok {
			if b, err := strconv.ParseBool(str); err == nil {
				return b
			}
		}
	}
	return defaultValue
}

func getStringSliceOption(options map[string]interface{}, key string) []string {
	if options == nil {
		return nil
	}
	value, exists := options[key]
	if !exists {
		return nil
	}

	switch v := value.(type) {
	case []string:
		return append([]string(nil), v...)
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			result = append(result, fmt.Sprintf("%v", item))
		}
		return result
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		return []string{v}
	default:
		return []string{fmt.Sprintf("%v", v)}
	}
}

func getEnvVar(key string) string {
	return os.Getenv(key)
}

// ParseProviderString parses a provider string like "openai:gpt-4" or "anthropic:claude-3-haiku"
func ParseProviderString(provider string) (providerName, model string, err error) {
	parts := strings.SplitN(provider, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid provider format, expected 'provider:model', got '%s'", provider)
	}

	providerName = strings.TrimSpace(parts[0])
	model = strings.TrimSpace(parts[1])

	if providerName == "" || model == "" {
		return "", "", fmt.Errorf("invalid provider format, both provider and model must be specified")
	}

	return providerName, model, nil
}

// NormalizeModelName normalizes model names to handle common variations
func NormalizeModelName(provider, model string) string {
	switch provider {
	case "openai":
		// Handle common OpenAI model aliases
		switch model {
		case "gpt4", "gpt-4-turbo":
			return "gpt-4-turbo-preview"
		case "gpt3.5", "gpt35":
			return "gpt-3.5-turbo"
		case "gpt-4-latest":
			return "gpt-4-turbo-preview"
		}

	case "anthropic":
		// Handle common Anthropic model aliases
		switch model {
		case "claude-3-opus":
			return "claude-3-opus-20240229"
		case "claude-3-sonnet":
			return "claude-3-sonnet-20240229"
		case "claude-3-haiku":
			return "claude-3-haiku-20240307"
		case "claude-2":
			return "claude-2.1"
		case "claude-instant":
			return "claude-instant-1.2"
		}
	}

	return model
}

// GetProviderDisplayName returns a human-readable provider name
func GetProviderDisplayName(provider, model string) string {
	switch provider {
	case "openai":
		return fmt.Sprintf("OpenAI %s", model)
	case "anthropic":
		return fmt.Sprintf("Anthropic %s", model)
	case "google":
		return fmt.Sprintf("Google %s", model)
	default:
		return fmt.Sprintf("%s:%s", provider, model)
	}
}

// EstimateTokens provides a rough estimate of token count for text
// This is a simple approximation and may not be accurate for all models
func EstimateTokens(text string) int {
	// Very rough estimate: ~4 characters per token for English text
	return len(text) / 4
}

// ValidateTemperature ensures temperature is in valid range
func ValidateTemperature(temp float64) error {
	if temp < 0.0 || temp > 2.0 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0, got %.2f", temp)
	}
	return nil
}

// ValidateMaxTokens ensures max tokens is positive
func ValidateMaxTokens(maxTokens int) error {
	if maxTokens <= 0 {
		return fmt.Errorf("max tokens must be positive, got %d", maxTokens)
	}
	return nil
}
