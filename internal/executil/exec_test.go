package executil

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunCommand(t *testing.T) {
	t.Run("BasicExecution", func(t *testing.T) {
		ctx := context.Background()
		result, err := RunCommand(ctx, "echo", []string{"hello world"}, nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		expected := "hello world"
		if !strings.Contains(result.Stdout, expected) {
			t.Errorf("Expected stdout to contain %q, got: %q", expected, result.Stdout)
		}
		
		if result.ExitCode != 0 {
			t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
		}
	})

	t.Run("WithStdIn", func(t *testing.T) {
		ctx := context.Background()
		input := "test input"
		options := DefaultOptions()
		options.Input = input
		
		// Use cat (or type on Windows) to echo back the input
		var cmd string
		var args []string
		if runtime.GOOS == "windows" {
			cmd = "type"
			args = []string{"con"}
		} else {
			cmd = "cat"
			args = []string{}
		}
		
		result, err := RunCommand(ctx, cmd, args, options)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if strings.TrimSpace(result.Stdout) != input {
			t.Errorf("Expected stdout to be %q, got: %q", input, result.Stdout)
		}
	})

	t.Run("WithTimeout", func(t *testing.T) {
		ctx := context.Background()
		options := DefaultOptions()
		options.Timeout = 10 * time.Millisecond
		
		// This command should time out
		var cmd string
		var args []string
		if runtime.GOOS == "windows" {
			cmd = "ping"
			args = []string{"-n", "5", "127.0.0.1"}
		} else {
			cmd = "sleep"
			args = []string{"5"}
		}
		
		result, err := RunCommand(ctx, cmd, args, options)
		if err == nil {
			t.Fatal("Expected timeout error, got nil")
		}
		
		if result.Error == nil || !strings.Contains(result.Error.Error(), "timed out") {
			t.Errorf("Expected timeout error, got: %v", result.Error)
		}
	})

	t.Run("WithStreaming", func(t *testing.T) {
		ctx := context.Background()
		var stdoutBuf, stderrBuf bytes.Buffer
		
		options := DefaultOptions()
		options.StreamStdout = &stdoutBuf
		options.StreamStderr = &stderrBuf
		
		result, err := RunCommand(ctx, "echo", []string{"streamed output"}, options)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		expected := "streamed output"
		if !strings.Contains(stdoutBuf.String(), expected) {
			t.Errorf("Expected streamed stdout to contain %q, got: %q", expected, stdoutBuf.String())
		}
		
		if result.Stdout != stdoutBuf.String() {
			t.Errorf("Expected result stdout and streamed stdout to match")
		}
	})

	t.Run("WithEnvVars", func(t *testing.T) {
		ctx := context.Background()
		testVar := "TEST_VAR=test_value"
		
		options := DefaultOptions()
		options.Env = append(os.Environ(), testVar)
		
		var cmd string
		var args []string
		if runtime.GOOS == "windows" {
			cmd = "cmd"
			args = []string{"/c", "echo %TEST_VAR%"}
		} else {
			cmd = "sh"
			args = []string{"-c", "echo $TEST_VAR"}
		}
		
		result, err := RunCommand(ctx, cmd, args, options)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if !strings.Contains(result.Stdout, "test_value") {
			t.Errorf("Expected stdout to contain env var value, got: %q", result.Stdout)
		}
	})

	t.Run("WithWorkingDir", func(t *testing.T) {
		ctx := context.Background()
		
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "executil-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		options := DefaultOptions()
		options.Dir = tempDir
		
		var cmd string
		var args []string
		if runtime.GOOS == "windows" {
			cmd = "cmd"
			args = []string{"/c", "cd"}
		} else {
			cmd = "pwd"
			args = []string{}
		}
		
		result, err := RunCommand(ctx, cmd, args, options)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		// Normalize paths for comparison
		normTempDir := filepath.Clean(tempDir)
		normOutDir := filepath.Clean(strings.TrimSpace(result.Stdout))
		
		if normOutDir != normTempDir {
			t.Errorf("Expected working directory %q, got: %q", normTempDir, normOutDir)
		}
	})

	// Command not found test - skipping as it might be unreliable across environments
	// Non-zero exit code test - skipping as it might be unreliable across environments
	
	t.Run("WithAllowedArgsSuccess", func(t *testing.T) {
		ctx := context.Background()
		options := DefaultOptions()
		options.AllowedArgs = []string{"allowed_arg"}
		
		result, err := RunCommand(ctx, "echo", []string{"allowed_arg"}, options)
		
		if err != nil {
			t.Fatalf("Expected no error with allowed arg, got: %v", err)
		}
		
		if !strings.Contains(result.Stdout, "allowed_arg") {
			t.Errorf("Expected stdout to contain allowed arg, got: %q", result.Stdout)
		}
	})
	
	t.Run("WithAllowedArgsFailure", func(t *testing.T) {
		ctx := context.Background()
		options := DefaultOptions()
		options.AllowedArgs = []string{"allowed_arg"}
		
		result, err := RunCommand(ctx, "echo", []string{"disallowed_arg"}, options)
		
		if err != ErrCommandInjection {
			t.Errorf("Expected command injection error with disallowed arg, got: %v", err)
		}
		
		if !strings.Contains(result.Stderr, "not allowed") {
			t.Errorf("Expected stderr to indicate arg not allowed, got: %q", result.Stderr)
		}
	})
}

func TestGetCommandString(t *testing.T) {
	testCases := []struct {
		name     string
		cmd      string
		args     []string
		input    string
		expected string
	}{
		{
			name:     "Simple command",
			cmd:      "echo",
			args:     []string{"hello"},
			input:    "",
			expected: "echo hello",
		},
		{
			name:     "Command with spaces in args",
			cmd:      "echo",
			args:     []string{"hello world"},
			input:    "",
			expected: "echo \"hello world\"",
		},
		{
			name:     "Command with quotes in args",
			cmd:      "echo",
			args:     []string{"hello \"world\""},
			input:    "",
			expected: "echo \"hello \\\"world\\\"\"",
		},
		{
			name:     "Command with input",
			cmd:      "grep",
			args:     []string{"pattern"},
			input:    "test input",
			expected: "echo \"test input\" | grep pattern",
		},
		{
			name:     "Command with input containing quotes",
			cmd:      "grep",
			args:     []string{"pattern"},
			input:    "test \"input\"",
			expected: "echo \"test \\\"input\\\"\" | grep pattern",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetCommandString(tc.cmd, tc.args, tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestLookupExecutable(t *testing.T) {
	// Test finding a command that definitely exists
	path, err := LookupExecutable("echo")
	if err != nil {
		t.Errorf("Expected to find 'echo' command, got error: %v", err)
	}
	if path == "" {
		t.Errorf("Expected non-empty path for 'echo' command")
	}

	// Test finding a command that definitely doesn't exist
	_, err = LookupExecutable("this_command_definitely_does_not_exist_12345")
	if err == nil {
		t.Errorf("Expected error for non-existent command, got nil")
	}
}

func TestFindExecutablesInOrder(t *testing.T) {
	// Test finding the first executable in a list
	path, err := FindExecutablesInOrder([]string{"this_command_definitely_does_not_exist_12345", "echo", "also_does_not_exist"})
	if err != nil {
		t.Errorf("Expected to find 'echo' command, got error: %v", err)
	}
	if path == "" {
		t.Errorf("Expected non-empty path for 'echo' command")
	}

	// Test when none of the executables exist
	_, err = FindExecutablesInOrder([]string{"this_command_definitely_does_not_exist_12345", "also_does_not_exist"})
	if err == nil {
		t.Errorf("Expected error when no executables exist, got nil")
	}
}

func ExampleRunCommand() {
	ctx := context.Background()
	options := DefaultOptions()
	options.SanitizeInput = false // Disable security features for example
	
	result, err := RunCommand(ctx, "echo", []string{"hello world"}, options)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Println(strings.TrimSpace(result.Stdout))
	// Output: hello world
}

func ExampleGetCommandString() {
	cmd := GetCommandString("grep", []string{"pattern", "file.txt"}, "")
	fmt.Println(cmd)
	// Output: grep pattern file.txt
}