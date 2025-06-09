package providers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSEScanner(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []SSEEvent
	}{
		{
			name: "single event",
			input: `event: message
data: hello world

`,
			expected: []SSEEvent{
				{Type: "message", Data: "hello world"},
			},
		},
		{
			name: "multiple events",
			input: `event: start
data: beginning

event: update
data: processing
id: 123

event: end
data: complete

`,
			expected: []SSEEvent{
				{Type: "start", Data: "beginning"},
				{Type: "update", Data: "processing", ID: "123"},
				{Type: "end", Data: "complete"},
			},
		},
		{
			name: "multiline data",
			input: `event: message
data: line 1
data: line 2
data: line 3

`,
			expected: []SSEEvent{
				{Type: "message", Data: "line 1\nline 2\nline 3"},
			},
		},
		{
			name: "data only",
			input: `data: just data

`,
			expected: []SSEEvent{
				{Data: "just data"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewSSEScanner(strings.NewReader(tt.input))

			var events []SSEEvent
			for scanner.Scan() {
				events = append(events, scanner.Event())
			}

			assert.NoError(t, scanner.Err())
			assert.Equal(t, tt.expected, events)
		})
	}
}

func TestSSEScanner_Empty(t *testing.T) {
	scanner := NewSSEScanner(bytes.NewReader([]byte("")))
	assert.False(t, scanner.Scan())
	assert.NoError(t, scanner.Err())
}

func TestGetStringOption(t *testing.T) {
	options := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	assert.Equal(t, "value1", getStringOption(options, "key1", "default"))
	assert.Equal(t, "default", getStringOption(options, "key2", "default"))
	assert.Equal(t, "default", getStringOption(options, "missing", "default"))
}

func TestGetIntOption(t *testing.T) {
	options := map[string]interface{}{
		"int":    42,
		"float":  42.5,
		"string": "42",
		"bad":    "not a number",
	}

	assert.Equal(t, 42, getIntOption(options, "int", 0))
	assert.Equal(t, 42, getIntOption(options, "float", 0))
	assert.Equal(t, 42, getIntOption(options, "string", 0))
	assert.Equal(t, 99, getIntOption(options, "bad", 99))
	assert.Equal(t, 99, getIntOption(options, "missing", 99))
}

func TestGetFloat64Option(t *testing.T) {
	options := map[string]interface{}{
		"float":  42.5,
		"int":    42,
		"string": "42.5",
		"bad":    "not a number",
	}

	assert.Equal(t, 42.5, getFloat64Option(options, "float", 0))
	assert.Equal(t, 42.0, getFloat64Option(options, "int", 0))
	assert.Equal(t, 42.5, getFloat64Option(options, "string", 0))
	assert.Equal(t, 99.9, getFloat64Option(options, "bad", 99.9))
	assert.Equal(t, 99.9, getFloat64Option(options, "missing", 99.9))
}

func TestGetBoolOption(t *testing.T) {
	options := map[string]interface{}{
		"bool":   true,
		"string": "true",
		"bad":    "not a bool",
	}

	assert.True(t, getBoolOption(options, "bool", false))
	assert.True(t, getBoolOption(options, "string", false))
	assert.False(t, getBoolOption(options, "bad", false))
	assert.True(t, getBoolOption(options, "missing", true))
}

func TestParseProviderString(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantProvider string
		wantModel    string
		wantErr      bool
	}{
		{
			name:         "valid provider string",
			input:        "openai:gpt-4",
			wantProvider: "openai",
			wantModel:    "gpt-4",
			wantErr:      false,
		},
		{
			name:         "valid with spaces",
			input:        " anthropic : claude-3-haiku ",
			wantProvider: "anthropic",
			wantModel:    "claude-3-haiku",
			wantErr:      false,
		},
		{
			name:    "missing colon",
			input:   "openai",
			wantErr: true,
		},
		{
			name:    "missing model",
			input:   "openai:",
			wantErr: true,
		},
		{
			name:    "missing provider",
			input:   ":gpt-4",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, model, err := ParseProviderString(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantProvider, provider)
				assert.Equal(t, tt.wantModel, model)
			}
		})
	}
}

func TestNormalizeModelName(t *testing.T) {
	tests := []struct {
		provider string
		model    string
		expected string
	}{
		// OpenAI normalization
		{"openai", "gpt4", "gpt-4-turbo-preview"},
		{"openai", "gpt-4-turbo", "gpt-4-turbo-preview"},
		{"openai", "gpt3.5", "gpt-3.5-turbo"},
		{"openai", "gpt35", "gpt-3.5-turbo"},
		{"openai", "gpt-4-latest", "gpt-4-turbo-preview"},
		{"openai", "gpt-4", "gpt-4"}, // unchanged

		// Anthropic normalization
		{"anthropic", "claude-3-opus", "claude-3-opus-20240229"},
		{"anthropic", "claude-3-sonnet", "claude-3-sonnet-20240229"},
		{"anthropic", "claude-3-haiku", "claude-3-haiku-20240307"},
		{"anthropic", "claude-2", "claude-2.1"},
		{"anthropic", "claude-instant", "claude-instant-1.2"},
		{"anthropic", "claude-3-opus-20240229", "claude-3-opus-20240229"}, // unchanged

		// Other providers (no normalization)
		{"google", "gemini-pro", "gemini-pro"},
		{"unknown", "model", "model"},
	}

	for _, tt := range tests {
		t.Run(tt.provider+":"+tt.model, func(t *testing.T) {
			result := NormalizeModelName(tt.provider, tt.model)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetProviderDisplayName(t *testing.T) {
	tests := []struct {
		provider string
		model    string
		expected string
	}{
		{"openai", "gpt-4", "OpenAI gpt-4"},
		{"anthropic", "claude-3-haiku", "Anthropic claude-3-haiku"},
		{"google", "gemini-pro", "Google gemini-pro"},
		{"custom", "model", "custom:model"},
	}

	for _, tt := range tests {
		t.Run(tt.provider+":"+tt.model, func(t *testing.T) {
			result := GetProviderDisplayName(tt.provider, tt.model)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"Hello world", 2}, // 11 chars / 4 = 2.75 ≈ 2
		{"", 0},            // empty string
		{"This is a longer test string with multiple words.", 12}, // 50 chars / 4 = 12.5 ≈ 12
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := EstimateTokens(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTemperature(t *testing.T) {
	tests := []struct {
		temp    float64
		wantErr bool
	}{
		{0.0, false},
		{0.5, false},
		{1.0, false},
		{2.0, false},
		{-0.1, true},
		{2.1, true},
		{-1.0, true},
		{3.0, true},
	}

	for _, tt := range tests {
		t.Run(string(rune(int(tt.temp*10))), func(t *testing.T) {
			err := ValidateTemperature(tt.temp)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMaxTokens(t *testing.T) {
	tests := []struct {
		maxTokens int
		wantErr   bool
	}{
		{1, false},
		{100, false},
		{4096, false},
		{0, true},
		{-1, true},
		{-100, true},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.maxTokens)), func(t *testing.T) {
			err := ValidateMaxTokens(tt.maxTokens)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
