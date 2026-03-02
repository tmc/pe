package providers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/llm"
)

func TestMLXBackends_EquivalentPromptResponses(t *testing.T) {
	binDir := t.TempDir()

	stub := `#!/bin/sh
prompt=""
while [ "$#" -gt 0 ]; do
	case "$1" in
		--prompt)
			shift
			prompt="$1"
			;;
	esac
	shift
done
printf "mlx-stub:%s\n" "$prompt"
`

	writeExecutable(t, filepath.Join(binDir, "mlx_lm.generate"), stub)
	writeExecutable(t, filepath.Join(binDir, "mlx-lm-generate"), stub)

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	providers := []string{
		"mlx-lm:test-model",
		"mlx-go:test-model",
	}

	tests := []struct {
		name   string
		prompt string
	}{
		{name: "simple", prompt: "Hello world"},
		{name: "quoted", prompt: `Say "hi" to Bob`},
		{name: "multispace", prompt: "A  B   C"},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(map[string]string)
			for _, spec := range providers {
				p, err := llm.GetProvider(spec)
				if err != nil {
					t.Fatalf("GetProvider(%q) failed: %v", spec, err)
				}

				resp, err := p.Generate(ctx, tt.prompt, llm.GenerateOptions{})
				if err != nil {
					t.Fatalf("Generate(%q) failed: %v", spec, err)
				}
				got[spec] = strings.TrimSpace(resp.Text)
			}

			want := "mlx-stub:" + tt.prompt
			if got["mlx-lm:test-model"] != want {
				t.Fatalf("mlx-lm output = %q, want %q", got["mlx-lm:test-model"], want)
			}
			if got["mlx-go:test-model"] != want {
				t.Fatalf("mlx-go output = %q, want %q", got["mlx-go:test-model"], want)
			}
		})
	}
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile(%q) failed: %v", path, err)
	}
}
