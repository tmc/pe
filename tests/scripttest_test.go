package tests

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"rsc.io/script"
	"rsc.io/script/scripttest"
)

// peCmd implements a script command that runs pe binary with restrictions
func peCmd(peBinary string) script.Cmd {
	return script.Command(
		script.CmdUsage{
			Summary: "run pe command",
			Args:    "args...",
		},
		func(s *script.State, args ...string) (script.WaitFunc, error) {
			cmd := exec.Command(peBinary, args...)
			cmd.Dir = s.Getwd()
			cmd.Env = append(os.Environ(), s.Environ()...)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Start(); err != nil {
				return nil, err
			}

			return func(*script.State) (string, string, error) {
				err := cmd.Wait()
				return stdout.String(), stderr.String(), err
			}, nil
		},
	)
}

func TestScripts(t *testing.T) {
	// Build the pe binary first
	cmd := exec.Command("go", "build", "-o", "../pe", "./cmd/pe")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build pe binary: %v", err)
	}

	pePath, _ := filepath.Abs("../pe")

	// Create custom commands - remove exec, add pe
	cmds := make(map[string]script.Cmd)
	for name, cmd := range scripttest.DefaultCmds() {
		// Skip exec to prevent arbitrary shell execution
		if name != "exec" {
			cmds[name] = cmd
		}
	}
	// Add our custom pe command
	cmds["pe"] = peCmd(pePath)

	// Create engine with custom commands (no exec, only pe)
	engine := &script.Engine{
		Cmds:  cmds,
		Conds: scripttest.DefaultConds(),
	}

	// Set up test environment
	env := []string{
		"PE_TEST_MODE=true",
		"PE_MOCK_PROVIDER=true",
	}

	// Run tests from testdata/script directory
	scriptDir := "testdata/script"

	// Check if directory exists
	if _, err := os.Stat(scriptDir); os.IsNotExist(err) {
		t.Skip("testdata/script directory not found")
		return
	}

	// Get all .txt files in the directory
	files, err := filepath.Glob(filepath.Join(scriptDir, "*.txt"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Skip("no test files found in testdata/script")
		return
	}

	// Run each test file
	for _, file := range files {
		file := file // capture loop variable
		name := strings.TrimSuffix(filepath.Base(file), ".txt")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			scripttest.Test(t, context.Background(), engine, env, file)
		})
	}
}
