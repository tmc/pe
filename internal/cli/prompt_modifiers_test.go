//go:build ignore

package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyModifiers(t *testing.T) {
	tests := []struct {
		name      string
		prompt    string
		modifiers []Modifier
		want      string
		wantErr   bool
	}{
		{
			name:      "no modifiers",
			prompt:    "Hello world",
			modifiers: []Modifier{},
			want:      "Hello world",
			wantErr:   false,
		},
		{
			name:   "temperature modifier",
			prompt: "Generate text",
			modifiers: []Modifier{
				{Type: "temperature", Value: 0.7},
			},
			want:    "Generate text",
			wantErr: false,
		},
		{
			name:   "style modifier",
			prompt: "Write a story",
			modifiers: []Modifier{
				{Type: "style", Value: "formal"},
			},
			want:    "Write a story in a formal style",
			wantErr: false,
		},
		{
			name:   "length modifier",
			prompt: "Summarize this",
			modifiers: []Modifier{
				{Type: "length", Value: "brief"},
			},
			want:    "Summarize this (keep it brief)",
			wantErr: false,
		},
		{
			name:   "multiple modifiers",
			prompt: "Write content",
			modifiers: []Modifier{
				{Type: "style", Value: "casual"},
				{Type: "length", Value: "detailed"},
			},
			want:    "Write content in a casual style (keep it detailed)",
			wantErr: false,
		},
		{
			name:   "tone modifier",
			prompt: "Explain this",
			modifiers: []Modifier{
				{Type: "tone", Value: "friendly"},
			},
			want:    "Explain this (use a friendly tone)",
			wantErr: false,
		},
		{
			name:   "format modifier",
			prompt: "List items",
			modifiers: []Modifier{
				{Type: "format", Value: "bullet points"},
			},
			want:    "List items\n\nFormat the response as: bullet points",
			wantErr: false,
		},
		{
			name:   "audience modifier",
			prompt: "Explain quantum physics",
			modifiers: []Modifier{
				{Type: "audience", Value: "5-year-old"},
			},
			want:    "Explain quantum physics (explain it for a 5-year-old)",
			wantErr: false,
		},
		{
			name:   "constraint modifier",
			prompt: "Generate ideas",
			modifiers: []Modifier{
				{Type: "constraint", Value: "must be eco-friendly"},
			},
			want:    "Generate ideas\n\nConstraint: must be eco-friendly",
			wantErr: false,
		},
		{
			name:   "example modifier",
			prompt: "Write a haiku",
			modifiers: []Modifier{
				{Type: "example", Value: "Cherry blossoms fall\nPetals dance on spring breeze\nBeauty ephemeral"},
			},
			want:    "Write a haiku\n\nExample:\nCherry blossoms fall\nPetals dance on spring breeze\nBeauty ephemeral",
			wantErr: false,
		},
		{
			name:   "invalid modifier type",
			prompt: "Test prompt",
			modifiers: []Modifier{
				{Type: "invalid", Value: "test"},
			},
			want:    "Test prompt",
			wantErr: true,
		},
		{
			name:   "complex combination",
			prompt: "Create a presentation",
			modifiers: []Modifier{
				{Type: "style", Value: "professional"},
				{Type: "length", Value: "10 slides"},
				{Type: "audience", Value: "executives"},
				{Type: "format", Value: "slide titles and key points"},
			},
			want:    "Create a presentation in a professional style (keep it 10 slides) (explain it for an executives)\n\nFormat the response as: slide titles and key points",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyModifiers(tt.prompt, tt.modifiers)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestValidateModifier(t *testing.T) {
	tests := []struct {
		name     string
		modifier Modifier
		wantErr  bool
	}{
		{
			name:     "valid temperature",
			modifier: Modifier{Type: "temperature", Value: 0.5},
			wantErr:  false,
		},
		{
			name:     "invalid temperature - too high",
			modifier: Modifier{Type: "temperature", Value: 2.5},
			wantErr:  true,
		},
		{
			name:     "invalid temperature - negative",
			modifier: Modifier{Type: "temperature", Value: -0.5},
			wantErr:  true,
		},
		{
			name:     "valid style",
			modifier: Modifier{Type: "style", Value: "formal"},
			wantErr:  false,
		},
		{
			name:     "valid length",
			modifier: Modifier{Type: "length", Value: "brief"},
			wantErr:  false,
		},
		{
			name:     "valid tone",
			modifier: Modifier{Type: "tone", Value: "professional"},
			wantErr:  false,
		},
		{
			name:     "valid format",
			modifier: Modifier{Type: "format", Value: "markdown"},
			wantErr:  false,
		},
		{
			name:     "valid audience",
			modifier: Modifier{Type: "audience", Value: "developers"},
			wantErr:  false,
		},
		{
			name:     "valid constraint",
			modifier: Modifier{Type: "constraint", Value: "no technical jargon"},
			wantErr:  false,
		},
		{
			name:     "valid example",
			modifier: Modifier{Type: "example", Value: "Example text"},
			wantErr:  false,
		},
		{
			name:     "invalid type",
			modifier: Modifier{Type: "unknown", Value: "value"},
			wantErr:  true,
		},
		{
			name:     "empty type",
			modifier: Modifier{Type: "", Value: "value"},
			wantErr:  true,
		},
		{
			name:     "empty value for non-temperature",
			modifier: Modifier{Type: "style", Value: ""},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModifier(tt.modifier)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseModifiers(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []Modifier
		wantErr bool
	}{
		{
			name:  "single modifier",
			input: []string{"temperature:0.7"},
			want: []Modifier{
				{Type: "temperature", Value: 0.7},
			},
			wantErr: false,
		},
		{
			name:  "multiple modifiers",
			input: []string{"style:formal", "length:brief"},
			want: []Modifier{
				{Type: "style", Value: "formal"},
				{Type: "length", Value: "brief"},
			},
			wantErr: false,
		},
		{
			name:  "modifier with spaces",
			input: []string{"constraint:must be family friendly"},
			want: []Modifier{
				{Type: "constraint", Value: "must be family friendly"},
			},
			wantErr: false,
		},
		{
			name:    "invalid format - no colon",
			input:   []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "invalid format - empty type",
			input:   []string{":value"},
			wantErr: true,
		},
		{
			name:    "invalid temperature value",
			input:   []string{"temperature:not-a-number"},
			wantErr: true,
		},
		{
			name:  "mixed valid and quotes",
			input: []string{"style:\"very formal\"", "example:This is an example"},
			want: []Modifier{
				{Type: "style", Value: "very formal"},
				{Type: "example", Value: "This is an example"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseModifiers(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestModifierPriority(t *testing.T) {
	// Test that modifiers are applied in the correct order
	prompt := "Base prompt"
	modifiers := []Modifier{
		{Type: "format", Value: "json"},      // Should be last
		{Type: "style", Value: "technical"},  // Should be first
		{Type: "constraint", Value: "brief"}, // Should be middle
	}

	result, err := ApplyModifiers(prompt, modifiers)
	assert.NoError(t, err)

	// Check order: style first, then constraint, then format
	assert.Contains(t, result, "technical style")
	assert.Contains(t, result, "Constraint: brief")
	assert.Contains(t, result, "Format the response as: json")

	// Ensure format instruction comes after constraint
	constraintIdx := strings.Index(result, "Constraint:")
	formatIdx := strings.Index(result, "Format the response as:")
	assert.True(t, formatIdx > constraintIdx, "Format modifier should come after constraint")
}
