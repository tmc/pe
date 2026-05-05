package module

import "sort"

// PruneDependencies returns modules reachable from the named roots.
func PruneDependencies(roots []string, modules []*Module) []*Module {
	byName := make(map[string]*Module)
	for _, mod := range modules {
		if mod != nil {
			byName[mod.Name] = mod
		}
	}
	seen := make(map[string]bool)
	var walk func(string)
	walk = func(name string) {
		if seen[name] {
			return
		}
		mod := byName[name]
		if mod == nil {
			return
		}
		seen[name] = true
		var deps []string
		for dep := range mod.Dependencies {
			deps = append(deps, dep)
		}
		sort.Strings(deps)
		for _, dep := range deps {
			walk(dep)
		}
	}
	for _, root := range roots {
		walk(root)
	}
	var names []string
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	pruned := make([]*Module, 0, len(names))
	for _, name := range names {
		pruned = append(pruned, byName[name])
	}
	return pruned
}
