// Package executil provides utilities for running external commands safely and effectively.
package executil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// Result represents the result of running a command.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
	Duration time.Duration
}

// CommandOptions configures how a command is run.
type CommandOptions struct {
	Timeout    time.Duration
	Dir        string
	Env        []string
	Input      string
	StreamStdout io.Writer
	StreamStderr io.Writer
}

// DefaultOptions returns a default set of command options.
func DefaultOptions() *CommandOptions {
	return &CommandOptions{
		Timeout: 30 * time.Second,
	}
}

// RunCommand runs a command with the given options and returns the result.
func RunCommand(ctx context.Context, name string, args []string, options *CommandOptions) (*Result, error) {
	if options == nil {
		options = DefaultOptions()
	}
	
	// Set up context with timeout if specified
	var cancel context.CancelFunc
	if options.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}
	
	// Create command
	cmd := exec.CommandContext(ctx, name, args...)
	
	// Set working directory if specified
	if options.Dir != "" {
		cmd.Dir = options.Dir
	}
	
	// Set environment if specified
	if len(options.Env) > 0 {
		cmd.Env = options.Env
	}
	
	// Set up stdout and stderr capture
	var stdout, stderr bytes.Buffer
	
	// If we're streaming output, use MultiWriter to capture and stream
	if options.StreamStdout != nil {
		cmd.Stdout = io.MultiWriter(&stdout, options.StreamStdout)
	} else {
		cmd.Stdout = &stdout
	}
	
	if options.StreamStderr != nil {
		cmd.Stderr = io.MultiWriter(&stderr, options.StreamStderr)
	} else {
		cmd.Stderr = &stderr
	}
	
	// Set up stdin if input is provided
	if options.Input != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
		}
		
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, options.Input)
		}()
	}
	
	// Run the command and time it
	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)
	
	// Create result
	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}
	
	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		result.Error = fmt.Errorf("command timed out after %v", options.Timeout)
		return result, result.Error
	}
	
	// Handle exit code and error
	if err != nil {
		result.Error = err
		
		// Extract exit code if available
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	}
	
	return result, nil
}

// GetCommandString returns a string representation of a command that can be run in a shell.
func GetCommandString(name string, args []string, input string) string {
	var cmdParts []string
	
	// Quote the executable name if it contains spaces
	if strings.ContainsAny(name, " \t\n\"'\\") {
		cmdParts = append(cmdParts, fmt.Sprintf("%q", name))
	} else {
		cmdParts = append(cmdParts, name)
	}
	
	// Quote arguments as needed
	for _, arg := range args {
		if strings.ContainsAny(arg, " \t\n\"'\\") {
			cmdParts = append(cmdParts, fmt.Sprintf("%q", arg))
		} else {
			cmdParts = append(cmdParts, arg)
		}
	}
	
	// Build command string with input handling
	cmd := strings.Join(cmdParts, " ")
	if input != "" {
		// Escape quotes in input
		escaped := strings.ReplaceAll(input, "\"", "\\\"")
		cmd = fmt.Sprintf("echo \"%s\" | %s", escaped, cmd)
	}
	
	return cmd
}

// LookupExecutable looks for an executable in the PATH.
func LookupExecutable(name string) (string, error) {
	return exec.LookPath(name)
}

// FindExecutablesInOrder looks for executables in the given order and returns the first one found.
func FindExecutablesInOrder(names []string) (string, error) {
	for _, name := range names {
		path, err := exec.LookPath(name)
		if err == nil {
			return path, nil
		}
	}
	
	return "", fmt.Errorf("none of the executables were found: %v", names)
}