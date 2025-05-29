package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [module]",
	Short: "Push a module to the registry",
	Long: `Push a prompt module to the GitHub gist registry.

This creates or updates a gist with your module content and
registers it in the root registry gist.

Requires GITHUB_TOKEN environment variable to be set.

Example:
  pe push tmc/hello
  pe push myorg/summarize --public`,
	Args: cobra.ExactArgs(1),
	RunE: runPush,
}

var (
	pushPublic bool
	pushUpdate bool
)

func init() {
	pushCmd.Flags().BoolVar(&pushPublic, "public", false, "Make the gist public")
	pushCmd.Flags().BoolVar(&pushUpdate, "update", false, "Update existing gist")
}

func runPush(cmd *cobra.Command, args []string) error {
	moduleName := args[0]
	
	// Check GitHub token
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}
	
	// Load module from local
	moduleDir := filepath.Join(".pe", "modules", strings.ReplaceAll(moduleName, "/", string(os.PathSeparator)))
	metaPath := filepath.Join(moduleDir, "module.json")
	
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return fmt.Errorf("reading module metadata: %w", err)
	}
	
	var module PromptModule
	if err := json.Unmarshal(data, &module); err != nil {
		return fmt.Errorf("parsing module metadata: %w", err)
	}
	
	// Read all files in the module
	files := make(map[string]GistFile)
	
	// Add module.json
	files["module.json"] = GistFile{
		Content: string(data),
	}
	
	// Add prompt file
	promptPath := filepath.Join(moduleDir, module.PromptFile)
	promptData, err := os.ReadFile(promptPath)
	if err != nil {
		return fmt.Errorf("reading prompt file: %w", err)
	}
	files[module.PromptFile] = GistFile{
		Content: string(promptData),
	}
	
	// Add any other files
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		return fmt.Errorf("reading module directory: %w", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "module.json" || entry.Name() == module.PromptFile {
			continue
		}
		
		content, err := os.ReadFile(filepath.Join(moduleDir, entry.Name()))
		if err != nil {
			continue
		}
		files[entry.Name()] = GistFile{
			Content: string(content),
		}
	}
	
	// Create or update gist
	var gistID string
	if module.GistID != "" && pushUpdate {
		// Update existing gist
		gistID = module.GistID
		if err := updateGist(token, gistID, module.Description, files); err != nil {
			return fmt.Errorf("updating gist: %w", err)
		}
		fmt.Printf("Updated gist: https://gist.github.com/%s\n", gistID)
	} else {
		// Create new gist
		gist, err := createGist(token, module.Description, pushPublic, files)
		if err != nil {
			return fmt.Errorf("creating gist: %w", err)
		}
		gistID = gist.ID
		
		// Update module with gist ID
		module.GistID = gistID
		updatedData, err := json.MarshalIndent(module, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling updated module: %w", err)
		}
		
		if err := os.WriteFile(metaPath, updatedData, 0644); err != nil {
			return fmt.Errorf("updating module metadata: %w", err)
		}
		
		fmt.Printf("Created gist: https://gist.github.com/%s\n", gistID)
	}
	
	// Register in root gist
	if err := registerModule(token, moduleName, module.Version, gistID); err != nil {
		// Just warn, don't fail
		fmt.Printf("Warning: failed to register in root gist: %v\n", err)
	}
	
	fmt.Printf("\nModule %s pushed successfully!\n", moduleName)
	fmt.Printf("Others can now use: pe run %s@latest\n", moduleName)
	
	return nil
}

func createGist(token, description string, public bool, files map[string]GistFile) (*Gist, error) {
	payload := map[string]interface{}{
		"description": description,
		"public":      public,
		"files":       files,
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequest("POST", "https://api.github.com/gists", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := http.DefaultClient.Do(req)
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

func updateGist(token, gistID, description string, files map[string]GistFile) error {
	payload := map[string]interface{}{
		"description": description,
		"files":       files,
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("https://api.github.com/gists/%s", gistID)
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}
	
	return nil
}

func registerModule(token, moduleName, version, gistID string) error {
	// Get root registry gist ID
	rootGistID := os.Getenv("PE_REGISTRY_GIST_ID")
	if rootGistID == "" {
		// Skip registration if no root gist is configured
		return nil
	}
	
	// Fetch current registry
	rootGist, err := getGist(token, rootGistID)
	if err != nil {
		return fmt.Errorf("fetching root registry: %w", err)
	}
	
	// Get or create index.json
	var index map[string]map[string]string
	if indexFile, ok := rootGist.Files["index.json"]; ok {
		if err := json.Unmarshal([]byte(indexFile.Content), &index); err != nil {
			return fmt.Errorf("parsing registry index: %w", err)
		}
	} else {
		index = make(map[string]map[string]string)
	}
	
	// Update index
	if index[moduleName] == nil {
		index[moduleName] = make(map[string]string)
	}
	index[moduleName][version] = gistID
	index[moduleName]["latest"] = gistID
	
	// Marshal updated index
	indexData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling index: %w", err)
	}
	
	// Update the root gist
	files := map[string]GistFile{
		"index.json": {
			Content: string(indexData),
		},
	}
	
	return updateGist(token, rootGistID, "PE Module Registry", files)
}

func getGist(token, gistID string) (*Gist, error) {
	url := fmt.Sprintf("https://api.github.com/gists/%s", gistID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
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