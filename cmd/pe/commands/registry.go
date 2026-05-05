// Package commands provides command registration helpers.
package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// Metadata describes a registered command.
type Metadata struct {
	Name         string
	Group        string
	Summary      string
	Experimental bool
}

// CommandGroup is a named collection of commands.
type CommandGroup struct {
	Name     string
	Summary  string
	Commands []Metadata
}

// CommandRegistry stores commands by group.
type CommandRegistry struct {
	groups   map[string]*CommandGroup
	commands map[string]*cobra.Command
	metadata map[string]Metadata
}

// NewRegistry creates an empty command registry.
func NewRegistry() *CommandRegistry {
	return &CommandRegistry{
		groups:   make(map[string]*CommandGroup),
		commands: make(map[string]*cobra.Command),
		metadata: make(map[string]Metadata),
	}
}

// AddGroup adds a command group if it does not already exist.
func (r *CommandRegistry) AddGroup(name, summary string) {
	if r.groups[name] != nil {
		return
	}
	r.groups[name] = &CommandGroup{Name: name, Summary: summary}
}

// Register adds a command to a group.
func (r *CommandRegistry) Register(group string, cmd *cobra.Command, meta Metadata) error {
	if cmd == nil {
		return fmt.Errorf("command is nil")
	}
	name := cmd.Name()
	if name == "" {
		return fmt.Errorf("command has empty name")
	}
	if _, ok := r.commands[name]; ok {
		return fmt.Errorf("command %s already registered", name)
	}
	if r.groups[group] == nil {
		r.AddGroup(group, "")
	}
	meta.Name = name
	meta.Group = group
	if meta.Summary == "" {
		meta.Summary = cmd.Short
	}
	r.commands[name] = cmd
	r.metadata[name] = meta
	r.groups[group].Commands = append(r.groups[group].Commands, meta)
	return nil
}

// Command returns a command by name.
func (r *CommandRegistry) Command(name string) (*cobra.Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// Discover returns metadata for all commands sorted by name.
func (r *CommandRegistry) Discover() []Metadata {
	names := make([]string, 0, len(r.metadata))
	for name := range r.metadata {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]Metadata, 0, len(names))
	for _, name := range names {
		out = append(out, r.metadata[name])
	}
	return out
}

// Search returns commands whose name, group, or summary contains query.
func (r *CommandRegistry) Search(query string) []Metadata {
	query = strings.ToLower(query)
	var out []Metadata
	for _, meta := range r.Discover() {
		if query == "" ||
			strings.Contains(strings.ToLower(meta.Name), query) ||
			strings.Contains(strings.ToLower(meta.Group), query) ||
			strings.Contains(strings.ToLower(meta.Summary), query) {
			out = append(out, meta)
		}
	}
	return out
}

// Groups returns groups sorted by name.
func (r *CommandRegistry) Groups() []CommandGroup {
	names := make([]string, 0, len(r.groups))
	for name := range r.groups {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]CommandGroup, 0, len(names))
	for _, name := range names {
		group := *r.groups[name]
		sort.Slice(group.Commands, func(i, j int) bool {
			return group.Commands[i].Name < group.Commands[j].Name
		})
		out = append(out, group)
	}
	return out
}

// AddTo attaches all registered commands to root in discovery order.
func (r *CommandRegistry) AddTo(root *cobra.Command) {
	for _, meta := range r.Discover() {
		root.AddCommand(r.commands[meta.Name])
	}
}

// Help renders grouped command help.
func (r *CommandRegistry) Help() string {
	var b strings.Builder
	for _, group := range r.Groups() {
		fmt.Fprintf(&b, "%s\n", group.Name)
		if group.Summary != "" {
			fmt.Fprintf(&b, "  %s\n", group.Summary)
		}
		for _, cmd := range group.Commands {
			fmt.Fprintf(&b, "  %-14s %s\n", cmd.Name, cmd.Summary)
		}
	}
	return b.String()
}
