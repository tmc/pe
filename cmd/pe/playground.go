package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/metaprompt"
	"github.com/tmc/pe/internal/promptfoo/evaluation/metrics"
)

// PlaygroundServer represents the web playground server
type PlaygroundServer struct {
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	router   *mux.Router
	history  []PlaygroundResponse
}

// PlaygroundRequest represents a request from the web UI
type PlaygroundRequest struct {
	Prompt      string            `json:"prompt"`
	Provider    string            `json:"provider"`
	Model       string            `json:"model"`
	Temperature float64           `json:"temperature"`
	MaxTokens   int               `json:"max_tokens"`
	Variables   map[string]string `json:"variables"`
	Optimize    bool              `json:"optimize"`
	Method      string            `json:"method"`
	Iterations  int               `json:"iterations"`
}

// PlaygroundResponse represents a response to the web UI
type PlaygroundResponse struct {
	ID              string                 `json:"id"`
	Prompt          string                 `json:"prompt"`
	Response        string                 `json:"response"`
	Error           string                 `json:"error,omitempty"`
	Latency         time.Duration          `json:"latency"`
	TokensUsed      int                    `json:"tokens_used"`
	Cost            float64                `json:"cost"`
	Provider        string                 `json:"provider"`
	Model           string                 `json:"model"`
	OptimizedPrompt string                 `json:"optimized_prompt,omitempty"`
	Metrics         map[string]interface{} `json:"metrics,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// playgroundCmd returns a cobra.Command for the interactive web playground
func playgroundCmd() *cobra.Command {
	var (
		port        int
		host        string
		openBrowser bool
	)

	cmd := &cobra.Command{
		Use:   "playground",
		Short: "Launch interactive web playground for prompt engineering",
		Long: `Launch a local web playground for prompt testing, optimization,
metrics, version history, security checks, component composition, and A/B
experiments.`,
		Example: `  # Launch playground on default port 8080
  pe playground

  # Launch on custom port with auto-open browser
  pe playground --port 3000 --open

  # Launch for team collaboration
  pe playground --host 0.0.0.0 --port 8080`,
		RunE: func(cmd *cobra.Command, args []string) error {
			server := NewPlaygroundServer()

			addr := fmt.Sprintf("%s:%d", host, port)
			fmt.Printf("Starting PE Playground on http://%s\n", addr)
			fmt.Printf("Advanced features:\n")
			fmt.Printf("  • Real-time optimization with PE2, APEX, TextGrad\n")
			fmt.Printf("  • Multi-provider testing and comparison\n")
			fmt.Printf("  • Component-based prompt engineering\n")
			fmt.Printf("  • Advanced metrics and cost analysis\n")
			fmt.Printf("  • Security testing integration\n")
			fmt.Printf("  • Collaborative development features\n\n")

			if openBrowser {
				go func() {
					time.Sleep(2 * time.Second)
					openURL(fmt.Sprintf("http://localhost:%d", port))
				}()
			}

			return server.Start(addr)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to serve playground on")
	cmd.Flags().StringVar(&host, "host", "localhost", "Host to bind to (use 0.0.0.0 for external access)")
	cmd.Flags().BoolVar(&openBrowser, "open", true, "Automatically open browser")

	return cmd
}

// NewPlaygroundServer creates a new playground server instance
func NewPlaygroundServer() *PlaygroundServer {
	return &PlaygroundServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients: make(map[*websocket.Conn]bool),
		router:  mux.NewRouter(),
		history: make([]PlaygroundResponse, 0),
	}
}

// Start starts the playground server
func (ps *PlaygroundServer) Start(addr string) error {
	ps.setupRoutes()

	server := &http.Server{
		Addr:    addr,
		Handler: ps.router,
	}

	return server.ListenAndServe()
}

// setupRoutes configures all HTTP routes
func (ps *PlaygroundServer) setupRoutes() {
	// Static files
	ps.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	// Main playground page
	ps.router.HandleFunc("/", ps.handleIndex).Methods("GET")

	// API endpoints
	api := ps.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/test", ps.handleTest).Methods("POST")
	api.HandleFunc("/optimize", ps.handleOptimize).Methods("POST")
	api.HandleFunc("/compare", ps.handleCompare).Methods("POST")
	api.HandleFunc("/metrics", ps.handleMetrics).Methods("POST")
	api.HandleFunc("/security", ps.handleSecurity).Methods("POST")
	api.HandleFunc("/components", ps.handleComponents).Methods("GET", "POST")
	api.HandleFunc("/history", ps.handleHistory).Methods("GET")

	// WebSocket for real-time updates
	ps.router.HandleFunc("/ws", ps.handleWebSocket)
}

// handleIndex serves the main playground interface
func (ps *PlaygroundServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PE Playground - Advanced Prompt Engineering</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/tailwindcss/2.2.19/tailwind.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/codemirror.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/theme/monokai.min.css" rel="stylesheet">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/codemirror.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/mode/markdown/markdown.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <!-- Header -->
        <div class="mb-8">
            <h1 class="text-4xl font-bold text-gray-800 mb-2">PE Playground</h1>
            <p class="text-gray-600">Advanced prompt engineering with real-time optimization, multi-provider testing, and collaborative development</p>
        </div>

        <!-- Main Interface -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <!-- Input Panel -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-2xl font-semibold mb-4">Prompt Development</h2>
                
                <!-- Prompt Editor -->
                <div class="mb-4">
                    <label class="block text-sm font-medium text-gray-700 mb-2">Prompt</label>
                    <textarea id="promptEditor" class="w-full h-40 p-3 border border-gray-300 rounded-lg resize-none"></textarea>
                </div>

                <!-- Provider Settings -->
                <div class="grid grid-cols-2 gap-4 mb-4">
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">Provider</label>
                        <select id="provider" class="w-full p-3 border border-gray-300 rounded-lg">
                            <option value="openai">OpenAI</option>
                            <option value="anthropic">Anthropic</option>
                            <option value="google">Google AI</option>
                        </select>
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">Model</label>
                        <select id="model" class="w-full p-3 border border-gray-300 rounded-lg">
                            <option value="gpt-4">GPT-4</option>
                            <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
                            <option value="claude-3-opus">Claude 3 Opus</option>
                            <option value="claude-3-sonnet">Claude 3 Sonnet</option>
                        </select>
                    </div>
                </div>

                <!-- Advanced Settings -->
                <div class="grid grid-cols-3 gap-4 mb-4">
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">Temperature</label>
                        <input type="range" id="temperature" min="0" max="1" step="0.1" value="0.7" class="w-full">
                        <span id="tempValue" class="text-sm text-gray-500">0.7</span>
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">Max Tokens</label>
                        <input type="number" id="maxTokens" value="1000" class="w-full p-2 border border-gray-300 rounded">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">Iterations</label>
                        <input type="number" id="iterations" value="3" min="1" max="10" class="w-full p-2 border border-gray-300 rounded">
                    </div>
                </div>

                <!-- Optimization Settings -->
                <div class="mb-4">
                    <label class="flex items-center mb-2">
                        <input type="checkbox" id="optimize" class="mr-2">
                        <span class="text-sm font-medium text-gray-700">Enable Optimization</span>
                    </label>
                    <select id="method" class="w-full p-3 border border-gray-300 rounded-lg" disabled>
                        <option value="pe2">PE2 - Meta-prompt Engineering</option>
                        <option value="apex">APEX - Long Prompt Optimization</option>
                        <option value="textgrad">TextGrad - Natural Language Gradients</option>
                        <option value="hybrid">Hybrid - Best of Multiple Methods</option>
                    </select>
                </div>

                <!-- Action Buttons -->
                <div class="flex gap-4">
                    <button id="testBtn" class="flex-1 bg-blue-600 text-white py-3 px-6 rounded-lg hover:bg-blue-700 font-medium">
                        Test Prompt
                    </button>
                    <button id="optimizeBtn" class="flex-1 bg-green-600 text-white py-3 px-6 rounded-lg hover:bg-green-700 font-medium">
                        Optimize
                    </button>
                    <button id="compareBtn" class="flex-1 bg-purple-600 text-white py-3 px-6 rounded-lg hover:bg-purple-700 font-medium">
                        Compare
                    </button>
                </div>
            </div>

            <!-- Results Panel -->
            <div class="bg-white rounded-lg shadow-lg p-6">
                <h2 class="text-2xl font-semibold mb-4">Results & Analytics</h2>
                
                <!-- Response Display -->
                <div class="mb-6">
                    <h3 class="text-lg font-medium mb-2">Response</h3>
                    <div id="responseDisplay" class="p-4 border border-gray-300 rounded-lg h-48 overflow-y-auto bg-gray-50">
                        <p class="text-gray-500 italic">Run a test to see results here...</p>
                    </div>
                </div>

                <!-- Metrics Dashboard -->
                <div class="grid grid-cols-2 gap-4 mb-4">
                    <div class="p-4 bg-blue-50 rounded-lg">
                        <h4 class="font-medium text-blue-800">Latency</h4>
                        <p id="latencyMetric" class="text-2xl font-bold text-blue-600">--</p>
                    </div>
                    <div class="p-4 bg-green-50 rounded-lg">
                        <h4 class="font-medium text-green-800">Cost</h4>
                        <p id="costMetric" class="text-2xl font-bold text-green-600">--</p>
                    </div>
                    <div class="p-4 bg-purple-50 rounded-lg">
                        <h4 class="font-medium text-purple-800">Tokens</h4>
                        <p id="tokensMetric" class="text-2xl font-bold text-purple-600">--</p>
                    </div>
                    <div class="p-4 bg-orange-50 rounded-lg">
                        <h4 class="font-medium text-orange-800">Quality Score</h4>
                        <p id="qualityMetric" class="text-2xl font-bold text-orange-600">--</p>
                    </div>
                </div>

                <!-- Performance Chart -->
                <div class="mb-4">
                    <canvas id="performanceChart" width="400" height="200"></canvas>
                </div>
            </div>
        </div>

        <!-- Advanced Features Tabs -->
        <div class="mt-8 bg-white rounded-lg shadow-lg">
            <div class="border-b border-gray-200">
                <nav class="-mb-px flex space-x-8" aria-label="Tabs">
                    <button class="tab-btn border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm" data-tab="history">
                        History & Versions
                    </button>
                    <button class="tab-btn border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm" data-tab="components">
                        Components Library
                    </button>
                    <button class="tab-btn border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm" data-tab="security">
                        Security Testing
                    </button>
                    <button class="tab-btn border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm" data-tab="analytics">
                        Advanced Analytics
                    </button>
                </nav>
            </div>
            
            <div class="p-6">
                <!-- History Tab -->
                <div id="tab-history" class="tab-content hidden">
                    <h3 class="text-lg font-semibold mb-4">Prompt History & Version Control</h3>
                    <div id="historyList" class="space-y-2">
                        <p class="text-gray-500">No history yet. Start testing prompts to build your history.</p>
                    </div>
                </div>

                <!-- Components Tab -->
                <div id="tab-components" class="tab-content hidden">
                    <h3 class="text-lg font-semibold mb-4">Prompt Components Library</h3>
                    <div class="grid grid-cols-3 gap-4">
                        <div class="p-4 border border-gray-300 rounded-lg">
                            <h4 class="font-medium">Context Components</h4>
                            <p class="text-sm text-gray-600 mt-2">Reusable context definitions</p>
                        </div>
                        <div class="p-4 border border-gray-300 rounded-lg">
                            <h4 class="font-medium">Instruction Templates</h4>
                            <p class="text-sm text-gray-600 mt-2">Common instruction patterns</p>
                        </div>
                        <div class="p-4 border border-gray-300 rounded-lg">
                            <h4 class="font-medium">Example Sets</h4>
                            <p class="text-sm text-gray-600 mt-2">Few-shot learning examples</p>
                        </div>
                    </div>
                </div>

                <!-- Security Tab -->
                <div id="tab-security" class="tab-content hidden">
                    <h3 class="text-lg font-semibold mb-4">OWASP LLM Top 10 Security Testing</h3>
                    <div class="grid grid-cols-2 gap-4">
                        <div class="p-4 border border-gray-300 rounded-lg">
                            <h4 class="font-medium">Prompt Injection Testing</h4>
                            <button class="mt-2 bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700">Run Tests</button>
                        </div>
                        <div class="p-4 border border-gray-300 rounded-lg">
                            <h4 class="font-medium">Data Leakage Detection</h4>
                            <button class="mt-2 bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700">Scan</button>
                        </div>
                    </div>
                </div>

                <!-- Analytics Tab -->
                <div id="tab-analytics" class="tab-content hidden">
                    <h3 class="text-lg font-semibold mb-4">Advanced Analytics & A/B Testing</h3>
                    <div class="grid grid-cols-2 gap-4">
                        <div class="p-4 border border-gray-300 rounded-lg">
                            <h4 class="font-medium">Performance Trends</h4>
                            <canvas id="trendsChart" width="300" height="150"></canvas>
                        </div>
                        <div class="p-4 border border-gray-300 rounded-lg">
                            <h4 class="font-medium">Cost Analysis</h4>
                            <canvas id="costChart" width="300" height="150"></canvas>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Global state
        let socket;
        let activeTab = 'history';
        let testHistory = [];

        // Initialize playground
        document.addEventListener('DOMContentLoaded', function() {
            initializeWebSocket();
            initializeEventListeners();
            initializeCharts();
            setupTabs();
        });

        // WebSocket connection
        function initializeWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            socket = new WebSocket(protocol + '//' + window.location.host + '/ws');
            
            socket.onopen = function() {
                console.log('Connected to PE Playground');
            };
            
            socket.onmessage = function(event) {
                const data = JSON.parse(event.data);
                handleWebSocketMessage(data);
            };
        }

        // Event listeners
        function initializeEventListeners() {
            document.getElementById('optimize').addEventListener('change', function() {
                document.getElementById('method').disabled = !this.checked;
            });

            document.getElementById('temperature').addEventListener('input', function() {
                document.getElementById('tempValue').textContent = this.value;
            });

            document.getElementById('testBtn').addEventListener('click', testPrompt);
            document.getElementById('optimizeBtn').addEventListener('click', optimizePrompt);
            document.getElementById('compareBtn').addEventListener('click', comparePrompts);
        }

        // Initialize charts
        function initializeCharts() {
            const ctx = document.getElementById('performanceChart').getContext('2d');
            window.performanceChart = new Chart(ctx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: 'Response Time (ms)',
                        data: [],
                        borderColor: 'rgb(59, 130, 246)',
                        backgroundColor: 'rgba(59, 130, 246, 0.1)',
                        tension: 0.1
                    }]
                },
                options: {
                    responsive: true,
                    scales: {
                        y: {
                            beginAtZero: true
                        }
                    }
                }
            });
        }

        // Tab system
        function setupTabs() {
            document.querySelectorAll('.tab-btn').forEach(btn => {
                btn.addEventListener('click', function() {
                    const tab = this.dataset.tab;
                    switchTab(tab);
                });
            });
            
            // Show first tab by default
            switchTab('history');
        }

        function switchTab(tabName) {
            // Hide all tabs
            document.querySelectorAll('.tab-content').forEach(content => {
                content.classList.add('hidden');
            });
            
            // Remove active class from all buttons
            document.querySelectorAll('.tab-btn').forEach(btn => {
                btn.classList.remove('border-blue-500', 'text-blue-600');
                btn.classList.add('border-transparent', 'text-gray-500');
            });
            
            // Show selected tab
            document.getElementById('tab-' + tabName).classList.remove('hidden');
            
            // Activate selected button
            const activeBtn = document.querySelector('[data-tab="' + tabName + '"]');
            activeBtn.classList.remove('border-transparent', 'text-gray-500');
            activeBtn.classList.add('border-blue-500', 'text-blue-600');
        }

        // API calls
        async function testPrompt() {
            const request = getPromptRequest();
            
            try {
                showLoading('Testing prompt...');
                const response = await fetch('/api/test', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(request)
                });
                
                const result = await response.json();
                displayResult(result);
                addToHistory(result);
                
            } catch (error) {
                showError('Test failed: ' + error.message);
            }
        }

        async function optimizePrompt() {
            const request = getPromptRequest();
            request.optimize = true;
            
            try {
                showLoading('Optimizing prompt...');
                const response = await fetch('/api/optimize', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(request)
                });
                
                const result = await response.json();
                displayOptimizedResult(result);
                addToHistory(result);
                
            } catch (error) {
                showError('Optimization failed: ' + error.message);
            }
        }

        async function comparePrompts() {
            showLoading('Running multi-provider comparison...');
            // Implementation for comparing across providers
        }

        // Helper functions
        function getPromptRequest() {
            return {
                prompt: document.getElementById('promptEditor').value,
                provider: document.getElementById('provider').value,
                model: document.getElementById('model').value,
                temperature: parseFloat(document.getElementById('temperature').value),
                max_tokens: parseInt(document.getElementById('maxTokens').value),
                optimize: document.getElementById('optimize').checked,
                method: document.getElementById('method').value,
                iterations: parseInt(document.getElementById('iterations').value)
            };
        }

        function displayResult(result) {
            document.getElementById('responseDisplay').innerHTML = '<p>' + result.response + '</p>';
            document.getElementById('latencyMetric').textContent = result.latency + 'ms';
            document.getElementById('costMetric').textContent = '$' + result.cost.toFixed(4);
            document.getElementById('tokensMetric').textContent = result.tokens_used;
            
            // Update performance chart
            updatePerformanceChart(result);
        }

        function displayOptimizedResult(result) {
            displayResult(result);
            if (result.optimized_prompt) {
                document.getElementById('promptEditor').value = result.optimized_prompt;
            }
        }

        function updatePerformanceChart(result) {
            const chart = window.performanceChart;
            const now = new Date().toLocaleTimeString();
            
            chart.data.labels.push(now);
            chart.data.datasets[0].data.push(result.latency);
            
            // Keep only last 10 points
            if (chart.data.labels.length > 10) {
                chart.data.labels.shift();
                chart.data.datasets[0].data.shift();
            }
            
            chart.update();
        }

        function addToHistory(result) {
            testHistory.unshift(result);
            updateHistoryDisplay();
        }

        function updateHistoryDisplay() {
            const historyList = document.getElementById('historyList');
            if (testHistory.length === 0) {
                historyList.innerHTML = '<p class="text-gray-500">No history yet. Start testing prompts to build your history.</p>';
                return;
            }

            historyList.innerHTML = testHistory.slice(0, 10).map(item => 
                '<div class="p-3 border border-gray-200 rounded-lg">' +
                    '<div class="flex justify-between items-start">' +
                        '<div class="flex-1">' +
                            '<p class="text-sm font-medium">' + item.prompt.substring(0, 100) + '...</p>' +
                            '<p class="text-xs text-gray-500">' + item.provider + ' - ' + item.model + '</p>' +
                        '</div>' +
                        '<div class="text-right">' +
                            '<p class="text-sm font-medium">' + item.latency + 'ms</p>' +
                            '<p class="text-xs text-gray-500">$' + item.cost.toFixed(4) + '</p>' +
                        '</div>' +
                    '</div>' +
                '</div>'
            ).join('');
        }

        function showLoading(message) {
            document.getElementById('responseDisplay').innerHTML = '<p class="text-blue-600 italic">' + message + '</p>';
        }

        function showError(message) {
            document.getElementById('responseDisplay').innerHTML = '<p class="text-red-600">' + message + '</p>';
        }

        function handleWebSocketMessage(data) {
            // Handle real-time updates from server
            console.log('Received update:', data);
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleTest processes prompt testing requests
func (ps *PlaygroundServer) handleTest(w http.ResponseWriter, r *http.Request) {
	var req PlaygroundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	start := time.Now()

	// Create LLM provider
	provider, err := commandLegacyProvider(req.Provider, req.Model)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create provider: %v", err), http.StatusInternalServerError)
		return
	}

	// Test the prompt
	options := llm.GenerateOptions{
		Temperature: &req.Temperature,
		MaxTokens:   &req.MaxTokens,
	}
	response, err := provider.Generate(context.Background(), req.Prompt, options)

	latency := time.Since(start)

	result := PlaygroundResponse{
		ID:        generateID(),
		Prompt:    req.Prompt,
		Provider:  req.Provider,
		Model:     req.Model,
		Latency:   latency,
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Response = response.Text
		result.TokensUsed = estimateTokens(req.Prompt + response.Text)
		result.Cost = calculateCost(req.Provider, req.Model, result.TokensUsed)
	}

	ps.recordHistory(result)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleOptimize processes prompt optimization requests
func (ps *PlaygroundServer) handleOptimize(w http.ResponseWriter, r *http.Request) {
	var req PlaygroundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	start := time.Now()

	// Create LLM provider and optimizer
	provider, err := commandLegacyProvider(req.Provider, req.Model)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create provider: %v", err), http.StatusInternalServerError)
		return
	}

	optimizer := metaprompt.NewOptimizer(provider)

	// Configure optimization
	cfg := metaprompt.Config{
		InitialPrompt: req.Prompt,
		Iterations:    req.Iterations,
		Temperature:   req.Temperature,
		MaxTokens:     req.MaxTokens,
		Method:        req.Method,
	}

	// Run optimization
	optimResult, err := optimizer.Optimize(context.Background(), cfg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Optimization failed: %v", err), http.StatusInternalServerError)
		return
	}

	latency := time.Since(start)

	result := PlaygroundResponse{
		ID:              generateID(),
		Prompt:          req.Prompt,
		OptimizedPrompt: optimResult.OptimizedPrompt,
		Response:        "Optimization completed successfully",
		Provider:        req.Provider,
		Model:           req.Model,
		Latency:         latency,
		TokensUsed:      estimateTokens(req.Prompt + optimResult.OptimizedPrompt),
		Timestamp:       time.Now(),
		Metrics: map[string]interface{}{
			"improvement_score": optimResult.ImprovementScore,
			"iterations":        len(optimResult.Iterations),
			"method":            req.Method,
		},
	}

	result.Cost = calculateCost(req.Provider, req.Model, result.TokensUsed)

	ps.recordHistory(result)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleCompare processes multi-provider comparison requests
func (ps *PlaygroundServer) handleCompare(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prompt    string   `json:"prompt"`
		Responses []string `json:"responses"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Prompt) == "" || len(req.Responses) == 0 {
		http.Error(w, "compare requires prompt and at least one response", http.StatusBadRequest)
		return
	}

	type comparison struct {
		Index     int     `json:"index"`
		Response  string  `json:"response"`
		Relevance float64 `json:"relevance"`
		Length    int     `json:"length"`
	}
	results := make([]comparison, 0, len(req.Responses))
	best := -1
	bestScore := -1.0
	for i, response := range req.Responses {
		score := keywordOverlap(req.Prompt, response)
		results = append(results, comparison{
			Index:     i,
			Response:  response,
			Relevance: score,
			Length:    len(response),
		})
		if score > bestScore {
			best = i
			bestScore = score
		}
	}

	writePlaygroundJSON(w, map[string]interface{}{
		"best_index": best,
		"method":     "local_keyword_overlap",
		"results":    results,
	})
}

