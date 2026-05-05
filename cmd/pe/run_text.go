package main

import (
	"fmt"
	"io"
	"os"

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
			out, err := f.Render(vars)
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
