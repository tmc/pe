package module

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCache(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	if cache == nil {
		t.Fatal("NewCache returned nil")
	}
	if cache.dir != tmpDir {
		t.Errorf("expected dir %s, got %s", tmpDir, cache.dir)
	}
}

func TestNewCache_DefaultDir(t *testing.T) {
	cache := NewCache("")
	if cache == nil {
		t.Fatal("NewCache returned nil")
	}
	if cache.dir == "" {
		t.Error("expected default dir to be set")
	}
}

func TestCache_PutAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	module := &Module{
		Name:        "test-module",
		Version:     "v1.0.0",
		Description: "A test module",
		Author:      "test",
	}

	// Put module
	err := cache.Put(module)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get module
	retrieved, err := cache.Get(module.Name, module.Version)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != module.Name {
		t.Errorf("expected name %s, got %s", module.Name, retrieved.Name)
	}
	if retrieved.Version != module.Version {
		t.Errorf("expected version %s, got %s", module.Version, retrieved.Version)
	}
}

func TestCache_RejectsTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	if err := cache.Put(&Module{Name: "../escape", Version: "v1.0.0"}); err == nil {
		t.Fatal("Put traversal module succeeded")
	}
	if _, err := cache.Get("../escape", "v1.0.0"); err == nil {
		t.Fatal("Get traversal module succeeded")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "..", "escape")); !os.IsNotExist(err) {
		t.Fatal("created path outside cache")
	}
}

func TestCache_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	_, err := cache.Get("nonexistent", "v1.0.0")
	if err == nil {
		t.Error("expected error for nonexistent module")
	}
}

func TestCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	// Put a module
	module := &Module{
		Name:    "test-module",
		Version: "v1.0.0",
	}
	if err := cache.Put(module); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Clear cache
	if err := cache.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify cache is empty
	_, err := cache.Get(module.Name, module.Version)
	if err == nil {
		t.Error("expected error after clearing cache")
	}
}

func TestLockFile_LoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "pe.lock")

	lf := &LockFile{}

	// Load non-existent file should initialize empty
	err := lf.Load(lockPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if lf.Modules == nil {
		t.Error("expected Modules to be initialized")
	}

	// Add a module
	lf.Add(&Module{
		Name:     "test-module",
		Version:  "v1.0.0",
		Checksum: "abc123",
	})

	// Save
	err = lf.Save(lockPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("lock file was not created")
	}

	// Load again
	lf2 := &LockFile{}
	err = lf2.Load(lockPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(lf2.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(lf2.Modules))
	}
	if entry, ok := lf2.Modules["test-module"]; !ok {
		t.Error("expected test-module in lock file")
	} else {
		if entry.Version != "v1.0.0" {
			t.Errorf("expected version v1.0.0, got %s", entry.Version)
		}
		if entry.Checksum != "abc123" {
			t.Errorf("expected checksum abc123, got %s", entry.Checksum)
		}
	}
}

func TestLockFile_Load_DefaultPath(t *testing.T) {
	lf := &LockFile{}
	// This will try to load from current dir, should not panic
	_ = lf.Load("")
}

func TestLockFile_Add(t *testing.T) {
	lf := &LockFile{}

	module := &Module{
		Name:     "test-module",
		Version:  "v1.0.0",
		Checksum: "abc123",
	}

	lf.Add(module)

	if lf.Modules == nil {
		t.Fatal("Modules should not be nil after Add")
	}

	entry, ok := lf.Modules["test-module"]
	if !ok {
		t.Fatal("expected module to be added")
	}

	if entry.Version != "v1.0.0" {
		t.Errorf("expected version v1.0.0, got %s", entry.Version)
	}
}

func TestNewDependencyGraph(t *testing.T) {
	graph := NewDependencyGraph()

	if graph == nil {
		t.Fatal("NewDependencyGraph returned nil")
	}
	if graph.modules == nil {
		t.Error("modules map should be initialized")
	}
	if graph.edges == nil {
		t.Error("edges map should be initialized")
	}
}

