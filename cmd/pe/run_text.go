package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/exectext"
)

func runTextCmd() *cobra.Command {
	var (
		vars  map[string]string
		check bool
	)
	cmd := &cobra.Command{
		Use:   "run-text [file|-]",
		Short: "Render executable text safely",
		Long: `Run-text validates and renders PE executable text.

Plain text is valid by default. Files may add a pe.text.v1 or pe.workflow.v1
front matter block to declare inputs, metadata, safety, and placement. This
command performs static validation and template rendering; it does not execute
providers, tools, shell commands, or network requests.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := readExecText(args[0])
			if err != nil {
				return err
			}
			if err := f.Validate(); err != nil {
				return err
			}
			if check {
				fmt.Fprintln(cmd.OutOrStdout(), "ok")
				return nil
			}
			imports, err := readExecTextImports(args[0], f)
			if err != nil {
				return err
			}
			out, err := f.RenderWithImports(vars, imports)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().StringToStringVar(&vars, "var", nil, "Template variables")
	cmd.Flags().BoolVar(&check, "check", false, "Validate without rendering")
	return cmd
}

func readExecText(name string) (*exectext.File, error) {
	var r io.Reader
	if name == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(name)
		if err != nil {
			return nil, fmt.Errorf("opening executable text: %w", err)
		}
		defer f.Close()
		r = f
	}
	file, err := exectext.Parse(r)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func readExecTextImports(name string, file *exectext.File) (map[string]string, error) {
	if name == "-" || len(file.Meta.Imports) == 0 {
		return nil, nil
	}
	base, err := filepath.Abs(filepath.Dir(name))
	if err != nil {
		return nil, fmt.Errorf("resolving executable text directory: %w", err)
	}
	imports := make(map[string]string)
	for alias, rel := range file.Meta.Imports {
		if rel == "" || filepath.IsAbs(rel) {
			return nil, fmt.Errorf("import %s escapes executable text directory", alias)
		}
		path, err := containedFilePath(base, rel)
		if err != nil {
			return nil, fmt.Errorf("import %s escapes executable text directory: %w", alias, err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading import %s: %w", alias, err)
		}
		imports[alias] = string(data)
	}
	return imports, nil
}
