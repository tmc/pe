package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewREPLSession(t *testing.T) {
	cmd := &cobra.Command{}
	session := NewREPLSession(cmd, "openai:gpt-4", "", 0.7)

	if session == nil {
		t.Fatal("NewREPLSession returned nil")
	}

	if session.provider != "openai:gpt-4" {
		t.Errorf("Expected provider 'openai:gpt-4', got %s", session.provider)
	}

	if session.temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", session.temperature)
	}

	if session.maxTokens != 4096 {
		t.Errorf("Expected maxTokens 4096, got %d", session.maxTokens)
	}

	if len(session.history) != 0 {
		t.Errorf("Expected empty history, got %d items", len(session.history))
	}
}

func TestREPLSession_getConfigDisplay(t *testing.T) {
	cmd := &cobra.Command{}

	tests := []struct {
		name       string
		configFile string
		want       string
	}{
		{
			name:       "with config file",
			configFile: "config.yaml",
			want:       "config.yaml",
		},
		{
			name:       "without config file",
			configFile: "",
			want:       "(none)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewREPLSession(cmd, "openai:gpt-4", tt.configFile, 0.7)
			got := session.getConfigDisplay()
			if got != tt.want {
				t.Errorf("getConfigDisplay() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestREPLSession_handleCommand(t *testing.T) {
	cmd := &cobra.Command{}
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	session := NewREPLSession(cmd, "openai:gpt-4", "", 0.7)

	tests := []struct {
		name      string
		input     string
		isCommand bool
	}{
		{
			name:      "help command",
			input:     ":help",
			isCommand: true,
		},
		{
			name:      "h command",
			input:     ":h",
			isCommand: true,
		},
		{
			name:      "provider command",
			input:     ":provider",
			isCommand: true,
		},
		{
			name:      "temp command",
			input:     ":temp",
			isCommand: true,
		},
		{
			name:      "tokens command",
			input:     ":tokens",
			isCommand: true,
		},
		{
			name:      "history command",
			input:     ":history",
			isCommand: true,
		},
		{
			name:      "clear command",
			input:     ":clear",
			isCommand: true,
		},
		{
			name:      "status command",
			input:     ":status",
			isCommand: true,
		},
		{
			name:      "models command",
			input:     ":models",
			isCommand: true,
		},
		{
			name:      "unknown command",
			input:     ":unknown",
			isCommand: true,
		},
		{
			name:      "not a command",
			input:     "hello world",
			isCommand: false,
		},
		{
			name:      "empty input",
			input:     "",
			isCommand: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()
			got := session.handleCommand(tt.input)
			if got != tt.isCommand {
				t.Errorf("handleCommand(%q) = %v, want %v", tt.input, got, tt.isCommand)
			}
		})
	}
}

func TestREPLSession_handleCommand_quitSetsDone(t *testing.T) {
	cmd := &cobra.Command{}
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	session := NewREPLSession(cmd, "openai:gpt-4", "", 0.7)

	if !session.handleCommand(":quit") {
		t.Fatal("handleCommand(:quit) = false, want true")
	}
	if !session.done {
		t.Fatal("session.done = false, want true")
	}
	if got := stdout.String(); !bytes.Contains([]byte(got), []byte("Goodbye!")) {
		t.Fatalf("stdout = %q, want goodbye message", got)
	}
}

func TestREPLSession_SaveDeniedByPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

capability {
    tools deny write
}
`), 0644); err != nil {
		t.Fatalf("writing pe.mod: %v", err)
	}
	if err := enforceRuntimeToolPolicy("write"); err == nil {
		t.Fatal("enforceRuntimeToolPolicy succeeded, want write policy error")
	}

	cmd := &cobra.Command{}
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	session := NewREPLSession(cmd, "openai:gpt-4", "", 0.7)
	session.history = []string{"hello"}

	if !session.handleCommand(":save session.yaml") {
		t.Fatal("handleCommand(:save) = false, want true")
	}
	combined := append(stdout.Bytes(), stderr.Bytes()...)
	if !bytes.Contains(combined, []byte("tool write is denied by pe.mod")) {
		t.Fatalf("stdout = %q, stderr = %q, want write policy error", stdout.String(), stderr.String())
	}
	if _, err := os.Stat("session.yaml"); !os.IsNotExist(err) {
		t.Fatalf("session.yaml stat error = %v, want not exist", err)
	}
}

func TestREPLSession_RunStopsOnQuit(t *testing.T) {
	cmd := &cobra.Command{}
	var stdout, stderr bytes.Buffer
	cmd.SetIn(bytes.NewBufferString(":quit\nthis should not run\n"))
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	session := NewREPLSession(cmd, "mock", "", 0.7)
	if err := session.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("Goodbye!")) {
		t.Fatalf("stdout missing goodbye:\n%s", output)
	}
	if bytes.Contains([]byte(output), []byte("this should not run")) {
		t.Fatalf("Run processed input after quit:\n%s", output)
	}
	if len(session.history) != 1 || session.history[0] != ":quit" {
		t.Fatalf("history = %#v, want only :quit", session.history)
	}
}

func TestREPLSession_handleCommand_provider(t *testing.T) {
	cmd := &cobra.Command{}
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)

	session := NewREPLSession(cmd, "openai:gpt-4", "", 0.7)

	// Test setting provider
	session.handleCommand(":provider anthropic:claude-3")

	if session.provider != "anthropic:claude-3" {
		t.Errorf("Expected provider 'anthropic:claude-3', got %s", session.provider)
	}
}

func TestREPLSession_handleCommand_temp(t *testing.T) {
	cmd := &cobra.Command{}
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)

	session := NewREPLSession(cmd, "openai:gpt-4", "", 0.7)

	// Test valid temperature
	session.handleCommand(":temp 0.9")
	if session.temperature != 0.9 {
		t.Errorf("Expected temperature 0.9, got %f", session.temperature)
	}

	// Test invalid temperature
	session.handleCommand(":temp 3.0")
	if session.temperature != 0.9 {
		t.Errorf("Temperature should not change for invalid value, got %f", session.temperature)
	}
}

func TestREPLSession_handleCommand_tokens(t *testing.T) {
	cmd := &cobra.Command{}
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)

	session := NewREPLSession(cmd, "openai:gpt-4", "", 0.7)

	// Test valid tokens
	session.handleCommand(":tokens 2048")
	if session.maxTokens != 2048 {
		t.Errorf("Expected maxTokens 2048, got %d", session.maxTokens)
	}

	// Test invalid tokens
	session.handleCommand(":tokens -100")
	if session.maxTokens != 2048 {
		t.Errorf("maxTokens should not change for invalid value, got %d", session.maxTokens)
	}
}

func TestREPLSession_printHistory(t *testing.T) {
	cmd := &cobra.Command{}
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	session := NewREPLSession(cmd, "openai:gpt-4", "", 0.7)

	// Test empty history
	session.printHistory()
	if !bytes.Contains(stdout.Bytes(), []byte("No history available")) {
		t.Error("Expected 'No history available' message")
	}

	// Add some history
	stdout.Reset()
	session.history = []string{"first", "second", "third"}
	session.printHistory()
	if !bytes.Contains(stdout.Bytes(), []byte("Command History")) {
		t.Error("Expected 'Command History' header")
	}
}

func TestREPLSession_printStatus(t *testing.T) {
	cmd := &cobra.Command{}
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	session := NewREPLSession(cmd, "openai:gpt-4", "config.yaml", 0.7)
	session.printStatus()

	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("Provider")) {
		t.Error("Expected 'Provider' in status output")
	}
	if !bytes.Contains([]byte(output), []byte("Temperature")) {
		t.Error("Expected 'Temperature' in status output")
	}
	if !bytes.Contains([]byte(output), []byte("Max Tokens")) {
		t.Error("Expected 'Max Tokens' in status output")
	}
}

func TestREPLSession_listModels(t *testing.T) {
	cmd := &cobra.Command{}
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	tests := []struct {
		name     string
		provider string
		contains string
	}{
		{
			name:     "openai provider",
			provider: "openai:gpt-4",
			contains: "gpt-4",
		},
		{
			name:     "anthropic provider",
			provider: "anthropic:claude-3",
			contains: "claude-3",
		},
		{
			name:     "unknown provider",
			provider: "unknown:model",
			contains: "No predefined models",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout.Reset()
			session := NewREPLSession(cmd, tt.provider, "", 0.7)
			session.listModels()

			if !bytes.Contains(stdout.Bytes(), []byte(tt.contains)) {
				t.Errorf("Expected output to contain %q, got %q", tt.contains, stdout.String())
			}
		})
	}
}

func TestREPLSessionStruct(t *testing.T) {
	session := REPLSession{
		provider:    "openai:gpt-4",
		temperature: 0.7,
		maxTokens:   4096,
		history:     []string{"test"},
		configFile:  "config.yaml",
		outputFile:  "output.txt",
	}

	if session.provider != "openai:gpt-4" {
		t.Error("REPLSession.provider mismatch")
	}
	if session.temperature != 0.7 {
		t.Error("REPLSession.temperature mismatch")
	}
	if len(session.history) != 1 {
		t.Error("REPLSession.history length mismatch")
	}
}
