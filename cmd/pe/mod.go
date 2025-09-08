package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/module"
	"github.com/tmc/pe/internal/pemod"
)

var modCmd = &cobra.Command{
	Use:   "mod",
	Short: "Module management for prompts",
	Long: `Manage prompt modules using GitHub gists as a registry.

PE uses a root gist that tracks forks containing prompt modules.
Each module is a gist containing prompt files and metadata.`,
}

var (
	modForce bool
)

func init() {
	modCmd.AddCommand(modInitCmd)
	modCmd.AddCommand(modListCmd)
	modCmd.AddCommand(modGetCmd) 
	modCmd.AddCommand(modDownloadCmd)
	modCmd.AddCommand(modTidyCmd)
	modCmd.AddCommand(modVendorCmd)
	modCmd.AddCommand(modSearchCmd)
	modCmd.AddCommand(modPublishCmd)
}

var modInitCmd = &cobra.Command{
	Use:   "init [module-name]",
	Short: "Initialize a new prompt module",
	Long: `Initialize a new prompt module with go.mod style dependency management.

This creates a go.mod file for managing prompt dependencies.

Example:
  pe mod init github.com/myorg/myproject
  pe mod init example.com/prompts`,
	Args: cobra.ExactArgs(1),
	RunE: runModInit,
}

var modListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available modules from the registry",
	RunE:  runModList,
}

var modGetCmd = &cobra.Command{
	Use:   "get [module]",
	Short: "Get module information",
	Args:  cobra.ExactArgs(1),
	RunE:  runModGet,
}

var modDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download modules specified in pe.mod",
	RunE:  runModDownload,
}

var modSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for modules in the registry",
	Args:  cobra.ExactArgs(1),
	RunE:  runModSearch,
}

var modPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a module to the registry",
	RunE:  runModPublish,
}

var modTidyCmd = &cobra.Command{
	Use:   "tidy",
	Short: "Add missing and remove unused modules",
	RunE:  runModTidy,
}

var modVendorCmd = &cobra.Command{
	Use:   "vendor",
	Short: "Copy dependencies to vendor directory",
	RunE:  runModVendor,
}

func init() {
	modInitCmd.Flags().BoolVar(&modForce, "force", false, "Overwrite existing module")
}

