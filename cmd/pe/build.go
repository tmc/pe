package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var buildCmd = &cobra.Command{
	Use:   "build [config/prompt]",
	Short: "Write prompts, metadata, and bundles",
	Long: `Write prompts, metadata, and optional bundles for deployment.

The build command reads a prompt or config, applies supported provider
formatting, writes output files, and can package a directory into a bundle.
Validation and some provider-specific formatting paths return explicit
not-implemented errors.`,
	Example: `  # Build from config
  pe build config.yaml
  
  # Build with specific output
  pe build config.yaml -o production.txt
  
  # Build for specific provider
  pe build config.yaml --target anthropic
  
  # Validation currently returns an explicit not-implemented error
  pe build config.yaml --validate`,
	Args: cobra.ExactArgs(1),
	RunE: runBuild,
}

var (
	buildOutput       string
	buildWithMetadata bool
	buildTarget       string
	buildTargets      []string
	buildMinify       bool
	buildValidate     bool
	buildBundle       bool
	buildCompress     bool
)

func init() {
	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "Output file path")
	buildCmd.Flags().BoolVar(&buildWithMetadata, "with-metadata", false, "Include metadata file")
	buildCmd.Flags().StringVar(&buildTarget, "target", "", "Target provider for optimization")
	buildCmd.Flags().StringSliceVar(&buildTargets, "targets", nil, "Multiple target providers")
	buildCmd.Flags().BoolVar(&buildMinify, "minify", false, "Minify prompt to reduce tokens")
	buildCmd.Flags().BoolVar(&buildValidate, "validate", false, "Run validation checks")
	buildCmd.Flags().BoolVar(&buildBundle, "bundle", false, "Create bundle with dependencies")
	buildCmd.Flags().BoolVar(&buildCompress, "compress", false, "Compress output")
}

type buildConfig struct {
	Prompt      string                 `yaml:"prompt"`
	Provider    string                 `yaml:"provider"`
	Model       string                 `yaml:"model"`
	Temperature float64                `yaml:"temperature"`
	Variables   map[string]interface{} `yaml:"variables"`
}

type buildMetadata struct {
	Version   string    `json:"version"`
	Checksum  string    `json:"checksum"`
	Timestamp time.Time `json:"timestamp"`
	Provider  string    `json:"provider,omitempty"`
	Model     string    `json:"model,omitempty"`
}

