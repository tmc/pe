// Package module provides a registry for PE modules
package module

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RegistryType defines the type of registry backend
type RegistryType string

const (
	// RegistryTypeGitHub uses GitHub releases as the registry
	RegistryTypeGitHub RegistryType = "github"
	// RegistryTypeGist uses GitHub gists as the registry
	RegistryTypeGist RegistryType = "gist"
	// RegistryTypeHTTP uses a simple HTTP server as the registry
	RegistryTypeHTTP RegistryType = "http"
	// RegistryTypeLocal uses a local directory as the registry
	RegistryTypeLocal RegistryType = "local"
)

// Registry interface for module registries
type Registry interface {
	// List returns all available modules
	List() ([]*Module, error)
	// Get retrieves a specific module by name
	Get(name string) (*Module, error)
	// Download downloads a module to the local cache
	Download(module *Module, destDir string) error
	// Publish publishes a module to the registry
	Publish(module *Module, sourceDir string) error
	// Search searches for modules matching a query
	Search(query string) ([]*Module, error)
	// Health checks whether the registry is reachable
	Health() error
}

// Module represents a PE module
type Module struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	Repository   string            `json:"repository,omitempty"`
	Homepage     string            `json:"homepage,omitempty"`
	License      string            `json:"license,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	Files        []string          `json:"files,omitempty"`
	Checksum     string            `json:"checksum,omitempty"`
	PublishedAt  time.Time         `json:"published_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// GitHubRegistry uses GitHub releases as a module registry
type GitHubRegistry struct {
	owner  string
	repo   string
	client *http.Client
	token  string
}

