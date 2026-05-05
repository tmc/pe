package compat

import "fmt"

// Alias maps an old name to a new name.
type Alias struct {
	Old string
	New string
}

// Layer resolves compatibility aliases.
type Layer struct {
	aliases map[string]string
}

// NewLayer creates a compatibility layer.
func NewLayer(aliases ...Alias) *Layer {
	l := &Layer{aliases: make(map[string]string)}
	for _, alias := range aliases {
		l.aliases[alias.Old] = alias.New
	}
	return l
}

// Resolve returns the replacement for name, or name if no alias exists.
func (l *Layer) Resolve(name string) string {
	if next, ok := l.aliases[name]; ok {
		return next
	}
	return name
}

// DeprecatedWarning returns a warning for old names.
func DeprecatedWarning(old, next string) string {
	if next == "" {
		return fmt.Sprintf("%s is deprecated", old)
	}
	return fmt.Sprintf("%s is deprecated; use %s", old, next)
}

// MigrationStep describes one compatibility migration.
type MigrationStep struct {
	From string `json:"from"`
	To   string `json:"to"`
	Note string `json:"note,omitempty"`
}

// Plan returns migration steps for aliases in order.
func Plan(aliases []Alias) []MigrationStep {
	steps := make([]MigrationStep, 0, len(aliases))
	for _, alias := range aliases {
		steps = append(steps, MigrationStep{
			From: alias.Old,
			To:   alias.New,
			Note: DeprecatedWarning(alias.Old, alias.New),
		})
	}
	return steps
}
