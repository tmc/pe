// Package executil provides utilities for safely executing external commands.
package executil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Result stores the outcome of an executed command.
type Result struct {
	// Command that was executed
	Command string
	// Args that were passed to the command
	Args []string
	// ExitCode of the command
	ExitCode int
	// Stdout captured from the command
	Stdout string
	// Stderr captured from the command
	Stderr string
	// CombinedOutput contains both stdout and stderr
	CombinedOutput string
	// Error if any occurred during execution
	Error error
	// Duration of command execution
	Duration time.Duration
}

// String returns a string representation of the command and its result.
func (r Result) String() string {
	status := "success"
	if r.Error != nil {
		status = fmt.Sprintf("error: %v", r.Error)
	} else if r.ExitCode != 0 {
		status = fmt.Sprintf("exit code: %d", r.ExitCode)
	}
	
	return fmt.Sprintf("Command: %s %s\nStatus: %s\nDuration: %v\nStdout: %s\nStderr: %s",
		r.Command, strings.Join(r.Args, " "), status, r.Duration, r.Stdout, r.Stderr)
}

// ExecOptions provides configuration options for command execution.
type ExecOptions struct {
	// Timeout for the command execution
	Timeout time.Duration
	// Dir sets the working directory for the command
	Dir string
	// Env sets environment variables for the command
	Env []string
	// Input provides stdin content for the command
	Input string
	// StdoutWriter receives stdout content as it's produced
	StdoutWriter io.Writer
	// StderrWriter receives stderr content as it's produced
	StderrWriter io.Writer
	// CombinedWriter receives both stdout and stderr as they're produced
	CombinedWriter io.Writer
	// Nice value for process priority (-20 to 19, lower is higher priority)
	Nice int
	// User ID to run command as (requires privileges)
	UID int
	// Group ID to run command as (requires privileges)
	GID int
	// DisableSanitization disables command sanitization (use with caution)
	DisableSanitization bool
	// AllowedCommands explicitly lists allowed commands (empty means all allowed)
	AllowedCommands []string
	// BlockedCommands explicitly lists blocked commands
	BlockedCommands []string
}

// DefaultExecOptions returns the default execution options.
func DefaultExecOptions() *ExecOptions {
	return &ExecOptions{
		Timeout:             60 * time.Second,
		Env:                 os.Environ(),
		DisableSanitization: false,
		BlockedCommands:     []string{"rm", "mkfs", "dd", "shutdown", "reboot", "halt"},
	}
}

