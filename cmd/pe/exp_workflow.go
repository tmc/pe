package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/exectext"
	"github.com/tmc/pe/internal/inference"
	"github.com/tmc/pe/internal/inference/providers/cgpt"
	"github.com/tmc/pe/internal/workflow"
)

func init() {
	expCmd.AddCommand(expWorkflowCmd())
}

func expWorkflowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Run and validate deterministic workflow scripts",
		Long: `Run and validate deterministic workflow scripts.

A workflow script is a Starlark program that composes model calls with
ordinary control flow. The script runs single-threaded and deterministically;
only model calls run concurrently, through parallel(), under a bounded worker
pool. Budgets are enforced before any call, and every call is recorded in a
strict JSON trace (schema pe.workflow.trace.v1).

Scripts are plain .star files or executable text with kind: pe.workflow.v1
front matter whose body is the script. Front matter input defaults seed the
args global, and safety.providers allow/deny lists are enforced on every
call, composed conservatively with pe.mod capability policy.`,
	}
	cmd.AddCommand(expWorkflowRunCmd(), expWorkflowValidateCmd())
	return cmd
}

func expWorkflowRunCmd() *cobra.Command {
	var (
		provider  string
		argsJSON  string
		vars      map[string]string
		workers   int
		maxCalls  int
		tracePath string
	)

	cmd := &cobra.Command{
		Use:   "run <script>",
		Short: "Run a workflow script",
		Long: `Run a workflow script and print its result.

The script result is the return value of main(args) when the script defines
main, otherwise the value of the global named result. String results print
verbatim; other values print as JSON.

Examples:
  pe exp workflow run review.star --provider ollama:llama3.2:3b
  pe exp workflow run audit.pe.md --args '{"topic": "error handling"}'
  pe exp workflow run pipeline.star --var topic=caching --trace trace.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, cmdArgs []string) error {
			file, src, err := readWorkflowScript(cmdArgs[0], cmd.InOrStdin())
			if err != nil {
				return err
			}
			args, err := workflowArgs(argsJSON, vars, file)
			if err != nil {
				return err
			}
			if err := enforceRuntimeProviderPolicy(provider); err != nil {
				return err
			}

			caller := newWorkflowCaller(provider)
			defer caller.Close()
			opts := workflow.Options{
				Caller:   caller,
				Args:     args,
				Workers:  workers,
				MaxCalls: maxCalls,
				Check: func(c workflow.Call) error {
					spec := c.Provider
					if spec == "" {
						spec = provider
					}
					if err := enforceRuntimeProviderPolicy(spec); err != nil {
						return err
					}
					return checkWorkflowProviderPolicy(file, spec)
				},
				Logf: func(format string, logArgs ...interface{}) {
					fmt.Fprintf(cmd.ErrOrStderr(), format+"\n", logArgs...)
				},
			}

			outcome, runErr := workflow.Run(cmd.Context(), cmdArgs[0], src, opts)
			if outcome != nil && tracePath != "" {
				if err := writeWorkflowTrace(cmd.OutOrStdout(), tracePath, outcome.Trace); err != nil {
					return errors.Join(runErr, err)
				}
			}
			if runErr != nil {
				return runErr
			}
			return printWorkflowValue(cmd.OutOrStdout(), outcome.Value)
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "cgpt", "default provider for calls that name none")
	cmd.Flags().StringVar(&argsJSON, "args", "", "script args as a JSON object")
	cmd.Flags().StringToStringVar(&vars, "var", nil, "script args as key=value strings")
	cmd.Flags().IntVar(&workers, "workers", workflow.DefaultWorkers, "bounded concurrent calls in parallel()")
	cmd.Flags().IntVar(&maxCalls, "max-calls", workflow.DefaultMaxCalls, "cap on total model calls (-1 disables)")
	cmd.Flags().StringVar(&tracePath, "trace", "", "trace output file, or - for stdout")
	return cmd
}

func expWorkflowValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <script>",
		Short: "Compile a workflow script without running it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file, src, err := readWorkflowScript(args[0], cmd.InOrStdin())
			if err != nil {
				return err
			}
			if file != nil {
				if err := file.Validate(); err != nil {
					return fmt.Errorf("front matter: %w", err)
				}
			}
			if err := workflow.Validate(args[0], src); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "OK: %s\n", args[0])
			return nil
		},
	}
}

// readWorkflowScript loads a workflow script from path or stdin ("-"). Plain
// files are used verbatim. Executable text with front matter must declare
// kind pe.workflow.v1; its body is the script and its metadata is returned
// for input defaults and safety policy.
func readWorkflowScript(path string, stdin io.Reader) (*exectext.File, []byte, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read workflow script: %w", err)
	}

	file, err := exectext.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("parse workflow script: %w", err)
	}
	switch file.Meta.Kind {
	case "pe.workflow.v1":
		return file, []byte(file.Body), nil
	case "":
		if hasFrontMatter(file) {
			return file, []byte(file.Body), nil
		}
		return nil, data, nil
	default:
		return nil, nil, fmt.Errorf("workflow scripts require kind pe.workflow.v1, got %s", file.Meta.Kind)
	}
}

// hasFrontMatter reports whether parsing recovered any metadata, meaning the
// body is no longer the raw file.
func hasFrontMatter(file *exectext.File) bool {
	m := file.Meta
	return m.Name != "" || m.Run != "" || len(m.Inputs) > 0 || len(m.Outputs) > 0 ||
		len(m.Metadata) > 0 || len(m.Budget) > 0 || len(m.Imports) > 0 ||
		len(m.Safety) > 0 || len(m.Placement) > 0
}

// workflowArgs merges --args JSON, --var overrides, and front matter input
// defaults, in increasing then decreasing precedence: vars beat JSON, and
// declared defaults fill remaining gaps.
func workflowArgs(argsJSON string, vars map[string]string, file *exectext.File) (map[string]interface{}, error) {
	args := make(map[string]interface{})
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return nil, fmt.Errorf("parse --args: %w", err)
		}
	}
	for k, v := range vars {
		args[k] = v
	}
	if file != nil {
		for name, input := range file.Meta.Inputs {
			if _, ok := args[name]; ok {
				continue
			}
			if input.Default != nil {
				args[name] = input.Default
			}
		}
	}
	if len(args) == 0 {
		return nil, nil
	}
	return args, nil
}

// checkWorkflowProviderPolicy enforces the script's own safety.providers
// allow/deny lists on a provider spec.
func checkWorkflowProviderPolicy(file *exectext.File, spec string) error {
	if file == nil {
		return nil
	}
	policy, ok := file.Meta.Safety["providers"]
	if !ok {
		return nil
	}
	base := providerBaseName(spec)
	if base == "" {
		base = "cgpt"
	}
	for _, deny := range policy.Deny {
		switch {
		case deny == spec || deny == base:
			return fmt.Errorf("provider %s denied by workflow safety policy", base)
		case deny == "remote" && isRemoteProvider(base):
			return fmt.Errorf("provider %s denied by workflow remote provider policy", base)
		case deny == "local" && isLocalProvider(base):
			return fmt.Errorf("provider %s denied by workflow local provider policy", base)
		}
	}
	if len(policy.Allow) == 0 {
		return nil
	}
	for _, allow := range policy.Allow {
		switch {
		case allow == spec || allow == base:
			return nil
		case allow == "remote" && isRemoteProvider(base):
			return nil
		case allow == "local" && isLocalProvider(base):
			return nil
		}
	}
	return fmt.Errorf("provider %s not in workflow safety allow list", base)
}

// workflowCaller adapts inference providers to the workflow.Caller interface.
// It resolves provider specs lazily and is safe for concurrent calls.
type workflowCaller struct {
	defaultSpec string

	mu        sync.Mutex
	providers map[string]inference.Provider
	created   []inference.Provider
}

func newWorkflowCaller(defaultSpec string) *workflowCaller {
	c := &workflowCaller{
		defaultSpec: defaultSpec,
		providers:   map[string]inference.Provider{"cgpt": cgpt.New()},
	}
	if os.Getenv("PE_TEST_MODE") == "true" || os.Getenv("PE_MOCK_PROVIDER") == "true" {
		c.providers["mock"] = &mockProvider{}
	}
	return c
}

func (c *workflowCaller) provider(spec string) (inference.Provider, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if p, ok := c.providers[spec]; ok {
		return p, nil
	}
	p, err := inference.CreateProviderFromSpec(spec, nil)
	if err != nil {
		return nil, err
	}
	c.providers[spec] = p
	c.created = append(c.created, p)
	return p, nil
}

func (c *workflowCaller) Call(ctx context.Context, call workflow.Call) (workflow.Result, error) {
	spec := call.Provider
	if spec == "" {
		spec = c.defaultSpec
	}
	p, err := c.provider(spec)
	if err != nil {
		return workflow.Result{}, err
	}
	resp, err := p.Complete(ctx, inference.Request{
		Prompt:       call.Prompt,
		Model:        call.Model,
		SystemPrompt: call.System,
		Temperature:  float32(call.Temperature),
		MaxTokens:    call.MaxTokens,
	})
	if err != nil {
		return workflow.Result{}, err
	}
	return workflow.Result{
		Content:  resp.Content,
		Provider: spec,
		Cost: workflow.Cost{
			PromptTokens:     resp.TokensUsed.PromptTokens,
			CompletionTokens: resp.TokensUsed.CompletionTokens,
			TotalTokens:      resp.TokensUsed.TotalTokens,
		},
	}, nil
}

func (c *workflowCaller) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var errs []error
	for _, p := range c.created {
		errs = append(errs, p.Close())
	}
	return errors.Join(errs...)
}

func writeWorkflowTrace(stdout io.Writer, path string, trace *workflow.Trace) error {
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

func printWorkflowValue(out io.Writer, value interface{}) error {
	if s, ok := value.(string); ok {
		fmt.Fprintln(out, s)
		return nil
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode result: %w", err)
	}
	fmt.Fprintln(out, string(data))
	return nil
}
