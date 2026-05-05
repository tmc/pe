package main

import (
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"
)

const commandGroupAnnotation = "pe.command.group"

var commandGroups = map[string]string{
	"run":          "core",
	"run-text":     "core",
	"build":        "core",
	"test":         "core",
	"ask":          "core",
	"doc":          "core",
	"init":         "core",
	"prompt":       "core",
	"edit":         "core",
	"work":         "core",
	"serve":        "core",
	"eval":         "evaluation",
	"eval-prompt":  "evaluation",
	"benchmark":    "evaluation",
	"diff":         "evaluation",
	"stats":        "evaluation",
	"view":         "evaluation",
	"optimize":     "optimization",
	"semantic":     "optimization",
	"evolve":       "optimization",
	"mod":          "module",
	"push":         "module",
	"get":          "module",
	"stream":       "pipeline",
	"filter":       "pipeline",
	"analyze":      "pipeline",
	"collect":      "pipeline",
	"reduce":       "pipeline",
	"extract":      "pipeline",
	"expand":       "pipeline",
	"compose":      "pipeline",
	"cat":          "pipeline",
	"fmt":          "utility",
	"vet":          "utility",
	"convert":      "utility",
	"template":     "utility",
	"interactive":  "utility",
	"watch":        "utility",
	"config":       "utility",
	"version":      "utility",
	"security":     "experimental",
	"profile":      "experimental",
	"experimental": "experimental",
	"exp":          "experimental",
	"plugin":       "plugin",
}

var commandAliases = map[string][]string{
	"eval":     {"evaluate"},
	"fmt":      {"format"},
	"vet":      {"validate"},
	"run-text": {"render-text"},
}

// applyRootMetadata attaches compatibility aliases, group metadata, and grouped help.
func applyRootMetadata(root *cobra.Command) {
	for _, cmd := range root.Commands() {
		group := applyCommandMetadata(cmd, "")
		name := cmd.Name()
		if group == "experimental" && cmd.Deprecated == "" {
			cmd.Deprecated = "experimental command; behavior may change"
		}
		for _, alias := range commandAliases[name] {
			if !hasString(cmd.Aliases, alias) {
				cmd.Aliases = append(cmd.Aliases, alias)
			}
		}
	}
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd != root && cmd.Long != "" {
			fmt.Fprintln(cmd.OutOrStdout(), cmd.Long)
			fmt.Fprintln(cmd.OutOrStdout())
		}
		_ = cmd.Usage()
		if cmd == root {
			writeCommandGroups(cmd.OutOrStdout(), cmd)
		}
	})
}

func applyCommandMetadata(cmd *cobra.Command, inheritedGroup string) string {
	name := cmd.Name()
	group := inheritedGroup
	if group == "" {
		group = commandGroups[name]
	}
	if group != "" {
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations[commandGroupAnnotation] = group
	}
	for _, child := range cmd.Commands() {
		applyCommandMetadata(child, group)
	}
	return group
}

func writeCommandGroups(w io.Writer, root *cobra.Command) {
	groups := make(map[string][]*cobra.Command)
	for _, cmd := range root.Commands() {
		group := cmd.Annotations[commandGroupAnnotation]
		if group == "" {
			continue
		}
		groups[group] = append(groups[group], cmd)
	}
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Command groups:")
	for _, name := range names {
		cmds := groups[name]
		sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name() < cmds[j].Name() })
		fmt.Fprintf(w, "  %s:\n", name)
		for _, cmd := range cmds {
			fmt.Fprintf(w, "    %-14s %s\n", cmd.Name(), cmd.Short)
		}
	}
}

func hasString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