func TestDependencyGraph_AddModule(t *testing.T) {
	graph := NewDependencyGraph()

	module := &Module{
		Name:    "test-module",
		Version: "v1.0.0",
		Dependencies: map[string]string{
			"dep1": "v1.0.0",
			"dep2": "v2.0.0",
		},
	}

	graph.AddModule(module)

	if _, ok := graph.modules["test-module"]; !ok {
		t.Error("module should be added to graph")
	}

	edges := graph.edges["test-module"]
	if len(edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(edges))
	}
}

func TestDependencyGraph_TopologicalSort(t *testing.T) {
	graph := NewDependencyGraph()

	// Add modules in order: A depends on B, B depends on C
	moduleC := &Module{Name: "C", Version: "v1.0.0"}
	moduleB := &Module{Name: "B", Version: "v1.0.0", Dependencies: map[string]string{"C": "v1.0.0"}}
	moduleA := &Module{Name: "A", Version: "v1.0.0", Dependencies: map[string]string{"B": "v1.0.0"}}

	graph.AddModule(moduleA)
	graph.AddModule(moduleB)
	graph.AddModule(moduleC)

	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	// C should come before B, B should come before A
	if len(sorted) != 3 {
		t.Errorf("expected 3 modules, got %d", len(sorted))
	}
}

func TestDependencyGraph_DetectCycles_NoCycle(t *testing.T) {
	graph := NewDependencyGraph()

	moduleA := &Module{Name: "A", Version: "v1.0.0", Dependencies: map[string]string{"B": "v1.0.0"}}
	moduleB := &Module{Name: "B", Version: "v1.0.0"}

	graph.AddModule(moduleA)
	graph.AddModule(moduleB)

	cycles, err := graph.DetectCycles()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cycles != nil {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestDependencyGraph_DetectCycles_HasCycle(t *testing.T) {
	graph := NewDependencyGraph()

	// Create a cycle: A -> B -> C -> A
	moduleA := &Module{Name: "A", Version: "v1.0.0", Dependencies: map[string]string{"B": "v1.0.0"}}
	moduleB := &Module{Name: "B", Version: "v1.0.0", Dependencies: map[string]string{"C": "v1.0.0"}}
	moduleC := &Module{Name: "C", Version: "v1.0.0", Dependencies: map[string]string{"A": "v1.0.0"}}

	graph.AddModule(moduleA)
	graph.AddModule(moduleB)
	graph.AddModule(moduleC)

	cycles, err := graph.DetectCycles()
	if err == nil {
		t.Error("expected error for cycle detection")
	}
	if cycles == nil {
		t.Error("expected cycles to be detected")
	}
}

func TestParseVersionConstraint(t *testing.T) {
	tests := []struct {
		input   string
		wantOp  string
		wantVer string
	}{
		{"", "*", ""},
		{"*", "*", ""},
		{"latest", "*", ""},
		{"1.0.0", "=", "1.0.0"},
		{"=1.0.0", "=", "1.0.0"},
		{"^1.0.0", "^", "1.0.0"},
		{"~1.2.0", "~", "1.2.0"},
		{">=1.0.0", ">=", "1.0.0"},
		{"<=1.0.0", "<=", "1.0.0"},
		{">1.0.0", ">", "1.0.0"},
		{"<1.0.0", "<", "1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			constraint, err := ParseVersionConstraint(tt.input)
			if err != nil {
				t.Fatalf("ParseVersionConstraint failed: %v", err)
			}

			if constraint.Op != tt.wantOp {
				t.Errorf("Op: expected %s, got %s", tt.wantOp, constraint.Op)
			}
			if constraint.Version != tt.wantVer {
				t.Errorf("Version: expected %s, got %s", tt.wantVer, constraint.Version)
			}
		})
	}
}

func TestVersionConstraint_Matches(t *testing.T) {
	tests := []struct {
		name       string
		op         string
		constraint string
		version    string
		want       bool
	}{
		{"wildcard matches all", "*", "", "1.0.0", true},
		{"exact match", "=", "1.0.0", "1.0.0", true},
		{"exact no match", "=", "1.0.0", "2.0.0", false},
		{"caret same major", "^", "1.0.0", "1.5.0", true},
		{"caret different major", "^", "1.0.0", "2.0.0", false},
		{"tilde same minor", "~", "1.2.0", "1.2.5", true},
		{"tilde different minor", "~", "1.2.0", "1.3.0", false},
		{"greater than", ">", "1.2.0", "1.2.1", true},
		{"greater than no match", ">", "1.2.0", "1.2.0", false},
		{"greater equal exact", ">=", "1.2.0", "1.2.0", true},
		{"less than", "<", "2.0.0", "1.9.9", true},
		{"less equal exact", "<=", "2.0.0", "2.0.0", true},
		{"v prefix", "=", "v1.2.3", "1.2.3", true},
		{"prerelease compares base", "=", "1.2.3-beta.1", "1.2.3", true},
		{"invalid version no match", "=", "1.2.3", "not-a-version", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := &VersionConstraint{Op: tt.op, Version: tt.constraint}
			got := constraint.Matches(tt.version)
			if got != tt.want {
				t.Errorf("Matches(%s): expected %v, got %v", tt.version, tt.want, got)
			}
		})
	}
}

