package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/inference"
	"github.com/tmc/pe/internal/inference/providers/cgpt"
	"github.com/tmc/pe/internal/prompt"
)

// This is how Russ Cox would write it:
// 1. Simple, direct, no unnecessary abstractions
// 2. Focus on the core use case
// 3. Let Unix handle composition (pipes, redirection)

func runCmdSimple() *cobra.Command {
	var (
		provider string
		vars     map[string]string
		stream   bool
	)

	cmd := &cobra.Command{
		Use:   "run [prompt or file]",
		Short: "Execute a prompt",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read the prompt
			input := args[0]
			var content string
			
			if input == "-" {
				// Read from stdin
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				content = string(data)
			} else if _, err := os.Stat(input); err == nil {
				// Read from file
				data, err := os.ReadFile(input)
				if err != nil {
					return err
				}
				content = string(data)
			} else {
				// Direct prompt
				content = input
			}

			// Parse prompt if it's a file
			parsed, _ := prompt.Parse(content)
			if parsed.Main != "" {
				content = parsed.Main
			}

			// Apply template variables if needed
			if len(vars) > 0 {
				tmpl, err := template.New("prompt").Parse(content)
				if err == nil {
					var buf strings.Builder
					tmpl.Execute(&buf, vars)
					content = buf.String()
				}
			}

			// Execute the prompt
			client := inference.NewClient()
			client.Register("cgpt", cgpt.New())
			
			req := inference.Request{
				Prompt: content,
				Stream: stream,
			}

			if stream {
				chunks, err := client.StreamWith(context.Background(), provider, req)
				if err != nil {
					return err
				}
				return inference.StreamToWriter(context.Background(), chunks, os.Stdout)
			}

			resp, err := client.CompleteWith(context.Background(), provider, req)
			if err != nil {
				return err
			}
			
			fmt.Print(resp.Content)
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "cgpt", "Provider to use")
	cmd.Flags().StringToStringVar(&vars, "var", nil, "Template variables")
	cmd.Flags().BoolVar(&stream, "stream", true, "Stream output")

	return cmd
}

// That's it. 100 lines instead of 1300.
// No ExecutionLog, no hashing, no complex abstractions.
// Just: read prompt, apply variables, send to LLM, print output.
//
// As Russ Cox says: "Software engineering is what happens to programming
// when you add time and other programmers." Keep it simple until you
// actually need the complexity.