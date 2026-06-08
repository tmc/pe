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
	tests := []Module{
		{Name: "../escape", Version: "v1.0.0", Files: []string{"prompt.txt"}},
		{Name: "example.com/prompts", Version: "v1.0.0", Files: []string{"../escape.txt"}},
	}
	for _, module := range tests {
		if err := registry.Download(&module, t.TempDir()); err == nil {
			t.Fatalf("Download(%+v) succeeded", module)
		}
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

func TestHTTPRegistrySendsBearerToken(t *testing.T) {
	module := Module{Name: "example.com/prompts", Version: "v1.0.0", Files: []string{"prompt.txt"}}
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		paths = append(paths, r.URL.Path)
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
	registry := NewHTTPRegistry(server.URL)
	registry.SetToken("test-token")
	if _, err := registry.List(); err != nil {
		t.Fatalf("List: %v", err)
	}
	if err := registry.Health(); err != nil {
		t.Fatalf("Health: %v", err)
	}
	if err := registry.Download(&module, t.TempDir()); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(paths) != 3 {
		t.Fatalf("paths = %v, want 3 requests", paths)
	}
}

func TestDefaultHTTPRegistryUsesToken(t *testing.T) {
	t.Setenv("PE_REGISTRY_TYPE", "http")
	t.Setenv("PE_REGISTRY_URL", "https://registry.example")
	t.Setenv("PE_REGISTRY_TOKEN", "test-token")
	registry, ok := DefaultRegistry().(*HTTPRegistry)
	if !ok {
		t.Fatalf("DefaultRegistry returned %T", registry)
	}
	if registry.token != "test-token" {
		t.Fatalf("token = %q", registry.token)
	}
}

func TestDefaultGitHubRegistryUsesToken(t *testing.T) {
	t.Setenv("PE_REGISTRY_TYPE", "github")
	t.Setenv("PE_REGISTRY_TOKEN", "test-token")
	registry, ok := DefaultRegistry().(*GitHubRegistry)
	if !ok {
		t.Fatalf("DefaultRegistry returned %T", registry)
	}
	if registry.token != "test-token" {
		t.Fatalf("token = %q", registry.token)
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

func TestLocalRegistryPublishListGetSearchDownload(t *testing.T) {
	root := t.TempDir()
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "prompt.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	module := &Module{
		Name:        "example.com/local",
		Version:     "v1.2.3",
		Description: "local prompts",
		Tags:        []string{"local", "chat"},
		Files:       []string{"prompt.txt"},
	}
	registry := NewLocalRegistry(root)
	if err := registry.Publish(module, source); err != nil {
		t.Fatalf("Publish: %v", err)
	}
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
	if err := registry.Download(module, dest); err != nil {
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

func TestLocalRegistrySkipsInvalidMetadataAndReportsMissing(t *testing.T) {
	root := t.TempDir()
	badDir := filepath.Join(root, "bad", "v1")
	if err := os.MkdirAll(badDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(badDir, "module.json"), []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}
	registry := NewLocalRegistry(root)
	modules, err := registry.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(modules) != 0 {
		t.Fatalf("modules = %#v", modules)
	}
	if _, err := registry.Get("missing/module"); err == nil {
		t.Fatal("Get missing module succeeded")
	}
	if _, err := registry.Get("bad"); err == nil {
		t.Fatal("Get invalid metadata succeeded")
	}
}

func TestHTTPRegistryErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/modules.json":
			http.Error(w, "nope", http.StatusTeapot)
		case "/modules/example.com/prompts/v1.0.0/missing.txt":
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	registry := NewHTTPRegistry(server.URL)
	if err := registry.Health(); err == nil {
		t.Fatal("Health accepted error status")
	}
	if _, err := registry.List(); err == nil {
		t.Fatal("List accepted error status")
	}
	if _, err := registry.Get("missing"); err == nil {
		t.Fatal("Get accepted list error")
	}
	if _, err := registry.Search("anything"); err == nil {
		t.Fatal("Search accepted list error")
	}
	if err := registry.Publish(&Module{}, t.TempDir()); err == nil {
		t.Fatal("HTTP Publish succeeded")
	}
	module := &Module{Name: "example.com/prompts", Version: "v1.0.0", Files: []string{"missing.txt"}}
	if err := registry.Download(module, t.TempDir()); err == nil {
		t.Fatal("Download accepted missing file")
	}
}

func TestDefaultRegistryVariants(t *testing.T) {
	t.Setenv("PE_REGISTRY_TYPE", "local")
	dir := t.TempDir()
	t.Setenv("PE_REGISTRY_DIR", dir)
	local, ok := DefaultRegistry().(*LocalRegistry)
	if !ok || local.rootDir != dir {
		t.Fatalf("local default = %#v", local)
	}

	t.Setenv("PE_REGISTRY_TYPE", "github")
	t.Setenv("PE_REGISTRY_OWNER", "owner")
	t.Setenv("PE_REGISTRY_REPO", "repo")
	t.Setenv("PE_REGISTRY_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "github-token")
	github, ok := DefaultRegistry().(*GitHubRegistry)
	if !ok || github.owner != "owner" || github.repo != "repo" || github.token != "github-token" {
		t.Fatalf("github default = %#v", github)
	}

	t.Setenv("PE_REGISTRY_TYPE", "")
	if _, ok := DefaultRegistry().(*LocalRegistry); !ok {
		t.Fatalf("empty registry type did not default to local")
	}

	t.Setenv("PE_REGISTRY_TYPE", "unsupported")
	if _, ok := DefaultRegistry().(*LocalRegistry); !ok {
		t.Fatalf("unsupported registry type did not default to local")
	}
}
