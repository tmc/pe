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

func runCmd() *cobra.Command {
	var (
		provider string
		vars     map[string]string
		stream   bool
	)

	cmd := &cobra.Command{
		Use:   "run [prompt or file]",
		Short: "Execute a prompt immediately (like go run)",
		Args:  cobra.ExactArgs(1),
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
			if err := enforceRuntimeProviderPolicy(provider); err != nil {
				return err
			}
			client := inference.NewClient()
			client.Register("cgpt", cgpt.New())

			// Register mock provider for testing
			if os.Getenv("PE_TEST_MODE") == "true" {
				// Create mock response based on content
				mockResp := "Mock response for: " + content

				// Handle specific test cases like the internal/providers mock does
				if strings.Contains(content, "2+2") {
					mockResp = "4"
				} else if strings.Contains(content, "pointer") {
					mockResp = "A pointer is a variable that stores the memory address of another variable."
				} else if strings.Contains(content, "capital") && strings.Contains(content, "France") {
					// Check if this is from the extract test (includes answer tags)
					if strings.Contains(content, "<answer>") {
						mockResp = content // Return as-is since it already has the answer
					} else {
						mockResp = "<answer>The capital of France is Paris</answer>"
					}
				} else if strings.Contains(content, "Translate Hello to Spanish") {
					mockResp = "Hola"
				} else if strings.Contains(content, "Count from 1 to 10") {
					mockResp = "1\n2\n3\n4\n5\n6\n7\n8\n9\n10"
				}

				client.Register("mock", &mockProvider{response: mockResp})
				if provider == "cgpt" {
					provider = "mock"
				}
			}
			if err := registerLLMProviderSpec(client, provider); err != nil {
				return err
			}

			req := inference.Request{
				Prompt: content,
				Stream: stream,
			}

			if stream {
				chunks, err := client.StreamWith(context.Background(), provider, req)
				if err != nil {
					return err
				}
				return inference.StreamToWriter(context.Background(), chunks, cmd.OutOrStdout())
			}

			resp, err := client.CompleteWith(context.Background(), provider, req)
			if err != nil {
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), resp.Content)
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "cgpt", "Provider to use")
	cmd.Flags().StringToStringVar(&vars, "var", nil, "Template variables")
	cmd.Flags().BoolVar(&stream, "stream", true, "Stream output")

	return cmd
}

// registerProviders registers the built-in providers used by run.
func registerProviders(client *inference.Client) {
	client.Register("cgpt", cgpt.New())

	if os.Getenv("PE_TEST_MODE") == "true" || os.Getenv("PE_MOCK_PROVIDER") == "true" {
		client.Register("mock", &mockProvider{})
	}
}

func registerLLMProviderSpec(client *inference.Client, spec string) error {
	if spec == "" || spec == "cgpt" || spec == "mock" {
		return nil
	}
	_, err := inference.RegisterProviderSpec(client, spec, nil)
	if err != nil {
		if !strings.Contains(spec, ":") {
			return fmt.Errorf("provider %q not found", spec)
		}
		return err
	}
	return nil
}

func enforceRuntimeProviderPolicy(provider string) error {
	file, ok, err := readRuntimePolicyFile()
	if err != nil || !ok {
		return err
	}
	if file.Capability == nil && file.Placement == nil {
		return nil
	}
	base := providerBaseName(provider)
	if base == "" {
		base = "cgpt"
	}
	if file.Capability != nil {
		for _, deny := range file.Capability.Providers.Deny {
			if deny == base {
				return fmt.Errorf("provider %s is denied by pe.mod", base)
			}
			if deny == "remote" && isRemoteProvider(base) {
				return fmt.Errorf("provider %s is denied by pe.mod remote provider policy", base)
			}
			if deny == "local" && isLocalProvider(base) {
				return fmt.Errorf("provider %s is denied by pe.mod local provider policy", base)
			}
		}
	}
	if file.Placement != nil && file.Placement.Network != nil && !*file.Placement.Network && isRemoteProvider(base) {
		return fmt.Errorf("provider %s requires network access denied by pe.mod placement", base)
	}
	return nil
}

func providerBaseName(provider string) string {
	if i := strings.Index(provider, ":"); i >= 0 {
		return provider[:i]
	}
	return provider
}

func isRemoteProvider(provider string) bool {
	switch provider {
	case "openai", "anthropic":
		return true
	default:
		return false
	}
}

func isLocalProvider(provider string) bool {
	switch provider {
	case "cgpt", "mock", "ollama":
		return true
	default:
		return false
	}
}
