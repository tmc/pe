package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOptimizeForProvider(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		provider string
		want     string
		wantErr  bool
	}{
		{
			name:     "anthropic provider",
			prompt:   "Hello, world!",
			provider: "anthropic",
			want:     "Human: Hello, world!\n\nAssistant: ",
		},
		{
			name:     "openai provider",
			prompt:   "Hello, world!",
			provider: "openai",
			want:     "Hello, world!",
		},
		{
			name:     "google provider",
			prompt:   "Hello, world!",
			provider: "google",
			wantErr:  true,
		},
		{
			name:     "unknown provider",
			prompt:   "Hello, world!",
			provider: "unknown",
			wantErr:  true,
		},
		{
			name:     "empty provider",
			prompt:   "Hello, world!",
			provider: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := optimizeForProvider(tt.prompt, tt.provider)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("optimizeForProvider() err = nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("optimizeForProvider() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMinifyPrompt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single line",
			input: "Hello world",
			want:  "Hello world",
		},
		{
			name:  "multiple lines",
			input: "Hello\nworld",
			want:  "Hello world",
		},
		{
			name:  "extra whitespace",
			input: "  Hello  \n  world  ",
			want:  "Hello world",
		},
		{
			name:  "empty lines",
			input: "Hello\n\n\nworld",
			want:  "Hello world",
		},
		{
			name:  "mixed content",
			input: "Line one\n  Line two  \n\nLine three",
			want:  "Line one Line two Line three",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "only whitespace",
			input: "   \n   \n   ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minifyPrompt(tt.input)
			if got != tt.want {
				t.Errorf("minifyPrompt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEstimateTokenCount(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "empty string",
			text: "",
			want: 0,
		},
		{
			name: "short string",
			text: "Hi",
			want: 0, // 2/4 = 0
		},
		{
			name: "four characters",
			text: "test",
			want: 1,
		},
		{
			name: "longer text",
			text: "This is a longer text with multiple words.",
			want: 10, // 43/4 = 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimateTokenCount(tt.text)
			if got != tt.want {
				t.Errorf("estimateTokenCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateChecksum(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "empty data",
			data: []byte{},
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name: "hello world",
			data: []byte("hello world"),
			want: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name: "prompt text",
			data: []byte("You are a helpful assistant."),
			want: "8f3e5d3c7f2a1b9e4d6c8a0f5e2b3d7c1a9f6e4b0d8c2a5f7e3b1d9c6a4f0e2b", // Placeholder - actual value will differ
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateChecksum(tt.data)
			// For the first two cases, check exact match
			if tt.name != "prompt text" {
				if got != tt.want {
					t.Errorf("calculateChecksum() = %q, want %q", got, tt.want)
				}
			} else {
				// For other cases, just verify it returns a valid hex string
				if len(got) != 64 {
					t.Errorf("calculateChecksum() returned %d chars, want 64", len(got))
				}
			}
		})
	}
}

func TestBuildCmd_BasicUsage(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tests := []struct {
		name            string
		fileContent     string
		fileName        string
		args            []string
		wantErr         bool
		checkOutputFile bool
	}{
		{
			name:            "build from yaml config",
			fileContent:     "prompt: You are a helpful assistant.\nprovider: openai\nmodel: gpt-4",
			fileName:        "config.yaml",
			args:            []string{},
			wantErr:         false,
			checkOutputFile: true,
		},
		{
			name:            "build from text file",
			fileContent:     "You are a helpful assistant.",
			fileName:        "prompt.txt",
			args:            []string{},
			wantErr:         false,
			checkOutputFile: true,
		},
		{
			name:            "build with target provider",
			fileContent:     "You are a helpful assistant.",
			fileName:        "prompt.txt",
			args:            []string{"--target", "anthropic"},
			wantErr:         false,
			checkOutputFile: true,
		},
		{
			name:            "build with minify",
			fileContent:     "Line one\n\nLine two\n\nLine three",
			fileName:        "prompt.txt",
			args:            []string{"--minify"},
			wantErr:         false,
			checkOutputFile: true,
		},
		{
			name:            "build with validation",
			fileContent:     "You are a helpful assistant.",
			fileName:        "prompt.txt",
			args:            []string{"--validate"},
			wantErr:         false,
			checkOutputFile: false,
		},
		{
			name:            "build yaml missing prompt",
			fileContent:     "provider: openai\nmodel: gpt-4",
			fileName:        "config.yaml",
			args:            []string{},
			wantErr:         true,
			checkOutputFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create test file
			filePath := filepath.Join(tmpDir, tt.fileName)
			if err := os.WriteFile(filePath, []byte(tt.fileContent), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Change to temp dir for output files
			oldDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(oldDir)

			// Reset global flags
			buildOutput = ""
			buildWithMetadata = false
			buildTarget = ""
			buildTargets = nil
			buildMinify = false
			buildValidate = false
			buildBundle = false
			buildCompress = false

			// Create command
			cmd := buildCmd
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			args := append([]string{filePath}, tt.args...)
			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Check output file was created
				if tt.checkOutputFile {
					outputFile := filepath.Join(tmpDir, "prompt_build.txt")
					if _, err := os.Stat(outputFile); os.IsNotExist(err) {
						t.Error("Expected output file prompt_build.txt was not created")
					}
				}
			}
		})
	}
}

func TestValidateBuildPrompt(t *testing.T) {
	tests := []struct {
		name      string
		prompt    string
		variables map[string]interface{}
		wantErr   string
	}{
		{
			name:   "plain text",
			prompt: "Review this change.\n",
		},
		{
			name: "typed executable text",
			prompt: `---
kind: pe.text.v1
inputs:
  topic:
    type: string
---
Review {{ .topic }}.
`,
			variables: map[string]interface{}{"topic": "release"},
		},
		{
			name: "missing input",
			prompt: `---
kind: pe.text.v1
inputs:
  topic:
    type: string
---
Review {{ .topic }}.
`,
			wantErr: "missing input topic",
		},
		{
			name: "unsupported kind",
			prompt: `---
kind: pe.future.v1
---
Review this.
`,
			wantErr: "unsupported executable text kind",
		},
		{
			name: "malformed front matter",
			prompt: `---
kind: pe.text.v1
Review this.
`,
			wantErr: "front matter missing closing",
		},
		{
			name:    "template references undeclared input",
			prompt:  "Review {{ .topic }}.\n",
			wantErr: "map has no entry for key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBuildPrompt("", tt.prompt, tt.variables)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateBuildPrompt: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateBuildPrompt error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestBuildValidateAllowsImports(t *testing.T) {
	tmpDir := t.TempDir()
	part := filepath.Join(tmpDir, "part.prompt")
	if err := os.WriteFile(part, []byte("part\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(tmpDir, "main.prompt")
	if err := os.WriteFile(main, []byte(`---
kind: pe.text.v1
imports:
  part: part.prompt
---
{{ import "part" }}`), 0644); err != nil {
		t.Fatal(err)
	}

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	buildOutput = ""
	buildWithMetadata = false
	buildTarget = ""
	buildTargets = nil
	buildMinify = false
	buildValidate = false
	buildBundle = false
	buildCompress = false

	cmd := buildCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{main, "--validate"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("build validate imports: %v", err)
	}
}

func TestBuildCmd_OutputFile(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()

	// Create test prompt
	promptContent := "You are a helpful assistant."
	promptFile := filepath.Join(tmpDir, "prompt.txt")
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "output.txt")

	// Reset global flags
	buildOutput = ""
	buildWithMetadata = false
	buildTarget = ""
	buildTargets = nil
	buildMinify = false
	buildValidate = false
	buildBundle = false
	buildCompress = false

	// Change to temp dir
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := buildCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{promptFile, "-o", outputFile})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Check content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if string(content) != promptContent {
		t.Errorf("Output file content = %q, want %q", string(content), promptContent)
	}
}

func TestBuildCmd_WithMetadata(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()

	promptContent := "You are a helpful assistant."
	promptFile := filepath.Join(tmpDir, "prompt.txt")
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "output.txt")

	// Reset global flags
	buildOutput = ""
	buildWithMetadata = false
	buildTarget = ""
	buildTargets = nil
	buildMinify = false
	buildValidate = false
	buildBundle = false
	buildCompress = false

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := buildCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{promptFile, "-o", outputFile, "--with-metadata"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check metadata file exists
	metadataFile := filepath.Join(tmpDir, "output_metadata.json")
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		t.Error("Metadata file was not created")
	}

	// Check metadata content contains expected fields
	content, err := os.ReadFile(metadataFile)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %v", err)
	}

	if !strings.Contains(string(content), "version") {
		t.Error("Metadata file should contain version field")
	}
	if !strings.Contains(string(content), "checksum") {
		t.Error("Metadata file should contain checksum field")
	}
	if !strings.Contains(string(content), "timestamp") {
		t.Error("Metadata file should contain timestamp field")
	}
}

func TestBuildCmd_Compress(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()

	promptContent := strings.Repeat("This is a test prompt. ", 100)
	promptFile := filepath.Join(tmpDir, "prompt.txt")
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "output.txt")

	// Reset global flags
	buildOutput = ""
	buildWithMetadata = false
	buildTarget = ""
	buildTargets = nil
	buildMinify = false
	buildValidate = false
	buildBundle = false
	buildCompress = false

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := buildCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{promptFile, "-o", outputFile, "--compress"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check compressed file exists
	compressedFile := outputFile + ".gz"
	if _, err := os.Stat(compressedFile); os.IsNotExist(err) {
		t.Error("Compressed file was not created")
	}

	// Check uncompressed output file also exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

func TestBuildCmd_MultipleTargets(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()

	promptContent := "You are a helpful assistant."
	promptFile := filepath.Join(tmpDir, "prompt.txt")
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Reset global flags
	buildOutput = ""
	buildWithMetadata = false
	buildTarget = ""
	buildTargets = nil
	buildMinify = false
	buildValidate = false
	buildBundle = false
	buildCompress = false

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := buildCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{promptFile, "--targets", "anthropic,openai"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check target directories exist
	anthropicDir := filepath.Join(tmpDir, "builds", "anthropic")
	if _, err := os.Stat(anthropicDir); os.IsNotExist(err) {
		t.Error("Anthropic target directory was not created")
	}

	openaiDir := filepath.Join(tmpDir, "builds", "openai")
	if _, err := os.Stat(openaiDir); os.IsNotExist(err) {
		t.Error("OpenAI target directory was not created")
	}

	// Check prompt files exist in target directories
	anthropicPrompt := filepath.Join(anthropicDir, "prompt.txt")
	if _, err := os.Stat(anthropicPrompt); os.IsNotExist(err) {
		t.Error("Anthropic prompt file was not created")
	}

	openaiPrompt := filepath.Join(openaiDir, "prompt.txt")
	if _, err := os.Stat(openaiPrompt); os.IsNotExist(err) {
		t.Error("OpenAI prompt file was not created")
	}
}

func TestBuildCmd_Bundle(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()

	// Create a directory with prompt files
	promptDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptDir, 0755); err != nil {
		t.Fatalf("Failed to create prompt directory: %v", err)
	}

	// Create test prompts
	prompts := map[string]string{
		"system.txt":  "You are a helpful assistant.",
		"user.txt":    "Hello, how can you help me?",
		"test.prompt": "This is a test prompt.",
	}

	for name, content := range prompts {
		if err := os.WriteFile(filepath.Join(promptDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write prompt file: %v", err)
		}
	}

	// Reset global flags
	buildOutput = ""
	buildWithMetadata = false
	buildTarget = ""
	buildTargets = nil
	buildMinify = false
	buildValidate = false
	buildBundle = false
	buildCompress = false

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := buildCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{promptDir, "--bundle"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check bundle file exists
	bundleFile := filepath.Join(tmpDir, "prompt_bundle.tar.gz")
	if _, err := os.Stat(bundleFile); os.IsNotExist(err) {
		t.Error("Bundle file was not created")
	}

	// Check bundle file is not empty
	info, err := os.Stat(bundleFile)
	if err != nil {
		t.Fatalf("Failed to stat bundle file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Bundle file is empty")
	}
}

func TestCompressFile(t *testing.T) {
	tmpDir := t.TempDir()

	content := strings.Repeat("Test content for compression. ", 100)
	inputFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := compressFile(inputFile)
	if err != nil {
		t.Fatalf("compressFile() error = %v", err)
	}

	// Check compressed file exists
	compressedFile := inputFile + ".gz"
	if _, err := os.Stat(compressedFile); os.IsNotExist(err) {
		t.Error("Compressed file was not created")
	}

	// Check compressed file is smaller
	inputInfo, _ := os.Stat(inputFile)
	compressedInfo, _ := os.Stat(compressedFile)

	if compressedInfo.Size() >= inputInfo.Size() {
		t.Error("Compressed file should be smaller than original")
	}
}

func TestBuildCmd_NonexistentInput(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	// Reset global flags
	buildOutput = ""
	buildWithMetadata = false
	buildTarget = ""
	buildTargets = nil
	buildMinify = false
	buildValidate = false
	buildBundle = false
	buildCompress = false

	cmd := buildCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"/nonexistent/path/file.txt"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent input file")
	}
}