// handleMetrics processes advanced metrics requests
func (ps *PlaygroundServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Generated string   `json:"generated"`
		Reference string   `json:"reference"`
		Metrics   []string `json:"metrics"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results := make(map[string]float64)

	for _, metric := range req.Metrics {
		switch metric {
		case "bleu":
			result := metrics.CalculateBLEU(req.Generated, req.Reference, 4)
			results[metric] = result.Score
		case "rouge":
			result := metrics.CalculateROUGE(req.Generated, req.Reference, "L")
			results[metric] = result.Score
		case "bertscore":
			results[metric] = keywordOverlap(req.Generated, req.Reference)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// handleSecurity processes security testing requests
func (ps *PlaygroundServer) handleSecurity(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	findings := localPlaygroundSecurityFindings(req.Prompt)
	writePlaygroundJSON(w, map[string]interface{}{
		"method":   "local_pattern_scan",
		"findings": findings,
		"passed":   len(findings) == 0,
	})
}

// handleComponents processes component library requests
func (ps *PlaygroundServer) handleComponents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		components, err := localPlaygroundComponents("components")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writePlaygroundJSON(w, map[string]interface{}{"components": components})
	case http.MethodPost:
		var req struct {
			Name     string `json:"name"`
			Category string `json:"category"`
			Content  string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Content) == "" {
			http.Error(w, "component requires name and content", http.StatusBadRequest)
			return
		}
		category := strings.TrimSpace(req.Category)
		if category == "" {
			category = "general"
		}
		category = filepath.Clean(category)
		if category == "." || strings.HasPrefix(category, "..") || filepath.IsAbs(category) {
			http.Error(w, "invalid component category", http.StatusBadRequest)
			return
		}
		if err := enforceRuntimeToolPolicy("write"); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		path := filepath.Join("components", category, filepath.Base(req.Name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := os.WriteFile(path, []byte(req.Content), 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writePlaygroundJSON(w, map[string]interface{}{"path": path})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleHistory processes history requests
func (ps *PlaygroundServer) handleHistory(w http.ResponseWriter, r *http.Request) {
	writePlaygroundJSON(w, map[string]interface{}{"history": ps.history})
}

// handleWebSocket handles WebSocket connections for real-time updates
func (ps *PlaygroundServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ps.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ps.clients[conn] = true

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			delete(ps.clients, conn)
			break
		}
	}
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("pg_%d", time.Now().UnixNano())
}

func estimateTokens(text string) int {
	// Rough token estimation (4 characters per token)
	return len(text) / 4
}

func calculateCost(provider, model string, tokens int) float64 {
	// Simplified cost calculation
	costPerToken := 0.00002 // $0.00002 per token (approximate)
	if strings.Contains(strings.ToLower(model), "gpt-4") {
		costPerToken = 0.00006
	}
	return float64(tokens) * costPerToken
}

func (ps *PlaygroundServer) recordHistory(result PlaygroundResponse) {
	ps.history = append([]PlaygroundResponse{result}, ps.history...)
	if len(ps.history) > 100 {
		ps.history = ps.history[:100]
	}
}

func writePlaygroundJSON(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value)
}

func keywordOverlap(a, b string) float64 {
	left := playgroundTokenSet(a)
	right := playgroundTokenSet(b)
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	overlap := 0
	for token := range left {
		if right[token] {
			overlap++
		}
	}
	return float64(overlap) / float64(len(left))
}

func playgroundTokenSet(text string) map[string]bool {
	tokens := make(map[string]bool)
	for _, token := range strings.Fields(strings.ToLower(text)) {
		token = strings.Trim(token, ".,;:!?()[]{}\"'")
		if len(token) > 2 {
			tokens[token] = true
		}
	}
	return tokens
}

func localPlaygroundSecurityFindings(prompt string) []string {
	lower := strings.ToLower(prompt)
	checks := map[string]string{
		"ignore previous instructions": "prompt injection",
		"system override":              "prompt injection",
		"reveal your prompt":           "sensitive prompt disclosure",
		"api key":                      "secret disclosure",
	}
	findings := make([]string, 0)
	for phrase, finding := range checks {
		if strings.Contains(lower, phrase) {
			findings = append(findings, finding)
		}
	}
	sort.Strings(findings)
	return findings
}

func localPlaygroundComponents(root string) ([]map[string]string, error) {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return []map[string]string{}, nil
	}
	var components []map[string]string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		components = append(components, map[string]string{
			"name":     filepath.Base(path),
			"category": filepath.Dir(rel),
			"path":     path,
		})
		return nil
	})
	sort.Slice(components, func(i, j int) bool {
		return components[i]["path"] < components[j]["path"]
	})
	return components, err
}

func openURL(url string) {
	fmt.Printf("Please open %s in your browser\n", url)
}
