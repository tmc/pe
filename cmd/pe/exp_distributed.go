package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/distributed"
)

type distributedTaskFile struct {
	Tasks []distributedTaskSpec `json:"tasks"`
}

type distributedTaskSpec struct {
	Name    string `json:"name"`
	Output  string `json:"output"`
	DelayMS int    `json:"delay_ms,omitempty"`
	Error   string `json:"error,omitempty"`
}

type distributedTaskResult struct {
	Name   string `json:"name"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

func newExpDistributedCmd() *cobra.Command {
	var workers int
	var timeoutMS int
	var format string

	cmd := &cobra.Command{
		Use:          "distributed <tasks.json>",
		Short:        "Run local deterministic tasks with bounded concurrency",
		SilenceUsage: true,
		Long: `Run local deterministic tasks with bounded concurrency.

The task file is JSON:

  {"tasks":[{"name":"one","output":"first","delay_ms":10}]}

This is a local scheduler prototype. It does not start daemons, open network
connections, or coordinate remote workers.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if workers < 1 {
				return fmt.Errorf("workers must be at least 1")
			}
			if timeoutMS < 0 {
				return fmt.Errorf("timeout-ms must be non-negative")
			}
			if format != "json" && format != "text" {
				return fmt.Errorf("format must be json or text")
			}

			specs, err := readDistributedTaskFile(args[0])
			if err != nil {
				return err
			}
			results, err := runDistributedTasks(cmd.Context(), specs, workers, time.Duration(timeoutMS)*time.Millisecond)
			if writeErr := writeDistributedResults(cmd.OutOrStdout(), results, format); writeErr != nil {
				return writeErr
			}
			return err
		},
	}
	cmd.Flags().IntVar(&workers, "workers", 1, "maximum concurrent local tasks")
	cmd.Flags().IntVar(&timeoutMS, "timeout-ms", 0, "optional timeout in milliseconds")
	cmd.Flags().StringVar(&format, "format", "json", "output format: json or text")
	return cmd
}

func readDistributedTaskFile(path string) ([]distributedTaskSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tasks: %w", err)
	}
	var file distributedTaskFile
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&file); err != nil {
		return nil, fmt.Errorf("decode tasks: %w", err)
	}
	if len(file.Tasks) == 0 {
		return nil, fmt.Errorf("tasks must not be empty")
	}
	for i, task := range file.Tasks {
		if task.Name == "" {
			return nil, fmt.Errorf("task %d has empty name", i)
		}
		if task.DelayMS < 0 {
			return nil, fmt.Errorf("task %q has negative delay_ms", task.Name)
		}
	}
	return file.Tasks, nil
}

func runDistributedTasks(ctx context.Context, specs []distributedTaskSpec, workers int, timeout time.Duration) ([]distributedTaskResult, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	tasks := make([]distributed.Task[distributedTaskResult], len(specs))
	for i, spec := range specs {
		spec := spec
		tasks[i] = distributed.Task[distributedTaskResult]{
			ID: spec.Name,
			Run: func(ctx context.Context) (distributedTaskResult, error) {
				result := distributedTaskResult{Name: spec.Name}
				if spec.DelayMS > 0 {
					timer := time.NewTimer(time.Duration(spec.DelayMS) * time.Millisecond)
					defer timer.Stop()
					select {
					case <-ctx.Done():
						result.Error = ctx.Err().Error()
						return result, ctx.Err()
					case <-timer.C:
					}
				}
				if spec.Error != "" {
					result.Error = spec.Error
					return result, fmt.Errorf("%s", spec.Error)
				}
				result.Output = spec.Output
				return result, nil
			},
		}
	}

	executionResults, err := distributed.RunLocal(ctx, workers, tasks)
	results := make([]distributedTaskResult, 0, len(executionResults))
	for _, executionResult := range executionResults {
		if executionResult.ID == "" {
			continue
		}
		result := executionResult.Value
		if result.Name == "" {
			result.Name = executionResult.ID
		}
		if executionResult.Err != nil && result.Error == "" {
			result.Error = executionResult.Err.Error()
		}
		results = append(results, result)
	}
	return results, err
}

func writeDistributedResults(w io.Writer, results []distributedTaskResult, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	case "text":
		for _, result := range results {
			if result.Error != "" {
				if _, err := fmt.Fprintf(w, "%s\tERROR\t%s\n", result.Name, result.Error); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(w, "%s\t%s\n", result.Name, result.Output); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("format must be json or text")
	}
}