func TestResolveVersionConflicts(t *testing.T) {
	modules := []*Module{
		{
			Name: "A",
			Dependencies: map[string]string{
				"shared": "1.0.0",
			},
		},
		{
			Name: "B",
			Dependencies: map[string]string{
				"shared": "1.0.0",
			},
		},
	}

	resolved, err := ResolveVersionConflicts(modules)
	if err != nil {
		t.Fatalf("ResolveVersionConflicts failed: %v", err)
	}

	if resolved["shared"] != "1.0.0" {
		t.Errorf("expected shared version 1.0.0, got %s", resolved["shared"])
	}
}

func TestResolveVersionConflicts_Different(t *testing.T) {
	modules := []*Module{
		{
			Name: "A",
			Dependencies: map[string]string{
				"shared": "1.0.0",
			},
		},
		{
			Name: "B",
			Dependencies: map[string]string{
				"shared": "2.0.0",
			},
		},
	}

	resolved, err := ResolveVersionConflicts(modules)
	if err != nil {
		t.Fatalf("ResolveVersionConflicts failed: %v", err)
	}

	// Should pick the "latest" (lexicographically larger)
	if resolved["shared"] != "2.0.0" {
		t.Errorf("expected shared version 2.0.0, got %s", resolved["shared"])
	}
}

func TestResolveDependencies(t *testing.T) {
	modules := []*Module{
		{Name: "A", Dependencies: map[string]string{"B": "1.0.0"}},
		{Name: "B", Dependencies: map[string]string{}},
	}

	sorted, err := ResolveDependencies(modules)
	if err != nil {
		t.Fatalf("ResolveDependencies failed: %v", err)
	}

	if len(sorted) != 2 {
		t.Errorf("expected 2 modules, got %d", len(sorted))
	}
}

func TestResolver_isVersionCompatible(t *testing.T) {
	resolver := &Resolver{}

	tests := []struct {
		moduleVer  string
		constraint string
		want       bool
	}{
		{"1.0.0", "", true},
		{"1.0.0", "*", true},
		{"1.0.0", "latest", true},
		{"1.0.0", "1.0.0", true},
		{"1.0.0", "2.0.0", false},
		{"1.5.0", "^1.0.0", true},
		{"2.0.0", "^1.0.0", false},
		{"1.2.5", "~1.2.0", true},
		{"1.3.0", "~1.2.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.moduleVer+"_"+tt.constraint, func(t *testing.T) {
			got := resolver.isVersionCompatible(tt.moduleVer, tt.constraint)
			if got != tt.want {
				t.Errorf("isVersionCompatible(%s, %s): expected %v, got %v",
					tt.moduleVer, tt.constraint, tt.want, got)
			}
		})
	}
}

func TestModule_Struct(t *testing.T) {
	module := Module{
		Name:        "test-module",
		Version:     "v1.0.0",
		Description: "A test module",
		Author:      "test",
		Repository:  "https://github.com/test/module",
		License:     "MIT",
		Tags:        []string{"test", "example"},
		Dependencies: map[string]string{
			"dep1": "v1.0.0",
		},
	}

	if module.Name != "test-module" {
		t.Error("Name mismatch")
	}
	if module.Version != "v1.0.0" {
		t.Error("Version mismatch")
	}
	if len(module.Tags) != 2 {
		t.Error("Tags length mismatch")
	}
	if len(module.Dependencies) != 1 {
		t.Error("Dependencies length mismatch")
	}
}
