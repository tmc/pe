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
	Use:   "init [module-name] --prompt='prompt text'",
	Short: "Initialize a new prompt module",
	Long: `Initialize a new prompt module that can be published to the registry.

This creates a local module structure that can be pushed to a GitHub gist.

Example:
  pe mod init tmc/hello --prompt='say hello in a random language'
  pe mod init myorg/summarize --prompt='summarize this: {{.text}}'`,
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

var modInitPrompt string

func init() {
	modInitCmd.Flags().StringVar(&modInitPrompt, "prompt", "", "The prompt text for the module")
	modInitCmd.MarkFlagRequired("prompt")
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
	
	// Validate module name format (org/name)
	parts := strings.Split(moduleName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("module name must be in format: org/name")
	}
	
	org, name := parts[0], parts[1]
	
	// Create module directory
	moduleDir := filepath.Join(".pe", "modules", org, name)
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("creating module directory: %w", err)
	}
	
	// Check if already exists
	metaPath := filepath.Join(moduleDir, "module.json")
	if _, err := os.Stat(metaPath); err == nil && !modForce {
		return fmt.Errorf("module already exists at %s (use --force to overwrite)", moduleDir)
	}
	
	// Create module metadata
	module := PromptModule{
		Module: Module{
			Name:        moduleName,
			Version:     "0.1.0",
			Description: fmt.Sprintf("Prompt module: %s", name),
			Author:      getCurrentUser(),
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		PromptFile: "prompt.txt",
		Files:      make(map[string]string),
	}
	
	// Create prompt file
	promptPath := filepath.Join(moduleDir, "prompt.txt")
	promptContent := modInitPrompt
	
	// If the prompt has variables, add defaults section
	if strings.Contains(promptContent, "{{.") {
		promptContent += "\n\n-- defaults --\n# Add your defaults here\n"
	}
	
	// Add a simple eval section
	promptContent += "\n-- evals --\n# Add your tests here\n$ \nExpected output\n"
	
	if err := os.WriteFile(promptPath, []byte(promptContent), 0644); err != nil {
		return fmt.Errorf("writing prompt file: %w", err)
	}
	
	// Save module metadata
	data, err := json.MarshalIndent(module, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling module: %w", err)
	}
	
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("writing module metadata: %w", err)
	}
	
	fmt.Printf("Initialized module %s at %s\n", moduleName, moduleDir)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit %s\n", promptPath)
	fmt.Printf("  2. Test with: pe vet %s\n", promptPath)
	fmt.Printf("  3. Publish with: pe push %s\n", moduleName)
	
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