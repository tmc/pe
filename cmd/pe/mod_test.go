package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModCmd_CommandStructure(t *testing.T) {
	// Test parent mod command
	if modCmd.Use != "mod" {
		t.Errorf("Unexpected Use: %s", modCmd.Use)
	}

	if modCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Verify subcommands exist
	subcommands := []string{"init", "list", "get", "download", "tidy", "vendor", "search", "publish"}
	for _, name := range subcommands {
		found := false
		for _, cmd := range modCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q to exist", name)
		}
	}
}

func TestModInitCmd_Structure(t *testing.T) {
	if modInitCmd.Use != "init [module-name]" {
		t.Errorf("Unexpected Use: %s", modInitCmd.Use)
	}
}

func TestModInitCmd_Success(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Run mod init
	err := runModInit(modInitCmd, []string{"example.com/test-module"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify pe.mod was created
	if _, err := os.Stat(filepath.Join(tmpDir, "pe.mod")); os.IsNotExist(err) {
		t.Error("Expected pe.mod to be created")
	}
}

func TestModInitCmd_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create existing pe.mod
	err := os.WriteFile(filepath.Join(tmpDir, "pe.mod"), []byte("module existing\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create pe.mod: %v", err)
	}

	// Run mod init without force flag
	modForce = false
	err = runModInit(modInitCmd, []string{"example.com/new-module"})
	if err == nil {
		t.Error("Expected error when pe.mod already exists")
	}
}

func TestModInitCmd_ForceOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create existing pe.mod
	err := os.WriteFile(filepath.Join(tmpDir, "pe.mod"), []byte("module existing\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create pe.mod: %v", err)
	}

	// Run mod init with force flag
	modForce = true
	defer func() { modForce = false }()

	err = runModInit(modInitCmd, []string{"example.com/new-module"})
	if err != nil {
		t.Errorf("Unexpected error with force flag: %v", err)
	}
}

func TestModTidyCmd_NoPeMod(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Should error when no pe.mod exists
	err := runModTidy(modTidyCmd, nil)
	if err == nil {
		t.Error("Expected error when pe.mod doesn't exist")
	}
}

