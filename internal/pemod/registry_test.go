package pemod

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGistRegistryListGetResolveDownloadAndCache(t *testing.T) {
	oldwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(oldwd) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	index := map[string]ModuleInfo{
		"example/module": {
			Name:        "example/module",
			Description: "demo prompts",
			Tags:        []string{"demo"},
			Versions: []ModuleVersion{
				{Version: "v0.9.0", GistID: "old", Published: now.Add(-time.Hour), Deprecated: true, Files: []string{"prompt.txt"}},
				{Version: "v1.0.0", GistID: "module-gist", Published: now, Files: []string{"prompt.txt"}, Checksum: "sum"},
			},
		},
	}
	indexBytes, _ := json.Marshal(index)
	registry, err := NewGistRegistry("root", "token")
	if err != nil {
		t.Fatal(err)
	}
	transport := &gistRoundTripper{responses: map[string]Gist{
		"/gists/root":        {ID: "root", Files: map[string]GistFile{"index.json": {Content: string(indexBytes)}}},
		"/gists/module-gist": {ID: "module-gist", Files: map[string]GistFile{"prompt.txt": {Content: "hello"}}},
	}}
	registry.client.Transport = transport

	modules, err := registry.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(modules) != 1 || modules[0].Name != "example/module" {
		t.Fatalf("modules = %#v", modules)
	}
	got, err := registry.Get(context.Background(), "example/module")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Description != "demo prompts" {
		t.Fatalf("module = %#v", got)
	}
	latest, err := registry.Resolve(context.Background(), "example/module", "latest")
	if err != nil {
		t.Fatalf("Resolve latest: %v", err)
	}
	if latest.Version != "v1.0.0" {
		t.Fatalf("latest = %#v", latest)
	}
	exact, err := registry.Resolve(context.Background(), "example/module", "v0.9.0")
	if err != nil || exact.Version != "v0.9.0" {
		t.Fatalf("exact = %#v err=%v", exact, err)
	}
	dest := t.TempDir()
	if err := registry.Download(context.Background(), "example/module", "v1.0.0", dest); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(dest, "prompt.txt")); err != nil || string(data) != "hello" {
		t.Fatalf("downloaded data = %q err=%v", data, err)
	}
	if err := registry.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if transport.count["/gists/root"] != 1 {
		t.Fatalf("root fetched %d times", transport.count["/gists/root"])
	}
}

func TestGistRegistryErrorsAndPublish(t *testing.T) {
	oldwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(oldwd) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	registry, err := NewGistRegistry("root", "")
	if err != nil {
		t.Fatal(err)
	}
	registry.client.Transport = &gistRoundTripper{status: http.StatusNotFound}
	if _, err := registry.List(context.Background()); err == nil || !strings.Contains(err.Error(), "fetching root gist") {
		t.Fatalf("List error = %v", err)
	}
	if err := registry.Publish(context.Background(), &PromptModule{Name: "x"}, true); err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Fatalf("Publish without token error = %v", err)
	}

	registry.token = "token"
	updated := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	transport := &gistRoundTripper{
		create: &Gist{ID: "created"},
		responses: map[string]Gist{
			"/gists/root": {
				ID:          "root",
				Description: "Root registry",
				Files: map[string]GistFile{
					"index.json": {Content: `{"existing/module":{"name":"existing/module","versions":[{"version":"v1","gist_id":"old"}]}}`},
				},
			},
		},
	}
	registry.client.Transport = transport
	module := &PromptModule{Name: "x", Version: "v1", Description: "d", Author: "a", PromptFile: "prompt", Files: map[string]string{"extra.txt": "extra"}}
	module.Updated = updated
	if err := registry.Publish(context.Background(), module, false); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if module.GistID != "created" {
		t.Fatalf("gist id = %q", module.GistID)
	}
	if transport.patchPath != "/gists/root" {
		t.Fatalf("patch path = %q", transport.patchPath)
	}
	var update struct {
		Files map[string]struct {
			Content string `json:"content"`
		} `json:"files"`
	}
	if err := json.Unmarshal([]byte(transport.patchBody), &update); err != nil {
		t.Fatalf("patch body: %v", err)
	}
	var index map[string]ModuleInfo
	if err := json.Unmarshal([]byte(update.Files["index.json"].Content), &index); err != nil {
		t.Fatalf("index content: %v", err)
	}
	if _, ok := index["existing/module"]; !ok {
		t.Fatalf("existing module missing from index %#v", index)
	}
	got := index["x"]
	if got.Name != "x" || len(got.Versions) != 1 || got.Versions[0].GistID != "created" {
		t.Fatalf("published index entry = %#v", got)
	}
	if got.Versions[0].Checksum == "" || got.Versions[0].Size != int64(len("prompt")+len("extra")) {
		t.Fatalf("published version = %#v", got.Versions[0])
	}
	if _, err := registry.Resolve(context.Background(), "missing", "latest"); err == nil {
		t.Fatal("Resolve missing succeeded")
	}
}

