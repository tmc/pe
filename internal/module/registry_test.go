package module

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestHTTPRegistryListGetSearchDownload(t *testing.T) {
	module := Module{
		Name:        "example.com/prompts",
		Version:     "v1.0.0",
		Description: "useful prompts",
		Tags:        []string{"chat"},
		Files:       []string{"prompt.txt"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/modules.json":
			_ = json.NewEncoder(w).Encode([]Module{module})
		case "/modules/example.com/prompts/v1.0.0/prompt.txt":
			_, _ = w.Write([]byte("hello"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	registry := NewHTTPRegistry(server.URL + "/")
	modules, err := registry.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(modules) != 1 || modules[0].Name != module.Name {
		t.Fatalf("modules = %#v", modules)
	}
	got, err := registry.Get(module.Name)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Version != module.Version {
		t.Fatalf("version = %q", got.Version)
	}
	found, err := registry.Search("chat")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(found) != 1 || found[0].Name != module.Name {
		t.Fatalf("search = %#v", found)
	}
	dest := t.TempDir()
	if err := registry.Download(&module, dest); err != nil {
		t.Fatalf("Download: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dest, module.Name, module.Version, "prompt.txt"))
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("downloaded file = %q", data)
	}
}

func TestHTTPRegistryRejectsTraversal(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	registry := NewHTTPRegistry(server.URL)
	module := &Module{Name: "../escape", Version: "v1.0.0", Files: []string{"prompt.txt"}}
	if err := registry.Download(module, t.TempDir()); err == nil {
		t.Fatalf("Download traversal module succeeded")
	}
}

func TestHTTPRegistryHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/modules.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()
	if err := NewHTTPRegistry(server.URL).Health(); err != nil {
		t.Fatalf("Health: %v", err)
	}
}

func TestLocalRegistryHealth(t *testing.T) {
	if err := NewLocalRegistry(t.TempDir()).Health(); err != nil {
		t.Fatalf("Health: %v", err)
	}
	file := filepath.Join(t.TempDir(), "registry")
	if err := os.WriteFile(file, []byte("not a dir"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := NewLocalRegistry(file).Health(); err == nil {
		t.Fatalf("Health accepted file path")
	}
}