// Module represents a prompt module
type Module struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	GistID      string    `json:"gist_id,omitempty"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

// PromptModule represents the full module with content
type PromptModule struct {
	Module
	PromptFile string            `json:"prompt_file"`
	Files      map[string]string `json:"files,omitempty"`
}

func runModList(cmd *cobra.Command, args []string) error {
	registry := module.DefaultRegistry()
	modules, err := registry.List()
	if err != nil {
		return fmt.Errorf("failed to list modules: %w", err)
	}
	
	if len(modules) == 0 {
		fmt.Println("No modules found in registry")
		return nil
	}
	
	fmt.Printf("Available modules:\n\n")
	for _, mod := range modules {
		fmt.Printf("  %s@%s - %s\n", mod.Name, mod.Version, mod.Description)
		if mod.Author != "" {
			fmt.Printf("    Author: %s\n", mod.Author)
		}
		if len(mod.Tags) > 0 {
			fmt.Printf("    Tags: %s\n", strings.Join(mod.Tags, ", "))
		}
	}
	
	return nil
}

func runModGet(cmd *cobra.Command, args []string) error {
	moduleName := args[0]
	
	registry := module.DefaultRegistry()
	mod, err := registry.Get(moduleName)
	if err != nil {
		return fmt.Errorf("failed to get module %s: %w", moduleName, err)
	}
	
	fmt.Printf("Module: %s@%s\n", mod.Name, mod.Version)
	fmt.Printf("Description: %s\n", mod.Description)
	if mod.Author != "" {
		fmt.Printf("Author: %s\n", mod.Author)
	}
	if mod.License != "" {
		fmt.Printf("License: %s\n", mod.License)
	}
	if len(mod.Dependencies) > 0 {
		fmt.Printf("Dependencies:\n")
		for dep, ver := range mod.Dependencies {
			fmt.Printf("  %s: %s\n", dep, ver)
		}
	}
	if len(mod.Files) > 0 {
		fmt.Printf("Files:\n")
		for _, file := range mod.Files {
			fmt.Printf("  - %s\n", file)
		}
	}
	
	return nil
}

func runModSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	
	registry := module.DefaultRegistry()
	modules, err := registry.Search(query)
	if err != nil {
		return fmt.Errorf("failed to search modules: %w", err)
	}
	
	if len(modules) == 0 {
		fmt.Printf("No modules found matching '%s'\n", query)
		return nil
	}
	
	fmt.Printf("Modules matching '%s':\n\n", query)
	for _, mod := range modules {
		fmt.Printf("  %s@%s - %s\n", mod.Name, mod.Version, mod.Description)
	}
	
	return nil
}

func runModPublish(cmd *cobra.Command, args []string) error {
	// Read module.json from current directory
	data, err := os.ReadFile("module.json")
	if err != nil {
		return fmt.Errorf("failed to read module.json: %w", err)
	}
	
	var mod module.Module
	if err := json.Unmarshal(data, &mod); err != nil {
		return fmt.Errorf("failed to parse module.json: %w", err)
	}
	
	// Validate module
	if mod.Name == "" {
		return fmt.Errorf("module name is required")
	}
	if mod.Version == "" {
		return fmt.Errorf("module version is required")
	}
	
	// Get list of files to publish
	if len(mod.Files) == 0 {
		// Default to all .prompt files
		files, err := filepath.Glob("*.prompt")
		if err == nil && len(files) > 0 {
			mod.Files = files
		}
	}
	
	// Set metadata
	mod.PublishedAt = time.Now()
	mod.UpdatedAt = time.Now()
	
	// Publish to registry
	registry := module.DefaultRegistry()
	if err := registry.Publish(&mod, "."); err != nil {
		return fmt.Errorf("failed to publish module: %w", err)
	}
	
	fmt.Printf("Successfully published %s@%s\n", mod.Name, mod.Version)
	return nil
}

func runModInit(cmd *cobra.Command, args []string) error {
	moduleName := args[0]

	// Check if pe.mod already exists
	if _, err := os.Stat("pe.mod"); err == nil && !modForce {
		return fmt.Errorf("pe.mod already exists (use --force to overwrite)")
	}

	// Create new pe.mod file
	file := &pemod.File{}

	// Set module path if provided
	if moduleName != "" {
		file.SetModule(moduleName)
	}

	// Set PE version
	file.SetPEVersion("1")

	// Format and write pe.mod file
	content := file.Format()
	if err := os.WriteFile("pe.mod", []byte(content), 0644); err != nil {
		return fmt.Errorf("writing pe.mod: %w", err)
	}

	// Create .pe directory for caching and metadata
	if err := os.MkdirAll(".pe", 0755); err != nil {
		return fmt.Errorf("creating .pe directory: %w", err)
	}

	// Create cache directories
	cacheDir := filepath.Join(".pe", "cache")
	if err := os.MkdirAll(filepath.Join(cacheDir, "modules"), 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	fmt.Printf("Created pe.mod for module %s\n", moduleName)

	return nil
}

// Registry functions disabled pending implementation
/*
func runModList(cmd *cobra.Command, args []string) error {
	// TODO: Implement fetching from root gist
	fmt.Println("Available modules:")
	fmt.Println("  (none - registry not configured)")
	fmt.Printf("\nRoot gist: %s\n", rootGistID)
	return nil
}
*/

func runModGet(cmd *cobra.Command, args []string) error {
	moduleName := args[0]

	// Check local first
	localPath := filepath.Join(".pe", "modules", strings.ReplaceAll(moduleName, "/", string(os.PathSeparator)))
	if info, err := os.Stat(localPath); err == nil && info.IsDir() {
		metaPath := filepath.Join(localPath, "module.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			return fmt.Errorf("reading module metadata: %w", err)
		}

		var module PromptModule
		if err := json.Unmarshal(data, &module); err != nil {
			return fmt.Errorf("parsing module metadata: %w", err)
		}

		fmt.Printf("Module: %s\n", module.Name)
		fmt.Printf("Version: %s\n", module.Version)
		fmt.Printf("Author: %s\n", module.Author)
		fmt.Printf("Description: %s\n", module.Description)
		fmt.Printf("Location: %s (local)\n", localPath)

		return nil
	}

	// TODO: Check registry
	return fmt.Errorf("module %s not found", moduleName)
}

func getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// Registry functions for gist-based module registry

type GistFile struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Language string `json:"language"`
	RawURL   string `json:"raw_url"`
	Size     int    `json:"size"`
	Content  string `json:"content"`
}

type Gist struct {
	ID          string              `json:"id"`
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	Files       map[string]GistFile `json:"files"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func fetchGist(gistID string) (*Gist, error) {
	url := fmt.Sprintf("https://api.github.com/gists/%s", gistID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add GitHub token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := http.DefaultClient.Do(req)
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

func runModDownload(cmd *cobra.Command, args []string) error {
	// Read pe.mod file
	content, err := os.ReadFile("pe.mod")
	if err != nil {
		return fmt.Errorf("reading pe.mod: %w (run 'pe mod init' first)", err)
	}

	// Parse pe.mod file
	file, err := pemod.Parse(strings.NewReader(string(content)))
	if err != nil {
		return fmt.Errorf("parsing pe.mod: %w", err)
	}

	// Create modules directory
	modulesDir := filepath.Join(".pe", "cache", "modules")
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return fmt.Errorf("creating modules directory: %w", err)
	}

	fmt.Println("Downloading modules...")

	// Download each required module
	for _, req := range file.Require {
		fmt.Printf("Downloading %s@%s...\n", req.Mod, req.Version)

		// Create module directory
		modDir := filepath.Join(modulesDir, string(req.Mod)+"@"+req.Version)
		if err := os.MkdirAll(modDir, 0755); err != nil {
			return fmt.Errorf("creating module directory: %w", err)
		}

		// TODO: Implement actual download from registry
		// For now, create a placeholder
		placeholder := fmt.Sprintf("# Module %s@%s\n# Downloaded on %s\n",
			req.Mod, req.Version, time.Now().Format(time.RFC3339))
		if err := os.WriteFile(filepath.Join(modDir, "module.info"), []byte(placeholder), 0644); err != nil {
			return fmt.Errorf("writing module info: %w", err)
		}
	}

	fmt.Printf("Downloaded %d modules\n", len(file.Require))
	return nil
}

func runModTidy(cmd *cobra.Command, args []string) error {
	// Check if pe.mod exists
	if _, err := os.Stat("pe.mod"); err != nil {
		return fmt.Errorf("pe.mod not found: run 'pe mod init' first")
	}

	// Read and parse pe.mod file
	content, err := os.ReadFile("pe.mod")
	if err != nil {
		return fmt.Errorf("reading pe.mod: %w", err)
	}

	file, err := pemod.Parse(strings.NewReader(string(content)))
	if err != nil {
		return fmt.Errorf("parsing pe.mod: %w", err)
	}

	fmt.Println("Analyzing prompt dependencies...")

	// TODO: Scan prompt files for pe://module/prompt references
	// For now, validate existing dependencies
	var updated bool

	// Remove unused dependencies (placeholder logic)
	var filteredRequires []pemod.Require
	for _, req := range file.Require {
		// TODO: Check if requirement is actually used
		filteredRequires = append(filteredRequires, req)
	}

	if len(filteredRequires) != len(file.Require) {
		file.Require = filteredRequires
		updated = true
	}

	// Write updated pe.mod if changes were made
	if updated {
		content := file.Format()
		if err := os.WriteFile("pe.mod", []byte(content), 0644); err != nil {
			return fmt.Errorf("writing pe.mod: %w", err)
		}
		fmt.Println("pe.mod updated")
	} else {
		fmt.Println("pe.mod is already tidy")
	}

	return nil
}

func runModVendor(cmd *cobra.Command, args []string) error {
	// Check if pe.mod exists
	if _, err := os.Stat("pe.mod"); err != nil {
		return fmt.Errorf("pe.mod not found: run 'pe mod init' first")
	}

	// Read and parse pe.mod file
	content, err := os.ReadFile("pe.mod")
	if err != nil {
		return fmt.Errorf("reading pe.mod: %w", err)
	}

	file, err := pemod.Parse(strings.NewReader(string(content)))
	if err != nil {
		return fmt.Errorf("parsing pe.mod: %w", err)
	}

	// Create vendor directory
	vendorDir := "vendor"
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		return fmt.Errorf("creating vendor directory: %w", err)
	}

	fmt.Println("Copying dependencies to vendor/...")

	// Create modules.txt file listing vendored modules
	var modulesList []string

	// Copy each required module from cache to vendor
	cacheDir := filepath.Join(".pe", "cache", "modules")
	for _, req := range file.Require {
		srcDir := filepath.Join(cacheDir, string(req.Mod)+"@"+req.Version)
		destDir := filepath.Join(vendorDir, string(req.Mod)+"@"+req.Version)

		// Check if module exists in cache
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			fmt.Printf("Warning: Module %s@%s not found in cache, run 'pe mod download' first\n", req.Mod, req.Version)
			continue
		}

		// Create destination directory
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("creating vendor module directory: %w", err)
		}

		// Copy module files (placeholder - would copy actual files)
		moduleInfo := fmt.Sprintf("# %s@%s\n# Vendored on %s\n", req.Mod, req.Version, time.Now().Format(time.RFC3339))
		if err := os.WriteFile(filepath.Join(destDir, "module.info"), []byte(moduleInfo), 0644); err != nil {
			return fmt.Errorf("writing vendored module info: %w", err)
		}

		modulesList = append(modulesList, fmt.Sprintf("%s@%s", req.Mod, req.Version))
		fmt.Printf("Vendored %s@%s\n", req.Mod, req.Version)
	}

	// Write modules.txt
	modulesContent := strings.Join(modulesList, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(vendorDir, "modules.txt"), []byte(modulesContent), 0644); err != nil {
		return fmt.Errorf("writing modules.txt: %w", err)
	}

	fmt.Printf("Vendored %d modules\n", len(modulesList))
	return nil
}
