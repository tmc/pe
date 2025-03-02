// Package executil provides utilities for running external commands safely and effectively.
// It includes features like timeouts, input/output handling, environment variable management,
// and security measures to prevent command injection.
package executil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// Result represents the result of running a command.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
	Duration time.Duration
	Command  string // String representation of the command that was run
}

// CommandOptions configures how a command is run.
type CommandOptions struct {
	Timeout       time.Duration
	Dir           string
	Env           []string
	Input         string
	StreamStdout  io.Writer
	StreamStderr  io.Writer
	// Security options
	DisableShell  bool       // If true, commands won't be run through a shell
	SanitizeInput bool       // If true, input will be sanitized to prevent injection
	AllowedArgs   []string   // If not empty, only these arguments are allowed
	BlockedArgs   []string   // Arguments that are never allowed
	GroupID       int        // Run command with this group ID (0 = don't change)
	UserID        int        // Run command with this user ID (0 = don't change)
	Nice          int        // Nice value for the process (-20 to 19, 0 = default priority)
}

// SecurityOption represents a security-related option for command execution.
type SecurityOption func(*CommandOptions)

// ErrCommandInjection is returned when a command or arguments are detected to be unsafe.
var ErrCommandInjection = errors.New("potential command injection detected")

// containsShellInjection checks if a string contains potential shell injection characters.
func containsShellInjection(s string, blockedArgs []string) bool {
	for _, blocked := range blockedArgs {
		if strings.Contains(s, blocked) {
			return true
		}
	}
	return false
}

// sanitizeInput removes potentially dangerous characters from input.
func sanitizeInput(input string) string {
	// Remove backticks, dollar signs, and other shell metacharacters
	re := regexp.MustCompile("[`$&|;()<>]")
	return re.ReplaceAllString(input, "")
}

// DefaultOptions returns a default set of command options.
func DefaultOptions() *CommandOptions {
	return &CommandOptions{
		Timeout:       30 * time.Second,
		DisableShell:  true,     // Safer default: don't use shell
		SanitizeInput: true,     // Safer default: sanitize input
		BlockedArgs:   []string{"$(", "`", "\\", "|", "&&", "||", ";", ">", "<", "&"},
	}
}

// WithSafeExecution returns a SecurityOption that configures the command to execute safely.
func WithSafeExecution() SecurityOption {
	return func(o *CommandOptions) {
		o.DisableShell = true
		o.SanitizeInput = true
	}
}

// WithAllowedArgs returns a SecurityOption that configures which arguments are allowed.
func WithAllowedArgs(args []string) SecurityOption {
	return func(o *CommandOptions) {
		o.AllowedArgs = args
	}
}

// WithBlockedArgs returns a SecurityOption that configures which arguments are blocked.
func WithBlockedArgs(args []string) SecurityOption {
	return func(o *CommandOptions) {
		o.BlockedArgs = args
	}
}

// WithCredentials returns a SecurityOption that configures the user and group IDs.
func WithCredentials(uid, gid int) SecurityOption {
	return func(o *CommandOptions) {
		o.UserID = uid
		o.GroupID = gid
	}
}

// WithNice returns a SecurityOption that configures the process priority.
func WithNice(nice int) SecurityOption {
	return func(o *CommandOptions) {
		if nice < -20 {
			nice = -20
		}
		if nice > 19 {
			nice = 19
		}
		o.Nice = nice
	}
}

// RunCommand runs a command with the given options and returns the result.
func RunCommand(ctx context.Context, name string, args []string, options *CommandOptions) (*Result, error) {
	if options == nil {
		options = DefaultOptions()
	}
	
	// Get the command string for the result
	cmdString := GetCommandString(name, args, options.Input)
	
	// Check for security issues before proceeding
	if options.SanitizeInput {
		// First check executable name
		if containsShellInjection(name, options.BlockedArgs) {
			return &Result{
				Command: cmdString,
				Stderr:  "Blocked unsafe command: " + name,
				Error:   ErrCommandInjection,
			}, ErrCommandInjection
		}
		
		// Then check each argument
		for i, arg := range args {
			if containsShellInjection(arg, options.BlockedArgs) {
				return &Result{
					Command: cmdString,
					Stderr:  fmt.Sprintf("Blocked unsafe argument %d: %s", i, arg),
					Error:   ErrCommandInjection,
				}, ErrCommandInjection
			}
		}
		
		// If allowed args are specified, verify all args are allowed
		if len(options.AllowedArgs) > 0 {
			for _, arg := range args {
				allowed := false
				for _, allowedArg := range options.AllowedArgs {
					if arg == allowedArg {
						allowed = true
						break
					}
				}
				if !allowed {
					return &Result{
						Command: cmdString,
						Stderr:  fmt.Sprintf("Argument not allowed: %s", arg),
						Error:   ErrCommandInjection,
					}, ErrCommandInjection
				}
			}
		}
		
		// Sanitize input if needed
		if options.Input != "" && options.SanitizeInput {
			options.Input = sanitizeInput(options.Input)
		}
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
	
	// Set process attributes for security
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	
	// Set user and group IDs if specified
	if options.UserID > 0 || options.GroupID > 0 {
		// Unix-specific code
		if options.UserID > 0 {
			cmd.SysProcAttr.Credential = &syscall.Credential{
				Uid: uint32(options.UserID),
			}
		}
		if options.GroupID > 0 {
			if cmd.SysProcAttr.Credential == nil {
				cmd.SysProcAttr.Credential = &syscall.Credential{}
			}
			cmd.SysProcAttr.Credential.Gid = uint32(options.GroupID)
		}
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
			return &Result{
				Command: cmdString,
				Stderr:  fmt.Sprintf("Failed to create stdin pipe: %v", err),
				Error:   err,
			}, fmt.Errorf("failed to create stdin pipe: %w", err)
		}
		
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, options.Input)
		}()
	}
	
	// Run the command and time it
	startTime := time.Now()
	
	// Set nice value if specified (Unix-specific)
	if options.Nice != 0 {
		cmd.SysProcAttr.Setpgid = true
		
		go func() {
			// Wait for the process to start
			time.Sleep(10 * time.Millisecond)
			if cmd.Process != nil {
				// Ignore error; nice is best-effort
				syscall.Setpriority(syscall.PRIO_PROCESS, cmd.Process.Pid, options.Nice)
			}
		}()
	}
	
	err := cmd.Run()
	duration := time.Since(startTime)
	
	// Create result
	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
		Command:  cmdString,
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