func TestGistRegistryRejectsFileTraversal(t *testing.T) {
	oldwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(oldwd) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	registry, err := NewGistRegistry("root", "token")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	index := map[string]ModuleInfo{
		"x": {
			Name: "x",
			Versions: []ModuleVersion{
				{Version: "v1", GistID: "module-gist", Published: now},
			},
		},
	}
	indexBytes, _ := json.Marshal(index)
	registry.client.Transport = &gistRoundTripper{responses: map[string]Gist{
		"/gists/root": {ID: "root", Files: map[string]GistFile{
			"index.json": {Content: string(indexBytes)},
		}},
		"/gists/module-gist": {ID: "module-gist", Files: map[string]GistFile{
			"../escape.txt": {Content: "bad"},
		}},
	}}
	if err := registry.Download(context.Background(), "x", "v1", t.TempDir()); err == nil || !strings.Contains(err.Error(), "invalid registry file path") {
		t.Fatalf("Download traversal error = %v", err)
	}

	registry.client.Transport = &gistRoundTripper{create: &Gist{ID: "created"}}
	module := &PromptModule{Name: "x", Version: "v1", Files: map[string]string{"../escape.txt": "bad"}}
	if err := registry.Publish(context.Background(), module, false); err == nil || !strings.Contains(err.Error(), "invalid registry file path") {
		t.Fatalf("Publish traversal error = %v", err)
	}
}

func TestRegistryCache(t *testing.T) {
	oldwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(oldwd) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	cache, err := newRegistryCache()
	if err != nil {
		t.Fatal(err)
	}
	modules := []ModuleInfo{{Name: "a/b", Versions: []ModuleVersion{{Version: "v1"}}}}
	if err := cache.setModuleList(modules); err != nil {
		t.Fatal(err)
	}
	gotList, err := cache.getModuleList()
	if err != nil || len(gotList) != 1 || gotList[0].Name != "a/b" {
		t.Fatalf("cached list = %#v err=%v", gotList, err)
	}
	if err := cache.setModule("a/b", &modules[0]); err != nil {
		t.Fatal(err)
	}
	got, err := cache.getModule("a/b")
	if err != nil || got.Name != "a/b" {
		t.Fatalf("cached module = %#v err=%v", got, err)
	}
	if err := cache.close(); err != nil {
		t.Fatal(err)
	}
}

type gistRoundTripper struct {
	responses map[string]Gist
	create    *Gist
	status    int
	count     map[string]int
	patchPath string
	patchBody string
}

func (rt *gistRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.count == nil {
		rt.count = make(map[string]int)
	}
	rt.count[req.URL.Path]++
	status := rt.status
	if status == 0 {
		status = http.StatusOK
	}
	var body string
	if req.Method == http.MethodPost {
		status = http.StatusCreated
		gist := rt.create
		if gist == nil {
			gist = &Gist{ID: "created"}
		}
		data, _ := json.Marshal(gist)
		body = string(data)
	} else if req.Method == http.MethodPatch {
		status = http.StatusOK
		reqBody, _ := io.ReadAll(req.Body)
		rt.patchPath = req.URL.Path
		rt.patchBody = string(reqBody)
		body = `{"id":"root"}`
	} else if gist, ok := rt.responses[req.URL.Path]; ok {
		data, _ := json.Marshal(gist)
		body = string(data)
	} else {
		status = http.StatusNotFound
		body = `{"message":"missing"}`
	}
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       ioNopCloser{strings.NewReader(body)},
		Request:    req,
	}, nil
}

type ioNopCloser struct{ *strings.Reader }

func (c ioNopCloser) Close() error { return nil }
