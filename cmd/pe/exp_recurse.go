package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/inference"
	"github.com/tmc/pe/internal/rlm"
)

func init() {
	expCmd.AddCommand(expRecurseCmd())
}

func expRecurseCmd() *cobra.Command {
	var (
		prompt    string
		provider  string
		model     string
		chunkSize int
		maxDepth  int
		maxTokens int
		workers   int
		maxChunks int
		rule      string
		tracePath string
	)

	cmd := &cobra.Command{
		Use:   "recurse <target-file>",
		Short: "Process a large input through bounded local recursive operations",
		Long: `Process a large input through bounded, local recursive operations.

The target file is split into bounded chunks. Each chunk is processed by a
single prompt through a configured provider, and the child results are merged
with a deterministic aggregation rule. Large input stays out of the prompt:
chunks are referenced by content key, and only the bounded chunk reaches the
provider.

This command is experimental and local-first. All budgets (depth, tokens,
workers, chunks) are enforced before any provider call. Every run can emit a
strict JSON trace (schema pe.rlm.trace.v1) for inspection and replay.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if prompt == "" {
				return fmt.Errorf("--prompt is required")
			}
			target, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read target: %w", err)
			}

			// Enforce module provider policy before constructing any provider,
			// matching the runtime enforcement applied by pe run.
			if err := enforceRuntimeProviderPolicy(provider); err != nil {
				return err
			}

			worker, err := newRecurseWorker(provider, model)
			if err != nil {
				return err
			}

			trace, runErr := rlm.Run(cmd.Context(), worker, target, rlm.Options{
				Prompt:         prompt,
				ChunkSize:      chunkSize,
				MaxDepth:       maxDepth,
				MaxTokens:      maxTokens,
				Workers:        workers,
				MaxChunks:      maxChunks,
				Rule:           rule,
				Command:        "pe exp recurse",
				TargetPath:     args[0],
				TargetManifest: "",
			})
			if trace != nil {
				if writeErr := writeRecurseTrace(cmd.OutOrStdout(), tracePath, trace); writeErr != nil {
					return writeErr
				}
			}
			return runErr
		},
	}

	cmd.Flags().StringVar(&prompt, "prompt", "", "instruction applied to each chunk (required)")
	cmd.Flags().StringVar(&provider, "provider", "cgpt", "provider to use")
	cmd.Flags().StringVar(&model, "model", "", "model to use")
	cmd.Flags().IntVar(&chunkSize, "chunk-size", rlm.DefaultChunkSize, "maximum bytes per chunk")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 1, "maximum recursive reduction depth")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 0, "per-worker token budget (0 disables)")
	cmd.Flags().IntVar(&workers, "workers", 4, "bounded local worker count")
	cmd.Flags().IntVar(&maxChunks, "max-chunks", 0, "cap on chunks processed (0 disables)")
	cmd.Flags().StringVar(&rule, "consensus", rlm.RuleMajority, "aggregation rule: majority or concat")
	cmd.Flags().StringVar(&tracePath, "trace", "-", "trace output file, or - for stdout")
	return cmd
}

// recurseWorker adapts an inference client to the rlm.Worker interface.
type recurseWorker struct {
	client   *inference.Client
	provider string
	model    string
}

func newRecurseWorker(provider, model string) (rlm.Worker, error) {
	client := inference.NewClient()
	registerProviders(client)
	if err := registerLLMProviderSpec(client, provider); err != nil {
		return nil, err
	}
	return &recurseWorker{client: client, provider: provider, model: model}, nil
}

func (w *recurseWorker) Name() string {
	if w.model != "" {
		return w.provider + ":" + w.model
	}
	return w.provider
}

func (w *recurseWorker) Run(ctx context.Context, prompt string, c rlm.Chunk) (string, rlm.Cost, error) {
	req := inference.Request{
		Prompt: prompt + "\n\n" + string(c.Data),
		Model:  w.model,
	}
	resp, err := w.client.CompleteWith(ctx, w.provider, req)
	if err != nil {
		return "", rlm.Cost{}, err
	}
	cost := rlm.Cost{
		PromptTokens:     resp.TokensUsed.PromptTokens,
		CompletionTokens: resp.TokensUsed.CompletionTokens,
		TotalTokens:      resp.TokensUsed.TotalTokens,
	}
	return resp.Content, cost, nil
}

func writeRecurseTrace(stdout io.Writer, path string, trace *rlm.Trace) error {
	var out io.Writer = stdout
	if path != "" && path != "-" {
		if err := enforceRuntimeToolPolicy("write"); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create trace output: %w", err)
		}
		defer f.Close()
		out = f
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(trace); err != nil {
		return fmt.Errorf("encode trace: %w", err)
	}
	return nil
}
