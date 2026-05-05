package pemod

import (
	"strings"
	"testing"
	"time"
)

func TestParseBasic(t *testing.T) {
	input := `pe 1`

	file, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if file.PE == nil || file.PE.Version != "1" {
		t.Errorf("Expected PE version 1, got %v", file.PE)
	}

	if len(file.Require) != 0 {
		t.Errorf("Expected no requirements, got %d", len(file.Require))
	}
}

func TestParseComplete(t *testing.T) {
	input := `module github.com/org/prompts

pe 1

require (
    github.com/org/math v1.0.0
    github.com/org/text v0.5.0 // indirect
)

replace github.com/org/math => ../local/math

exclude github.com/org/vulnerable v1.0.0

retract v0.9.0 // Contains security issue
retract [v0.8.0,v0.8.5] // Range retraction

trust github.com/org/math abc123fingerprint
trust github.com/org/trusted def456fingerprint valid 2024-01-01 2024-12-31

sign {
    key ~/.pe/keys/main.pem
    algorithm ed25519
    hsm pkcs11 slot=0 pin=1234
}

registry {
    default registry.pe.dev
    mirror cache.example.com
    private internal.company.com
    auth registry.pe.dev token123
}

security {
    require-signatures true
    allow-unsigned-dev true
    scan-content true
    policy vulnerability-check vulnerability severity=high action=fail
    policy license-check license severity=medium action=warn
}`

	file, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test module
	if file.Module == nil || file.Module.Mod != "github.com/org/prompts" {
		t.Errorf("Expected module github.com/org/prompts, got %v", file.Module)
	}

	// Test PE version
	if file.PE == nil || file.PE.Version != "1" {
		t.Errorf("Expected PE version 1, got %v", file.PE)
	}

	// Test requirements
	if len(file.Require) != 2 {
		t.Fatalf("Expected 2 requirements, got %d", len(file.Require))
	}

	req1 := file.Require[0]
	if req1.Mod != "github.com/org/math" || req1.Version != "v1.0.0" || req1.Indirect {
		t.Errorf("Expected direct requirement github.com/org/math v1.0.0, got %+v", req1)
	}

	req2 := file.Require[1]
	if req2.Mod != "github.com/org/text" || req2.Version != "v0.5.0" || !req2.Indirect {
		t.Errorf("Expected indirect requirement github.com/org/text v0.5.0, got %+v", req2)
	}

	// Test replacements
	if len(file.Replace) != 1 {
		t.Fatalf("Expected 1 replacement, got %d", len(file.Replace))
	}

	repl := file.Replace[0]
	if repl.Old != "github.com/org/math" || repl.New != "../local/math" {
		t.Errorf("Expected replacement github.com/org/math => ../local/math, got %+v", repl)
	}

	// Test exclusions
	if len(file.Exclude) != 1 {
		t.Fatalf("Expected 1 exclusion, got %d", len(file.Exclude))
	}

	exc := file.Exclude[0]
	if exc.Mod != "github.com/org/vulnerable" || exc.Version != "v1.0.0" {
		t.Errorf("Expected exclusion github.com/org/vulnerable v1.0.0, got %+v", exc)
	}

	// Test retractions
	if len(file.Retract) != 2 {
		t.Fatalf("Expected 2 retractions, got %d", len(file.Retract))
	}

	ret1 := file.Retract[0]
	if ret1.Low != "v0.9.0" || ret1.High != "v0.9.0" || ret1.Reason != "Contains security issue" {
		t.Errorf("Expected retraction v0.9.0 with reason, got %+v", ret1)
	}

	ret2 := file.Retract[1]
	if ret2.Low != "v0.8.0" || ret2.High != "v0.8.5" || ret2.Reason != "Range retraction" {
		t.Errorf("Expected range retraction [v0.8.0,v0.8.5], got %+v", ret2)
	}

	// Test trust
	if len(file.Trust) != 2 {
		t.Fatalf("Expected 2 trust entries, got %d", len(file.Trust))
	}

	trust1 := file.Trust[0]
	if trust1.Mod != "github.com/org/math" || trust1.Fingerprint != "abc123fingerprint" {
		t.Errorf("Expected trust github.com/org/math abc123fingerprint, got %+v", trust1)
	}

	trust2 := file.Trust[1]
	if trust2.Mod != "github.com/org/trusted" || trust2.Fingerprint != "def456fingerprint" {
		t.Errorf("Expected trust with validity period, got %+v", trust2)
	}
	expectedFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedUntil := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	if !trust2.ValidFrom.Equal(expectedFrom) || !trust2.ValidUntil.Equal(expectedUntil) {
		t.Errorf("Expected validity period 2024-01-01 to 2024-12-31, got %v to %v",
			trust2.ValidFrom, trust2.ValidUntil)
	}

	// Test signing configuration
	if file.Sign == nil {
		t.Fatal("Expected signing configuration")
	}
	if file.Sign.KeyPath != "~/.pe/keys/main.pem" || file.Sign.Algorithm != "ed25519" {
		t.Errorf("Expected key and algorithm, got %+v", file.Sign)
	}
	if file.Sign.HSM == nil || file.Sign.HSM.Provider != "pkcs11" {
		t.Errorf("Expected HSM configuration, got %+v", file.Sign.HSM)
	}

	// Test registry configuration
	if file.Registry == nil {
		t.Fatal("Expected registry configuration")
	}
	if file.Registry.Default != "registry.pe.dev" {
		t.Errorf("Expected default registry.pe.dev, got %s", file.Registry.Default)
	}
	if len(file.Registry.Mirrors) != 1 || file.Registry.Mirrors[0] != "cache.example.com" {
		t.Errorf("Expected mirror cache.example.com, got %v", file.Registry.Mirrors)
	}

	// Test security configuration
	if file.Security == nil {
		t.Fatal("Expected security configuration")
	}
	if !file.Security.RequireSignatures || !file.Security.AllowUnsignedDev || !file.Security.ScanContent {
		t.Errorf("Expected security flags to be true, got %+v", file.Security)
	}
	if len(file.Security.Policies) != 2 {
		t.Errorf("Expected 2 security policies, got %d", len(file.Security.Policies))
	}
}