// RunCommand runs a command with the given options and returns the result.
func RunCommand(ctx context.Context, command string, args []string, options *ExecOptions) (Result, error) {
	if options == nil {
		options = DefaultExecOptions()
	}
	
	result := Result{
		Command: command,
		Args:    args,
	}
	
	// Check if command is allowed/blocked
	if err := validateCommand(command, options); err != nil {
		result.Error = err
		return result, err
	}
	
	// Apply timeout if specified
	var cancel context.CancelFunc
	if options.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}
	
	// Create command
	cmd := exec.CommandContext(ctx, command, args...)
	
	// Set working directory if specified
	if options.Dir != "" {
		cmd.Dir = options.Dir
	}
	
	// Set environment variables if specified
	if options.Env != nil {
		cmd.Env = options.Env
	}
	
	// Set up process attributes for UID/GID/Nice if specified
	if options.UID > 0 || options.GID > 0 || options.Nice != 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		
		if options.UID > 0 || options.GID > 0 {
			cmd.SysProcAttr.Credential = &syscall.Credential{}
			
			if options.UID > 0 {
				cmd.SysProcAttr.Credential.Uid = uint32(options.UID)
			}
			
			if options.GID > 0 {
				cmd.SysProcAttr.Credential.Gid = uint32(options.GID)
			}
		}
		
		if options.Nice != 0 {
			cmd.SysProcAttr.Setpgid = true
		}
	}
	
	// Set up input/output pipes
	var stdout, stderr bytes.Buffer
	var combinedOutput bytes.Buffer
	
	// Create stdin pipe if input is provided
	if options.Input != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			result.Error = fmt.Errorf("failed to create stdin pipe: %w", err)
			return result, result.Error
		}
		
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, options.Input)
		}()
	}
	
	// Set up stdout
	if options.StdoutWriter != nil {
		cmd.Stdout = io.MultiWriter(options.StdoutWriter, &stdout, &combinedOutput)
	} else {
		cmd.Stdout = io.MultiWriter(&stdout, &combinedOutput)
	}
	
	// Set up stderr
	if options.StderrWriter != nil {
		cmd.Stderr = io.MultiWriter(options.StderrWriter, &stderr, &combinedOutput)
	} else {
		cmd.Stderr = io.MultiWriter(&stderr, &combinedOutput)
	}
	
	// Add combined writer if specified
	if options.CombinedWriter != nil {
		cmd.Stdout = io.MultiWriter(cmd.Stdout, options.CombinedWriter)
		cmd.Stderr = io.MultiWriter(cmd.Stderr, options.CombinedWriter)
	}
	
	// Track execution time
	startTime := time.Now()
	
	// Start the command
	if err := cmd.Start(); err != nil {
		result.Error = fmt.Errorf("failed to start command: %w", err)
		return result, result.Error
	}
	
	// If nice value is specified, try to set it after process has started
	if options.Nice != 0 && cmd.Process != nil {
		if err := syscall.Setpriority(syscall.PRIO_PROCESS, cmd.Process.Pid, options.Nice); err != nil {
			// Just log the error but continue execution
			fmt.Fprintf(os.Stderr, "Warning: failed to set nice value: %v\n", err)
		}
	}
	
	// Wait for the command to complete
	err := cmd.Wait()
	result.Duration = time.Since(startTime)
	
	// Populate result
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.CombinedOutput = combinedOutput.String()
	
	// Handle errors and exit code
	if err != nil {
		result.Error = err
		
		// Extract exit code if possible
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		}
		
		// Check for timeout
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("command timed out after %v: %w", options.Timeout, err)
		}
	}
	
	return result, result.Error
}

// RunCommandSimple runs a command with default options.
func RunCommandSimple(command string, args ...string) (string, error) {
	result, err := RunCommand(context.Background(), command, args, nil)
	if err != nil {
		return result.CombinedOutput, err
	}
	return result.Stdout, nil
}

// RunCommandWithTimeout runs a command with a specified timeout.
func RunCommandWithTimeout(timeout time.Duration, command string, args ...string) (string, error) {
	options := DefaultExecOptions()
	options.Timeout = timeout
	
	result, err := RunCommand(context.Background(), command, args, options)
	if err != nil {
		return result.CombinedOutput, err
	}
	return result.Stdout, nil
}

// RunCommandWithInput runs a command with input provided via stdin.
func RunCommandWithInput(input string, command string, args ...string) (string, error) {
	options := DefaultExecOptions()
	options.Input = input
	
	result, err := RunCommand(context.Background(), command, args, options)
	if err != nil {
		return result.CombinedOutput, err
	}
	return result.Stdout, nil
}

// RunCommandStreaming runs a command and streams output to the provided writers.
func RunCommandStreaming(stdout, stderr io.Writer, command string, args ...string) (Result, error) {
	options := DefaultExecOptions()
	options.StdoutWriter = stdout
	options.StderrWriter = stderr
	
	return RunCommand(context.Background(), command, args, options)
}

// validateCommand checks if the command is allowed to run.
func validateCommand(command string, options *ExecOptions) error {
	// Skip sanitization if disabled
	if options.DisableSanitization {
		return nil
	}
	
	// Extract command name without path
	cmdName := command
	if lastSlash := strings.LastIndex(command, "/"); lastSlash != -1 {
		cmdName = command[lastSlash+1:]
	}
	
	// Check if command is in blocked list
	for _, blocked := range options.BlockedCommands {
		if cmdName == blocked {
			return fmt.Errorf("command '%s' is blocked for security reasons", cmdName)
		}
	}
	
	// Check if command is in allowed list (if the list is non-empty)
	if len(options.AllowedCommands) > 0 {
		allowed := false
		for _, allowedCmd := range options.AllowedCommands {
			if cmdName == allowedCmd {
				allowed = true
				break
			}
		}
		
		if !allowed {
			return fmt.Errorf("command '%s' is not in the allowed list", cmdName)
		}
	}
	
	return nil
}