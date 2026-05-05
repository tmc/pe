package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeCmdDefaultsToLocalhost(t *testing.T) {
	cmd := serveCmd()
	flag := cmd.Flags().Lookup("addr")
	if flag == nil {
		t.Fatal("addr flag not found")
	}
	if flag.DefValue != defaultServeAddr {
		t.Fatalf("addr default = %q, want %q", flag.DefValue, defaultServeAddr)
	}
}

func TestValidateServeAddr(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{name: "default localhost", addr: "127.0.0.1:8080"},
		{name: "explicit localhost", addr: "localhost:8080"},
		{name: "explicit external", addr: "0.0.0.0:8080"},
		{name: "empty", addr: "", wantErr: true},
		{name: "missing host", addr: ":8080", wantErr: true},
		{name: "missing port", addr: "127.0.0.1", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServeAddr(tt.addr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateServeAddr(%q) error = %v, wantErr %v", tt.addr, err, tt.wantErr)
			}
		})
	}
}

func TestServeHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	newServeServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"status":"ok"`) {
		t.Fatalf("body = %q, want status ok", got)
	}
}

func TestServeFormat(t *testing.T) {
	body := bytes.NewBufferString(`{"prompt":"hello","style":"anthropic"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/format", body)
	rec := httptest.NewRecorder()

	newServeServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got formatResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.Prompt != "Human: hello\n" {
		t.Fatalf("prompt = %q, want anthropic formatting", got.Prompt)
	}
}

func TestServeFormatBadRequests(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed json", body: `{`},
		{name: "empty prompt", body: `{"prompt":"   "}`},
		{name: "bad style", body: `{"prompt":"hello","style":"bogus"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/format", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			newServeServer().ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
			if got := rec.Body.String(); !strings.Contains(got, `"error"`) {
				t.Fatalf("body = %q, want error response", got)
			}
		})
	}
}