func TestModTidyCmd_ReportsMissingAndUnusedDeps(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

require (
	github.com/example/used v1.0.0
	github.com/example/unused v1.0.0
)
`), 0644)
	if err != nil {
		t.Fatalf("writing pe.mod: %v", err)
	}
	if err := os.WriteFile("review.prompt", []byte("use pe://github.com/example/used/review\n"), 0644); err != nil {
		t.Fatalf("writing review.prompt: %v", err)
	}
	if err := os.WriteFile("summarize.yaml", []byte("prompt: pe://github.com/example/missing/summarize\n"), 0644); err != nil {
		t.Fatalf("writing summarize.yaml: %v", err)
	}
	if err := os.Mkdir(".pe", 0755); err != nil {
		t.Fatalf("creating .pe: %v", err)
	}
	if err := os.WriteFile(filepath.Join(".pe", "ignored.prompt"), []byte("pe://github.com/example/ignored/prompt\n"), 0644); err != nil {
		t.Fatalf("writing ignored prompt: %v", err)
	}

	var out bytes.Buffer
	cmd := *modTidyCmd
	cmd.SetOut(&out)
	if err := runModTidy(&cmd, nil); err != nil {
		t.Fatalf("runModTidy: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"found 2 module references in 2 files",
		"missing dependency: github.com/example/missing (referenced by summarize.yaml)",
		"unused dependency: github.com/example/unused",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tidy output missing %q\noutput:\n%s", want, got)
		}
	}
	content, err := os.ReadFile("pe.mod")
	if err != nil {
		t.Fatalf("reading pe.mod: %v", err)
	}
	if !strings.Contains(string(content), "github.com/example/unused v1.0.0") {
		t.Fatalf("dry-run tidy changed pe.mod:\n%s", content)
	}
}

func TestModTidyCmd_WriteUpdatesSafeDeps(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	oldWrite := modTidyWrite
	modTidyWrite = true
	defer func() { modTidyWrite = oldWrite }()

	err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

require (
	github.com/example/keep v1.0.0
	github.com/example/remove v1.0.0
)
`), 0644)
	if err != nil {
		t.Fatalf("writing pe.mod: %v", err)
	}
	files := map[string]string{
		"keep.prompt":    "pe://github.com/example/keep/run\n",
		"add.prompt":     "pe://github.com/example/add@v1.2.3/run\n",
		"noversion.yaml": "prompt: pe://github.com/example/noversion/run\n",
	}
	for name, body := range files {
		if err := os.WriteFile(name, []byte(body), 0644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
	if err := os.Mkdir(".hidden", 0755); err != nil {
		t.Fatalf("creating hidden dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(".hidden", "ignored.prompt"), []byte("pe://github.com/example/hidden@v1.0.0/run\n"), 0644); err != nil {
		t.Fatalf("writing hidden prompt: %v", err)
	}

	var out bytes.Buffer
	cmd := *modTidyCmd
	cmd.SetOut(&out)
	if err := runModTidy(&cmd, nil); err != nil {
		t.Fatalf("runModTidy: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"removed dependency: github.com/example/remove",
		"added dependency: github.com/example/add v1.2.3",
		"skipped dependency: github.com/example/noversion (no single explicit version in references)",
		"pe.mod updated",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tidy output missing %q\noutput:\n%s", want, got)
		}
	}
	content, err := os.ReadFile("pe.mod")
	if err != nil {
		t.Fatalf("reading pe.mod: %v", err)
	}
	want := `module example.com/app

pe 1

require (
	github.com/example/add v1.2.3
	github.com/example/keep v1.0.0
)
`
	if string(content) != want {
		t.Fatalf("pe.mod mismatch\nwant:\n%s\ngot:\n%s", want, content)
	}
}

func TestModuleFromPERef(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want string
		ok   bool
	}{
		{
			name: "simple module",
			ref:  "pe://basic/speak-like-a-pirate",
			want: "basic",
			ok:   true,
		},
		{
			name: "github module",
			ref:  "pe://github.com/example/prompts/review)",
			want: "github.com/example/prompts",
			ok:   true,
		},
		{
			name: "github module with version",
			ref:  "pe://github.com/example/prompts@v1.2.3/review",
			want: "github.com/example/prompts",
			ok:   true,
		},
		{
			name: "bad scheme",
			ref:  "http://github.com/example/prompts/review",
			ok:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := moduleFromPERef(tt.ref)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("moduleFromPERef(%q) = %q, %v; want %q, %v", tt.ref, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestModVendorCmd_NoPeMod(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Should error when no pe.mod exists
	err := runModVendor(modVendorCmd, nil)
	if err == nil {
		t.Error("Expected error when pe.mod doesn't exist")
	}
}

func TestModDownloadCmd_NoPeMod(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Should error when no pe.mod exists
	err := runModDownload(modDownloadCmd, nil)
	if err == nil {
		t.Error("Expected error when pe.mod doesn't exist")
	}
}

func TestModListCmd_Structure(t *testing.T) {
	if modListCmd.Use != "list" {
		t.Errorf("Unexpected Use: %s", modListCmd.Use)
	}
}

func TestModGetCmd_Structure(t *testing.T) {
	if modGetCmd.Use != "get [module]" {
		t.Errorf("Unexpected Use: %s", modGetCmd.Use)
	}
}

func TestModSearchCmd_Structure(t *testing.T) {
	if modSearchCmd.Use != "search [query]" {
		t.Errorf("Unexpected Use: %s", modSearchCmd.Use)
	}
}

func TestModPublishCmd_Structure(t *testing.T) {
	if modPublishCmd.Use != "publish" {
		t.Errorf("Unexpected Use: %s", modPublishCmd.Use)
	}
}
