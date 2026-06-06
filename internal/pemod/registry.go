package pemod

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PromptModule represents a prompt module with associated metadata and files
type PromptModule struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	GistID      string            `json:"gist_id,omitempty"`
	Created     time.Time         `json:"created"`
	Updated     time.Time         `json:"updated"`
	PromptFile  string            `json:"prompt_file,omitempty"` // Main prompt content
	Files       map[string]string `json:"files,omitempty"`      // Additional files (filename -> content)
	Tags        []string          `json:"tags,omitempty"`
	License     string            `json:"license,omitempty"`
	Repository  string            `json:"repository,omitempty"`
}

// Gist represents a GitHub gist response
type Gist struct {
	ID          string                `json:"id"`
	Description string                `json:"description"`
	Public      bool                  `json:"public"`
	Files       map[string]GistFile   `json:"files"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Owner       *GistOwner            `json:"owner,omitempty"`
}

// GistFile represents a file within a gist
type GistFile struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Language string `json:"language"`
	RawURL   string `json:"raw_url"`
	Size     int    `json:"size"`
	Content  string `json:"content"`
}

// GistOwner represents the owner of a gist
type GistOwner struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Type  string `json:"type"`
}

// ModuleRegistry provides access to module registry operations
type ModuleRegistry interface {
	// List returns all available modules in the registry
	List(ctx context.Context) ([]ModuleInfo, error)
	
	// Get retrieves information about a specific module
	Get(ctx context.Context, name string) (*ModuleInfo, error)
	
	// Resolve finds the best version for a module given version constraints
	Resolve(ctx context.Context, name, version string) (*ModuleVersion, error)
	
	// Download downloads a module to the specified directory
	Download(ctx context.Context, name, version string, targetDir string) error
	
	// Publish publishes a module to the registry
	Publish(ctx context.Context, module *PromptModule, public bool) error
	
	// Close closes any resources used by the registry
	Close() error
}

// ModuleInfo contains basic information about a module
type ModuleInfo struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Versions    []ModuleVersion   `json:"versions"`
	Tags        []string          `json:"tags,omitempty"`
	License     string            `json:"license,omitempty"`
	Repository  string            `json:"repository,omitempty"`
	Created     time.Time         `json:"created"`
	Updated     time.Time         `json:"updated"`
}

// ModuleVersion contains information about a specific version of a module
type ModuleVersion struct {
	Version     string            `json:"version"`
	GistID      string            `json:"gist_id"`
	Published   time.Time         `json:"published"`
	Files       []string          `json:"files"`
	Size        int64             `json:"size"`
	Checksum    string            `json:"checksum"`
	Deprecated  bool              `json:"deprecated,omitempty"`
	PreRelease  bool              `json:"pre_release,omitempty"`
}

// GistRegistry implements ModuleRegistry using GitHub gists
type GistRegistry struct {
	rootGistID string
	token      string
	client     *http.Client
	cache      *registryCache
}

// NewGistRegistry creates a new gist-based registry
func NewGistRegistry(rootGistID, token string) (*GistRegistry, error) {
	if rootGistID == "" {
		// Use default public registry if available
		rootGistID = os.Getenv("PE_REGISTRY_GIST_ID")
		if rootGistID == "" {
			// Use well-known public registry (would be created by PE maintainers)
			rootGistID = "pe-public-registry"
		}
	}
	
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	
	cache, err := newRegistryCache()
	if err != nil {
		return nil, fmt.Errorf("creating registry cache: %w", err)
	}
	
	return &GistRegistry{
		rootGistID: rootGistID,
		token:      token,
		client:     &http.Client{Timeout: 30 * time.Second},
		cache:      cache,
	}, nil
}

// List returns all available modules in the registry
func (r *GistRegistry) List(ctx context.Context) ([]ModuleInfo, error) {
	// Check cache first
	if modules, err := r.cache.getModuleList(); err == nil && len(modules) > 0 {
		return modules, nil
	}
	
	// Fetch from GitHub
	gist, err := r.fetchGist(ctx, r.rootGistID)
	if err != nil {
		return nil, fmt.Errorf("fetching root gist: %w", err)
	}
	
	indexFile, ok := gist.Files["index.json"]
	if !ok {
		return nil, fmt.Errorf("registry index not found in root gist")
	}
	
	var index map[string]ModuleInfo
	if err := json.Unmarshal([]byte(indexFile.Content), &index); err != nil {
		return nil, fmt.Errorf("parsing registry index: %w", err)
	}
	
	modules := make([]ModuleInfo, 0, len(index))
	for _, module := range index {
		modules = append(modules, module)
	}
	
	// Cache the result
	r.cache.setModuleList(modules)
	
	return modules, nil
}

// Get retrieves information about a specific module
func (r *GistRegistry) Get(ctx context.Context, name string) (*ModuleInfo, error) {
	// Check cache first
	if module, err := r.cache.getModule(name); err == nil {
		return module, nil
	}
	
	modules, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	
	for _, module := range modules {
		if module.Name == name {
			r.cache.setModule(name, &module)
			return &module, nil
		}
	}
	
	return nil, fmt.Errorf("module %s not found in registry", name)
}

// Resolve finds the best version for a module given version constraints
func (r *GistRegistry) Resolve(ctx context.Context, name, version string) (*ModuleVersion, error) {
	module, err := r.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	
	if version == "latest" || version == "" {
		// Find the latest stable version
		var latest *ModuleVersion
		for i := range module.Versions {
			v := &module.Versions[i]
			if !v.Deprecated && !v.PreRelease {
				if latest == nil || v.Published.After(latest.Published) {
					latest = v
				}
			}
		}
		if latest != nil {
			return latest, nil
		}
		
		// Fall back to any version if no stable version found
		if len(module.Versions) > 0 {
			return &module.Versions[0], nil
		}
		
		return nil, fmt.Errorf("no versions available for module %s", name)
	}
	
	// Look for exact version match
	for i := range module.Versions {
		v := &module.Versions[i]
		if v.Version == version {
			return v, nil
		}
	}
	
	return nil, fmt.Errorf("version %s not found for module %s", version, name)
}

// Download downloads a module to the specified directory
func (r *GistRegistry) Download(ctx context.Context, name, version string, targetDir string) error {
	moduleVersion, err := r.Resolve(ctx, name, version)
	if err != nil {
		return err
	}
	
	// Fetch the module gist
	gist, err := r.fetchGist(ctx, moduleVersion.GistID)
	if err != nil {
		return fmt.Errorf("fetching module gist: %w", err)
	}
	
	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("creating target directory: %w", err)
	}
	
	// Save all files from the gist
	for filename, file := range gist.Files {
		filePath := filepath.Join(targetDir, filename)
		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			return fmt.Errorf("saving %s: %w", filename, err)
		}
	}
	
	// Create module metadata
	moduleInfo := map[string]interface{}{
		"name":      name,
		"version":   moduleVersion.Version,
		"gist_id":   moduleVersion.GistID,
		"published": moduleVersion.Published,
		"checksum":  moduleVersion.Checksum,
	}
	
	metaBytes, _ := json.MarshalIndent(moduleInfo, "", "  ")
	metaPath := filepath.Join(targetDir, "module.json")
	if err := os.WriteFile(metaPath, metaBytes, 0644); err != nil {
		return fmt.Errorf("saving module metadata: %w", err)
	}
	
	return nil
}

// Publish publishes a module to the registry
func (r *GistRegistry) Publish(ctx context.Context, module *PromptModule, public bool) error {
	if r.token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required for publishing")
	}
	
	// Create gist with module files
	gistReq := map[string]interface{}{
		"description": fmt.Sprintf("PE Module: %s", module.Name),
		"public":      public,
		"files":       make(map[string]interface{}),
	}
	
	// Add prompt file
	if module.PromptFile != "" {
		gistReq["files"].(map[string]interface{})["prompt.txt"] = map[string]interface{}{
			"content": module.PromptFile,
		}
	}
	
	// Add other files
	for filename, content := range module.Files {
		gistReq["files"].(map[string]interface{})[filename] = map[string]interface{}{
			"content": content,
		}
	}
	
	// Add module metadata
	metadata := map[string]interface{}{
		"name":        module.Name,
		"version":     module.Version,
		"description": module.Description,
		"author":      module.Author,
		"created":     module.Created,
		"updated":     module.Updated,
	}
	metaBytes, _ := json.MarshalIndent(metadata, "", "  ")
	gistReq["files"].(map[string]interface{})["module.json"] = map[string]interface{}{
		"content": string(metaBytes),
	}
	
	// Create the gist
	gistResp, err := r.createGist(ctx, gistReq)
	if err != nil {
		return fmt.Errorf("creating module gist: %w", err)
	}
	
	module.GistID = gistResp.ID
	
	// Update registry index
	return r.updateRegistryIndex(ctx, module)
}

// Close closes any resources used by the registry
func (r *GistRegistry) Close() error {
	return r.cache.close()
}

// fetchGist fetches a gist from GitHub
func (r *GistRegistry) fetchGist(ctx context.Context, gistID string) (*Gist, error) {
	url := fmt.Sprintf("https://api.github.com/gists/%s", gistID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	if r.token != "" {
		req.Header.Set("Authorization", "token "+r.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}
	
	var gist Gist
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return nil, err
	}
	
	return &gist, nil
}

// createGist creates a new gist on GitHub
func (r *GistRegistry) createGist(ctx context.Context, gistReq map[string]interface{}) (*Gist, error) {
	if r.token == "" {
		return nil, fmt.Errorf("GitHub token required for creating gist")
	}
	
	reqBytes, err := json.Marshal(gistReq)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/gists", strings.NewReader(string(reqBytes)))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "token "+r.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}
	
	var gist Gist
	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return nil, err
	}
	
	return &gist, nil
}

// updateRegistryIndex updates the registry index with the new module
func (r *GistRegistry) updateRegistryIndex(ctx context.Context, module *PromptModule) error {
	return fmt.Errorf("registry indexing is not yet implemented")
}

// registryCache provides caching for registry operations
type registryCache struct {
	cacheDir string
}

// newRegistryCache creates a new registry cache
func newRegistryCache() (*registryCache, error) {
	cacheDir := filepath.Join(".pe", "cache", "registry")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}
	
	return &registryCache{
		cacheDir: cacheDir,
	}, nil
}

// getModuleList retrieves cached module list
func (c *registryCache) getModuleList() ([]ModuleInfo, error) {
	path := filepath.Join(c.cacheDir, "modules.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var modules []ModuleInfo
	if err := json.Unmarshal(data, &modules); err != nil {
		return nil, err
	}
	
	return modules, nil
}

// setModuleList caches the module list
func (c *registryCache) setModuleList(modules []ModuleInfo) error {
	path := filepath.Join(c.cacheDir, "modules.json")
	data, err := json.MarshalIndent(modules, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// getModule retrieves cached module info
func (c *registryCache) getModule(name string) (*ModuleInfo, error) {
	path := filepath.Join(c.cacheDir, "module_"+strings.ReplaceAll(name, "/", "_")+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var module ModuleInfo
	if err := json.Unmarshal(data, &module); err != nil {
		return nil, err
	}
	
	return &module, nil
}

// setModule caches module info
func (c *registryCache) setModule(name string, module *ModuleInfo) error {
	path := filepath.Join(c.cacheDir, "module_"+strings.ReplaceAll(name, "/", "_")+".json")
	data, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// close closes the cache
func (c *registryCache) close() error {
	return nil
}
