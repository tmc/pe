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
)

var modCmd = &cobra.Command{
	Use:   "mod",
	Short: "Module management for prompts",
	Long: `Manage prompt modules using GitHub gists as a registry.

PE uses a root gist that tracks forks containing prompt modules.
Each module is a gist containing prompt files and metadata.`,
}

var (
	// Root gist ID that contains the registry
	rootGistID = "YOUR_ROOT_GIST_ID" // TODO: Set this to actual root gist
	modForce   bool
)

func init() {
	modCmd.AddCommand(modInitCmd)
	modCmd.AddCommand(modListCmd)
	modCmd.AddCommand(modGetCmd)
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

func runModInit(cmd *cobra.Command, args []string) error {
	moduleName := args[0]
	
	// Check if go.mod already exists
	if _, err := os.Stat("go.mod"); err == nil && !modForce {
		return fmt.Errorf("go.mod already exists")
	}
	
	// Create go.mod content similar to go mod init
	goModContent := fmt.Sprintf(`module %s

go 1.21

require (
	// Prompt dependencies will be added here
)
`, moduleName)
	
	// Write go.mod file
	if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("writing go.mod: %w", err)
	}
	
	// Also create a .pe directory for prompt-specific metadata
	if err := os.MkdirAll(".pe", 0755); err != nil {
		return fmt.Errorf("creating .pe directory: %w", err)
	}
	
	// Create pe.json for prompt-specific configuration
	peConfig := map[string]interface{}{
		"module":  moduleName,
		"version": "0.1.0",
		"created": time.Now().Format(time.RFC3339),
	}
	
	data, err := json.MarshalIndent(peConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling pe config: %w", err)
	}
	
	if err := os.WriteFile(".pe/config.json", data, 0644); err != nil {
		return fmt.Errorf("writing pe config: %w", err)
	}
	
	fmt.Printf("Created go.mod\n")
	
	return nil
}

func runModList(cmd *cobra.Command, args []string) error {
	// TODO: Implement fetching from root gist
	fmt.Println("Available modules:")
	fmt.Println("  (none - registry not configured)")
	fmt.Printf("\nRoot gist: %s\n", rootGistID)
	return nil
}

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