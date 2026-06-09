package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/templates"
)

func TestTemplateCmd_CommandStructure(t *testing.T) {
	cmd := templateCmd()

	if cmd.Use != "template" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Verify subcommands exist
	subcommandNames := []string{"list", "search", "show", "apply", "create", "validate", "export", "import", "interactive"}
	for _, name := range subcommandNames {
		found := false
		for _, c := range cmd.Commands() {
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

func TestTemplateListCmd_FlagParsing(t *testing.T) {
	cmd := templateListCmd()

	flags := []string{"category", "tag", "format"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestTemplateSearchCmd_CommandStructure(t *testing.T) {
	cmd := templateSearchCmd()

	if cmd.Use != "search [query]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestTemplateShowCmd_CommandStructure(t *testing.T) {
	cmd := templateShowCmd()

	if cmd.Use != "show [template_name]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestTemplateApplyCmd_CommandStructure(t *testing.T) {
	cmd := templateApplyCmd()

	if cmd.Use != "apply [template_name]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestTemplateCreateCmd_CommandStructure(t *testing.T) {
	cmd := templateCreateCmd()

	if cmd.Use != "create [template_name]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestTemplateValidateCmd_CommandStructure(t *testing.T) {
	cmd := templateValidateCmd()

	if cmd.Use != "validate [file...]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestTemplateExportCmd_CommandStructure(t *testing.T) {
	cmd := templateExportCmd()

	if cmd.Use != "export [template_name...]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestTemplateImportCmd_CommandStructure(t *testing.T) {
	cmd := templateImportCmd()

	if cmd.Use != "import [file...]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestTemplateInteractiveCmd_CommandStructure(t *testing.T) {
	cmd := templateInteractiveCmd()

	if cmd.Use != "interactive" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestTemplateListSearchShowAndOutput(t *testing.T) {
	cmd := templateTestCmd()
	for _, format := range []string{"table", "json", "yaml"} {
		if err := runTemplateList(cmd, "", "", format); err != nil {
			t.Fatalf("list %s: %v", format, err)
		}
		if err := runTemplateSearch(cmd, "summary", format); err != nil {
			t.Fatalf("search %s: %v", format, err)
		}
	}
	if err := runTemplateList(cmd, "writing", "", "table"); err != nil {
		t.Fatal(err)
	}
	if err := runTemplateList(cmd, "", "analysis", "table"); err != nil {
		t.Fatal(err)
	}
	if err := runTemplateList(cmd, "", "", "bad"); err == nil {
		t.Fatal("bad list format succeeded")
	}
	if err := runTemplateShow(cmd, "summarization", "json", false); err != nil {
		t.Fatal(err)
	}
	if err := runTemplateShow(cmd, "summarization", "yaml", true); err != nil {
		t.Fatal(err)
	}
	if err := runTemplateShow(cmd, "summarization", "bad", true); err == nil {
		t.Fatal("bad show format succeeded")
	}
	if err := runTemplateShow(cmd, "missing", "json", true); err == nil {
		t.Fatal("missing template showed")
	}
	out := cmd.OutOrStdout().(*bytes.Buffer).String()
	if !strings.Contains(out, "summarization") {
		t.Fatalf("output = %s", out)
	}
}

func TestTemplateApplyCreateValidateExportImport(t *testing.T) {
	tmpDir := t.TempDir()
	cmd := templateTestCmd()
	varsJSON := filepath.Join(tmpDir, "vars.json")
	if err := os.WriteFile(varsJSON, []byte(`{"text":"Long article","max_points":3}`), 0644); err != nil {
		t.Fatal(err)
	}
	outFile := filepath.Join(tmpDir, "applied.txt")
	if err := runTemplateApply(cmd, "summarization", varsJSON, outFile, false); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(outFile); err != nil || !strings.Contains(string(data), "Long article") {
		t.Fatalf("applied = %q err=%v", data, err)
	}
	varsYAML := filepath.Join(tmpDir, "vars.yaml")
	if err := os.WriteFile(varsYAML, []byte("text: YAML article\nmax_points: 2\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if vars, err := loadVariablesFromFile(varsYAML); err != nil || vars["text"] != "YAML article" {
		t.Fatalf("vars = %#v err=%v", vars, err)
	}
	if _, err := loadVariablesFromFile(filepath.Join(tmpDir, "missing.json")); err == nil {
		t.Fatal("missing vars loaded")
	}
	badVars := filepath.Join(tmpDir, "bad.vars")
	if err := os.WriteFile(badVars, []byte(":"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadVariablesFromFile(badVars); err == nil {
		t.Fatal("bad vars loaded")
	}
	if err := runTemplateApply(cmd, "summarization", "", "", false); err == nil {
		t.Fatal("apply without vars succeeded")
	}
	if err := runTemplateApply(cmd, "missing", varsJSON, "", false); err == nil {
		t.Fatal("missing apply succeeded")
	}

	createJSON := filepath.Join(tmpDir, "created.json")
	if err := runTemplateCreate(cmd, "mine", createJSON, true, "", "", "custom", nil); err != nil {
		t.Fatal(err)
	}
	createYAML := filepath.Join(tmpDir, "created.yaml")
	if err := runTemplateCreate(cmd, "mine-yaml", createYAML, true, "", "", "custom", nil); err != nil {
		t.Fatal(err)
	}
	createPrompt := filepath.Join(tmpDir, "created.prompt")
	if err := runTemplateCreate(cmd, "mine-prompt", createPrompt, false, "Analyze {{ .topic }} carefully.", "", "custom", []string{"analysis"}); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(createPrompt); err != nil || string(data) != "Analyze {{ .topic }} carefully." {
		t.Fatalf("created prompt = %q err=%v", data, err)
	}
	createNonInteractiveYAML := filepath.Join(tmpDir, "created-noninteractive.yaml")
	if err := runTemplateCreate(cmd, "mine-vars", createNonInteractiveYAML, false, "Summarize {{ .text }} for {{ .audience }}.", "Summarizer", "writing", []string{"summary"}); err != nil {
		t.Fatal(err)
	}
	created, err := templates.NewTemplateLibrary("").LoadTemplateFromFile(createNonInteractiveYAML)
	if err != nil {
		t.Fatal(err)
	}
	if created.Description != "Summarizer" || created.Category != "writing" {
		t.Fatalf("created template metadata = %#v", created)
	}
	if _, ok := created.Variables["text"]; !ok {
		t.Fatalf("created variables = %#v, want text", created.Variables)
	}
	if _, ok := created.Variables["audience"]; !ok {
		t.Fatalf("created variables = %#v, want audience", created.Variables)
	}
	dotlessYAML := filepath.Join(tmpDir, "dotless.yaml")
	if err := runTemplateCreate(cmd, "dotless", dotlessYAML, false, "Legacy {{ audience }} placeholder.", "", "custom", nil); err != nil {
		t.Fatal(err)
	}
	dotless, err := templates.NewTemplateLibrary("").LoadTemplateFromFile(dotlessYAML)
	if err != nil {
		t.Fatal(err)
	}
	if len(dotless.Variables) != 0 {
		t.Fatalf("dotless variables = %#v, want none", dotless.Variables)
	}
	if err := runTemplateCreate(cmd, "mine", "", false, "", "", "custom", nil); err == nil {
		t.Fatal("non-interactive create without prompt succeeded")
	}
	if err := runTemplateValidate(cmd, nil); err == nil {
		t.Fatal("validate without files succeeded")
	}
	if err := runTemplateValidate(cmd, []string{createYAML}); err != nil {
		t.Fatal(err)
	}
	invalid := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(invalid, []byte("name: bad\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runTemplateValidate(cmd, []string{invalid}); err == nil {
		t.Fatal("invalid template validated")
	}
	exportDir := filepath.Join(tmpDir, "export")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := runTemplateExport(cmd, []string{"summarization"}, "json", exportDir); err != nil {
		t.Fatal(err)
	}
	if err := runTemplateExport(cmd, nil, "yaml", exportDir); err != nil {
		t.Fatal(err)
	}
	if err := runTemplateExport(cmd, []string{"missing"}, "json", exportDir); err == nil {
		t.Fatal("missing export succeeded")
	}
	if err := runTemplateImport(cmd, nil); err == nil {
		t.Fatal("import without files succeeded")
	}
	if err := runTemplateImport(cmd, []string{createYAML, filepath.Join(tmpDir, "missing.yaml")}); err != nil {
		t.Fatal(err)
	}
}

func TestWriteTemplateFileDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte(templatePolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	err := writeTemplateFile("template.prompt", []byte("Hello {{ .name }}"))
	if err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("template write error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("template.prompt"); !os.IsNotExist(err) {
		t.Fatalf("template wrote output despite write policy: %v", err)
	}
}

func TestTemplateCreateDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte(templatePolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := templateTestCmd()
	err := runTemplateCreate(cmd, "mine", "mine.prompt", false, "Hello {{ .name }}", "", "custom", nil)
	if err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("template create error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("mine.prompt"); !os.IsNotExist(err) {
		t.Fatalf("template create wrote output despite write policy: %v", err)
	}
}

func TestTemplateApplyStdoutAllowedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte(templatePolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}
	varsJSON := filepath.Join(tmpDir, "vars.json")
	if err := os.WriteFile(varsJSON, []byte(`{"text":"Long article","max_points":3}`), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := templateTestCmd()
	if err := runTemplateApply(cmd, "summarization", varsJSON, "", false); err != nil {
		t.Fatalf("template apply stdout: %v", err)
	}
	if _, err := os.Stat("applied.txt"); !os.IsNotExist(err) {
		t.Fatalf("template apply created unexpected output: %v", err)
	}
}

func templatePolicyTestModule() string {
	return `module example.com/prompts

pe 1

capability {
    tools deny write
}
`
}

func TestTemplateInteractiveAndVariableCollection(t *testing.T) {
	for _, tt := range []struct {
		name  string
		v     templates.Variable
		input string
		want  interface{}
	}{
		{"string", templates.Variable{Name: "name", Type: "string", Required: true}, "Alice\n", "Alice"},
		{"number", templates.Variable{Name: "n", Type: "number", Required: true}, "2.5\n", 2.5},
		{"boolean", templates.Variable{Name: "ok", Type: "boolean", Required: true}, "yes\n", true},
		{"default", templates.Variable{Name: "def", Type: "string", Default: "default"}, "\n", "default"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cmd := templateTestCmd()
			cmd.SetIn(strings.NewReader(tt.input))
			vars, err := collectVariablesInteractively(cmd, &templates.Template{Variables: map[string]templates.Variable{tt.name: tt.v}})
			if err != nil {
				t.Fatal(err)
			}
			if vars[tt.name] != tt.want {
				t.Fatalf("vars = %#v", vars)
			}
		})
	}
	cmd := templateTestCmd()
	cmd.SetIn(strings.NewReader("\n"))
	if _, err := collectVariablesInteractively(cmd, &templates.Template{Variables: map[string]templates.Variable{"required": {Name: "required", Type: "string", Required: true}}}); err == nil {
		t.Fatal("missing required var succeeded")
	}
	cmd = templateTestCmd()
	cmd.SetIn(strings.NewReader("not-number\n"))
	if _, err := collectVariablesInteractively(cmd, &templates.Template{Variables: map[string]templates.Variable{"n": {Name: "n", Type: "number", Required: true}}}); err == nil {
		t.Fatal("bad number succeeded")
	}
	if vars, err := collectVariablesInteractively(templateTestCmd(), &templates.Template{}); err != nil || len(vars) != 0 {
		t.Fatalf("empty vars = %#v err=%v", vars, err)
	}
	if err := runTemplateInteractive(templateTestCmd(), nil); err == nil {
		t.Fatal("interactive without loaded templates succeeded")
	}
}

func templateTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(""))
	return cmd
}
