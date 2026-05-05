package module

import "testing"

func TestParseModulePath(t *testing.T) {
	tests := []struct {
		name        string
		ref         string
		wantPath    string
		wantVersion string
		wantErr     bool
	}{
		{name: "path only", ref: "example.com/prompts", wantPath: "example.com/prompts"},
		{name: "path and version", ref: "example.com/prompts@v1.2.3", wantPath: "example.com/prompts", wantVersion: "v1.2.3"},
		{name: "scoped path", ref: "github.com/acme/prompts/foo@latest", wantPath: "github.com/acme/prompts/foo", wantVersion: "latest"},
		{name: "empty", ref: "", wantErr: true},
		{name: "empty version", ref: "example.com/prompts@", wantErr: true},
		{name: "absolute", ref: "/example/prompts", wantErr: true},
		{name: "traversal", ref: "example.com/../prompts", wantErr: true},
		{name: "doubled slash", ref: "example.com//prompts", wantErr: true},
		{name: "backslash", ref: `example.com\prompts`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseModulePath(tt.ref)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseModulePath(%q) succeeded", tt.ref)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseModulePath(%q): %v", tt.ref, err)
			}
			if got.Path != tt.wantPath || got.Version != tt.wantVersion {
				t.Fatalf("ParseModulePath(%q) = %#v, want path %q version %q", tt.ref, got, tt.wantPath, tt.wantVersion)
			}
		})
	}
}
