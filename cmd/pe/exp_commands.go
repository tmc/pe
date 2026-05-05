package main

import (
	"github.com/spf13/cobra"
)

func init() {
	// Register all experimental subcommands
	expCmd.AddCommand(expDistributedCmd)
	expCmd.AddCommand(expAttestCmd)
	expCmd.AddCommand(expCacheCmd)
	expCmd.AddCommand(expComposeCmd)

	expCmd.AddCommand(expTransformCmd)
	expCmd.AddCommand(expMergeCmd)
	expCmd.AddCommand(expBatchCmd)
	expCmd.AddCommand(expSweepCmd)
	expCmd.AddCommand(expScheduleCmd)
	expCmd.AddCommand(expTraceCmd)
	expCmd.AddCommand(expExplainCmd)
	expCmd.AddCommand(expLintCmd)
	expCmd.AddCommand(expImportCmd)
	expCmd.AddCommand(expExportCmd)
	expCmd.AddCommand(expSyncCmd)
	expCmd.AddCommand(expWorkflowCmd)
	expCmd.AddCommand(expHookCmd)
	expCmd.AddCommand(expReportCmd)

	// Move implementation real commands here later if needed
	// For now these are just stubs unless they link to existing implementations
}

func createStubCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Command '%s' is an experimental prototype.\n", use)
			cmd.Println("This feature is planned but not yet fully implemented.")
		},
	}
}

// Disabled/Moved commands
var expDistributedCmd = newExpDistributedCmd()
var expComposeCmd = createStubCmd("compose", "Compose prompts from verified components with type-safe composition")

// Planned commands
var expTransformCmd = createStubCmd("transform", "Transform prompts between formats and styles")
var expMergeCmd = createStubCmd("merge", "Intelligently merge prompts and prompt components")
var expBatchCmd = createStubCmd("batch", "Process multiple prompts efficiently")
var expSweepCmd = createStubCmd("sweep", "Parameter sweeping for optimization")
var expScheduleCmd = createStubCmd("schedule", "Schedule and orchestrate prompt execution")
var expTraceCmd = createStubCmd("trace", "Detailed execution tracing")
var expExplainCmd = createStubCmd("explain", "Explain prompt behavior and provider responses")
var expLintCmd = createStubCmd("lint", "Lint prompts for best practices")
var expImportCmd = createStubCmd("import", "Import prompts from external sources")
var expExportCmd = createStubCmd("export", "Export prompts to external formats")
var expSyncCmd = createStubCmd("sync", "Synchronize prompts with external systems")
var expWorkflowCmd = createStubCmd("workflow", "Define and execute complex workflows")
var expHookCmd = createStubCmd("hook", "Manage lifecycle hooks")
var expReportCmd = createStubCmd("report", "Generate comprehensive reports")
