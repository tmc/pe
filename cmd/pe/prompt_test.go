package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromptCmd_CommandStructure(t *testing.T) {
	if promptCmd.Use != "prompt" {
		t.Errorf("Unexpected Use: %s", promptCmd.Use)
	}

	if promptCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Verify subcommands exist
	subcommandNames := []string{"init", "edit", "info", "tidy", "help"}
	for _, name := range subcommandNames {
		found := false
		for _, cmd := range promptCmd.Commands() {
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

func TestPromptInitCmd_DefaultFilename(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := promptInitCmd()
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should create prompt.prompt
	content, err := os.ReadFile(filepath.Join(tmpDir, "prompt.prompt"))
	if err != nil {
		t.Fatalf("Expected prompt.prompt to be created: %v", err)
	}

	// Should contain shebang
	if !strings.HasPrefix(string(content), "#!/usr/bin/env pe run") {
		t.Error("Expected shebang in prompt file")
	}
}

func TestPromptInitCmd_CustomFilename(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := promptInitCmd()
	err := cmd.RunE(cmd, []string{"custom"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should create custom.prompt (adds extension)
	if _, err := os.Stat(filepath.Join(tmpDir, "custom.prompt")); os.IsNotExist(err) {
		t.Error("Expected custom.prompt to be created")
	}
}

func TestPromptInitCmd_WithExtension(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := promptInitCmd()
	err := cmd.RunE(cmd, []string{"test.prompt"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should create test.prompt
	if _, err := os.Stat(filepath.Join(tmpDir, "test.prompt")); os.IsNotExist(err) {
		t.Error("Expected test.prompt to be created")
	}
}

func TestPromptInitCmd_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create existing file
	err := os.WriteFile(filepath.Join(tmpDir, "existing.prompt"), []byte("existing"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	cmd := promptInitCmd()
	err = cmd.RunE(cmd, []string{"existing"})
	if err == nil {
		t.Error("Expected error when file exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %v", err)
	}
}

func TestPromptInitCmd_DeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeDenyWritePeMod(t)

	cmd := promptInitCmd()
	err := cmd.RunE(cmd, []string{"blocked"})
	if err == nil {
		t.Fatal("prompt init succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("prompt init error = %v, want write policy error", err)
	}
	if _, err := os.Stat("blocked.prompt"); !os.IsNotExist(err) {
		t.Fatalf("blocked.prompt stat error = %v, want not exist", err)
	}
}

func TestPromptInitCmd_Flags(t *testing.T) {
	cmd := promptInitCmd()

	// Check flag existence
	flags := []string{"provider", "system", "with-defaults", "with-tests", "with-variants", "force"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestPromptInfoCmd_Structure(t *testing.T) {
	cmd := promptInfoCmd()
	if cmd.Use != "info [file]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestPromptInfoCmd_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := promptInfoCmd()
	err := cmd.RunE(cmd, []string{"nonexistent.prompt"})
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestPromptInfoCmd_Success(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create a test prompt
	content := `#!/usr/bin/env pe run
Hello {{.name}}!

-- defaults --
name = World
`
	err := os.WriteFile(filepath.Join(tmpDir, "test.prompt"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	cmd := promptInfoCmd()
	err = cmd.RunE(cmd, []string{"test.prompt"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPromptTidyCmd_Structure(t *testing.T) {
	cmd := promptTidyCmd()
	if cmd.Use != "tidy [file]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestPromptEditCmd_Structure(t *testing.T) {
	cmd := promptEditCmd()
	if cmd.Use != "edit [file]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestPromptHelpCmd_Structure(t *testing.T) {
	cmd := promptHelpCmd()
	if !strings.HasPrefix(cmd.Use, "help") {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestPromptInitCmd_AllOptionsAndForce(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := promptInitCmd()
	cmd.Flags().Set("provider", "openai")
	cmd.Flags().Set("system", "You are helpful")
	cmd.Flags().Set("with-tests", "true")
	cmd.Flags().Set("with-variants", "true")
	if err := cmd.RunE(cmd, []string{"full.prompt"}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile("full.prompt")
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, want := range []string{"--provider=openai", "---defaults---", "---system---", "---tests---", "---variants---"} {
		if !strings.Contains(text, want) {
			t.Fatalf("content missing %q:\n%s", want, text)
		}
	}
	cmd = promptInitCmd()
	cmd.Flags().Set("force", "true")
	if err := cmd.RunE(cmd, []string{"full.prompt"}); err != nil {
		t.Fatal(err)
	}
}

func TestPromptEditCmd_UpdatesProviderAndDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "edit.prompt")
	content := `#!/usr/bin/env pe run --provider=cgpt
Hello {{.name}}

---defaults---
name: World
unused: gone

---tests---
- name: smoke
`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := promptEditCmd()
	cmd.Flags().Set("set-provider", "openai")
	cmd.Flags().Set("set-default", "name=Alice,task=Explain")
	cmd.Flags().Set("remove-default", "unused")
	if err := cmd.RunE(cmd, []string{file}); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	for _, want := range []string{"--provider=openai", "name: Alice", "task: Explain", "---tests---"} {
		if !strings.Contains(text, want) {
			t.Fatalf("edited content missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "unused:") {
		t.Fatalf("unused default still present:\n%s", text)
	}

	noDefaults := filepath.Join(tmpDir, "new-defaults.prompt")
	if err := os.WriteFile(noDefaults, []byte("#!/usr/bin/env pe run\nHello {{.task}}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd = promptEditCmd()
	cmd.Flags().Set("set-default", "task=Write")
	if err := cmd.RunE(cmd, []string{noDefaults}); err != nil {
		t.Fatal(err)
	}
	out, err = os.ReadFile(noDefaults)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "---defaults---") || !strings.Contains(string(out), "task: Write") {
		t.Fatalf("new defaults content:\n%s", out)
	}
}

func TestPromptEditCmd_DeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeDenyWritePeMod(t)
	const original = `#!/usr/bin/env pe run --provider=cgpt
Hello {{.name}}

---defaults---
name: World
`
	if err := os.WriteFile("edit.prompt", []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := promptEditCmd()
	cmd.Flags().Set("set-provider", "openai")
	err := cmd.RunE(cmd, []string{"edit.prompt"})
	if err == nil {
		t.Fatal("prompt edit succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("prompt edit error = %v, want write policy error", err)
	}
	out, err := os.ReadFile("edit.prompt")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != original {
		t.Fatalf("prompt changed after denied edit:\n%s", out)
	}
}

func TestPromptInfoTidyHelpAndHelpers(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "info.prompt")
	content := `#!/usr/bin/env pe run --provider=openai
# Describe line one
# Describe line two
Hello {{.NAME}} {{.NAME | printf "%s"}} {{.TASK}}

---defaults---
NAME: World
UNUSED: remove

---system---
System prompt

---tests---
- name: smoke

---variants---
fast:
  Be brief
`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	info := analyzePromptFile(content)
	if info["executable"] != true || info["provider"] != "openai" || info["test_count"] != 1 {
		t.Fatalf("info = %#v", info)
	}
	if got := extractDescription(content); got != "Describe line one\nDescribe line two" {
		t.Fatalf("description = %q", got)
	}
	if defaults := extractDefaults(content); defaults["NAME"] != "World" || defaults["UNUSED"] != "remove" {
		t.Fatalf("defaults = %#v", defaults)
	}
	vars := extractPromptVariables(content)
	if !contains(vars, "NAME") || !contains(vars, "TASK") {
		t.Fatalf("vars = %#v", vars)
	}
	if got := findVariables(`{{.A}} {{.A}} {{.B | printf "%s"}}`); len(got) != 2 || !contains(got, "A") || !contains(got, "B") {
		t.Fatalf("findVariables = %#v", got)
	}
	if got := trimEmptyLines([]string{"", " a ", "", ""}); len(got) != 1 || got[0] != " a " {
		t.Fatalf("trimmed = %#v", got)
	}
	if main := extractMainContent(strings.Split(content, "\n")); !strings.Contains(main, "Hello") || strings.Contains(main, "---defaults---") {
		t.Fatalf("main = %q", main)
	}

	cmd := promptInfoCmd()
	cmd.Flags().Set("verbose", "true")
	if err := cmd.RunE(cmd, []string{file}); err != nil {
		t.Fatal(err)
	}
	cmd = promptInfoCmd()
	cmd.Flags().Set("json", "true")
	if err := cmd.RunE(cmd, []string{file}); err != nil {
		t.Fatal(err)
	}
	cmd = promptTidyCmd()
	cmd.Flags().Set("remove-unused", "true")
	if err := cmd.RunE(cmd, []string{file}); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "UNUSED:") || !strings.Contains(string(out), "NAME: World") {
		t.Fatalf("tidied content:\n%s", out)
	}
	cmd = promptHelpCmd()
	if err := cmd.RunE(cmd, []string{file}); err != nil {
		t.Fatal(err)
	}
	cmd = promptHelpCmd()
	cmd.Flags().Set("as-script", "true")
	if err := cmd.RunE(cmd, []string{file}); err != nil {
		t.Fatal(err)
	}
	if err := tidyPromptFile(filepath.Join(tmpDir, "missing.prompt"), false, true); err == nil {
		t.Fatal("missing tidy succeeded")
	}
}

func TestTidyPromptFile_RemoveUnusedDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeDenyWritePeMod(t)
	const original = `Hello {{.NAME}}

---defaults---
NAME: World
UNUSED: remove
`
	if err := os.WriteFile("tidy.prompt", []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	err := tidyPromptFile("tidy.prompt", true, true)
	if err == nil {
		t.Fatal("tidyPromptFile succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("tidyPromptFile error = %v, want write policy error", err)
	}
	out, err := os.ReadFile("tidy.prompt")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != original {
		t.Fatalf("prompt changed after denied tidy:\n%s", out)
	}
}

func writeDenyWritePeMod(t *testing.T) {
	t.Helper()
	if err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

capability {
    tools deny write
}
`), 0644); err != nil {
		t.Fatalf("writing pe.mod: %v", err)
	}
}
