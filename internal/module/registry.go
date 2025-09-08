// Package module provides a registry for PE modules
package module

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
}

// Module represents a PE module
type Module struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Repository  string            `json:"repository,omitempty"`
	Homepage    string            `json:"homepage,omitempty"`
	License     string            `json:"license,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	Files       []string          `json:"files,omitempty"`
	Checksum    string            `json:"checksum,omitempty"`
	PublishedAt time.Time         `json:"published_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// GitHubRegistry uses GitHub releases as a module registry
type GitHubRegistry struct {
	owner  string
	repo   string
	client *http.Client
}

// NewGitHubRegistry creates a new GitHub-based registry
func NewGitHubRegistry(owner, repo string) *GitHubRegistry {
	return &GitHubRegistry{
		owner:  owner,
		repo:   repo,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// List returns all modules from GitHub releases
func (r *GitHubRegistry) List() ([]*Module, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", r.owner, r.repo)
	
	resp, err := r.client.Get(url)
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
	moduleDir := filepath.Join(destDir, module.Name, module.Version)
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}
	
	// Download module files
	for _, file := range module.Files {
		url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
			r.owner, r.repo, module.Version, file)
		
		if err := r.downloadFile(url, filepath.Join(moduleDir, file)); err != nil {
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
	
	return results, nil
}

func (r *GitHubRegistry) fetchModuleMetadata(url string) (*Module, error) {
	resp, err := r.client.Get(url)
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
	resp, err := r.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	
	_, err = io.Copy(out, resp.Body)
	return err
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
			data, err := os.ReadFile(path)
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

// Get retrieves a specific module by name
func (r *LocalRegistry) Get(name string) (*Module, error) {
	// First, try to find any version of the module
	moduleDir := filepath.Join(r.rootDir, name)
	
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
	data, err := os.ReadFile(modulePath)
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
	sourceDir := filepath.Join(r.rootDir, module.Name, module.Version)
	targetDir := filepath.Join(destDir, module.Name, module.Version)
	
	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	
	// Copy all files
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		
		targetPath := filepath.Join(targetDir, relPath)
		
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		
		// Copy file
		source, err := os.Open(path)
		if err != nil {
			return err
		}
		defer source.Close()
		
		target, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer target.Close()
		
		_, err = io.Copy(target, source)
		return err
	})
}

// Publish adds a module to the local registry
func (r *LocalRegistry) Publish(module *Module, sourceDir string) error {
	targetDir := filepath.Join(r.rootDir, module.Name, module.Version)
	
	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}
	
	// Copy all files
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		
		targetPath := filepath.Join(targetDir, relPath)
		
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		
		// Copy file
		source, err := os.Open(path)
		if err != nil {
			return err
		}
		defer source.Close()
		
		target, err := os.Create(targetPath)
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

// Search searches for modules in the local registry
func (r *LocalRegistry) Search(query string) ([]*Module, error) {
	modules, err := r.List()
	if err != nil {
		return nil, err
	}
	
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
	
	return results, nil
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
		return NewGitHubRegistry(owner, repo)
		
	case "local":
		dir := os.Getenv("PE_REGISTRY_DIR")
		if dir == "" {
			home, _ := os.UserHomeDir()
			dir = filepath.Join(home, ".pe", "registry")
		}
		return NewLocalRegistry(dir)
		
	default:
		// Default to local registry
		home, _ := os.UserHomeDir()
		dir := filepath.Join(home, ".pe", "registry")
		return NewLocalRegistry(dir)
	}
}