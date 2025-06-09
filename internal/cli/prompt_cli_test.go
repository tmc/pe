//go:build ignore

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPromptCLI(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		wantErr bool
		check   func(t *testing.T, cli *PromptCLI)
	}{
		{
			name:    "basic prompt",
			prompt:  `Write a {{style}} poem about {{topic}}`,
			wantErr: false,
			check: func(t *testing.T, cli *PromptCLI) {
				assert.Equal(t, "Write a {{style}} poem about {{topic}}", cli.Prompt)
				assert.Len(t, cli.Variables, 2)
				assert.Equal(t, "style", cli.Variables[0].Name)
				assert.Equal(t, "topic", cli.Variables[1].Name)
			},
		},
		{
			name: "prompt with metadata",
			prompt: `---
name: poem-writer
description: A tool to write poems
version: 1.0.0
defaults:
  style: haiku
---
Write a {{style}} poem about {{topic}}`,
			wantErr: false,
			check: func(t *testing.T, cli *PromptCLI) {
				assert.Equal(t, "poem-writer", cli.Metadata.Name)
				assert.Equal(t, "A tool to write poems", cli.Metadata.Description)
				assert.Equal(t, "1.0.0", cli.Metadata.Version)
				assert.Equal(t, "haiku", cli.Metadata.Defaults["style"])
			},
		},
		{
			name: "prompt with typed variables",
			prompt: `---
variables:
  - name: count
    type: int
    default: 5
  - name: verbose
    type: bool
    default: false
---
Generate {{count}} items{{if verbose}} with details{{end}}`,
			wantErr: false,
			check: func(t *testing.T, cli *PromptCLI) {
				assert.Len(t, cli.Variables, 2)
				assert.Equal(t, "int", cli.Variables[0].Type)
				assert.Equal(t, float64(5), cli.Variables[0].Default)
				assert.Equal(t, "bool", cli.Variables[1].Type)
				assert.Equal(t, false, cli.Variables[1].Default)
			},
		},
		{
			name: "prompt with output variables",
			prompt: `---
output_vars:
  - name: summary
    pattern: 'Summary: (.+)'
  - name: keywords
    pattern: 'Keywords: (.+)'
    type: list
---
Analyze {{text}} and provide a summary and keywords`,
			wantErr: false,
			check: func(t *testing.T, cli *PromptCLI) {
				assert.Len(t, cli.OutputVars, 2)
				assert.Equal(t, "summary", cli.OutputVars[0].Name)
				assert.Equal(t, "keywords", cli.OutputVars[1].Name)
				assert.Equal(t, "list", cli.OutputVars[1].Type)
			},
		},
		{
			name: "prompt with examples",
			prompt: `---
examples:
  - name: "Write a haiku about nature"
    flags:
      style: haiku
      topic: nature
  - name: "Write a sonnet about love"
    flags:
      style: sonnet
      topic: love
---
Write a {{style}} poem about {{topic}}`,
			wantErr: false,
			check: func(t *testing.T, cli *PromptCLI) {
				assert.Len(t, cli.Metadata.Examples, 2)
				assert.Equal(t, "Write a haiku about nature", cli.Metadata.Examples[0].Name)
				assert.Equal(t, "haiku", cli.Metadata.Examples[0].Flags["style"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli, err := NewPromptCLI(tt.prompt)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cli)
				if tt.check != nil {
					tt.check(t, cli)
				}
			}
		})
	}
}

func TestPromptCLI_ToCommand(t *testing.T) {
	prompt := `---
name: test-tool
description: A test tool
defaults:
  name: World
---
Hello {{name}}!`

	cli, err := NewPromptCLI(prompt)
	require.NoError(t, err)

	cmd := cli.ToCommand()
	assert.NotNil(t, cmd)
	assert.Equal(t, "test-tool", cmd.Use)
	assert.Equal(t, "A test tool", cmd.Short)

	// Check that flags were added
	nameFlag := cmd.Flags().Lookup("name")
	assert.NotNil(t, nameFlag)
	assert.Equal(t, "World", nameFlag.DefValue)
}

