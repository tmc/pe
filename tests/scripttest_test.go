package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"rsc.io/script"
	"rsc.io/script/scripttest"
)

func TestScripts(t *testing.T) {
	// Build the pe binary first
	cmd := exec.Command("go", "build", "-o", "../pe", "./cmd/pe")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build pe binary: %v", err)
	}
	
	// Create engine with default commands
	engine := &script.Engine{
		Cmds:  scripttest.DefaultCmds(),
		Conds: scripttest.DefaultConds(),
	}

	// Set up test environment
	pePath, _ := filepath.Abs("../pe")
	peDir := filepath.Dir(pePath)
	env := []string{
		"PE_TEST_MODE=true",
		"PE_MOCK_PROVIDER=true",
		fmt.Sprintf("PATH=%s:%s", peDir, os.Getenv("PATH")),
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