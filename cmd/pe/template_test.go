package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tmc/pe/internal/template"
)

func testTemplate() {
	// Create a simple test
	tmpl := template.NewTemplate("Hello, {{name}}!", map[string]interface{}{
		"name": "World",
	})
	
	result, err := tmpl.Process()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Template result: %s\n", result)
	
	// Test conditional logic
	condTmpl := template.NewTemplate("#if showGreeting then Hello, {{name}}! #else Goodbye, {{name}}! #endif", 
		map[string]interface{}{
			"showGreeting": true,
			"name":         "World",
		})
	
	condResult, err := condTmpl.Process()
	if err != nil {
		fmt.Printf("Error in conditional template: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Conditional template result: %s\n", condResult)
	
	// Test file inclusion
	tempDir, err := os.MkdirTemp("", "template-test")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("File content"), 0644); err != nil {
		fmt.Printf("Failed to write test file: %v\n", err)
		os.Exit(1)
	}
	
	fileTmpl := template.NewTemplate("#include \""+testFile+"\"", nil)
	fileResult, err := fileTmpl.Process()
	if err != nil {
		fmt.Printf("Error in file template: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("File inclusion template result: %s\n", fileResult)
	
	fmt.Println("All template tests completed successfully!")
}