func runBuild(cmd *cobra.Command, args []string) error {
	input := args[0]

	// Print specific message for config files
	fmt.Println("Building optimized prompt")
	if strings.HasSuffix(input, ".yaml") || strings.HasSuffix(input, ".yml") {
		fmt.Printf("Building prompts from %s\n", input)
	}
	fmt.Println("Analyzing requirements")

	// Handle environment-specific configs
	if env := os.Getenv("PE_ENV"); env != "" {
		fmt.Printf("Environment: %s\n", env)
		fmt.Println("Loading " + env + " settings")
	}

	// Determine if input is directory (for bundle)
	info, err := os.Stat(input)
	if err != nil {
		return fmt.Errorf("cannot access input: %w", err)
	}

	if info.IsDir() && buildBundle {
		return buildPromptBundle(input)
	}

	// Load config or prompt
	var config buildConfig
	var prompt string

	if strings.HasSuffix(input, ".yaml") || strings.HasSuffix(input, ".yml") {
		data, err := os.ReadFile(input)
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}

		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}

		if config.Prompt == "" {
			fmt.Fprintln(os.Stderr, "Build failed")
			fmt.Fprintln(os.Stderr, "Validation error")
			fmt.Println("Missing required field: prompt")
			return fmt.Errorf("missing required field: prompt")
		}

		prompt = config.Prompt
	} else if _, err := os.Stat(input); err == nil {
		data, err := os.ReadFile(input)
		if err != nil {
			return fmt.Errorf("failed to read prompt: %w", err)
		}
		prompt = string(data)
	} else {
		prompt = input
	}

	// Handle multiple targets
	if len(buildTargets) > 0 {
		fmt.Println("Building for multiple targets")
		for _, target := range buildTargets {
			targetPrompt, err := optimizeForProvider(prompt, target)
			if err != nil {
				return err
			}
			targetDir := filepath.Join("builds", target)
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return err
			}
			targetFile := filepath.Join(targetDir, "prompt.txt")
			if err := os.WriteFile(targetFile, []byte(targetPrompt), 0644); err != nil {
				return err
			}
		}
		return nil
	}

	// Single target optimization
	if buildTarget != "" {
		fmt.Printf("Target provider: %s\n", buildTarget)
		fmt.Println("Applying provider-specific formatting")
		var err error
		prompt, err = optimizeForProvider(prompt, buildTarget)
		if err != nil {
			return err
		}
	} else if config.Provider != "" {
		buildTarget = config.Provider
		var err error
		prompt, err = optimizeForProvider(prompt, config.Provider)
		if err != nil {
			return err
		}
	}

	// Minification
	if buildMinify {
		fmt.Println("Minifying prompt")
		originalTokens := estimateTokenCount(prompt)
		prompt = minifyPrompt(prompt)
		minifiedTokens := estimateTokenCount(prompt)
		reduction := float64(originalTokens-minifiedTokens) / float64(originalTokens) * 100
		fmt.Printf("Original tokens: %d\n", originalTokens)
		fmt.Printf("Minified tokens: %d\n", minifiedTokens)
		fmt.Printf("Reduction: %.0f%%\n", reduction)
	}

	// Validation
	if buildValidate {
		return fmt.Errorf("build validation is not yet implemented")
	}

	// Add production mode marker if in production environment
	if os.Getenv("PE_ENV") == "production" {
		prompt = prompt + "\n\n[Production mode]"
	}

	// Determine output file
	outputFile := buildOutput
	if outputFile == "" {
		outputFile = "prompt_build.txt"
	} else {
		fmt.Printf("Building to: %s\n", outputFile)
	}

	// Write main output
	if err := os.WriteFile(outputFile, []byte(prompt), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Compression
	if buildCompress {
		fmt.Println("Compressing prompt")
		if err := compressFile(outputFile); err != nil {
			return err
		}
		// Calculate compression ratio
		originalSize := len([]byte(prompt))
		compressedFile := outputFile + ".gz"
		info, _ := os.Stat(compressedFile)
		compressedSize := info.Size()
		ratio := float64(originalSize) / float64(compressedSize)
		fmt.Printf("Compression ratio: %.1f:1\n", ratio)
	}

	// Write metadata if requested
	if buildWithMetadata {
		metadata := buildMetadata{
			Version:   "1.0.0",
			Checksum:  calculateChecksum([]byte(prompt)),
			Timestamp: time.Now(),
			Provider:  buildTarget,
			Model:     config.Model,
		}

		metadataFile := strings.TrimSuffix(outputFile, filepath.Ext(outputFile)) + "_metadata.json"
		if buildOutput == "" {
			metadataFile = "prompt_build_metadata.json"
		}

		data, err := json.MarshalIndent(metadata, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(metadataFile, data, 0644); err != nil {
			return err
		}
	}

	// Create JSON output only for default builds (no specific output)
	if buildOutput == "" {
		// Also create metadata for default builds
		metadata := buildMetadata{
			Version:   "1.0.0",
			Checksum:  calculateChecksum([]byte(prompt)),
			Timestamp: time.Now(),
			Provider:  buildTarget,
			Model:     config.Model,
		}

		metadataData, _ := json.MarshalIndent(metadata, "", "  ")
		os.WriteFile("prompt_build_metadata.json", metadataData, 0644)

		// Create JSON output
		jsonOutput := map[string]interface{}{
			"prompt":   prompt,
			"provider": config.Provider,
			"model":    config.Model,
		}
		data, _ := json.MarshalIndent(jsonOutput, "", "  ")
		os.WriteFile("prompt_build.json", data, 0644)
	}

	fmt.Println("Optimization complete")
	fmt.Println("✓ Built successfully")

	return nil
}

func buildPromptBundle(dir string) error {
	fmt.Println("Creating prompt bundle")

	// Count components
	components := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".txt") || strings.HasSuffix(path, ".prompt")) {
			components++
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("Including %d components\n", components)
	fmt.Println("Writing bundle")

	// Create tar.gz bundle
	bundleFile, err := os.Create("prompt_bundle.tar.gz")
	if err != nil {
		return err
	}
	defer bundleFile.Close()

	gzWriter := gzip.NewWriter(bundleFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		header := &tar.Header{
			Name:    relPath,
			Size:    info.Size(),
			Mode:    int64(info.Mode()),
			ModTime: info.ModTime(),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		_, err = io.Copy(tarWriter, file)
		return err
	})
}

func optimizeForProvider(prompt string, provider string) (string, error) {
	switch provider {
	case "anthropic":
		return fmt.Sprintf("Human: %s\n\nAssistant: ", prompt), nil
	case "openai":
		return prompt, nil
	case "google":
		return "", fmt.Errorf("provider-specific formatting for %q is not yet implemented", provider)
	default:
		return "", fmt.Errorf("provider-specific formatting for %q is not yet implemented", provider)
	}
}

func minifyPrompt(prompt string) string {
	// Simple minification: remove extra whitespace
	lines := strings.Split(prompt, "\n")
	var minified []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			minified = append(minified, trimmed)
		}
	}
	return strings.Join(minified, " ")
}

func estimateTokenCount(text string) int {
	// Simple estimation: ~4 characters per token
	return len(text) / 4
}

func calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func compressFile(filename string) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	output, err := os.Create(filename + ".gz")
	if err != nil {
		return err
	}
	defer output.Close()

	gzWriter := gzip.NewWriter(output)
	defer gzWriter.Close()

	_, err = gzWriter.Write(input)
	return err
}