func TestParseInlineRequire(t *testing.T) {
	input := `pe 1

require github.com/org/math v1.0.0`

	file, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(file.Require) != 1 {
		t.Fatalf("Expected 1 requirement, got %d", len(file.Require))
	}

	req := file.Require[0]
	if req.Mod != "github.com/org/math" || req.Version != "v1.0.0" {
		t.Errorf("Expected github.com/org/math v1.0.0, got %+v", req)
	}
}

func TestParseVersionRange(t *testing.T) {
	input := `pe 1

retract [v1.0.0,v1.0.9] // Security issues`

	file, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(file.Retract) != 1 {
		t.Fatalf("Expected 1 retraction, got %d", len(file.Retract))
	}

	ret := file.Retract[0]
	if ret.Low != "v1.0.0" || ret.High != "v1.0.9" || ret.Reason != "Security issues" {
		t.Errorf("Expected range retraction [v1.0.0,v1.0.9], got %+v", ret)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "missing PE version",
			input: `module test`,
			want:  "pe.mod must specify PE version",
		},
		{
			name:  "invalid PE version",
			input: `pe invalid`,
			want:  "invalid PE version",
		},
		{
			name:  "invalid require",
			input: "pe 1\nrequire",
			want:  "require directive needs module and version",
		},
		{
			name:  "invalid replace",
			input: "pe 1\nreplace old new",
			want:  "replace directive needs old => new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tt.input))
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Expected error containing %q, got %q", tt.want, err.Error())
			}
		})
	}
}

func TestFormat(t *testing.T) {
	input := `module github.com/org/prompts

pe 1

require (
	github.com/org/math v1.0.0
	github.com/org/text v0.5.0 // indirect
)

replace github.com/org/math => ../local/math

exclude github.com/org/vulnerable v1.0.0

retract v0.9.0 // Contains security issue

trust github.com/org/math abc123fingerprint

sign {
	key ~/.pe/keys/main.pem
	algorithm ed25519
}

registry {
	default registry.pe.dev
	mirror cache.example.com
}

security {
	require-signatures true
	allow-unsigned-dev true
	scan-content true
}`

	file, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	formatted := file.Format()

	// Verify key components are present
	expectedParts := []string{
		"module github.com/org/prompts",
		"pe 1",
		"require (",
		"github.com/org/math v1.0.0",
		"github.com/org/text v0.5.0 // indirect",
		"replace github.com/org/math => ../local/math",
		"exclude github.com/org/vulnerable v1.0.0",
		"retract v0.9.0 // Contains security issue",
		"trust github.com/org/math abc123fingerprint",
		"sign {",
		"registry {",
		"security {",
	}

	for _, part := range expectedParts {
		if !strings.Contains(formatted, part) {
			t.Errorf("Formatted output missing: %s\nGot:\n%s", part, formatted)
		}
	}
}

