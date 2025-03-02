package tests

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tmc/pe/internal/executil"
)

func TestRunCommand(t *testing.T) {
	// Test successful command execution
	t.Run("SuccessfulExecution", func(t *testing.T) {
		ctx := context.Background()
		result, err := executil.RunCommand(ctx, "echo", []string{"test output"}, nil)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		if !strings.Contains(result.Stdout, "test output") {
			t.Errorf("Expected stdout to contain 'test output', got: %s", result.Stdout)
		}
		
		if result.ExitCode != 0 {
			t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
		}
	})
	
	// Test command with input
	t.Run("CommandWithInput", func(t *testing.T) {
		ctx := context.Background()
		opts := executil.DefaultOptions()
		opts.Input = "input text"
		
		result, err := executil.RunCommand(ctx, "cat", []string{}, opts)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		if !strings.Contains(result.Stdout, "input text") {
			t.Errorf("Expected stdout to contain input text, got: %s", result.Stdout)
		}
	})
	
	// Test command with timeout
	t.Run("CommandTimeout", func(t *testing.T) {
		ctx := context.Background()
		opts := executil.DefaultOptions()
		opts.Timeout = 100 * time.Millisecond
		
		result, err := executil.RunCommand(ctx, "sleep", []string{"1"}, opts)
		
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
		
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		if result.Error == nil || !strings.Contains(result.Error.Error(), "timed out") {
			t.Errorf("Expected timeout error, got: %v", result.Error)
		}
	})
	
	// Test non-existent command
	t.Run("NonExistentCommand", func(t *testing.T) {
		ctx := context.Background()
		result, err := executil.RunCommand(ctx, "nonexistentcommand", []string{}, nil)
		
		if err == nil {
			t.Error("Expected error for non-existent command, got nil")
		}
		
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		if result.Error == nil {
			t.Error("Expected error in result, got nil")
		}
	})
	
	// Test command with working directory
	t.Run("CommandWithWorkDir", func(t *testing.T) {
		ctx := context.Background()
		opts := executil.DefaultOptions()
		opts.Dir = "/"
		
		result, err := executil.RunCommand(ctx, "pwd", []string{}, opts)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		if !strings.Contains(result.Stdout, "/") {
			t.Errorf("Expected stdout to indicate root directory, got: %s", result.Stdout)
		}
	})
	
	// Test command with environment variables
	t.Run("CommandWithEnv", func(t *testing.T) {
		ctx := context.Background()
		opts := executil.DefaultOptions()
		opts.Env = []string{"TESTVAR=testvalue"}
		
		result, err := executil.RunCommand(ctx, "env", []string{}, opts)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		if !strings.Contains(result.Stdout, "TESTVAR=testvalue") {
			t.Errorf("Expected stdout to contain environment variable, got: %s", result.Stdout)
		}
	})
}

func TestGetCommandString(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		args     []string
		input    string
		expected string
	}{
		{
			name:     "SimpleCommand",
			cmd:      "echo",
			args:     []string{"hello"},
			input:    "",
			expected: "echo hello",
		},
		{
			name:     "CommandWithQuotedArgs",
			cmd:      "echo",
			args:     []string{"hello world"},
			input:    "",
			expected: "echo \"hello world\"",
		},
		{
			name:     "CommandWithInput",
			cmd:      "cat",
			args:     []string{},
			input:    "input text",
			expected: "echo \"input text\" | cat",
		},
		{
			name:     "CommandWithQuotesInInput",
			cmd:      "grep",
			args:     []string{"pattern"},
			input:    "text with \"quotes\"",
			expected: "echo \"text with \\\"quotes\\\"\" | grep pattern",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executil.GetCommandString(tt.cmd, tt.args, tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFindExecutables(t *testing.T) {
	// Create a temporary directory for test executables
	tmpDir := t.TempDir()
	
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)
	
	// Add temp dir to PATH
	os.Setenv("PATH", tmpDir+":"+originalPath)
	
	// Create test executables
	exes := []string{"exe1", "exe2", "exe3"}
	for _, exe := range exes {
		path := filepath.Join(tmpDir, exe)
		if err := ioutil.WriteFile(path, []byte("#!/bin/sh\necho test"), 0755); err != nil {
			t.Fatalf("Failed to create test executable %s: %v", exe, err)
		}
	}
	
	// Test LookupExecutable
	t.Run("LookupExecutable", func(t *testing.T) {
		path, err := executil.LookupExecutable("exe1")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if !strings.Contains(path, "exe1") {
			t.Errorf("Expected path to contain 'exe1', got: %s", path)
		}
		
		// Test non-existent executable
		_, err = executil.LookupExecutable("nonexistentexe")
		if err == nil {
			t.Error("Expected error for non-existent executable, got nil")
		}
	})
	
	// Test FindExecutablesInOrder
	t.Run("FindExecutablesInOrder", func(t *testing.T) {
		// Should find the first one
		path, err := executil.FindExecutablesInOrder([]string{"exe1", "exe2", "exe3"})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if !strings.Contains(path, "exe1") {
			t.Errorf("Expected path to contain 'exe1', got: %s", path)
		}
		
		// Should find the second one if first doesn't exist
		path, err = executil.FindExecutablesInOrder([]string{"nonexistentexe", "exe2", "exe3"})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if !strings.Contains(path, "exe2") {
			t.Errorf("Expected path to contain 'exe2', got: %s", path)
		}
		
		// Should return error if none exist
		_, err = executil.FindExecutablesInOrder([]string{"nonexistentexe1", "nonexistentexe2"})
		if err == nil {
			t.Error("Expected error when no executables found, got nil")
		}
	})
}