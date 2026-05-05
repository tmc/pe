package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/exectext"
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
	modForce     bool
	modTidyWrite bool
	modTidyJSON  bool
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
	modCmd.AddCommand(modVetCmd)
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

var modVetCmd = &cobra.Command{
	Use:   "vet [file...]",
	Short: "Validate pe.mod capability policy",
	Long: `Vet parses pe.mod and checks the static capability contract.

With file arguments, vet also checks executable text front matter against
module policy requirements such as typed inputs and reviewed imports.`,
	RunE: runModVet,
}

func init() {
	modInitCmd.Flags().BoolVar(&modForce, "force", false, "Overwrite existing module")
	modTidyCmd.Flags().BoolVarP(&modTidyWrite, "write", "w", false, "Update pe.mod; add only refs with one explicit version")
	modTidyCmd.Flags().BoolVar(&modTidyJSON, "json", false, "Write dependency audit report as JSON")
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

func runModGetLocal(cmd *cobra.Command, args []string) error {
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

	resp, err := githubHTTPClient.Do(req)
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

		// Download from registry
		registry := module.DefaultRegistry()
		mod, err := registry.Get(string(req.Mod))
		if err != nil {
			// Try without version if not found
			fmt.Printf("Warning: %s - using placeholder\n", err)
			placeholder := fmt.Sprintf("# Module %s@%s\n# Downloaded on %s\n",
				req.Mod, req.Version, time.Now().Format(time.RFC3339))
			if err := os.WriteFile(filepath.Join(modDir, "module.info"), []byte(placeholder), 0644); err != nil {
				return fmt.Errorf("writing module info: %w", err)
			}
			continue
		}

		// Download the module files
		if err := registry.Download(mod, filepath.Join(".pe", "cache")); err != nil {
			return fmt.Errorf("downloading module %s: %w", req.Mod, err)
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

	out := cmd.OutOrStdout()
	if !modTidyJSON {
		fmt.Fprintln(out, "Analyzing prompt dependencies...")
	}

	refs, err := scanPEModuleReferences(".")
	if err != nil {
		return err
	}

	required := make(map[string]bool)
	for _, req := range file.Require {
		required[string(req.Mod)] = true
	}

	var missing []string
	for mod := range refs {
		if !required[mod] {
			missing = append(missing, mod)
		}
	}
	sort.Strings(missing)

	var unused []string
	for _, req := range file.Require {
		mod := string(req.Mod)
		if _, ok := refs[mod]; !ok {
			unused = append(unused, mod)
		}
	}
	sort.Strings(unused)

	report := modTidyReport{
		References: countPEModuleReferences(refs),
		Files:      countPEModuleReferenceFiles(refs),
	}
	for _, mod := range missing {
		report.Missing = append(report.Missing, modTidyDependency{
			Module:   mod,
			Files:    peReferenceFiles(refs[mod]),
			Versions: peReferenceVersions(refs[mod]),
		})
	}
	report.Unused = append(report.Unused, unused...)

	if !modTidyJSON {
		fmt.Fprintf(out, "found %d module references in %d files\n", report.References, report.Files)
	}
	for _, mod := range missing {
		if !modTidyJSON {
			fmt.Fprintf(out, "missing dependency: %s (referenced by %s)\n", mod, strings.Join(peReferenceFiles(refs[mod]), ", "))
		}
	}
	for _, mod := range unused {
		if !modTidyJSON {
			fmt.Fprintf(out, "unused dependency: %s\n", mod)
		}
	}
	if len(missing) == 0 && len(unused) == 0 {
		if !modTidyJSON {
			fmt.Fprintln(out, "pe.mod is tidy")
		}
		return writeModTidyReport(out, report)
	}
	if !modTidyWrite {
		return writeModTidyReport(out, report)
	}

	for _, mod := range unused {
		file.RemoveRequire(mod)
		report.Removed = append(report.Removed, mod)
		if !modTidyJSON {
			fmt.Fprintf(out, "removed dependency: %s\n", mod)
		}
	}
	for _, mod := range missing {
		version, ok := singlePEReferenceVersion(refs[mod])
		if !ok {
			report.Skipped = append(report.Skipped, modTidySkipped{
				Module: mod,
				Reason: "no single explicit version in references",
			})
			if !modTidyJSON {
				fmt.Fprintf(out, "skipped dependency: %s (no single explicit version in references)\n", mod)
			}
			continue
		}
		file.AddRequire(mod, version)
		report.Added = append(report.Added, modTidyRequirement{Module: mod, Version: version})
		if !modTidyJSON {
			fmt.Fprintf(out, "added dependency: %s %s\n", mod, version)
		}
	}
	formatted := file.Format()
	if err := os.WriteFile("pe.mod", []byte(formatted), 0644); err != nil {
		return fmt.Errorf("writing pe.mod: %w", err)
	}
	report.Updated = true
	if !modTidyJSON {
		fmt.Fprintln(out, "pe.mod updated")
	}

	return writeModTidyReport(out, report)
}

var peRefRE = regexp.MustCompile(`pe://[^\s"'<>]+`)

type modTidyReport struct {
	References int                  `json:"references"`
	Files      int                  `json:"files"`
	Missing    []modTidyDependency  `json:"missing,omitempty"`
	Unused     []string             `json:"unused,omitempty"`
	Added      []modTidyRequirement `json:"added,omitempty"`
	Removed    []string             `json:"removed,omitempty"`
	Skipped    []modTidySkipped     `json:"skipped,omitempty"`
	Updated    bool                 `json:"updated"`
}

type modTidyDependency struct {
	Module   string   `json:"module"`
	Files    []string `json:"files"`
	Versions []string `json:"versions,omitempty"`
}

type modTidyRequirement struct {
	Module  string `json:"module"`
	Version string `json:"version"`
}

type modTidySkipped struct {
	Module string `json:"module"`
	Reason string `json:"reason"`
}

type peModuleReference struct {
	File    string
	Version string
}

func writeModTidyReport(w io.Writer, report modTidyReport) error {
	if !modTidyJSON {
		return nil
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func scanPEModuleReferences(root string) (map[string][]peModuleReference, error) {
	refs := make(map[string][]peModuleReference)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			switch name {
			case ".git", ".pe", ".beads", "vendor", "node_modules":
				return filepath.SkipDir
			}
			if name != "." && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() || !isPEReferenceFile(path) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		for _, raw := range peRefRE.FindAllString(string(data), -1) {
			mod, version, ok := moduleAndVersionFromPERef(raw)
			if !ok {
				continue
			}
			file := filepath.ToSlash(path)
			if file == "." {
				file = name
			}
			if !containsPEReference(refs[mod], file, version) {
				refs[mod] = append(refs[mod], peModuleReference{File: file, Version: version})
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning module references: %w", err)
	}
	for mod := range refs {
		sort.Slice(refs[mod], func(i, j int) bool {
			if refs[mod][i].File != refs[mod][j].File {
				return refs[mod][i].File < refs[mod][j].File
			}
			return refs[mod][i].Version < refs[mod][j].Version
		})
	}
	return refs, nil
}

func isPEReferenceFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".prompt", ".txt", ".md", ".yaml", ".yml", ".json", ".toml", ".star", ".tmpl", ".tpl":
		return true
	default:
		return false
	}
}

func moduleFromPERef(ref string) (string, bool) {
	mod, _, ok := moduleAndVersionFromPERef(ref)
	return mod, ok
}

func moduleAndVersionFromPERef(ref string) (string, string, bool) {
	ref = strings.TrimRight(ref, ".,;:)]}")
	u, err := url.Parse(ref)
	if err != nil || u.Scheme != "pe" || u.Host == "" {
		return "", "", false
	}
	parts := strings.FieldsFunc(strings.Trim(u.Path, "/"), func(r rune) bool { return r == '/' })
	var mod string
	if strings.Contains(u.Host, ".") {
		if len(parts) >= 2 {
			mod = u.Host + "/" + parts[0] + "/" + parts[1]
		} else if len(parts) == 1 {
			mod = u.Host + "/" + parts[0]
		} else {
			mod = u.Host
		}
	} else {
		mod = u.Host
	}
	mod, version := splitPEModuleVersion(mod)
	return mod, version, true
}

func splitPEModuleVersion(mod string) (string, string) {
	i := strings.LastIndex(mod, "@")
	if i < 0 {
		return mod, ""
	}
	version := mod[i+1:]
	if version == "" || !strings.HasPrefix(version, "v") {
		return mod, ""
	}
	return mod[:i], version
}

func singlePEReferenceVersion(refs []peModuleReference) (string, bool) {
	var version string
	for _, ref := range refs {
		if ref.Version == "" {
			return "", false
		}
		if version == "" {
			version = ref.Version
			continue
		}
		if version != ref.Version {
			return "", false
		}
	}
	return version, version != ""
}

func countPEModuleReferences(refs map[string][]peModuleReference) int {
	var n int
	for _, files := range refs {
		n += len(files)
	}
	return n
}

func countPEModuleReferenceFiles(refs map[string][]peModuleReference) int {
	files := make(map[string]bool)
	for _, refs := range refs {
		for _, ref := range refs {
			files[ref.File] = true
		}
	}
	return len(files)
}

func peReferenceFiles(refs []peModuleReference) []string {
	seen := make(map[string]bool)
	var files []string
	for _, ref := range refs {
		if !seen[ref.File] {
			seen[ref.File] = true
			files = append(files, ref.File)
		}
	}
	return files
}

func peReferenceVersions(refs []peModuleReference) []string {
	seen := make(map[string]bool)
	var versions []string
	for _, ref := range refs {
		if ref.Version == "" || seen[ref.Version] {
			continue
		}
		seen[ref.Version] = true
		versions = append(versions, ref.Version)
	}
	sort.Strings(versions)
	return versions
}

func containsPEReference(list []peModuleReference, file, version string) bool {
	for _, item := range list {
		if item.File == file && item.Version == version {
			return true
		}
	}
	return false
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

func runModVet(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile("pe.mod")
	if err != nil {
		return fmt.Errorf("reading pe.mod: %w", err)
	}
	modFile, err := pemod.Parse(strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("parsing pe.mod: %w", err)
	}
	if err := vetCapabilityPolicy(modFile); err != nil {
		return err
	}
	for _, name := range args {
		if err := vetExecutableTextFile(modFile, name); err != nil {
			return err
		}
	}
	fmt.Fprintln(cmd.OutOrStdout(), "ok")
	return nil
}

func vetCapabilityPolicy(file *pemod.File) error {
	if file.Capability == nil && file.Placement == nil && file.Policy == nil {
		return nil
	}
	if file.Policy != nil && file.Policy.Composition != "" && file.Policy.Composition != "strict" {
		return fmt.Errorf("unsupported policy composition %s", file.Policy.Composition)
	}
	return nil
}

func vetExecutableTextFile(modFile *pemod.File, name string) error {
	f, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("opening executable text %s: %w", name, err)
	}
	defer f.Close()
	text, err := exectext.Parse(f)
	if err != nil {
		return fmt.Errorf("parsing executable text %s: %w", name, err)
	}
	if err := text.Validate(); err != nil {
		return fmt.Errorf("validating executable text %s: %w", name, err)
	}
	if modFile.Policy != nil && modFile.Policy.RequireTypedIO && text.Meta.Kind != "" && len(text.Meta.Inputs) == 0 {
		return fmt.Errorf("%s: policy requires typed inputs", name)
	}
	for dim, vals := range text.Meta.Safety {
		if err := deniedByModule(modFile, dim, vals.Allow); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

func deniedByModule(modFile *pemod.File, dim string, allowed []string) error {
	if modFile.Capability == nil || len(allowed) == 0 {
		return nil
	}
	var deny []string
	name := dim
	if strings.HasSuffix(name, "s") {
		name = strings.TrimSuffix(name, "s")
	}
	switch dim {
	case "data":
		deny = modFile.Capability.Data.Deny
	case "prompts":
		deny = modFile.Capability.Prompts.Deny
	case "providers":
		deny = modFile.Capability.Providers.Deny
	case "tools":
		deny = modFile.Capability.Tools.Deny
	default:
		return nil
	}
	denied := make(map[string]bool)
	for _, v := range deny {
		denied[v] = true
	}
	for _, v := range allowed {
		if denied[v] {
			return fmt.Errorf("%s %s is denied by pe.mod", name, v)
		}
	}
	return nil
}
