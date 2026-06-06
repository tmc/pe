package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/promptfoo"
)

// viewCmd returns a cobra.Command for the 'view' subcommand.
//
// view displays evaluation results in a browser
//
// Usage:
//
//	pe view
//	pe view [evalId]
//	pe view -f [file]
func viewCmd() *cobra.Command {
	var fileName string
	var port int
	var promptfooView bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "view [evalId]",
		Short: "View evaluation results in browser UI",
		Long: `View local evaluation results in a browser UI.

Use --promptfoo to explicitly delegate to the promptfoo CLI viewer.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if promptfooView {
				return runPromptfooView(args, yes)
			}

			// Case 1: File specified with -f flag
			if fileName != "" {
				return viewFile(fileName, port)
			}

			// Case 2: Eval ID specified as positional argument
			if len(args) > 0 {
				evalId := args[0]

				// Check if the file exists in the standard location
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("error getting user home directory: %v", err)
				}

				evalFile := filepath.Join(homeDir, ".promptfoo", "evals", evalId+".json")
				if _, err := os.Stat(evalFile); err == nil {
					return viewFile(evalFile, port)
				}

				return fmt.Errorf("evaluation %q not found in %s", evalId, filepath.Dir(evalFile))
			}

			// Case 3: No arguments, list available evaluations
			return listEvaluations(cmd)
		},
	}

	cmd.Flags().StringVarP(&fileName, "file", "f", "", "Path to evaluation results file")
	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to use for local viewer")
	cmd.Flags().BoolVar(&promptfooView, "promptfoo", false, "Open the promptfoo CLI viewer instead of the local viewer")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Pass -y to promptfoo when used with --promptfoo")

	return cmd
}

// viewFile displays the evaluation results from a file in a browser
func viewFile(filePath string, port int) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Parse the JSON to confirm it's valid
	var results promptfoo.EvaluationResult
	if err := json.Unmarshal(data, &results); err != nil {
		return fmt.Errorf("error parsing evaluation results: %v", err)
	}

	// Extract the evaluation ID
	evalId := results.EvalID
	if evalId == "" {
		evalId = filepath.Base(filePath)
	}

	fmt.Printf("Starting viewer for evaluation ID: %s\n", evalId)
	fmt.Printf("Press Ctrl+C to stop the server\n\n")

	// Create a simple HTTP server to serve the file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve the HTML viewer
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, generateHtmlViewer(evalId))
	})

	http.HandleFunc("/data.json", func(w http.ResponseWriter, r *http.Request) {
		// Serve the JSON data
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	// Determine the URL
	url := fmt.Sprintf("http://localhost:%d", port)
	fmt.Printf("View results at: %s\n", url)

	// Open the browser
	go func() {
		time.Sleep(500 * time.Millisecond) // Give the server a moment to start
		openBrowser(url)
	}()

	// Start the server
	return http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

// listEvaluations lists available evaluations in the standard location
func listEvaluations(cmd *cobra.Command) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %v", err)
	}

	evalsDir := filepath.Join(homeDir, ".promptfoo", "evals")

	// Check if the directory exists
	if _, err := os.Stat(evalsDir); os.IsNotExist(err) {
		fmt.Fprintln(cmd.OutOrStdout(), "No evaluations found. Run 'pe eval --save-db' to save an evaluation.")
		return nil
	}

	// List files in the evals directory
	files, err := os.ReadDir(evalsDir)
	if err != nil {
		return fmt.Errorf("error reading evaluations directory: %v", err)
	}

	if len(files) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No evaluations found. Run 'pe eval --save-db' to save an evaluation.")
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Available evaluations:")
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			evalId := filepath.Base(file.Name())
			evalId = evalId[:len(evalId)-5] // Remove .json extension
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", evalId)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), "\nTo view an evaluation, run:")
	fmt.Fprintln(cmd.OutOrStdout(), "  pe view <evalId>")
	fmt.Fprintln(cmd.OutOrStdout(), "  pe view -f <file>")

	return nil
}

func runPromptfooView(args []string, yes bool) error {
	if _, err := exec.LookPath("npx"); err != nil {
		return fmt.Errorf("npx not found for promptfoo viewer")
	}

	promptfooArgs := []string{"promptfoo", "view"}
	if len(args) > 0 {
		promptfooArgs = append(promptfooArgs, args[0])
	}
	if yes {
		promptfooArgs = append(promptfooArgs, "-y")
	}

	cmd := exec.Command("npx", promptfooArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// openBrowser opens the default browser with the provided URL
func openBrowser(url string) {
	var err error

	switch os.Getenv("GOOS") {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default: // Linux and others
		err = exec.Command("xdg-open", url).Start()
	}

	if err != nil {
		fmt.Printf("Error opening browser: %v\n", err)
		fmt.Printf("Please open %s in your browser\n", url)
	}
}

// generateHtmlViewer generates a simple HTML viewer for the evaluation results
func generateHtmlViewer(evalId string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Evaluation Results: ` + evalId + `</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        h1, h2, h3 {
            color: #1a73e8;
        }
        .container {
            display: grid;
            grid-template-columns: 250px 1fr;
            gap: 20px;
        }
        .sidebar {
            background-color: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
        }
        .sidebar ul {
            list-style-type: none;
            padding: 0;
        }
        .sidebar li {
            margin-bottom: 10px;
            cursor: pointer;
            padding: 8px 12px;
            border-radius: 4px;
        }
        .sidebar li:hover, .sidebar li.active {
            background-color: #e8f0fe;
            color: #1a73e8;
        }
        .content {
            background-color: #fff;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }
        .result-card {
            border: 1px solid #e1e4e8;
            border-radius: 6px;
            margin-bottom: 20px;
            overflow: hidden;
        }
        .result-header {
            background-color: #f6f8fa;
            padding: 12px 15px;
            border-bottom: 1px solid #e1e4e8;
            font-weight: 600;
        }
        .result-body {
            padding: 15px;
        }
        .result-section {
            margin-bottom: 15px;
        }
        .result-section h4 {
            margin-top: 0;
            margin-bottom: 8px;
        }
        .success {
            color: #28a745;
        }
        .failure {
            color: #dc3545;
        }
        pre {
            background-color: #f6f8fa;
            border-radius: 3px;
            padding: 12px;
            overflow-x: auto;
        }
        .stats {
            background-color: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
        }
        .stat-box {
            background-color: #fff;
            padding: 12px;
            border-radius: 6px;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
            text-align: center;
        }
        .stat-value {
            font-size: 24px;
            font-weight: 600;
            margin: 5px 0;
        }
        .stat-label {
            color: #6c757d;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <h1>Evaluation Results: <span id="eval-id">` + evalId + `</span></h1>
    
    <div id="loading">Loading evaluation data...</div>
    
    <div id="results-container" style="display: none;">
        <div id="stats" class="stats">
            <h2>Statistics</h2>
            <div class="stats-grid">
                <div class="stat-box">
                    <div class="stat-label">Success Rate</div>
                    <div id="pass-rate" class="stat-value">0%</div>
                </div>
                <div class="stat-box">
                    <div class="stat-label">Successes</div>
                    <div id="successes" class="stat-value">0</div>
                </div>
                <div class="stat-box">
                    <div class="stat-label">Failures</div>
                    <div id="failures" class="stat-value">0</div>
                </div>
                <div class="stat-box">
                    <div class="stat-label">Total Tests</div>
                    <div id="total-tests" class="stat-value">0</div>
                </div>
            </div>
        </div>
        
        <div class="container">
            <div class="sidebar">
                <h3>Test Cases</h3>
                <ul id="test-list"></ul>
            </div>
            
            <div id="content" class="content">
                <p>Select a test case from the sidebar to view details.</p>
            </div>
        </div>
    </div>

    <script>
        // Fetch the evaluation data
        fetch('/data.json')
            .then(response => response.json())
            .then(data => {
                // Hide loading indicator
                document.getElementById('loading').style.display = 'none';
                document.getElementById('results-container').style.display = 'block';
                
                // Set evaluation ID
                document.getElementById('eval-id').textContent = data.evalId || '` + evalId + `';
                
                // Update statistics
                updateStatistics(data);
                
                // Generate test list
                generateTestList(data);
            })
            .catch(error => {
                document.getElementById('loading').textContent = 'Error loading evaluation data: ' + error.message;
            });
        
        function updateStatistics(data) {
            const resultsData = data.results || {};
            const stats = resultsData.stats || {};
            
            // Calculate statistics - handle both number types (int or float)
            let successes = stats.successes || 0;
            let failures = stats.failures || 0;
            
            // Convert to numbers in case they're strings
            successes = Number(successes);
            failures = Number(failures);
            
            const totalTests = successes + failures;
            const passRate = totalTests > 0 ? ((successes / totalTests) * 100).toFixed(2) : '0.00';
            
            console.log('Statistics:', { successes, failures, totalTests, passRate });
            
            // Update the UI
            document.getElementById('pass-rate').textContent = passRate + '%';
            document.getElementById('successes').textContent = successes;
            document.getElementById('failures').textContent = failures;
            document.getElementById('total-tests').textContent = totalTests;
        }
        
        function generateTestList(data) {
            console.log('Data structure:', JSON.stringify(data, null, 2).substring(0, 1000));
            
            const results = (data.results && data.results.results) || [];
            console.log('Results count:', results.length);
            
            const testList = document.getElementById('test-list');
            
            // Group results by test case
            const testCases = {};
            
            results.forEach(result => {
                const vars = result.vars || {};
                // Use test variables for grouping
                const testKey = JSON.stringify(vars);
                
                if (!testCases[testKey]) {
                    testCases[testKey] = {
                        vars,
                        results: []
                    };
                }
                
                testCases[testKey].results.push(result);
            });
            
            console.log('Unique test cases:', Object.keys(testCases).length);
            
            // Create list items for each test case
            Object.entries(testCases).forEach(([key, testCase], index) => {
                const li = document.createElement('li');
                const vars = testCase.vars;
                
                // Create a readable name for the test
                let testName = '';
                if (vars.input && vars.language) {
                    testName = vars.input + ' (' + vars.language + ')';
                } else {
                    testName = 'Test ' + (index + 1);
                }
                
                li.textContent = testName;
                li.dataset.key = key;
                
                li.addEventListener('click', () => {
                    // Remove active class from all list items
                    document.querySelectorAll('#test-list li').forEach(item => {
                        item.classList.remove('active');
                    });
                    
                    // Add active class to clicked item
                    li.classList.add('active');
                    
                    // Show test details
                    showTestDetails(testCase);
                });
                
                testList.appendChild(li);
            });
            
            // Select first test by default if available
            if (testList.firstChild) {
                testList.firstChild.click();
            }
        }
        
        function showTestDetails(testCase) {
            const content = document.getElementById('content');
            content.innerHTML = '';
            
            // Create test details header
            const header = document.createElement('h2');
            const vars = testCase.vars;
            
            // Create a readable name for the test
            let testName = '';
            if (vars.input && vars.language) {
                testName = vars.input + ' (' + vars.language + ')';
            } else {
                testName = 'Test Details';
            }
            
            header.textContent = testName;
            content.appendChild(header);
            
            // Variables section
            const varsSection = document.createElement('div');
            varsSection.className = 'result-card';
            
            const varsHeader = document.createElement('div');
            varsHeader.className = 'result-header';
            varsHeader.textContent = 'Variables';
            varsSection.appendChild(varsHeader);
            
            const varsBody = document.createElement('div');
            varsBody.className = 'result-body';
            
            const varsPre = document.createElement('pre');
            varsPre.textContent = JSON.stringify(vars, null, 2);
            varsBody.appendChild(varsPre);
            
            varsSection.appendChild(varsBody);
            content.appendChild(varsSection);
            
            // Results section
            const results = testCase.results || [];
            results.forEach(result => {
                const resultCard = document.createElement('div');
                resultCard.className = 'result-card';
                
                // Result header
                const resultHeader = document.createElement('div');
                resultHeader.className = 'result-header';
                
                const prompt = result.prompt && result.prompt.label ? result.prompt.label : 'Unknown Prompt';
                const provider = result.provider && result.provider.id ? result.provider.id : 'Unknown Provider';
                
                resultHeader.textContent = provider + ' (' + prompt + ')';
                resultCard.appendChild(resultHeader);
                
                // Result body
                const resultBody = document.createElement('div');
                resultBody.className = 'result-body';
                
                // Status section
                const statusSection = document.createElement('div');
                statusSection.className = 'result-section';
                
                const statusHeader = document.createElement('h4');
                statusHeader.textContent = 'Status';
                statusSection.appendChild(statusHeader);
                
                const statusContent = document.createElement('div');
                const success = result.success;
                statusContent.className = success ? 'success' : 'failure';
                statusContent.textContent = success ? '✓ Pass' : '✗ Fail';
                statusSection.appendChild(statusContent);
                
                resultBody.appendChild(statusSection);
                
                // Response section
                const responseSection = document.createElement('div');
                responseSection.className = 'result-section';
                
                const responseHeader = document.createElement('h4');
                responseHeader.textContent = 'Response';
                responseSection.appendChild(responseHeader);
                
                const response = result.response || {};
                const output = response.output || 'No output';
                
                const responsePre = document.createElement('pre');
                responsePre.textContent = output;
                responseSection.appendChild(responsePre);
                
                resultBody.appendChild(responseSection);
                
                // Add result body to card
                resultCard.appendChild(resultBody);
                
                // Add the result card to the content
                content.appendChild(resultCard);
            });
        }
    </script>
</body>
</html>`
}