func TestAddRemoveRequire(t *testing.T) {
	file := &File{
		PE: &PEVersion{Version: "1"},
	}

	// Add requirement
	file.AddRequire("github.com/org/math", "v1.0.0")
	if len(file.Require) != 1 {
		t.Fatalf("Expected 1 requirement after add, got %d", len(file.Require))
	}

	req := file.GetRequire("github.com/org/math")
	if req == nil || req.Version != "v1.0.0" {
		t.Errorf("Expected requirement v1.0.0, got %+v", req)
	}

	// Update existing requirement
	file.AddRequire("github.com/org/math", "v1.1.0")
	if len(file.Require) != 1 {
		t.Fatalf("Expected 1 requirement after update, got %d", len(file.Require))
	}

	req = file.GetRequire("github.com/org/math")
	if req == nil || req.Version != "v1.1.0" {
		t.Errorf("Expected updated requirement v1.1.0, got %+v", req)
	}

	// Remove requirement
	file.RemoveRequire("github.com/org/math")
	if len(file.Require) != 0 {
		t.Fatalf("Expected 0 requirements after remove, got %d", len(file.Require))
	}

	req = file.GetRequire("github.com/org/math")
	if req != nil {
		t.Errorf("Expected nil after remove, got %+v", req)
	}
}

func TestSetModule(t *testing.T) {
	file := &File{}

	file.SetModule("github.com/org/test")
	if file.Module == nil || file.Module.Mod != "github.com/org/test" {
		t.Errorf("Expected module github.com/org/test, got %+v", file.Module)
	}
}

func TestSetPEVersion(t *testing.T) {
	file := &File{}

	file.SetPEVersion("1.2")
	if file.PE == nil || file.PE.Version != "1.2" {
		t.Errorf("Expected PE version 1.2, got %+v", file.PE)
	}
}

func TestValidPEVersion(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1", true},
		{"1.2", true},
		{"1.2.3", true},
		{"invalid", false},
		{"1.2.3.4", false},
		{"", false},
		{"1.a", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			valid := isValidPEVersion(tt.version)
			if valid != tt.valid {
				t.Errorf("isValidPEVersion(%q) = %v, want %v", tt.version, valid, tt.valid)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	input := `module github.com/org/prompts

pe 1

require (
	github.com/org/math v1.0.0
	github.com/org/text v0.5.0 // indirect
)

replace github.com/org/math => ../local/math

trust github.com/org/math abc123fingerprint`

	// Parse
	file, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Format
	formatted := file.Format()

	// Parse again
	file2, err := Parse(strings.NewReader(formatted))
	if err != nil {
		t.Fatalf("Second parse failed: %v", err)
	}

	// Compare
	if file.Module.Mod != file2.Module.Mod {
		t.Errorf("Module mismatch: %s vs %s", file.Module.Mod, file2.Module.Mod)
	}

	if file.PE.Version != file2.PE.Version {
		t.Errorf("PE version mismatch: %s vs %s", file.PE.Version, file2.PE.Version)
	}

	if len(file.Require) != len(file2.Require) {
		t.Errorf("Require count mismatch: %d vs %d", len(file.Require), len(file2.Require))
	}

	if len(file.Replace) != len(file2.Replace) {
		t.Errorf("Replace count mismatch: %d vs %d", len(file.Replace), len(file2.Replace))
	}

	if len(file.Trust) != len(file2.Trust) {
		t.Errorf("Trust count mismatch: %d vs %d", len(file.Trust), len(file2.Trust))
	}
}

func TestParseCapabilities(t *testing.T) {
	input := `module github.com/acme/prompts

pe 1

capability {
    data allow repo docs public
    data deny secrets credentials
    prompts allow local reviewed
    providers allow local test
    providers deny remote
    tools allow read search verify write
    tools deny shell network
}

placement {
    run local
    workspace isolated
    network false
    data-class public => providers local test remote
    data-class repo-internal => providers local test
}

policy {
    composition strict
    require-typed-io true
    require-reviewed-imports true
}
`
	file, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if file.Capability == nil {
		t.Fatal("Capability is nil")
	}
	if got := strings.Join(file.Capability.Providers.Deny, ","); got != "remote" {
		t.Fatalf("Providers.Deny = %q", got)
	}
	if file.Placement == nil || file.Placement.Network == nil || *file.Placement.Network {
		t.Fatalf("Placement.Network = %v", file.Placement)
	}
	if len(file.Placement.DataClassRules) != 2 {
		t.Fatalf("DataClassRules = %d", len(file.Placement.DataClassRules))
	}
	if file.Policy == nil || file.Policy.Composition != "strict" || !file.Policy.RequireTypedIO {
		t.Fatalf("Policy = %+v", file.Policy)
	}

	formatted := file.Format()
	for _, want := range []string{
		"capability {",
		"providers deny remote",
		"placement {",
		"network false",
		"policy {",
		"require-reviewed-imports true",
	} {
		if !strings.Contains(formatted, want) {
			t.Fatalf("formatted file missing %q:\n%s", want, formatted)
		}
	}
}
