package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/observability"
)

func instrumentCommandTracing(root *cobra.Command) {
	for _, cmd := range root.Commands() {
		instrumentCommandTracing(cmd)
	}
	instrumentCommand(root)
}

func instrumentCommand(cmd *cobra.Command) {
	if cmd == nil || cmd.Annotations["pe.trace.instrumented"] == "true" {
		return
	}
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations["pe.trace.instrumented"] = "true"
	group := cmd.Annotations[commandGroupAnnotation]
	name := cmd.Name()

	if runE := cmd.RunE; runE != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			ctx, span := observability.GetGlobalTracer().StartSpan(cmd.Context(), traceOperation(cmd))
			if span != nil {
				observability.SetTag(ctx, "command", name)
				observability.SetTag(ctx, "group", group)
				observability.SetTag(ctx, "path", cmd.CommandPath())
				cmd.SetContext(ctx)
				defer observability.GetGlobalTracer().FinishSpan(span)
			}
			err := runE(cmd, args)
			if err != nil {
				observability.SetError(ctx, err)
				observability.AddEvent(ctx, "error", err.Error(), "error")
			}
			return err
		}
	}
	if run := cmd.Run; run != nil {
		cmd.Run = func(cmd *cobra.Command, args []string) {
			ctx, span := observability.GetGlobalTracer().StartSpan(cmd.Context(), traceOperation(cmd))
			if span != nil {
				observability.SetTag(ctx, "command", name)
				observability.SetTag(ctx, "group", group)
				observability.SetTag(ctx, "path", cmd.CommandPath())
				cmd.SetContext(ctx)
				defer observability.GetGlobalTracer().FinishSpan(span)
			}
			run(cmd, args)
		}
	}
}

func traceOperation(cmd *cobra.Command) string {
	if cmd == nil {
		return "command"
	}
	return fmt.Sprintf("command.%s", cmd.CommandPath())
}
