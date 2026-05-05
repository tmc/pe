package module

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalRegistryRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	dest := t.TempDir()
	registry := NewLocalRegistry(root)

	module := &Module{Name: "../escape", Version: "v1.0.0"}
	if err := registry.Publish(module, t.TempDir()); err == nil {
		t.Fatal("Publish traversal module succeeded")
	}
	if err := registry.Download(module, dest); err == nil {
		t.Fatal("Download traversal module succeeded")
	}
	if _, err := registry.Get("../escape"); err == nil {
		t.Fatal("Get traversal module succeeded")
	}
	if _, err := os.Stat(filepath.Join(root, "..", "escape")); !os.IsNotExist(err) {
		t.Fatal("created path outside registry root")
	}
}

func TestGitHubRegistryDownloadRejectsTraversal(t *testing.T) {
	dest := t.TempDir()
	registry := NewGitHubRegistry("owner", "repo")

	tests := []Module{
		{Name: "../escape", Version: "v1.0.0"},
		{Name: "safe", Version: "v1.0.0", Files: []string{"../escape.txt"}},
	}
	for _, module := range tests {
		if err := registry.Download(&module, dest); err == nil {
			t.Fatalf("Download(%+v) succeeded", module)
		}
	}
	if _, err := os.Stat(filepath.Join(dest, "..", "escape")); !os.IsNotExist(err) {
		t.Fatal("created path outside download directory")
	}
}
