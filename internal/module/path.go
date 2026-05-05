package module

import (
	"fmt"
	"path"
	"strings"
)

// ModulePath is a parsed module reference of the form path or path@version.
type ModulePath struct {
	Path    string
	Version string
}

// ParseModulePath parses a module reference.
func ParseModulePath(ref string) (ModulePath, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ModulePath{}, fmt.Errorf("module path is empty")
	}
	if strings.Contains(ref, "\\") {
		return ModulePath{}, fmt.Errorf("module path %q contains backslash", ref)
	}

	mod := ModulePath{Path: ref}
	if i := strings.LastIndexByte(ref, '@'); i >= 0 {
		mod.Path = ref[:i]
		mod.Version = ref[i+1:]
		if mod.Version == "" {
			return ModulePath{}, fmt.Errorf("module path %q has empty version", ref)
		}
	}
	if !validModulePath(mod.Path) {
		return ModulePath{}, fmt.Errorf("invalid module path %q", mod.Path)
	}
	return mod, nil
}

func validModulePath(p string) bool {
	if p == "" || strings.HasPrefix(p, "/") || strings.HasSuffix(p, "/") {
		return false
	}
	if path.Clean(p) != p {
		return false
	}
	for _, elem := range strings.Split(p, "/") {
		if elem == "" || elem == "." || elem == ".." {
			return false
		}
	}
	return true
}
