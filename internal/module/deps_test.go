package module

import "testing"

func TestPruneDependencies(t *testing.T) {
	modules := []*Module{
		{Name: "app", Dependencies: map[string]string{"used": "v1.0.0"}},
		{Name: "used", Dependencies: map[string]string{"transitive": "v1.0.0"}},
		{Name: "transitive"},
		{Name: "unused"},
	}
	got := PruneDependencies([]string{"app"}, modules)
	names := moduleNames(got)
	want := []string{"app", "transitive", "used"}
	if len(names) != len(want) {
		t.Fatalf("PruneDependencies returned %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("PruneDependencies returned %v, want %v", names, want)
		}
	}
}

func TestPruneDependenciesIgnoresMissingRootsAndDeps(t *testing.T) {
	modules := []*Module{
		{Name: "app", Dependencies: map[string]string{"missing": "v1.0.0"}},
		{Name: "unused"},
		nil,
	}
	got := PruneDependencies([]string{"missing-root", "app"}, modules)
	names := moduleNames(got)
	if len(names) != 1 || names[0] != "app" {
		t.Fatalf("PruneDependencies returned %v, want [app]", names)
	}
}

func moduleNames(modules []*Module) []string {
	names := make([]string, 0, len(modules))
	for _, mod := range modules {
		names = append(names, mod.Name)
	}
	return names
}
