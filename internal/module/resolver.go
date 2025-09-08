// Package module provides dependency resolution for PE modules
package module

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolver handles module resolution and dependency management
type Resolver struct {
	registry Registry
	cache    *Cache
	lockFile *LockFile
}

// NewResolver creates a new module resolver
func NewResolver(registry Registry, cacheDir string) *Resolver {
	return &Resolver{
		registry: registry,
		cache:    NewCache(cacheDir),
		lockFile: &LockFile{},
	}
}

// Resolve resolves a module and its dependencies
func (r *Resolver) Resolve(moduleName string, version string) (*Module, error) {
	// Check cache first
	if cached, err := r.cache.Get(moduleName, version); err == nil {
		return cached, nil
	}
	
	// Query registry
	module, err := r.registry.Get(moduleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get module %s: %w", moduleName, err)
	}
	
	// Check version compatibility
	if version != "" && !r.isVersionCompatible(module.Version, version) {
		return nil, fmt.Errorf("module %s version %s is not compatible with requested %s", 
			moduleName, module.Version, version)
	}
	
	// Download to cache
	if err := r.registry.Download(module, r.cache.dir); err != nil {
		return nil, fmt.Errorf("failed to download module %s: %w", moduleName, err)
	}
	
	// Cache the module
	if err := r.cache.Put(module); err != nil {
		return nil, fmt.Errorf("failed to cache module %s: %w", moduleName, err)
	}
	
	// Resolve dependencies recursively
	if err := r.resolveDependencies(module); err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies for %s: %w", moduleName, err)
	}
	
	return module, nil
}

// ResolveDependencies resolves all dependencies for a module
func (r *Resolver) resolveDependencies(module *Module) error {
	for depName, depVersion := range module.Dependencies {
		if _, err := r.Resolve(depName, depVersion); err != nil {
			return fmt.Errorf("failed to resolve dependency %s@%s: %w", depName, depVersion, err)
		}
	}
	return nil
}

// isVersionCompatible checks if a module version is compatible with a constraint
func (r *Resolver) isVersionCompatible(moduleVersion, constraint string) bool {
	// Simple version matching for now
	// TODO: Implement semantic versioning comparison
	if constraint == "" || constraint == "*" || constraint == "latest" {
		return true
	}
	
	// Exact match
	if moduleVersion == constraint {
		return true
	}
	
	// Range matching (e.g., "^1.0.0", "~1.2.0")
	if strings.HasPrefix(constraint, "^") {
		// Compatible with same major version
		major1 := strings.Split(moduleVersion, ".")[0]
		major2 := strings.Split(constraint[1:], ".")[0]
		return major1 == major2
	}
	
	if strings.HasPrefix(constraint, "~") {
		// Compatible with same minor version
		parts1 := strings.Split(moduleVersion, ".")
		parts2 := strings.Split(constraint[1:], ".")
		if len(parts1) >= 2 && len(parts2) >= 2 {
			return parts1[0] == parts2[0] && parts1[1] == parts2[1]
		}
	}
	
	return false
}

// Cache manages local module cache
type Cache struct {
	dir string
}

// NewCache creates a new module cache
func NewCache(dir string) *Cache {
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".pe", "cache")
	}
	os.MkdirAll(dir, 0755)
	return &Cache{dir: dir}
}

// Get retrieves a module from cache
func (c *Cache) Get(name, version string) (*Module, error) {
	modulePath := filepath.Join(c.dir, name, version, "module.json")
	
	data, err := os.ReadFile(modulePath)
	if err != nil {
		return nil, fmt.Errorf("module not in cache: %w", err)
	}
	
	var module Module
	if err := json.Unmarshal(data, &module); err != nil {
		return nil, fmt.Errorf("failed to parse cached module: %w", err)
	}
	
	return &module, nil
}

// Put stores a module in cache
func (c *Cache) Put(module *Module) error {
	modulePath := filepath.Join(c.dir, module.Name, module.Version, "module.json")
	
	if err := os.MkdirAll(filepath.Dir(modulePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	data, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal module: %w", err)
	}
	
	if err := os.WriteFile(modulePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write to cache: %w", err)
	}
	
	return nil
}

// Clear removes all cached modules
func (c *Cache) Clear() error {
	return os.RemoveAll(c.dir)
}

// LockFile manages module lock files (similar to go.sum)
type LockFile struct {
	Modules map[string]LockEntry `json:"modules"`
}

// LockEntry represents a locked module version
type LockEntry struct {
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
}

// Load loads a lock file from disk
func (l *LockFile) Load(path string) error {
	if path == "" {
		path = "pe.lock"
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize empty lock file
			l.Modules = make(map[string]LockEntry)
			return nil
		}
		return fmt.Errorf("failed to read lock file: %w", err)
	}
	
	if err := json.Unmarshal(data, l); err != nil {
		return fmt.Errorf("failed to parse lock file: %w", err)
	}
	
	if l.Modules == nil {
		l.Modules = make(map[string]LockEntry)
	}
	
	return nil
}

// Save saves the lock file to disk
func (l *LockFile) Save(path string) error {
	if path == "" {
		path = "pe.lock"
	}
	
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}
	
	return nil
}

// Add adds or updates a module in the lock file
func (l *LockFile) Add(module *Module) {
	if l.Modules == nil {
		l.Modules = make(map[string]LockEntry)
	}
	
	l.Modules[module.Name] = LockEntry{
		Version:  module.Version,
		Checksum: module.Checksum,
	}
}

// DependencyGraph manages module dependencies
type DependencyGraph struct {
	modules map[string]*Module
	edges   map[string][]string
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		modules: make(map[string]*Module),
		edges:   make(map[string][]string),
	}
}

