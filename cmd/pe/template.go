package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/tmc/pe/internal/templates"
)

// templateCmd returns a cobra.Command for the 'template' subcommand.
//
// template manages prompt templates including listing, searching, and applying templates.
//
// Usage:
//
//	pe template list
//	pe template search summarization
//	pe template apply summarization --vars vars.json
func templateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage prompt templates",
		Long: `Template manages a library of reusable prompt templates.
Templates provide structured, parameterized prompts for common use cases
like summarization, code review, creative writing, and data analysis.`,
	}

	cmd.AddCommand(templateListCmd())
	cmd.AddCommand(templateSearchCmd())
	cmd.AddCommand(templateShowCmd())
	cmd.AddCommand(templateApplyCmd())
	cmd.AddCommand(templateCreateCmd())
	cmd.AddCommand(templateValidateCmd())
	cmd.AddCommand(templateExportCmd())
	cmd.AddCommand(templateImportCmd())
	cmd.AddCommand(templateInteractiveCmd())

	return cmd
}

// templateListCmd lists available templates
func templateListCmd() *cobra.Command {
	var category string
	var tag string
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateList(cmd, category, tag, format)
		},
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "Filter by category")
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "Filter by tag")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format: table, json, yaml")

	return cmd
}

// templateSearchCmd searches for templates
func templateSearchCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search templates",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateSearch(cmd, args[0], format)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format: table, json, yaml")

	return cmd
}

// templateShowCmd shows template details
func templateShowCmd() *cobra.Command {
	var format string
	var showExamples bool

	cmd := &cobra.Command{
		Use:   "show [template_name]",
		Short: "Show template details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateShow(cmd, args[0], format, showExamples)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "yaml", "Output format: yaml, json")
	cmd.Flags().BoolVarP(&showExamples, "examples", "e", true, "Show usage examples")

	return cmd
}

// templateApplyCmd applies a template
func templateApplyCmd() *cobra.Command {
	var varsFile string
	var outputFile string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "apply [template_name]",
		Short: "Apply a template with variables",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateApply(cmd, args[0], varsFile, outputFile, interactive)
		},
	}

	cmd.Flags().StringVarP(&varsFile, "vars", "v", "", "Variables file (JSON or YAML)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive variable input")

	return cmd
}

// templateCreateCmd creates a new template
func templateCreateCmd() *cobra.Command {
	var interactive bool
	var outputFile string
	var prompt string
	var description string
	var category string
	var tags []string

	cmd := &cobra.Command{
		Use:   "create [template_name]",
		Short: "Create a new template",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runTemplateCreate(cmd, name, outputFile, interactive, prompt, description, category, tags)
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", true, "Interactive template creation")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for the template")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Prompt text for non-interactive template creation")
	cmd.Flags().StringVar(&description, "description", "", "Template description")
	cmd.Flags().StringVar(&category, "category", "custom", "Template category")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{"user-created"}, "Template tags")

	return cmd
}

// templateValidateCmd validates templates
func templateValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [file...]",
		Short: "Validate template files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateValidate(cmd, args)
		},
	}

	return cmd
}

// templateExportCmd exports templates
func templateExportCmd() *cobra.Command {
	var format string
	var outputDir string

	cmd := &cobra.Command{
		Use:   "export [template_name...]",
		Short: "Export templates to files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateExport(cmd, args, format, outputDir)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "yaml", "Export format: yaml, json")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "d", ".", "Output directory")

	return cmd
}

// templateImportCmd imports templates
func templateImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [file...]",
		Short: "Import templates from files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateImport(cmd, args)
		},
	}

	return cmd
}

// Implementation functions

func runTemplateList(cmd *cobra.Command, category, tag, format string) error {
	// Initialize template library
	library := templates.GetDefaultLibrary()
	if err := library.LoadBuiltinTemplates(); err != nil {
		return fmt.Errorf("failed to load templates: %v", err)
	}

	var templateList []*templates.Template

	// Filter templates
	if category != "" {
		templateList = library.GetByCategory(category)
	} else if tag != "" {
		templateList = library.GetByTag(tag)
	} else {
		templateList = library.ListTemplates()
	}

	return outputTemplateList(cmd, templateList, format)
}