// NewGitHubRegistry creates a new GitHub-based registry
func NewGitHubRegistry(owner, repo string) *GitHubRegistry {
	return &GitHubRegistry{
		owner:  owner,
		repo:   repo,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// HTTPRegistry uses a static HTTP endpoint as a module registry.
type HTTPRegistry struct {
	baseURL string
	client  *http.Client
	token   string
}

// NewHTTPRegistry creates a new HTTP registry.
func NewHTTPRegistry(baseURL string) *HTTPRegistry {
	return &HTTPRegistry{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// SetToken configures bearer-token authentication for GitHub requests.
func (r *GitHubRegistry) SetToken(token string) {
	r.token = token
}

// SetToken configures bearer-token authentication for HTTP registry requests.
func (r *HTTPRegistry) SetToken(token string) {
	r.token = token
}

// List returns all modules from GitHub releases
func (r *GitHubRegistry) List() ([]*Module, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", r.owner, r.repo)

	resp, err := r.do(http.MethodGet, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []struct {
		TagName     string    `json:"tag_name"`
		Name        string    `json:"name"`
		Body        string    `json:"body"`
		PublishedAt time.Time `json:"published_at"`
		Assets      []struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases: %w", err)
	}

	var modules []*Module
	for _, release := range releases {
		// Look for module.json in assets
		for _, asset := range release.Assets {
			if asset.Name == "module.json" {
				module, err := r.fetchModuleMetadata(asset.DownloadURL)
				if err != nil {
					continue
				}
				modules = append(modules, module)
				break
			}
		}
	}

	return modules, nil
}

// Health checks whether the GitHub registry is reachable.
func (r *GitHubRegistry) Health() error {
	req, err := http.NewRequest(http.MethodHead, fmt.Sprintf("https://api.github.com/repos/%s/%s", r.owner, r.repo), nil)
	if err != nil {
		return err
	}
	r.authorize(req)
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	return nil
}

// Get retrieves a specific module by name
func (r *GitHubRegistry) Get(name string) (*Module, error) {
	modules, err := r.List()
	if err != nil {
		return nil, err
	}

	for _, module := range modules {
		if module.Name == name {
			return module, nil
		}
	}

	return nil, fmt.Errorf("module %s not found", name)
}

// Download downloads a module from GitHub
func (r *GitHubRegistry) Download(module *Module, destDir string) error {
	// Create destination directory
	moduleDir, err := containedPath(destDir, module.Name, module.Version)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}

	// Download module files
	for _, file := range module.Files {
		url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
			r.owner, r.repo, module.Version, file)

		destPath, err := containedPath(moduleDir, file)
		if err != nil {
			return err
		}
		if err := r.downloadFile(url, destPath); err != nil {
			return fmt.Errorf("failed to download %s: %w", file, err)
		}
	}

	// Save module metadata
	metadataPath := filepath.Join(moduleDir, "module.json")
	data, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal module metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save module metadata: %w", err)
	}

	if err := verifyModuleChecksum(moduleDir, module); err != nil {
		return err
	}

	return nil
}

// Publish publishes a module to GitHub (requires authentication)
func (r *GitHubRegistry) Publish(module *Module, sourceDir string) error {
	// This would require GitHub authentication and create a new release
	// For now, return an error indicating manual publishing is required
	return fmt.Errorf("automated publishing not yet implemented - please create a GitHub release manually")
}

// Search searches for modules
func (r *GitHubRegistry) Search(query string) ([]*Module, error) {
	modules, err := r.List()
	if err != nil {
		return nil, err
	}
	return searchModules(modules, query), nil
}

func (r *GitHubRegistry) fetchModuleMetadata(url string) (*Module, error) {
	resp, err := r.do(http.MethodGet, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var module Module
	if err := json.NewDecoder(resp.Body).Decode(&module); err != nil {
		return nil, err
	}

	return &module, nil
}

func (r *GitHubRegistry) downloadFile(url, destPath string) error {
	resp, err := r.do(http.MethodGet, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	out, err := os.Create(destPath) // #nosec G304 -- destPath is contained under the module download directory.
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (r *GitHubRegistry) do(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	r.authorize(req)
	return r.client.Do(req)
}

func (r *GitHubRegistry) authorize(req *http.Request) {
	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	}
}

// List returns all modules from the HTTP registry index.
func (r *HTTPRegistry) List() ([]*Module, error) {
	var modules []*Module
	if err := r.getJSON("/modules.json", &modules); err != nil {
		return nil, err
	}
	return modules, nil
}

// Health checks whether the HTTP registry index is reachable.
func (r *HTTPRegistry) Health() error {
	resp, err := r.do(http.MethodGet, "/modules.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP registry returned status %d", resp.StatusCode)
	}
	return nil
}

// Get retrieves a module by name from the HTTP registry.
func (r *HTTPRegistry) Get(name string) (*Module, error) {
	modules, err := r.List()
	if err != nil {
		return nil, err
	}
	for _, module := range modules {
		if module.Name == name {
			return module, nil
		}
	}
	return nil, fmt.Errorf("module %s not found", name)
}

// Download downloads a module from the HTTP registry.
func (r *HTTPRegistry) Download(module *Module, destDir string) error {
	moduleDir, err := containedPath(destDir, module.Name, module.Version)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}
	for _, file := range module.Files {
		destPath, err := containedPath(moduleDir, file)
		if err != nil {
			return err
		}
		urlPath := "/" + strings.TrimLeft(pathForModuleFile(module, file), "/")
		if err := r.downloadFile(urlPath, destPath); err != nil {
			return fmt.Errorf("failed to download %s: %w", file, err)
		}
	}
	data, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal module metadata: %w", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "module.json"), data, 0644); err != nil {
		return err
	}
	return verifyModuleChecksum(moduleDir, module)
}

// Publish returns an error because the HTTP registry is read-only.
func (r *HTTPRegistry) Publish(module *Module, sourceDir string) error {
	return fmt.Errorf("HTTP registry is read-only")
}

// Search searches modules in the HTTP registry index.
func (r *HTTPRegistry) Search(query string) ([]*Module, error) {
	modules, err := r.List()
	if err != nil {
		return nil, err
	}
	return searchModules(modules, query), nil
}

func (r *HTTPRegistry) getJSON(urlPath string, v interface{}) error {
	resp, err := r.do(http.MethodGet, urlPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP registry returned status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func (r *HTTPRegistry) downloadFile(urlPath, destPath string) error {
	resp, err := r.do(http.MethodGet, urlPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP registry returned status %d", resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	out, err := os.Create(destPath) // #nosec G304 -- destPath is contained under the module download directory.
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func (r *HTTPRegistry) do(method, urlPath string) (*http.Response, error) {
	req, err := http.NewRequest(method, r.baseURL+urlPath, nil)
	if err != nil {
		return nil, err
	}
	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	}
	return r.client.Do(req)
}

func pathForModuleFile(module *Module, file string) string {
	return path.Join("modules", module.Name, module.Version, file)
}

// LocalRegistry uses a local directory as a module registry
type LocalRegistry struct {
	rootDir string
}

// NewLocalRegistry creates a new local directory-based registry
func NewLocalRegistry(rootDir string) *LocalRegistry {
	return &LocalRegistry{
		rootDir: rootDir,
	}
}

// List returns all modules from the local directory
func (r *LocalRegistry) List() ([]*Module, error) {
	var modules []*Module

	// Walk through module directories
	err := filepath.Walk(r.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for module.json files
		if filepath.Base(path) == "module.json" {
			data, err := os.ReadFile(path) // #nosec G304 -- path comes from walking the configured local registry root.
			if err != nil {
				return nil // Skip invalid modules
			}

			var module Module
			if err := json.Unmarshal(data, &module); err != nil {
				return nil // Skip invalid modules
			}

			modules = append(modules, &module)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk registry directory: %w", err)
	}

	return modules, nil
}

// Health checks whether the local registry directory is accessible.
func (r *LocalRegistry) Health() error {
	info, err := os.Stat(r.rootDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("registry path is not a directory")
	}
	return nil
}

// Get retrieves a specific module by name
func (r *LocalRegistry) Get(name string) (*Module, error) {
	// First, try to find any version of the module
	moduleDir, err := containedPath(r.rootDir, name)
	if err != nil {
		return nil, err
	}

	// Check if module directory exists
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("module %s not found", name)
	}

	// List all versions
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read module directory: %w", err)
	}

	// Find the latest version (simple approach - take the last one alphabetically)
	var latestVersion string
	for _, entry := range entries {
		if entry.IsDir() {
			latestVersion = entry.Name()
		}
	}

	if latestVersion == "" {
		return nil, fmt.Errorf("no versions found for module %s", name)
	}

	// Read the module.json from the latest version
	modulePath := filepath.Join(moduleDir, latestVersion, "module.json")
	data, err := os.ReadFile(modulePath) // #nosec G304 -- modulePath is contained under the local registry root.
	if err != nil {
		return nil, fmt.Errorf("failed to read module: %w", err)
	}

	var module Module
	if err := json.Unmarshal(data, &module); err != nil {
		return nil, fmt.Errorf("failed to parse module metadata: %w", err)
	}

	return &module, nil
}

// Download copies a module from the local registry
func (r *LocalRegistry) Download(module *Module, destDir string) error {
	sourceDir, err := containedPath(r.rootDir, module.Name, module.Version)
	if err != nil {
		return err
	}
	targetDir, err := containedPath(destDir, module.Name, module.Version)
	if err != nil {
		return err
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy all files
	if err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		targetPath, err := containedPath(targetDir, relPath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file
		source, err := os.Open(path) // #nosec G304 -- path comes from walking the contained source module directory.
		if err != nil {
			return err
		}
		defer source.Close()

		target, err := os.Create(targetPath) // #nosec G304 -- targetPath is contained under the destination module directory.
		if err != nil {
			return err
		}
		defer target.Close()

		_, err = io.Copy(target, source)
		return err
	}); err != nil {
		return err
	}
	return verifyModuleChecksum(targetDir, module)
}

// Publish adds a module to the local registry
func (r *LocalRegistry) Publish(module *Module, sourceDir string) error {
	targetDir, err := containedPath(r.rootDir, module.Name, module.Version)
	if err != nil {
		return err
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}

	// Copy all files
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		targetPath, err := containedPath(targetDir, relPath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file
		source, err := os.Open(path) // #nosec G304 -- path comes from walking the caller-supplied source directory.
		if err != nil {
			return err
		}
		defer source.Close()

		target, err := os.Create(targetPath) // #nosec G304 -- targetPath is contained under the local registry root.
		if err != nil {
			return err
		}
		defer target.Close()

		_, err = io.Copy(target, source)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to copy module files: %w", err)
	}

	if module.Checksum == "" {
		sum, err := DirectoryChecksum(targetDir)
		if err != nil {
			return fmt.Errorf("failed to compute module checksum: %w", err)
		}
		module.Checksum = sum
	} else if err := verifyModuleChecksum(targetDir, module); err != nil {
		return err
	}

	// Save module metadata
	metadataPath := filepath.Join(targetDir, "module.json")
	data, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal module metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save module metadata: %w", err)
	}

	return nil
}

func verifyModuleChecksum(root string, module *Module) error {
	if module.Checksum == "" {
		return nil
	}
	sum, err := DirectoryChecksum(root)
	if err != nil {
		return fmt.Errorf("computing checksum for %s@%s: %w", module.Name, module.Version, err)
	}
	if sum != module.Checksum {
		return fmt.Errorf("checksum mismatch for %s@%s", module.Name, module.Version)
	}
	return nil
}

// DirectoryChecksum returns the deterministic sha256 checksum for a module
// directory. The checksum covers regular files except module.json, sorted by
// slash-separated relative path, and refuses symlinks.
func DirectoryChecksum(root string) (string, error) {
	var names []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root || entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlink %s", path)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if filepath.ToSlash(rel) == "module.json" {
			return nil
		}
		names = append(names, rel)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return "", err
		}
		hash.Write([]byte(filepath.ToSlash(name)))
		hash.Write([]byte{0})
		hash.Write(data)
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Search searches for modules in the local registry
func (r *LocalRegistry) Search(query string) ([]*Module, error) {
	modules, err := r.List()
	if err != nil {
		return nil, err
	}
	return searchModules(modules, query), nil
}

func searchModules(modules []*Module, query string) []*Module {
	query = strings.ToLower(query)
	var results []*Module

	for _, module := range modules {
		if strings.Contains(strings.ToLower(module.Name), query) ||
			strings.Contains(strings.ToLower(module.Description), query) {
			results = append(results, module)
			continue
		}

		for _, tag := range module.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, module)
				break
			}
		}
	}

	return results
}

