package main

import "testing"

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
