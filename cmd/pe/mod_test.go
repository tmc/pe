package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/module"
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
	subcommands := []string{"init", "list", "get", "download", "tidy", "vendor", "verify", "search", "publish", "vet"}
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

func TestModTidyCmd_JSONReportsWriteDecisions(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	oldWrite, oldJSON := modTidyWrite, modTidyJSON
	modTidyWrite = true
	modTidyJSON = true
	defer func() {
		modTidyWrite = oldWrite
		modTidyJSON = oldJSON
	}()

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
		"add.prompt":        "pe://github.com/example/add@v1.2.3/run\n",
		"conflict-a.prompt": "pe://github.com/example/conflict@v1.0.0/run\n",
		"conflict-b.prompt": "pe://github.com/example/conflict@v1.1.0/run\n",
		"keep.prompt":       "pe://github.com/example/keep/run\n",
		"noversion.yaml":    "prompt: pe://github.com/example/noversion/run\n",
	}
	for name, body := range files {
		if err := os.WriteFile(name, []byte(body), 0644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}

	var out bytes.Buffer
	cmd := *modTidyCmd
	cmd.SetOut(&out)
	if err := runModTidy(&cmd, nil); err != nil {
		t.Fatalf("runModTidy: %v", err)
	}
	if strings.Contains(out.String(), "Analyzing prompt dependencies") {
		t.Fatalf("JSON output contains text report:\n%s", out.String())
	}
	var got modTidyReport
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decoding JSON report: %v\n%s", err, out.String())
	}
	if got.References != 5 || got.Files != 5 || !got.Updated {
		t.Fatalf("report counts/updated = %+v", got)
	}
	if len(got.Missing) != 3 ||
		got.Missing[0].Module != "github.com/example/add" ||
		got.Missing[1].Module != "github.com/example/conflict" ||
		got.Missing[2].Module != "github.com/example/noversion" {
		t.Fatalf("missing order = %+v", got.Missing)
	}
	if len(got.Added) != 1 || got.Added[0] != (modTidyRequirement{Module: "github.com/example/add", Version: "v1.2.3"}) {
		t.Fatalf("added = %+v", got.Added)
	}
	if len(got.Removed) != 1 || got.Removed[0] != "github.com/example/remove" {
		t.Fatalf("removed = %+v", got.Removed)
	}
	if len(got.Skipped) != 2 ||
		got.Skipped[0].Module != "github.com/example/conflict" ||
		got.Skipped[1].Module != "github.com/example/noversion" {
		t.Fatalf("skipped = %+v", got.Skipped)
	}
	content, err := os.ReadFile("pe.mod")
	if err != nil {
		t.Fatalf("reading pe.mod: %v", err)
	}
	if strings.Contains(string(content), "github.com/example/conflict") ||
		strings.Contains(string(content), "github.com/example/noversion") {
		t.Fatalf("unsafe refs were written:\n%s", content)
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

func TestModVerifyCmd_VerifiesChecksums(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeVerifyPeMod(t)
	dir := filepath.Join(".pe", "cache", "modules", "example.com", "mod@v1.0.0")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prompt.pe"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	sum, err := moduleDirectoryChecksum(dir)
	if err != nil {
		t.Fatal(err)
	}
	writeModuleJSON(t, dir, sum)

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := runModVerify(cmd, nil); err != nil {
		t.Fatalf("runModVerify: %v", err)
	}
	if !strings.Contains(out.String(), "verified example.com/mod@v1.0.0") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestModVerifyCmd_DetectsTamper(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeVerifyPeMod(t)
	dir := filepath.Join(".pe", "cache", "modules", "example.com", "mod@v1.0.0")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(dir, "prompt.pe")
	if err := os.WriteFile(name, []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	sum, err := moduleDirectoryChecksum(dir)
	if err != nil {
		t.Fatal(err)
	}
	writeModuleJSON(t, dir, sum)
	if err := os.WriteFile(name, []byte("changed\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err = runModVerify(&cobra.Command{}, nil)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("runModVerify error = %v, want checksum mismatch", err)
	}
}

func TestModVerifyCmd_MissingModule(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeVerifyPeMod(t)
	err := runModVerify(&cobra.Command{}, nil)
	if err == nil || !strings.Contains(err.Error(), "not found in cache") {
		t.Fatalf("runModVerify error = %v, want missing cache error", err)
	}
}

func TestModuleDirectoryChecksumRejectsSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "target"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", filepath.Join(tmpDir, "link")); err != nil {
		t.Skipf("symlink not available: %v", err)
	}
	_, err := moduleDirectoryChecksum(tmpDir)
	if err == nil || !strings.Contains(err.Error(), "refusing symlink") {
		t.Fatalf("moduleDirectoryChecksum error = %v, want symlink error", err)
	}
}

func writeVerifyPeMod(t *testing.T) {
	t.Helper()
	data := []byte("module example.com/app\n\npe 1\n\nrequire example.com/mod v1.0.0\n")
	if err := os.WriteFile("pe.mod", data, 0644); err != nil {
		t.Fatal(err)
	}
}

func writeModuleJSON(t *testing.T, dir, sum string) {
	t.Helper()
	data, err := json.Marshal(module.Module{
		Name:     "example.com/mod",
		Version:  "v1.0.0",
		Checksum: sum,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "module.json"), data, 0644); err != nil {
		t.Fatal(err)
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

func TestRunModVetCapabilities(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	mod := `module example.com/prompts

pe 1

capability {
    providers deny remote
    tools deny shell network
}

policy {
    composition strict
    require-typed-io true
}
`
	if err := os.WriteFile("pe.mod", []byte(mod), 0644); err != nil {
		t.Fatal(err)
	}
	prompt := `---
kind: pe.text.v1
inputs:
  topic:
    type: string
safety:
  providers:
    allow: [local]
---
Review {{ .topic }}.
`
	if err := os.WriteFile("review.prompt", []byte(prompt), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{}
	if err := runModVet(cmd, []string{"review.prompt"}); err != nil {
		t.Fatalf("mod vet failed: %v", err)
	}
}

func TestRunModVetDeniedProvider(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	mod := `pe 1

capability {
    providers deny remote
}
`
	if err := os.WriteFile("pe.mod", []byte(mod), 0644); err != nil {
		t.Fatal(err)
	}
	prompt := `---
kind: pe.text.v1
safety:
  providers:
    allow: [remote]
---
Review.
`
	if err := os.WriteFile("review.prompt", []byte(prompt), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{}
	err := runModVet(cmd, []string{"review.prompt"})
	if err == nil || !strings.Contains(err.Error(), "provider remote is denied") {
		t.Fatalf("mod vet error = %v", err)
	}
}
