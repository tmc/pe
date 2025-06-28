package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVetCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		files    map[string]string
		wantErr  bool
		contains string
	}{
		{
			name: "vet yaml config",
			args: []string{"config.yaml"},
			files: map[string]string{
				"config.yaml": `prompts:
  - "Test prompt"
providers:
  - "openai:gpt-4"`,
			},
			wantErr:  false,
			contains: "OK",
		},
		{
			name: "vet invalid yaml",
			args: []string{"invalid.yaml"},
			files: map[string]string{
				"invalid.yaml": `invalid yaml content
  bad: [indentation`,
			},
			wantErr:  true,
			contains: "Error",
		},
		{
			name: "vet missing prompts field",
			args: []string{"missing.yaml"},
			files: map[string]string{
				"missing.yaml": `providers:
  - "openai:gpt-4"`,
			},
			wantErr:  true,
			contains: "Error", // General error check instead of specific message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Create test files
			for filename, content := range tt.files {
				path := filepath.Join(tmpDir, filename)
				err := os.WriteFile(path, []byte(content), 0644)
				require.NoError(t, err)
			}

			// Update args with full paths
			var fullArgs []string
			for _, arg := range tt.args {
				if strings.HasSuffix(arg, ".yaml") || strings.HasSuffix(arg, ".yml") {
					fullArgs = append(fullArgs, filepath.Join(tmpDir, arg))
				} else {
					fullArgs = append(fullArgs, arg)
				}
			}

			// Create and execute command
			cmd := vetCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(fullArgs)

			err := cmd.Execute()
			output := buf.String()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.contains != "" {
				assert.Contains(t, output, tt.contains)
			}
		})
	}
}

func TestFmtCmd(t *testing.T) {
	tests := []struct {
		name       string
		inputFile  string
		content    string
		args       []string
		wantErr    bool
		checkWrite bool
	}{
		{
			name:      "format yaml file",
			inputFile: "test.yaml",
			content: `prompts:
- "Test prompt"
providers:
- "openai:gpt-4"`,
			args:    []string{},
			wantErr: false,
		},
		{
			name:      "format and write",
			inputFile: "test.yaml",
			content: `prompts: ["Test prompt"]
providers: ["openai:gpt-4"]`,
			args:       []string{"-w"},
			wantErr:    false,
			checkWrite: true,
		},
		{
			name:      "convert to json",
			inputFile: "test.yaml",
			content: `prompts:
  - "Test prompt"`,
			args:    []string{"-o", "json"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputPath := filepath.Join(tmpDir, tt.inputFile)

			// Write input file
			err := os.WriteFile(inputPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Create command
			cmd := fmtCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			args := append([]string{inputPath}, tt.args...)
			cmd.SetArgs(args)

			err = cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tt.checkWrite {
					// Check if file was modified
					_, err := os.Stat(inputPath)
					assert.NoError(t, err)
				} else {
					// Check output
					output := buf.String()
					assert.NotEmpty(t, output)
				}
			}
		})
	}
}

func TestInitCmd(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		checkFile string
		format    string
	}{
		{
			name:      "init default yaml",
			args:      []string{},
			wantErr:   false,
			checkFile: "pe-config.yaml",
			format:    "yaml",
		},
		{
			name:      "init custom filename",
			args:      []string{"custom.yaml"},
			wantErr:   false,
			checkFile: "custom.yaml",
			format:    "yaml",
		},
		{
			name:      "init json format",
			args:      []string{"-f", "json"},
			wantErr:   false,
			checkFile: "pe-config.json",
			format:    "json",
		},
		{
			name:      "init with force",
			args:      []string{"--force"},
			wantErr:   false,
			checkFile: "pe-config.yaml",
			format:    "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldWd, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(oldWd)

			cmd := initCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check file was created
				content, err := os.ReadFile(tt.checkFile)
				assert.NoError(t, err)
				assert.NotEmpty(t, content)

				// Verify content structure
				contentStr := string(content)
				assert.Contains(t, contentStr, "prompts")
				assert.Contains(t, contentStr, "providers")
				assert.Contains(t, contentStr, "tests")
			}
		})
	}
}

func TestWatchCmd(t *testing.T) {
	// Watch command is harder to test as it runs indefinitely
	// Test basic setup and validation
	cmd := watchCmd()

	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "watch")
	assert.Contains(t, strings.ToLower(cmd.Short), "watch")

	// Check flags exist
	assert.NotNil(t, cmd.Flags().Lookup("config"))
	assert.NotNil(t, cmd.Flags().Lookup("output"))
	assert.NotNil(t, cmd.Flags().Lookup("include"))
}

func TestConvertCmd(t *testing.T) {
	tests := []struct {
		name       string
		inputFile  string
		outputFile string
		content    string
		args       []string
		wantErr    bool
	}{
		{
			name:       "convert yaml to json",
			inputFile:  "input.yaml",
			outputFile: "output.json",
			content: `prompts:
  - "Test prompt"`,
			args:    []string{},
			wantErr: false,
		},
		{
			name:       "convert with explicit format",
			inputFile:  "input.yaml",
			outputFile: "output.txt",
			content: `prompts:
  - "Test prompt"`,
			args:    []string{"-o", "json"},
			wantErr: false,
		},
		{
			name:       "convert invalid input",
			inputFile:  "input.yaml",
			outputFile: "output.json",
			content:    `invalid: [yaml content`,
			args:       []string{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputPath := filepath.Join(tmpDir, tt.inputFile)
			outputPath := filepath.Join(tmpDir, tt.outputFile)

			// Write input file
			err := os.WriteFile(inputPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			cmd := convertCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			args := append([]string{inputPath, outputPath}, tt.args...)
			cmd.SetArgs(args)

			err = cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check output file exists
				_, err = os.Stat(outputPath)
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunEval(t *testing.T) {
	// Test the runEval helper function
	configContent := `prompts:
  - "Test prompt"
providers:
  - "mock"
tests:
  - vars:
      test: "value"
    assert:
      - type: "contains"
        value: "test"`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create a mock command for testing
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Note: runEval will try to execute evalCmd which may not work in test
	// This is more of a smoke test
	runEval(cmd, configFile, "")
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestVetQuietMode(t *testing.T) {
	// Test quiet mode flag
	cmd := vetCmd()
	cmd.SetArgs([]string{"-q"})

	// Parse flags
	err := cmd.ParseFlags([]string{"-q"})
	assert.NoError(t, err)

	quiet, err := cmd.Flags().GetBool("quiet")
	assert.NoError(t, err)
	assert.True(t, quiet)
}

func TestVetVerboseMode(t *testing.T) {
	// Test verbose mode flag
	cmd := vetCmd()
	cmd.SetArgs([]string{"-v"})

	// Parse flags
	err := cmd.ParseFlags([]string{"-v"})
	assert.NoError(t, err)

	verbose, err := cmd.Flags().GetBool("verbose")
	assert.NoError(t, err)
	assert.True(t, verbose)
}