func TestPromptCLI_Execute(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		args    []string
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "simple execution",
			prompt:  `Hello {{.name}}!`,
			args:    []string{"--name", "Test"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "Hello Test!")
			},
		},
		{
			name: "execution with defaults",
			prompt: `---
defaults:
  greeting: Hi
---
{{.greeting}} {{.name}}!`,
			args:    []string{"--name", "User"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "Hi User!")
			},
		},
		{
			name:    "execution with file input",
			prompt:  `Process file: {{.input}}`,
			args:    []string{"--input", "test.txt"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "Process file:")
			},
		},
		{
			name: "execution with output format",
			prompt: `---
output_format: json
---
{"message": "Hello {{.name}}"}`,
			args:    []string{"--name", "World"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, `"message": "Hello World"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file for file input test
			if strings.Contains(tt.prompt, "file:input") {
				tmpDir := t.TempDir()
				testFile := filepath.Join(tmpDir, "test.txt")
				err := os.WriteFile(testFile, []byte("test content"), 0644)
				require.NoError(t, err)
				// Update args with full path
				for i, arg := range tt.args {
					if arg == "test.txt" {
						tt.args[i] = testFile
					}
				}
			}

			cli, err := NewPromptCLI(tt.prompt)
			require.NoError(t, err)

			cmd := cli.ToCommand()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			// Override the RunE function to capture output
			originalRunE := cmd.RunE
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				// Execute the template without calling the provider
				vars := make(map[string]interface{})
				cmd.Flags().VisitAll(func(f *pflag.Flag) {
					vars[f.Name] = f.Value.String()
				})

				// Simple Go template execution
				tmpl, err := template.New("test").Parse(cli.Prompt)
				if err != nil {
					return err
				}
				return tmpl.Execute(&buf, vars)
			}
			defer func() { cmd.RunE = originalRunE }()

			err = cmd.Execute()
			output := buf.String()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.check != nil {
					tt.check(t, output)
				}
			}
		})
	}
}

func TestParseMetadata(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantMeta PromptMetadata
		wantBody string
		wantErr  bool
	}{
		{
			name: "basic metadata",
			content: `---
name: test
description: Test tool
---
Body content`,
			wantMeta: PromptMetadata{
				Name:        "test",
				Description: "Test tool",
			},
			wantBody: "Body content",
			wantErr:  false,
		},
		{
			name: "full metadata",
			content: `---
name: full-test
description: Full test tool
version: 1.2.3
author: Test Author
model: gpt-4
temperature: 0.7
max_tokens: 1000
system_prompt: "You are a helpful assistant"
prefill: "Here is my response:"
---
Template body`,
			wantMeta: PromptMetadata{
				Name:         "full-test",
				Description:  "Full test tool",
				Version:      "1.2.3",
				Author:       "Test Author",
				Model:        "gpt-4",
				Temperature:  0.7,
				MaxTokens:    1000,
				SystemPrompt: "You are a helpful assistant",
				Prefill:      "Here is my response:",
			},
			wantBody: "Template body",
			wantErr:  false,
		},
		{
			name:     "no metadata",
			content:  "Just a template",
			wantMeta: PromptMetadata{},
			wantBody: "Just a template",
			wantErr:  false,
		},
		{
			name: "invalid yaml",
			content: `---
name: test
invalid yaml [here
---
Body`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, body, err := parseMetadata(tt.content)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMeta, meta)
				assert.Equal(t, tt.wantBody, body)
			}
		})
	}
}

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "simple variables",
			template: "Hello {{.name}}, welcome to {{.place}}!",
			want:     []string{"name", "place"},
		},
		{
			name:     "duplicate variables",
			template: "{{.name}} is {{.name}}'s name",
			want:     []string{"name"},
		},
		{
			name:     "no variables",
			template: "Hello world!",
			want:     []string{},
		},
		{
			name:     "complex template",
			template: "{{.greeting}} {{.name}}! Today is {{.day}} and the weather is {{.weather}}.",
			want:     []string{"greeting", "name", "day", "weather"},
		},
		{
			name:     "mixed case variables",
			template: "{{.firstName}} {{.LastName}} {{.user_id}}",
			want:     []string{"firstName", "LastName", "user_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := extractVariables(tt.template)
			got := make([]string, len(vars))
			for i, v := range vars {
				got[i] = v.Name
			}
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestValidateOutputFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"empty format", "", false},
		{"json format", "json", false},
		{"yaml format", "yaml", false},
		{"markdown format", "markdown", false},
		{"csv format", "csv", false},
		{"tsv format", "tsv", false},
		{"xml format", "xml", false},
		{"invalid format", "invalid", true},
		{"mixed case", "JSON", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOutputFormat(tt.format)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildExamplesHelp(t *testing.T) {
	examples := []Example{
		{
			Name: "Simple example",
			Flags: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
		},
		{
			Name: "Complex example",
			Flags: map[string]interface{}{
				"file":    "input.txt",
				"verbose": true,
			},
		},
	}

	help := buildExamplesHelp("test-tool", examples)
	assert.Contains(t, help, "Examples:")
	assert.Contains(t, help, "Simple example")
	assert.Contains(t, help, "test-tool --name=\"John\" --age=\"30\"")
	assert.Contains(t, help, "Complex example")
	assert.Contains(t, help, "--file=\"input.txt\" --verbose=\"true\"")
}

func TestAddFlagsToCommand(t *testing.T) {
	cli := &PromptCLI{
		Variables: []Variable{
			{Name: "name", Type: "string", Description: "User name", Default: "World"},
			{Name: "count", Type: "int", Description: "Item count", Default: 5},
			{Name: "verbose", Type: "bool", Description: "Verbose output", Default: false},
			{Name: "rate", Type: "float", Description: "Rate value", Default: 0.5},
			{Name: "config", Type: "file", Description: "Config file"},
		},
	}

	cmd := &cobra.Command{}
	cli.addFlagsToCommand(cmd)

	// Check string flag
	nameFlag := cmd.Flags().Lookup("name")
	assert.NotNil(t, nameFlag)
	assert.Equal(t, "World", nameFlag.DefValue)
	assert.Equal(t, "User name", nameFlag.Usage)

	// Check int flag
	countFlag := cmd.Flags().Lookup("count")
	assert.NotNil(t, countFlag)
	assert.Equal(t, "5", countFlag.DefValue)

	// Check bool flag
	verboseFlag := cmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)

	// Check float flag
	rateFlag := cmd.Flags().Lookup("rate")
	assert.NotNil(t, rateFlag)
	assert.Equal(t, "0.5", rateFlag.DefValue)

	// Check file flag
	configFlag := cmd.Flags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "", configFlag.DefValue)
}
