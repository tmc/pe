// Package template provides utilities for working with templated prompts.
package template

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileProcessor processes file inclusion directives in templates.
type FileProcessor struct {
	// Base directory for resolving relative paths
	BaseDir string
	// Maximum recursion depth to prevent infinite loops
	MaxDepth int
	// Current recursion depth
	currentDepth int
	// Processed file cache to avoid redundant processing
	processedFiles map[string]string
}

// NewFileProcessor creates a new FileProcessor with the given base directory.
func NewFileProcessor(baseDir string) *FileProcessor {
	return &FileProcessor{
		BaseDir:        baseDir,
		MaxDepth:       10,
		currentDepth:   0,
		processedFiles: make(map[string]string),
	}
}

// Process processes file inclusion directives in the template.
// Supports #include "/path/to/file.txt" syntax.
func (p *FileProcessor) Process(text string) (string, error) {
	// First handle {{file "/path/to/file"}} (legacy format)
	legacyRe := regexp.MustCompile(`{{file\s+"([^"]+)"}}`)
	processed := legacyRe.ReplaceAllStringFunc(text, func(match string) string {
		fileMatch := legacyRe.FindStringSubmatch(match)
		if len(fileMatch) < 2 {
			return match // Keep original if no file path found
		}
		
		content, err := p.includeFile(fileMatch[1])
		if err != nil {
			// Return an error placeholder that will be easy to identify
			return fmt.Sprintf("ERROR_READING_FILE:%s", fileMatch[1])
		}
		
		return content
	})
	
	// Then handle #include "/path/to/file.txt" (new format)
	includeRe := regexp.MustCompile(`#include\s+"([^"]+)"`)
	processed = includeRe.ReplaceAllStringFunc(processed, func(match string) string {
		fileMatch := includeRe.FindStringSubmatch(match)
		if len(fileMatch) < 2 {
			return match // Keep original if no file path found
		}
		
		content, err := p.includeFile(fileMatch[1])
		if err != nil {
			// Return an error placeholder that will be easy to identify
			return fmt.Sprintf("ERROR_READING_FILE:%s", fileMatch[1])
		}
		
		return content
	})
	
	// Check for any file error placeholders
	if strings.Contains(processed, "ERROR_READING_FILE:") {
		re := regexp.MustCompile(`ERROR_READING_FILE:([^\s]+)`)
		match := re.FindStringSubmatch(processed)
		if len(match) >= 2 {
			return "", fmt.Errorf("error reading file: %s", match[1])
		}
	}
	
	return processed, nil
}

// includeFile reads the content of a file and processes any nested includes.
func (p *FileProcessor) includeFile(filePath string) (string, error) {
	// Check recursion depth
	if p.currentDepth >= p.MaxDepth {
		return "", fmt.Errorf("maximum inclusion depth reached (%d)", p.MaxDepth)
	}
	
	// Check cache first
	if content, exists := p.processedFiles[filePath]; exists {
		return content, nil
	}
	
	// Resolve relative paths
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(p.BaseDir, filePath)
	}
	
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	
	// Process nested includes
	p.currentDepth++
	processed, err := p.Process(string(content))
	p.currentDepth--
	
	if err != nil {
		return "", err
	}
	
	// Cache the result
	p.processedFiles[filePath] = processed
	
	return processed, nil
}

// ExtractFileInclusions identifies all file inclusions in a template.
func ExtractFileInclusions(text string) []string {
	var files []string
	seen := make(map[string]bool)
	
	// Check for {{file "/path/to/file"}} syntax (legacy)
	legacyRe := regexp.MustCompile(`{{file\s+"([^"]+)"}}`)
	legacyMatches := legacyRe.FindAllStringSubmatch(text, -1)
	
	for _, match := range legacyMatches {
		if len(match) > 1 {
			filePath := match[1]
			if !seen[filePath] {
				files = append(files, filePath)
				seen[filePath] = true
			}
		}
	}
	
	// Check for #include "/path/to/file.txt" syntax (new)
	includeRe := regexp.MustCompile(`#include\s+"([^"]+)"`)
	includeMatches := includeRe.FindAllStringSubmatch(text, -1)
	
	for _, match := range includeMatches {
		if len(match) > 1 {
			filePath := match[1]
			if !seen[filePath] {
				files = append(files, filePath)
				seen[filePath] = true
			}
		}
	}
	
	return files
}

// ValidateFileInclusions checks if all included files exist.
func ValidateFileInclusions(text string, baseDir string) []string {
	files := ExtractFileInclusions(text)
	var missing []string
	
	for _, filePath := range files {
		// Resolve relative paths
		resolvedPath := filePath
		if !filepath.IsAbs(filePath) {
			resolvedPath = filepath.Join(baseDir, filePath)
		}
		
		// Check if file exists
		if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
			missing = append(missing, filePath)
		}
	}
	
	return missing
}