// DefaultRegistry returns the default registry based on environment configuration
func DefaultRegistry() Registry {
	// Check for environment variable
	registryType := os.Getenv("PE_REGISTRY_TYPE")

	switch registryType {
	case "github":
		owner := os.Getenv("PE_REGISTRY_OWNER")
		repo := os.Getenv("PE_REGISTRY_REPO")
		if owner == "" {
			owner = "pe-modules"
		}
		if repo == "" {
			repo = "registry"
		}
		registry := NewGitHubRegistry(owner, repo)
		if token := os.Getenv("PE_REGISTRY_TOKEN"); token != "" {
			registry.SetToken(token)
		} else if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			registry.SetToken(token)
		}
		return registry

	case "local":
		dir := os.Getenv("PE_REGISTRY_DIR")
		if dir == "" {
			home, _ := os.UserHomeDir()
			dir = filepath.Join(home, ".pe", "registry")
		}
		return NewLocalRegistry(dir)

	case "http":
		url := os.Getenv("PE_REGISTRY_URL")
		if url == "" {
			url = "https://pe.dev/registry"
		}
		registry := NewHTTPRegistry(url)
		registry.SetToken(os.Getenv("PE_REGISTRY_TOKEN"))
		return registry

	default:
		// Default to local registry
		home, _ := os.UserHomeDir()
		dir := filepath.Join(home, ".pe", "registry")
		return NewLocalRegistry(dir)
	}
}
