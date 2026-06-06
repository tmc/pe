package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkCmd_CommandStructure(t *testing.T) {
	if workCmd.Use != "work" {
		t.Errorf("Unexpected Use: %s", workCmd.Use)
	}

	if workCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Verify subcommands exist
	subcommandNames := []string{"init", "use", "edit", "sync", "list"}
	for _, name := range subcommandNames {
		found := false
		for _, c := range workCmd.Commands() {
			if c.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q to exist", name)
		}
	}
}

func TestWorkInitCmd_CommandStructure(t *testing.T) {
	if workInitCmd.Use != "init [directories...]" {
		t.Errorf("Unexpected Use: %s", workInitCmd.Use)
	}
}

func TestWorkUseCmd_FlagParsing(t *testing.T) {
	if workUseCmd.Flags().Lookup("remove") == nil {
		t.Error("Expected --remove flag to exist")
	}
}

func TestWorkEditCmd_FlagParsing(t *testing.T) {
	flags := []string{"replace", "dropreplace"}
	for _, name := range flags {
		if workEditCmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestRunWorkInit(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	err := runWorkInit(workInitCmd, []string{"./prompts", "./shared"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify pe.work was created
	if _, err := os.Stat("pe.work"); os.IsNotExist(err) {
		t.Error("Expected pe.work to be created")
	}

	// Read and verify content
	content, err := os.ReadFile("pe.work")
	if err != nil {
		t.Fatalf("Failed to read pe.work: %v", err)
	}

	if !strings.Contains(string(content), "prompts") {
		t.Error("Expected pe.work to contain 'prompts'")
	}
	if !strings.Contains(string(content), "shared") {
		t.Error("Expected pe.work to contain 'shared'")
	}
}

func TestRunWorkInit_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create existing pe.work
	if err := os.WriteFile("pe.work", []byte("pe 1\n"), 0644); err != nil {
		t.Fatalf("Failed to create pe.work: %v", err)
	}

	err := runWorkInit(workInitCmd, []string{"./prompts"})
	if err == nil {
		t.Error("Expected error when pe.work already exists")
	}
}

func TestRunWorkUse_NoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	err := runWorkUse(workUseCmd, []string{"./prompts"})
	if err == nil {
		t.Error("Expected error when no pe.work exists")
	}
}

func TestRunWorkUse_AddDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Initialize workspace
	if err := runWorkInit(workInitCmd, []string{}); err != nil {
		t.Fatalf("Failed to init workspace: %v", err)
	}

	// Add directory
	workRemove = false
	err := runWorkUse(workUseCmd, []string{"./new-prompts"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify content
	content, err := os.ReadFile("pe.work")
	if err != nil {
		t.Fatalf("Failed to read pe.work: %v", err)
	}

	if !strings.Contains(string(content), "new-prompts") {
		t.Error("Expected pe.work to contain 'new-prompts'")
	}
}

func TestRunWorkEdit_NoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	err := runWorkEdit(workEditCmd, []string{})
	if err == nil {
		t.Error("Expected error when no pe.work exists")
	}
}

func TestRunWorkEdit_ReplaceRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := runWorkInit(workInitCmd, []string{"./prompts"}); err != nil {
		t.Fatalf("Failed to init workspace: %v", err)
	}

	workReplace = []string{"example.com/old=./local"}
	workDropReplace = nil
	t.Cleanup(func() {
		workReplace = nil
		workDropReplace = nil
	})
	if err := runWorkEdit(workEditCmd, nil); err != nil {
		t.Fatalf("runWorkEdit replace: %v", err)
	}

	ws, err := loadWorkspace()
	if err != nil {
		t.Fatalf("loadWorkspace: %v", err)
	}
	if len(ws.Replace) != 1 || ws.Replace[0].Old != "example.com/old" || ws.Replace[0].New != "./local" {
		t.Fatalf("replace = %#v", ws.Replace)
	}

	workReplace = nil
	workDropReplace = []string{"example.com/old"}
	if err := runWorkEdit(workEditCmd, nil); err != nil {
		t.Fatalf("runWorkEdit drop: %v", err)
	}
	ws, err = loadWorkspace()
	if err != nil {
		t.Fatalf("loadWorkspace after drop: %v", err)
	}
	if len(ws.Replace) != 0 {
		t.Fatalf("replace after drop = %#v", ws.Replace)
	}
}

func TestRunWorkSync_NoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	err := runWorkSync(workSyncCmd, []string{})
	if err == nil {
		t.Error("Expected error when no pe.work exists")
	}
}