func runTemplateSearch(cmd *cobra.Command, query, format string) error {
	library := templates.GetDefaultLibrary()
	if err := library.LoadBuiltinTemplates(); err != nil {
		return fmt.Errorf("failed to load templates: %v", err)
	}

	templateList := library.Search(query)
	return outputTemplateList(cmd, templateList, format)
}

func runTemplateShow(cmd *cobra.Command, name, format string, showExamples bool) error {
	library := templates.GetDefaultLibrary()
	if err := library.LoadBuiltinTemplates(); err != nil {
		return fmt.Errorf("failed to load templates: %v", err)
	}

	template, err := library.GetTemplate(name)
	if err != nil {
		return err
	}

	// Optionally hide examples
	if !showExamples {
		template.Examples = nil
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(template, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	case "yaml":
		data, err := yaml.Marshal(template)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

func runTemplateApply(cmd *cobra.Command, name, varsFile, outputFile string, interactive bool) error {
	library := templates.GetDefaultLibrary()
	if err := library.LoadBuiltinTemplates(); err != nil {
		return fmt.Errorf("failed to load templates: %v", err)
	}

	template, err := library.GetTemplate(name)
	if err != nil {
		return err
	}

	var variables map[string]interface{}

	if interactive {
		variables, err = collectVariablesInteractively(cmd, template)
		if err != nil {
			return err
		}
	} else if varsFile != "" {
		variables, err = loadVariablesFromFile(varsFile)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("either --vars or --interactive must be specified")
	}

	result, err := library.ApplyTemplate(name, variables)
	if err != nil {
		return err
	}

	if outputFile != "" {
		return os.WriteFile(outputFile, []byte(result), 0644)
	}

	fmt.Fprintln(cmd.OutOrStdout(), result)
	return nil
}

func runTemplateCreate(cmd *cobra.Command, name, outputFile string, interactive bool, prompt, description, category string, tags []string) error {
	if !interactive || prompt != "" {
		return runTemplateCreateNonInteractive(cmd, name, outputFile, prompt, description, category, tags)
	}

	// For now, create a basic template structure
	template := &templates.Template{
		Name:        name,
		Description: "Custom template",
		Category:    "custom",
		Tags:        []string{"user-created"},
		Version:     "1.0.0",
		Prompt:      "{{input}}",
		Variables: map[string]templates.Variable{
			"input": {
				Name:        "input",
				Description: "Input text",
				Type:        "string",
				Required:    true,
			},
		},
	}

	var err error
	var data []byte

	if outputFile != "" {
		ext := strings.ToLower(filepath.Ext(outputFile))
		if ext == ".json" {
			data, err = json.MarshalIndent(template, "", "  ")
		} else {
			data, err = yaml.Marshal(template)
		}
		if err != nil {
			return err
		}
		return os.WriteFile(outputFile, data, 0644)
	}

	data, err = yaml.Marshal(template)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

func runTemplateCreateNonInteractive(cmd *cobra.Command, name, outputFile, prompt, description, category string, tags []string) error {
	if name == "" {
		return fmt.Errorf("template name is required")
	}
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("template prompt is required for non-interactive creation")
	}
	if outputFile == "" {
		outputFile = name + ".prompt"
	}
	if description == "" {
		description = "Custom template"
	}
	if category == "" {
		category = "custom"
	}
	if len(tags) == 0 {
		tags = []string{"user-created"}
	}

	ext := strings.ToLower(filepath.Ext(outputFile))
	switch ext {
	case ".prompt", ".txt", ".md":
		if err := os.WriteFile(outputFile, []byte(prompt), 0644); err != nil {
			return err
		}
	default:
		template := newTemplateDefinition(name, prompt, description, category, tags)
		var (
			data []byte
			err  error
		)
		if ext == ".json" {
			data, err = json.MarshalIndent(template, "", "  ")
		} else {
			data, err = yaml.Marshal(template)
		}
		if err != nil {
			return err
		}
		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			return err
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created template %s at %s\n", name, outputFile)
	return nil
}

func newTemplateDefinition(name, prompt, description, category string, tags []string) *templates.Template {
	now := time.Now().UTC()
	variables := inferTemplateVariables(prompt)
	return &templates.Template{
		Name:        name,
		Description: description,
		Category:    category,
		Tags:        append([]string(nil), tags...),
		Version:     "1.0.0",
		CreatedAt:   now,
		UpdatedAt:   now,
		Prompt:      prompt,
		Variables:   variables,
	}
}

var templateVariablePattern = regexp.MustCompile(`\{\{\s*\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

func inferTemplateVariables(prompt string) map[string]templates.Variable {
	matches := templateVariablePattern.FindAllStringSubmatch(prompt, -1)
	if len(matches) == 0 {
		return map[string]templates.Variable{}
	}

	names := make(map[string]bool)
	for _, match := range matches {
		names[match[1]] = true
	}

	ordered := make([]string, 0, len(names))
	for name := range names {
		ordered = append(ordered, name)
	}
	sort.Strings(ordered)

	variables := make(map[string]templates.Variable, len(ordered))
	for _, name := range ordered {
		variables[name] = templates.Variable{
			Name:        name,
			Description: name,
			Type:        "string",
			Required:    true,
		}
	}
	return variables
}

func runTemplateValidate(cmd *cobra.Command, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files specified")
	}

	library := templates.NewTemplateLibrary("")
	errors := 0

	for _, file := range files {
		fmt.Fprintf(cmd.OutOrStdout(), "Validating %s... ", file)

		template, err := library.LoadTemplateFromFile(file)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "✗ FAILED: %v\n", err)
			errors++
			continue
		}

		// Basic validation
		if template.Name == "" {
			fmt.Fprintf(cmd.OutOrStdout(), "✗ FAILED: missing name\n")
			errors++
			continue
		}

		if template.Prompt == "" {
			fmt.Fprintf(cmd.OutOrStdout(), "✗ FAILED: missing prompt\n")
			errors++
			continue
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✓ PASSED\n")
	}

	if errors > 0 {
		return fmt.Errorf("%d file(s) failed validation", errors)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nAll files passed validation!\n")
	return nil
}

func runTemplateExport(cmd *cobra.Command, names []string, format, outputDir string) error {
	library := templates.GetDefaultLibrary()
	if err := library.LoadBuiltinTemplates(); err != nil {
		return fmt.Errorf("failed to load templates: %v", err)
	}

	if len(names) == 0 {
		// Export all templates
		templateList := library.ListTemplates()
		for _, template := range templateList {
			names = append(names, template.Name)
		}
	}

	for _, name := range names {
		data, err := library.ExportTemplate(name, format)
		if err != nil {
			return fmt.Errorf("failed to export template %s: %v", name, err)
		}

		ext := ".yaml"
		if format == "json" {
			ext = ".json"
		}

		filename := filepath.Join(outputDir, name+ext)
		if err := os.WriteFile(filename, data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %v", filename, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Exported: %s\n", filename)
	}

	return nil
}

func runTemplateImport(cmd *cobra.Command, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files specified")
	}

	library := templates.GetDefaultLibrary()
	imported := 0

	for _, file := range files {
		template, err := library.LoadTemplateFromFile(file)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to load %s: %v\n", file, err)
			continue
		}

		if err := library.AddTemplate(template); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to add template %s: %v\n", template.Name, err)
			continue
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Imported: %s\n", template.Name)
		imported++
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nImported %d template(s)\n", imported)
	return nil
}

// Helper functions

func outputTemplateList(cmd *cobra.Command, templateList []*templates.Template, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(templateList, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))

	case "yaml":
		data, err := yaml.Marshal(templateList)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))

	case "table":
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCATEGORY\tTAGS\tDESCRIPTION")
		fmt.Fprintln(w, "----\t--------\t----\t-----------")

		for _, template := range templateList {
			tags := strings.Join(template.Tags, ",")
			if len(tags) > 20 {
				tags = tags[:17] + "..."
			}
			desc := template.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", template.Name, template.Category, tags, desc)
		}

		w.Flush()

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// templateInteractiveCmd provides interactive template selection and application
func templateInteractiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Interactive template selection and application",
		Long:  `Select and apply templates interactively with guided variable input.`,
		RunE:  runTemplateInteractive,
	}

	return cmd
}

func runTemplateInteractive(cmd *cobra.Command, args []string) error {
	library := templates.GetDefaultLibrary()
	templateList := library.ListTemplates()

	if len(templateList) == 0 {
		return fmt.Errorf("no templates available")
	}

	// Show available templates
	fmt.Fprintf(cmd.OutOrStdout(), "Available templates:\n\n")
	for i, tmpl := range templateList {
		fmt.Fprintf(cmd.OutOrStdout(), "%d. %s - %s\n", i+1, tmpl.Name, tmpl.Description)
	}

	// Ask user to select a template
	fmt.Fprintf(cmd.OutOrStdout(), "\nSelect a template (1-%d): ", len(templateList))

	var selection int
	if _, err := fmt.Fscanf(cmd.InOrStdin(), "%d", &selection); err != nil {
		return fmt.Errorf("invalid selection: %v", err)
	}

	if selection < 1 || selection > len(templateList) {
		return fmt.Errorf("selection out of range")
	}

	selectedTemplate := templateList[selection-1]
	fmt.Fprintf(cmd.OutOrStdout(), "\nSelected: %s\n\n", selectedTemplate.Name)

	// Collect variables interactively
	vars, err := collectVariablesInteractively(cmd, selectedTemplate)
	if err != nil {
		return err
	}

	// Apply the template
	result, err := library.ApplyTemplate(selectedTemplate.Name, vars)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nGenerated prompt:\n\n%s\n", result)
	return nil
}

func collectVariablesInteractively(cmd *cobra.Command, template *templates.Template) (map[string]interface{}, error) {
	vars := make(map[string]interface{})

	if len(template.Variables) == 0 {
		return vars, nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Please provide values for the following variables:\n\n")

	scanner := bufio.NewScanner(cmd.InOrStdin())

	for name, variable := range template.Variables {
		fmt.Fprintf(cmd.OutOrStdout(), "%s", name)
		if variable.Description != "" {
			fmt.Fprintf(cmd.OutOrStdout(), " (%s)", variable.Description)
		}
		if variable.Default != nil {
			fmt.Fprintf(cmd.OutOrStdout(), " [default: %v]", variable.Default)
		}
		if variable.Required {
			fmt.Fprintf(cmd.OutOrStdout(), " *")
		}
		fmt.Fprintf(cmd.OutOrStdout(), ": ")

		// Read user input
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("error reading input: %v", err)
			}
			return nil, fmt.Errorf("unexpected end of input")
		}

		input := strings.TrimSpace(scanner.Text())

		// Use default if no input and default exists
		if input == "" && variable.Default != nil {
			vars[name] = variable.Default
		} else if input == "" && variable.Required {
			return nil, fmt.Errorf("variable '%s' is required", name)
		} else if input != "" {
			// Parse based on type
			switch variable.Type {
			case "number":
				if val, err := strconv.ParseFloat(input, 64); err == nil {
					vars[name] = val
				} else {
					return nil, fmt.Errorf("invalid number for '%s': %v", name, err)
				}
			case "boolean":
				vars[name] = strings.ToLower(input) == "true" || input == "1" || strings.ToLower(input) == "yes"
			default:
				vars[name] = input
			}
		}
	}

	return vars, nil
}

func loadVariablesFromFile(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var variables map[string]interface{}

	// Try JSON first, then YAML
	if err := json.Unmarshal(data, &variables); err != nil {
		if err := yaml.Unmarshal(data, &variables); err != nil {
			return nil, fmt.Errorf("failed to parse variables file as JSON or YAML: %v", err)
		}
	}

	return variables, nil
}
