package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// viewCmd returns a cobra.Command for the 'view' subcommand.
//
// view opens the promptfoo viewer to visualize evaluation results
//
// Usage:
//
//	pe view
//	pe view [evalId]
func viewCmd() *cobra.Command {
	var fileName string
	
	cmd := &cobra.Command{
		Use:   "view [evalId]",
		Short: "View evaluation results in browser UI",
		Long:  `View evaluation results in the promptfoo browser-based UI.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if fileName != "" {
				// Import the file first
				importCmd := exec.Command("npx", "promptfoo", "import", fileName)
				importCmd.Stderr = os.Stderr
				output, err := importCmd.Output()
				if err != nil {
					return fmt.Errorf("error importing results: %v", err)
				}
				fmt.Println(string(output))
			}
			
			return runView(cmd, args)
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation and auto-open the URL")
	cmd.Flags().StringVarP(&fileName, "file", "f", "", "Path to evaluation results file")
	return cmd
}

func runView(cmd *cobra.Command, args []string) error {
	// Check if promptfoo is installed
	if _, err := exec.LookPath("npx"); err != nil {
		return fmt.Errorf("npx not found. Please install Node.js and npm: https://nodejs.org/")
	}

	// Build the command to run promptfoo view
	promptfooCmd := exec.Command("npx", "promptfoo", "view")
	
	// Add the eval ID if provided
	if len(args) > 0 {
		promptfooCmd.Args = append(promptfooCmd.Args, args[0])
	}
	
	// Add the -y flag if specified
	yes, _ := cmd.Flags().GetBool("yes")
	if yes {
		promptfooCmd.Args = append(promptfooCmd.Args, "-y")
	}
	
	// Connect the command's stdio to our process
	promptfooCmd.Stdin = os.Stdin
	promptfooCmd.Stdout = os.Stdout
	promptfooCmd.Stderr = os.Stderr
	
	// Run the command
	return promptfooCmd.Run()
}