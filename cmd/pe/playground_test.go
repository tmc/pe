package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPlaygroundCmd_FlagParsing(t *testing.T) {
	cmd := playgroundCmd()

	flags := []string{"port", "host", "open"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestPlaygroundCmd_CommandStructure(t *testing.T) {
	cmd := playgroundCmd()

	if cmd.Use != "playground" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

func TestPlaygroundCmd_FlagDefaults(t *testing.T) {
	cmd := playgroundCmd()

	// Check port default
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		t.Errorf("Failed to get port flag: %v", err)
	}
	if port != 8080 {
		t.Errorf("Expected default port 8080, got %d", port)
	}

	// Check host default
	host, err := cmd.Flags().GetString("host")
	if err != nil {
		t.Errorf("Failed to get host flag: %v", err)
	}
	if host != "localhost" {
		t.Errorf("Expected default host 'localhost', got %s", host)
	}
}

func TestNewPlaygroundServer(t *testing.T) {
	server := NewPlaygroundServer()
	if server == nil {
		t.Fatal("NewPlaygroundServer returned nil")
	}

	if server.clients == nil {
		t.Error("Expected clients map to be initialized")
	}

	if server.router == nil {
		t.Error("Expected router to be initialized")
	}
}

func TestUnimplementedEndpointsFailClosed(t *testing.T) {
	server := NewPlaygroundServer()
	server.setupRoutes()

	tests := []struct {
		name     string
		method   string
		path     string
		body     string
		wantBody string
	}{
		{name: "compare", method: http.MethodPost, path: "/api/compare", wantBody: "comparison is not yet implemented"},
		{name: "security", method: http.MethodPost, path: "/api/security", wantBody: "security testing is not yet implemented"},
		{name: "components", method: http.MethodGet, path: "/api/components", wantBody: "component library is not yet implemented"},
		{name: "history", method: http.MethodGet, path: "/api/history", wantBody: "prompt history is not yet implemented"},
		{name: "bertscore", method: http.MethodPost, path: "/api/metrics", body: `{"generated":"a","reference":"b","metrics":["bertscore"]}`, wantBody: "bertscore is not yet implemented"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			server.router.ServeHTTP(rec, req)
			if rec.Code != http.StatusNotImplemented {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestPlaygroundRequestStruct(t *testing.T) {
	req := PlaygroundRequest{
		Prompt:      "Test prompt",
		Provider:    "openai",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   1024,
		Variables:   map[string]string{"key": "value"},
		Optimize:    true,
		Method:      "pe2",
		Iterations:  5,
	}

	if req.Prompt != "Test prompt" {
		t.Error("PlaygroundRequest.Prompt mismatch")
	}
	if req.Provider != "openai" {
		t.Error("PlaygroundRequest.Provider mismatch")
	}
	if req.Optimize != true {
		t.Error("PlaygroundRequest.Optimize mismatch")
	}
}

func TestPlaygroundResponseStruct(t *testing.T) {
	resp := PlaygroundResponse{
		ID:              "test-123",
		Prompt:          "Test",
		Response:        "Response",
		Latency:         100,
		TokensUsed:      50,
		Cost:            0.001,
		Provider:        "openai",
		Model:           "gpt-4",
		OptimizedPrompt: "Optimized test",
		Metrics:         map[string]interface{}{"score": 0.95},
	}

	if resp.ID != "test-123" {
		t.Error("PlaygroundResponse.ID mismatch")
	}
	if resp.TokensUsed != 50 {
		t.Error("PlaygroundResponse.TokensUsed mismatch")
	}
	if resp.OptimizedPrompt != "Optimized test" {
		t.Error("PlaygroundResponse.OptimizedPrompt mismatch")
	}
}
