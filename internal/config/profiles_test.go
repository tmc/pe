package config

import "testing"

func TestProfileSetResolve(t *testing.T) {
	profiles := ProfileSet{Profiles: map[string]Profile{
		"base": {
			Values: map[string]interface{}{
				"providers.default": "openai",
				"app.verbose":       false,
			},
		},
		"dev": {
			Inherits: "base",
			Values: map[string]interface{}{
				"providers.default": "anthropic",
			},
		},
	}}

	values, err := profiles.Resolve("dev")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if values["providers.default"] != "anthropic" {
		t.Fatalf("providers.default = %v, want anthropic", values["providers.default"])
	}
	if values["app.verbose"] != false {
		t.Fatalf("app.verbose = %v, want false", values["app.verbose"])
	}
}

func TestProfileSetResolveCycle(t *testing.T) {
	profiles := ProfileSet{Profiles: map[string]Profile{
		"a": {Inherits: "b"},
		"b": {Inherits: "a"},
	}}
	if _, err := profiles.Resolve("a"); err == nil {
		t.Fatal("Resolve cycle succeeded")
	}
}

func TestConfigManagerApplyProfile(t *testing.T) {
	manager, err := NewManager(WithConfigPaths())
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	profiles := ProfileSet{Profiles: map[string]Profile{
		"ci": {
			Values: map[string]interface{}{
				"providers.default": "anthropic",
				"app.verbose":       true,
			},
		},
	}}

	if err := manager.ApplyProfile(profiles, "ci"); err != nil {
		t.Fatalf("ApplyProfile: %v", err)
	}
	cfg := manager.Get()
	if cfg.Providers.Default != "anthropic" {
		t.Fatalf("providers.default = %q, want anthropic", cfg.Providers.Default)
	}
	if !cfg.App.Verbose {
		t.Fatal("app.verbose = false, want true")
	}
}
