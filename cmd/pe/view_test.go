package main

import (
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestViewCmd_FlagParsing(t *testing.T) {
	cmd := viewCmd()

	// Test flag existence
	if cmd.Flags().Lookup("file") == nil {
		t.Error("Expected --file flag to exist")
	}
	if cmd.Flags().Lookup("port") == nil {
		t.Error("Expected --port flag to exist")
	}
	if cmd.Flags().Lookup("promptfoo") == nil {
		t.Error("Expected --promptfoo flag to exist")
	}
	if cmd.Flags().Lookup("yes") == nil {
		t.Error("Expected --yes flag to exist")
	}
}

func TestViewCmd_CommandStructure(t *testing.T) {
	cmd := viewCmd()

	if cmd.Use != "view [evalId]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

func TestViewCmd_FlagDefaults(t *testing.T) {
	cmd := viewCmd()

	// Check port default
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		t.Errorf("Failed to get port flag: %v", err)
	}
	if port != 8080 {
		t.Errorf("Expected default port 8080, got %d", port)
	}
}

func TestViewCmd_MissingEvalDoesNotRunPromptfoo(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cmd := viewCmd()
	cmd.SetArgs([]string{"missing-eval"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("view missing eval succeeded")
	}
	if !strings.Contains(err.Error(), `evaluation "missing-eval" not found`) {
		t.Fatalf("error = %v, want local not found", err)
	}
}

func TestRunPromptfooViewRequiresNpx(t *testing.T) {
	oldLookPath := promptfooViewLookPath
	oldRun := promptfooViewRun
	defer func() {
		promptfooViewLookPath = oldLookPath
		promptfooViewRun = oldRun
	}()

	promptfooViewLookPath = func(string) (string, error) {
		return "", exec.ErrNotFound
	}
	promptfooViewRun = func(string, []string) error {
		t.Fatal("promptfoo view command ran without npx")
		return nil
	}

	err := runPromptfooView(nil, false)
	if err == nil {
		t.Fatal("runPromptfooView succeeded without npx")
	}
	if err.Error() != "npx not found for promptfoo viewer" {
		t.Fatalf("error = %v, want npx not found", err)
	}
}

func TestRunPromptfooViewUsesExplicitArgv(t *testing.T) {
	oldLookPath := promptfooViewLookPath
	oldRun := promptfooViewRun
	defer func() {
		promptfooViewLookPath = oldLookPath
		promptfooViewRun = oldRun
	}()

	var gotName string
	var gotArgs []string
	promptfooViewLookPath = func(name string) (string, error) {
		if name != "npx" {
			t.Fatalf("look path name = %q, want npx", name)
		}
		return "/usr/bin/npx", nil
	}
	promptfooViewRun = func(name string, args []string) error {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return nil
	}

	if err := runPromptfooView([]string{"eval-123"}, true); err != nil {
		t.Fatalf("runPromptfooView: %v", err)
	}

	if gotName != "npx" {
		t.Fatalf("command name = %q, want npx", gotName)
	}
	wantArgs := []string{"promptfoo", "view", "eval-123", "-y"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestViewCmdPromptfooFlagDelegatesExplicitly(t *testing.T) {
	oldLookPath := promptfooViewLookPath
	oldRun := promptfooViewRun
	defer func() {
		promptfooViewLookPath = oldLookPath
		promptfooViewRun = oldRun
	}()

	promptfooViewLookPath = func(string) (string, error) {
		return "/usr/bin/npx", nil
	}
	promptfooViewRun = func(string, []string) error {
		return errors.New("sentinel promptfoo delegation")
	}

	cmd := viewCmd()
	cmd.SetArgs([]string{"--promptfoo", "eval-123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("view --promptfoo succeeded")
	}
	if err.Error() != "sentinel promptfoo delegation" {
		t.Fatalf("error = %v, want sentinel promptfoo delegation", err)
	}
}