// AddModule adds a module to the graph
func (g *DependencyGraph) AddModule(module *Module) {
	g.modules[module.Name] = module
	
	// Add edges for dependencies
	for depName := range module.Dependencies {
		g.edges[module.Name] = append(g.edges[module.Name], depName)
	}
}

// TopologicalSort returns modules in dependency order
func (g *DependencyGraph) TopologicalSort() ([]*Module, error) {
	// Calculate in-degrees
	inDegree := make(map[string]int)
	for name := range g.modules {
		inDegree[name] = 0
	}
	for _, deps := range g.edges {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}
	
	// Find modules with no dependencies
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}
	
	var sorted []*Module
	visited := make(map[string]bool)
	
	for len(queue) > 0 {
		// Pop from queue
		name := queue[0]
		queue = queue[1:]
		
		if visited[name] {
			continue
		}
		visited[name] = true
		
		// Add to sorted list
		if module, ok := g.modules[name]; ok {
			sorted = append(sorted, module)
		}
		
		// Reduce in-degree of dependencies
		for _, dep := range g.edges[name] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}
	
	// Check for cycles
	if len(sorted) != len(g.modules) {
		return nil, fmt.Errorf("circular dependency detected")
	}
	
	return sorted, nil
}

// DetectCycles detects circular dependencies
func (g *DependencyGraph) DetectCycles() ([][]string, error) {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}
	
	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)
		
		for _, neighbor := range g.edges[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				// Found a cycle
				cycleStart := 0
				for i, n := range path {
					if n == neighbor {
						cycleStart = i
						break
					}
				}
				cycle := append([]string{}, path[cycleStart:]...)
				cycles = append(cycles, cycle)
				return true
			}
		}
		
		path = path[:len(path)-1]
		recStack[node] = false
		return false
	}
	
	for node := range g.modules {
		if !visited[node] {
			dfs(node)
		}
	}
	
	if len(cycles) > 0 {
		return cycles, fmt.Errorf("found %d circular dependencies", len(cycles))
	}
	
	return nil, nil
}

// ResolveDependencies resolves all dependencies and returns them in order
func ResolveDependencies(modules []*Module) ([]*Module, error) {
	graph := NewDependencyGraph()
	
	// Build dependency graph
	for _, module := range modules {
		graph.AddModule(module)
	}
	
	// Check for cycles
	if cycles, err := graph.DetectCycles(); err != nil {
		// Format cycle information
		var cycleStrs []string
		for _, cycle := range cycles {
			cycleStrs = append(cycleStrs, strings.Join(cycle, " -> "))
		}
		return nil, fmt.Errorf("circular dependencies detected: %s", strings.Join(cycleStrs, "; "))
	}
	
	// Sort modules in dependency order
	sorted, err := graph.TopologicalSort()
	if err != nil {
		return nil, err
	}
	
	return sorted, nil
}

// VersionConstraint represents a version constraint
type VersionConstraint struct {
	Op      string // "=", ">", ">=", "<", "<=", "^", "~"
	Version string
}

// ParseVersionConstraint parses a version constraint string
func ParseVersionConstraint(constraint string) (*VersionConstraint, error) {
	constraint = strings.TrimSpace(constraint)
	
	if constraint == "" || constraint == "*" || constraint == "latest" {
		return &VersionConstraint{Op: "*", Version: ""}, nil
	}
	
	// Check for operators
	for _, op := range []string{">=", "<=", "^", "~", ">", "<", "="} {
		if strings.HasPrefix(constraint, op) {
			version := strings.TrimSpace(constraint[len(op):])
			return &VersionConstraint{Op: op, Version: version}, nil
		}
	}
	
	// No operator means exact match
	return &VersionConstraint{Op: "=", Version: constraint}, nil
}

// Matches checks if a version matches the constraint
func (c *VersionConstraint) Matches(version string) bool {
	if c.Op == "*" {
		return true
	}
	
	// Simple string comparison for now
	// TODO: Implement proper semantic versioning
	switch c.Op {
	case "=":
		return version == c.Version
	case "^":
		// Compatible with same major version
		major1 := strings.Split(version, ".")[0]
		major2 := strings.Split(c.Version, ".")[0]
		return major1 == major2
	case "~":
		// Compatible with same minor version
		parts1 := strings.Split(version, ".")
		parts2 := strings.Split(c.Version, ".")
		if len(parts1) >= 2 && len(parts2) >= 2 {
			return parts1[0] == parts2[0] && parts1[1] == parts2[1]
		}
	default:
		// For now, just do exact matching for other operators
		return version == c.Version
	}
	
	return false
}

// ResolveVersionConflicts resolves version conflicts between dependencies
func ResolveVersionConflicts(modules []*Module) (map[string]string, error) {
	versions := make(map[string][]string)
	
	// Collect all version requirements
	for _, module := range modules {
		for depName, depVersion := range module.Dependencies {
			versions[depName] = append(versions[depName], depVersion)
		}
	}
	
	resolved := make(map[string]string)
	
	// Resolve conflicts
	for name, versionList := range versions {
		// Remove duplicates
		unique := make(map[string]bool)
		for _, v := range versionList {
			unique[v] = true
		}
		
		if len(unique) == 1 {
			// No conflict
			for v := range unique {
				resolved[name] = v
			}
		} else {
			// Try to find compatible version
			// For now, just take the latest version
			// TODO: Implement proper version resolution
			var latest string
			for v := range unique {
				if latest == "" || v > latest {
					latest = v
				}
			}
			resolved[name] = latest
		}
	}
	
	return resolved, nil
}