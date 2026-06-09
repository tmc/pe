package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
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
	subcommands := []string{"init", "list", "get", "download", "tidy", "vendor", "verify", "graph", "upgrade", "audit", "search", "publish", "vet"}
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

func TestModInitCmd_ForceOverwriteDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	const original = `module existing

pe 1

capability {
    tools deny write
}
`
	if err := os.WriteFile("pe.mod", []byte(original), 0644); err != nil {
		t.Fatalf("writing pe.mod: %v", err)
	}

	modForce = true
	defer func() { modForce = false }()

	err := runModInit(modInitCmd, []string{"example.com/new-module"})
	if err == nil {
		t.Fatal("runModInit succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("runModInit error = %v, want write policy error", err)
	}

	data, err := os.ReadFile("pe.mod")
	if err != nil {
		t.Fatalf("reading pe.mod: %v", err)
	}
	if string(data) != original {
		t.Fatalf("pe.mod changed after denied init:\n%s", data)
	}
	if _, err := os.Stat(".pe"); !os.IsNotExist(err) {
		t.Fatalf(".pe stat error = %v, want not exist", err)
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

func TestModTidyCmd_WriteDeniedByPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	oldWrite := modTidyWrite
	modTidyWrite = true
	defer func() { modTidyWrite = oldWrite }()

	const original = `module example.com/app

pe 1

require github.com/example/remove v1.0.0

capability {
    tools deny write
}
`
	if err := os.WriteFile("pe.mod", []byte(original), 0644); err != nil {
		t.Fatalf("writing pe.mod: %v", err)
	}
	if err := os.WriteFile("add.prompt", []byte("pe://github.com/example/add@v1.2.3/run\n"), 0644); err != nil {
		t.Fatalf("writing add.prompt: %v", err)
	}

	var out bytes.Buffer
	cmd := *modTidyCmd
	cmd.SetOut(&out)
	err := runModTidy(&cmd, nil)
	if err == nil {
		t.Fatal("runModTidy succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("runModTidy error = %v, want write policy error", err)
	}
	data, err := os.ReadFile("pe.mod")
	if err != nil {
		t.Fatalf("reading pe.mod: %v", err)
	}
	if string(data) != original {
		t.Fatalf("pe.mod changed after denied tidy:\n%s", data)
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

func TestModVendorCmd_CopiesCachedModule(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeVerifyPeMod(t)
	srcDir := filepath.Join(".pe", "cache", "modules", "example.com", "mod@v1.0.0")
	nestedDir := filepath.Join(srcDir, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "prompt.pe"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "extra.pe"), []byte("extra\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := runModVendor(modVendorCmd, nil); err != nil {
		t.Fatalf("runModVendor: %v", err)
	}
	got, err := os.ReadFile(filepath.Join("vendor", "example.com", "mod@v1.0.0", "nested", "extra.pe"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "extra\n" {
		t.Fatalf("vendored content = %q", got)
	}
	modules, err := os.ReadFile(filepath.Join("vendor", "modules.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(modules)) != "example.com/mod@v1.0.0" {
		t.Fatalf("modules.txt = %q", modules)
	}
}

func TestModVendorCmd_SkipsMissingCachedModule(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeVerifyPeMod(t)
	if err := runModVendor(modVendorCmd, nil); err != nil {
		t.Fatalf("runModVendor: %v", err)
	}
	modules, err := os.ReadFile(filepath.Join("vendor", "modules.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(modules)) != "" {
		t.Fatalf("modules.txt = %q", modules)
	}
	if _, err := os.Stat(filepath.Join("vendor", "example.com", "mod@v1.0.0")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing module was vendored: %v", err)
	}
}

func TestModVendorCmd_DeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	const mod = `module example.com/app

pe 1

require example.com/mod v1.0.0

capability {
    tools deny write
}
`
	if err := os.WriteFile("pe.mod", []byte(mod), 0644); err != nil {
		t.Fatalf("writing pe.mod: %v", err)
	}

	err := runModVendor(modVendorCmd, nil)
	if err == nil {
		t.Fatal("runModVendor succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("runModVendor error = %v, want write policy error", err)
	}
	if _, err := os.Stat("vendor"); !os.IsNotExist(err) {
		t.Fatalf("vendor stat error = %v, want not exist", err)
	}
}

func TestCopyModuleDirRejectsSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src")
	dst := filepath.Join(tmpDir, "dst")
	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "target"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", filepath.Join(src, "link")); err != nil {
		t.Skipf("symlink not available: %v", err)
	}
	err := copyModuleDir(src, dst)
	if err == nil || !strings.Contains(err.Error(), "refusing symlink") {
		t.Fatalf("copyModuleDir error = %v, want symlink error", err)
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

func TestModDownloadCmd_DownloadsToCanonicalCache(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	registryDir := filepath.Join(tmpDir, "registry")
	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "prompt.pe"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mod := &module.Module{
		Name:    "example.com/mod",
		Version: "v1.0.0",
		Files:   []string{"prompt.pe"},
	}
	if err := module.NewLocalRegistry(registryDir).Publish(mod, sourceDir); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	t.Setenv("PE_REGISTRY_TYPE", "local")
	t.Setenv("PE_REGISTRY_DIR", registryDir)
	if err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

require example.com/mod v1.0.0
`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := runModDownload(modDownloadCmd, nil); err != nil {
		t.Fatalf("runModDownload: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(".pe", "cache", "modules", "example.com", "mod@v1.0.0", "prompt.pe"))
	if err != nil {
		t.Fatalf("read cached prompt: %v", err)
	}
	if string(got) != "hello\n" {
		t.Fatalf("cached prompt = %q", got)
	}
	if _, err := os.Stat(filepath.Join(".pe", "cache", "modules", "example.com", "mod@v1.0.0", "module.json")); err != nil {
		t.Fatalf("missing cached module metadata: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".pe", "cache", "example.com", "mod", "v1.0.0", "module.json")); err != nil {
		t.Fatalf("missing resolver cache metadata: %v", err)
	}
}

func TestModDownloadCmd_FailsOnMissingModule(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	t.Setenv("PE_REGISTRY_TYPE", "local")
	t.Setenv("PE_REGISTRY_DIR", filepath.Join(tmpDir, "registry"))
	if err := os.MkdirAll(filepath.Join(tmpDir, "registry"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

require example.com/missing v1.0.0
`), 0644); err != nil {
		t.Fatal(err)
	}

	err := runModDownload(modDownloadCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "getting module example.com/missing") {
		t.Fatalf("runModDownload error = %v, want missing module error", err)
	}
	if _, statErr := os.Stat(filepath.Join(".pe", "cache", "modules", "example.com", "missing@v1.0.0", "module.info")); !os.IsNotExist(statErr) {
		t.Fatalf("placeholder module.info exists or stat failed differently: %v", statErr)
	}
}

func TestModDownloadCmd_FailsOnVersionMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	registryDir := filepath.Join(tmpDir, "registry")
	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "prompt.pe"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mod := &module.Module{
		Name:    "example.com/mod",
		Version: "v1.0.0",
		Files:   []string{"prompt.pe"},
	}
	if err := module.NewLocalRegistry(registryDir).Publish(mod, sourceDir); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	t.Setenv("PE_REGISTRY_TYPE", "local")
	t.Setenv("PE_REGISTRY_DIR", registryDir)
	if err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

require example.com/mod v1.2.0
`), 0644); err != nil {
		t.Fatal(err)
	}

	err := runModDownload(modDownloadCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "registry returned version v1.0.0, want v1.2.0") {
		t.Fatalf("runModDownload error = %v, want version mismatch", err)
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
	sum, err := module.DirectoryChecksum(dir)
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
	sum, err := module.DirectoryChecksum(dir)
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
	_, err := module.DirectoryChecksum(tmpDir)
	if err == nil || !strings.Contains(err.Error(), "refusing symlink") {
		t.Fatalf("DirectoryChecksum error = %v, want symlink error", err)
	}
}

func TestModGraphCmd_PrintsCachedDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeVerifyPeMod(t)
	modDir := filepath.Join(".pe", "cache", "modules", "example.com", "mod@v1.0.0")
	depDir := filepath.Join(".pe", "cache", "modules", "example.com", "dep@v1.2.0")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(depDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeModuleJSONWithDependencies(t, modDir, map[string]string{"example.com/dep": "v1.2.0"})
	writeModuleJSONWithDependencies(t, depDir, nil)

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := runModGraph(cmd, nil); err != nil {
		t.Fatalf("runModGraph: %v", err)
	}
	got := strings.TrimSpace(out.String())
	want := strings.Join([]string{
		"example.com/app example.com/mod@v1.0.0",
		"example.com/mod@v1.0.0 example.com/dep@v1.2.0",
	}, "\n")
	if got != want {
		t.Fatalf("graph = %q, want %q", got, want)
	}
}

func TestModUpgradeCmd_UpdatesRequirements(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	t.Setenv("PE_REGISTRY_TYPE", "local")
	t.Setenv("PE_REGISTRY_DIR", filepath.Join(tmpDir, "registry"))
	registry := module.NewLocalRegistry(filepath.Join(tmpDir, "registry"))
	if err := registry.Publish(&module.Module{Name: "example.com/mod", Version: "v1.2.0"}, t.TempDir()); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	data := []byte("module example.com/app\n\npe 1\n\nrequire example.com/mod v1.0.0\n")
	if err := os.WriteFile("pe.mod", data, 0644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := runModUpgrade(cmd, nil); err != nil {
		t.Fatalf("runModUpgrade: %v", err)
	}
	if !strings.Contains(out.String(), "upgraded example.com/mod v1.0.0 => v1.2.0") {
		t.Fatalf("output = %q", out.String())
	}
	got, err := os.ReadFile("pe.mod")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "require example.com/mod v1.2.0") {
		t.Fatalf("pe.mod = %s", got)
	}
}

func TestModUpgradeCmd_DeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	t.Setenv("PE_REGISTRY_TYPE", "local")
	t.Setenv("PE_REGISTRY_DIR", filepath.Join(tmpDir, "registry"))
	registry := module.NewLocalRegistry(filepath.Join(tmpDir, "registry"))
	if err := registry.Publish(&module.Module{Name: "example.com/mod", Version: "v1.2.0"}, t.TempDir()); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	const original = `module example.com/app

pe 1

require example.com/mod v1.0.0

capability {
    tools deny write
}
`
	if err := os.WriteFile("pe.mod", []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	err := runModUpgrade(cmd, nil)
	if err == nil {
		t.Fatal("runModUpgrade succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("runModUpgrade error = %v, want write policy error", err)
	}
	if out.String() != "" {
		t.Fatalf("runModUpgrade output = %q, want none", out.String())
	}
	got, err := os.ReadFile("pe.mod")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Fatalf("pe.mod changed after denied upgrade:\n%s", got)
	}
}

func TestModUpgradeCmd_RejectsUnrequiredModule(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	t.Setenv("PE_REGISTRY_TYPE", "local")
	t.Setenv("PE_REGISTRY_DIR", filepath.Join(tmpDir, "registry"))
	if err := os.WriteFile("pe.mod", []byte("module example.com/app\n\npe 1\n\nrequire example.com/mod v1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	err := runModUpgrade(&cobra.Command{}, []string{"example.com/other"})
	if err == nil || !strings.Contains(err.Error(), "is not required") {
		t.Fatalf("runModUpgrade error = %v, want not required", err)
	}
}

func TestModAuditCmd_OK(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte("module example.com/app\n\npe 1\n\nrequire example.com/mod v1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := runModAudit(cmd, nil); err != nil {
		t.Fatalf("runModAudit: %v", err)
	}
	if strings.TrimSpace(out.String()) != "module audit ok" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestModAuditCmd_RequiresTrustedSignatures(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

require example.com/mod v1.0.0

security {
	require-signatures true
}
`), 0644); err != nil {
		t.Fatal(err)
	}
	err := runModAudit(&cobra.Command{}, nil)
	if err == nil || !strings.Contains(err.Error(), "module audit found 1 issue") {
		t.Fatalf("runModAudit error = %v, want signature finding", err)
	}
}

func TestModAuditCmd_FindsVulnerabilities(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte("module example.com/app\n\npe 1\n\nrequire example.com/mod v1.0.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	db := filepath.Join(tmpDir, "vulns.json")
	if err := os.WriteFile(db, []byte(`[{"module":"example.com/mod","version":"v1.0.0","id":"PE-1","severity":"high","summary":"bad prompt"}]`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PE_VULN_DB", db)
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	err := runModAudit(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "module audit found 1 issue") {
		t.Fatalf("runModAudit error = %v, want vulnerability finding", err)
	}
	if !strings.Contains(out.String(), "vulnerability PE-1 (high): bad prompt") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestModuleRegistrySeed(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "examples", "modules", "templates", "registry-seed", "modules.json"))
	if err != nil {
		t.Fatal(err)
	}
	var modules []module.Module
	if err := json.Unmarshal(data, &modules); err != nil {
		t.Fatalf("decode registry seed: %v", err)
	}
	if len(modules) == 0 {
		t.Fatal("registry seed is empty")
	}
	for _, mod := range modules {
		if mod.Name == "" || mod.Version == "" || len(mod.Files) == 0 {
			t.Fatalf("incomplete registry seed module: %+v", mod)
		}
		for _, file := range mod.Files {
			if _, err := os.Stat(filepath.Join("..", "..", "examples", "modules", "templates", "basic-module", file)); err != nil {
				t.Fatalf("seed file %s: %v", file, err)
			}
		}
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
	writeModuleJSONWithDependencies(t, dir, nil, sum)
}

func writeModuleJSONWithDependencies(t *testing.T, dir string, deps map[string]string, checksum ...string) {
	t.Helper()
	sum := ""
	if len(checksum) > 0 {
		sum = checksum[0]
	}
	data, err := json.Marshal(module.Module{
		Name:         "example.com/mod",
		Version:      "v1.0.0",
		Dependencies: deps,
		Checksum:     sum,
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

func TestRunModVetRequireTypedIORejectsPlainText(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	mod := `module example.com/prompts

pe 1

policy {
    composition strict
    require-typed-io true
}
`
	if err := os.WriteFile("pe.mod", []byte(mod), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("plain.prompt", []byte("plain text prompt\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{}
	err := runModVet(cmd, []string{"plain.prompt"})
	if err == nil || !strings.Contains(err.Error(), "policy requires typed inputs") {
		t.Fatalf("runModVet error = %v, want typed inputs error", err)
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

func TestRunModVetStrictDependencyPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	mod := `module example.com/prompts

pe 1

require example.com/base v1.0.0

capability {
    providers deny remote
    tools deny shell
}

placement {
    network false
}

policy {
    composition strict
}
`
	if err := os.WriteFile("pe.mod", []byte(mod), 0644); err != nil {
		t.Fatal(err)
	}
	depDir := filepath.Join(".pe", "cache", "modules", "example.com", "base@v1.0.0")
	if err := os.MkdirAll(depDir, 0755); err != nil {
		t.Fatal(err)
	}
	dep := `module example.com/base

pe 1

capability {
    providers allow local
    tools allow read
}

placement {
    network false
}
`
	if err := os.WriteFile(filepath.Join(depDir, "pe.mod"), []byte(dep), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{}
	if err := runModVet(cmd, nil); err != nil {
		t.Fatalf("runModVet: %v", err)
	}
}

func TestRunModVetStrictDependencyDeniedProvider(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	mod := `module example.com/prompts

pe 1

require example.com/base v1.0.0

capability {
    providers deny remote
}

policy {
    composition strict
}
`
	if err := os.WriteFile("pe.mod", []byte(mod), 0644); err != nil {
		t.Fatal(err)
	}
	depDir := filepath.Join(".pe", "cache", "modules", "example.com", "base@v1.0.0")
	if err := os.MkdirAll(depDir, 0755); err != nil {
		t.Fatal(err)
	}
	dep := `module example.com/base

pe 1

capability {
    providers allow remote
}
`
	if err := os.WriteFile(filepath.Join(depDir, "pe.mod"), []byte(dep), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{}
	err := runModVet(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "dependency provider remote is denied") {
		t.Fatalf("runModVet error = %v, want dependency provider denial", err)
	}
}

func TestRunModVetStrictDependencyDeniedNetwork(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	mod := `module example.com/prompts

pe 1

require example.com/base v1.0.0

placement {
    network false
}

policy {
    composition strict
}
`
	if err := os.WriteFile("pe.mod", []byte(mod), 0644); err != nil {
		t.Fatal(err)
	}
	depDir := filepath.Join(".pe", "cache", "modules", "example.com", "base@v1.0.0")
	if err := os.MkdirAll(depDir, 0755); err != nil {
		t.Fatal(err)
	}
	dep := `module example.com/base

pe 1

placement {
    network true
}
`
	if err := os.WriteFile(filepath.Join(depDir, "pe.mod"), []byte(dep), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{}
	err := runModVet(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "dependency network access is denied") {
		t.Fatalf("runModVet error = %v, want dependency network denial", err)
	}
}