func TestRunWorkSync_WithWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create directory structure
	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("Failed to create prompts dir: %v", err)
	}

	// Create a prompt file
	if err := os.WriteFile(filepath.Join(promptsDir, "test.prompt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create prompt file: %v", err)
	}

	// Initialize workspace
	if err := runWorkInit(workInitCmd, []string{"./prompts"}); err != nil {
		t.Fatalf("Failed to init workspace: %v", err)
	}

	// Sync
	err := runWorkSync(workSyncCmd, []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunWorkList_NoWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	err := runWorkList(workListCmd, []string{})
	if err == nil {
		t.Error("Expected error when no pe.work exists")
	}
}

func TestRunWorkList_WithWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create directory structure
	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("Failed to create prompts dir: %v", err)
	}

	// Create a prompt file
	if err := os.WriteFile(filepath.Join(promptsDir, "test.prompt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create prompt file: %v", err)
	}

	// Initialize workspace
	if err := runWorkInit(workInitCmd, []string{"./prompts"}); err != nil {
		t.Fatalf("Failed to init workspace: %v", err)
	}

	// List
	err := runWorkList(workListCmd, []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestLoadWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create pe.work
	content := `pe 1

use (
	./prompts
	./shared
)
`
	if err := os.WriteFile("pe.work", []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create pe.work: %v", err)
	}

	ws, err := loadWorkspace()
	if err != nil {
		t.Fatalf("Failed to load workspace: %v", err)
	}

	if len(ws.Use) != 2 {
		t.Errorf("Expected 2 use directories, got %d", len(ws.Use))
	}
}

func TestLoadWorkspace_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	_, err := loadWorkspace()
	if err == nil {
		t.Error("Expected error when pe.work not found")
	}
}

func TestSaveWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	ws := &Workspace{
		Version: "1",
		Use:     []string{"./prompts", "./shared"},
		Replace: []Replace{{Old: "example.com/old", New: "./local"}},
	}

	err := saveWorkspace(ws)
	if err != nil {
		t.Fatalf("Failed to save workspace: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile("pe.work")
	if err != nil {
		t.Fatalf("Failed to read pe.work: %v", err)
	}

	if !strings.Contains(string(content), "prompts") {
		t.Error("Expected pe.work to contain 'prompts'")
	}
	if !strings.Contains(string(content), "example.com/old => ./local") {
		t.Error("Expected pe.work to contain replace directive")
	}

	loaded, err := loadWorkspace()
	if err != nil {
		t.Fatalf("Failed to load pe.work: %v", err)
	}
	if len(loaded.Replace) != 1 || loaded.Replace[0].Old != "example.com/old" || loaded.Replace[0].New != "./local" {
		t.Fatalf("loaded replace = %#v", loaded.Replace)
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		slice []string
		s     string
		want  bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{nil, "a", false},
	}

	for _, tt := range tests {
		got := containsString(tt.slice, tt.s)
		if got != tt.want {
			t.Errorf("containsString(%v, %q) = %v, want %v", tt.slice, tt.s, got, tt.want)
		}
	}
}

func TestRemoveStrings(t *testing.T) {
	tests := []struct {
		slice  []string
		remove []string
		want   []string
	}{
		{[]string{"a", "b", "c"}, []string{"b"}, []string{"a", "c"}},
		{[]string{"a", "b", "c"}, []string{"d"}, []string{"a", "b", "c"}},
		{[]string{"a", "b", "c"}, []string{"a", "c"}, []string{"b"}},
		{[]string{}, []string{"a"}, nil},
	}

	for _, tt := range tests {
		got := removeStrings(tt.slice, tt.remove)
		if len(got) != len(tt.want) {
			t.Errorf("removeStrings(%v, %v) = %v, want %v", tt.slice, tt.remove, got, tt.want)
		}
	}
}

func TestWorkspaceStruct(t *testing.T) {
	ws := Workspace{
		Version: "1",
		Use:     []string{"./prompts", "./shared"},
		Replace: []Replace{
			{Old: "example.com/old", New: "example.com/new"},
		},
	}

	if ws.Version != "1" {
		t.Error("Workspace.Version mismatch")
	}
	if len(ws.Use) != 2 {
		t.Error("Workspace.Use length mismatch")
	}
	if len(ws.Replace) != 1 {
		t.Error("Workspace.Replace length mismatch")
	}
}

func TestReplaceStruct(t *testing.T) {
	r := Replace{
		Old: "example.com/original",
		New: "./local",
	}

	if r.Old != "example.com/original" {
		t.Error("Replace.Old mismatch")
	}
	if r.New != "./local" {
		t.Error("Replace.New mismatch")
	}
}
