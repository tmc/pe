package main

import (
	commandreg "github.com/tmc/pe/cmd/pe/commands"

	"github.com/spf13/cobra"
)

type rootCommandSpec struct {
	group string
	cmd   *cobra.Command
}

func registerRootCommands(root *cobra.Command) {
	registry := commandreg.NewRegistry()
	for _, spec := range rootCommandSpecs() {
		mustRegisterRootCommand(registry, spec)
	}
	registry.AddTo(root)
}

func mustRegisterRootCommand(registry *commandreg.CommandRegistry, spec rootCommandSpec) {
	if err := registry.Register(spec.group, spec.cmd, commandreg.Metadata{}); err != nil {
		panic(err)
	}
}

func rootCommandSpecs() []rootCommandSpec {
	return []rootCommandSpec{
		{group: "core", cmd: runCmd()},
		{group: "core", cmd: runTextCmd()},
		{group: "core", cmd: buildCmd},
		{group: "core", cmd: testCmd()},
		{group: "core", cmd: askCmd()},
		{group: "core", cmd: docCmd()},
		{group: "core", cmd: peInitCmd()},
		{group: "core", cmd: promptCmd},
		{group: "core", cmd: editCmd},
		{group: "core", cmd: workCmd},
		{group: "core", cmd: serveCmd()},

		{group: "evaluation", cmd: evalCmd()},
		{group: "evaluation", cmd: evalPromptCmd},
		{group: "evaluation", cmd: benchmarkCmd()},
		{group: "evaluation", cmd: diffCmd()},
		{group: "evaluation", cmd: statsCmd()},
		{group: "evaluation", cmd: viewCmd()},

		{group: "optimization", cmd: optimizeCmd()},
		{group: "optimization", cmd: semanticCmd()},
		{group: "optimization", cmd: evolveCmd()},
		{group: "optimization", cmd: textgradCmd()},
		{group: "optimization", cmd: pe2Cmd()},
		{group: "optimization", cmd: gasoCmd()},

		{group: "module", cmd: modCmd},
		{group: "module", cmd: pushCmd},
		{group: "module", cmd: getCmd},

		{group: "pipeline", cmd: streamCmd()},
		{group: "pipeline", cmd: filterCmd()},
		{group: "pipeline", cmd: analyzeCmd()},
		{group: "pipeline", cmd: collectCmd()},
		{group: "pipeline", cmd: reduceCmd()},
		{group: "pipeline", cmd: extractCmd()},
		{group: "pipeline", cmd: expandCmd()},
		{group: "pipeline", cmd: composeCmd},
		{group: "pipeline", cmd: catCmd()},

		{group: "utility", cmd: vetCmd()},
		{group: "utility", cmd: promptFmtCmd()},
		{group: "utility", cmd: convertCmd()},
		{group: "utility", cmd: watchCmd()},
		{group: "utility", cmd: templateCmd()},
		{group: "utility", cmd: interactiveCmd()},
		{group: "utility", cmd: versionCmd()},
		{group: "utility", cmd: configCmd()},

		{group: "experimental", cmd: experimentalCmd()},
		{group: "experimental", cmd: expCmd},
		{group: "experimental", cmd: securityCmd()},
		{group: "experimental", cmd: profileCmd()},

		{group: "plugin", cmd: pluginCmd()},
	}
}
