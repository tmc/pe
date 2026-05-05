package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestServeCmdShutdownOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := serveCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetContext(ctx)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--addr", "127.0.0.1:0"})

	done := make(chan error, 1)
	go func() {
		done <- cmd.Execute()
	}()

	waitServeAddr(t, &stdout, &stderr, done)

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down after context cancel")
	}
}

func TestServeCmdOperatesLocalAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := serveCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetContext(ctx)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--addr", "127.0.0.1:0"})

	done := make(chan error, 1)
	go func() {
		done <- cmd.Execute()
	}()

	addr := waitServeAddr(t, &stdout, &stderr, done)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get("http://" + addr + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	assertJSONStatus(t, resp, http.StatusOK, `"status":"ok"`)

	resp, err = client.Post("http://"+addr+"/api/v1/format", "application/json", strings.NewReader(`{"prompt":"hello","style":"anthropic"}`))
	if err != nil {
		t.Fatalf("POST /api/v1/format: %v", err)
	}
	assertJSONStatus(t, resp, http.StatusOK, `"prompt":"Human: hello\n"`)

	resp, err = client.Post("http://"+addr+"/api/v1/render", "application/json", strings.NewReader(`{"prompt":"Hello {{.name}}","variables":{"name":"PE"}}`))
	if err != nil {
		t.Fatalf("POST /api/v1/render: %v", err)
	}
	assertJSONStatus(t, resp, http.StatusOK, `"prompt":"Hello PE"`)

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down after context cancel")
	}
}

func waitServeAddr(t *testing.T, stdout, stderr *bytes.Buffer, done <-chan error) string {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		line := strings.TrimSpace(stdout.String())
		if strings.HasPrefix(line, "serving PE API on http://") {
			return strings.TrimPrefix(line, "serving PE API on http://")
		}
		select {
		case err := <-done:
			t.Fatalf("Execute returned before serving: %v, stderr=%q", err, stderr.String())
		case <-deadline:
			t.Fatalf("server did not start, stdout=%q stderr=%q", stdout.String(), stderr.String())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func assertJSONStatus(t *testing.T, resp *http.Response, status int, contains string) {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if resp.StatusCode != status {
		t.Fatalf("status = %d, want %d: %s", resp.StatusCode, status, string(body))
	}
	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}
	if !strings.Contains(string(body), contains) {
		t.Fatalf("body = %q, want %q", string(body), contains)
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

func TestServeMethodRejection(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		allow  string
	}{
		{name: "healthz post", method: http.MethodPost, path: "/healthz", allow: http.MethodGet},
		{name: "format get", method: http.MethodGet, path: "/api/v1/format", allow: http.MethodPost},
		{name: "render get", method: http.MethodGet, path: "/api/v1/render", allow: http.MethodPost},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			newServeServer().ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
			}
			if got := rec.Header().Get("Allow"); got != tt.allow {
				t.Fatalf("Allow = %q, want %q", got, tt.allow)
			}
			if got := rec.Body.String(); !strings.Contains(got, `"error":"method not allowed"`) {
				t.Fatalf("body = %q, want structured method error", got)
			}
		})
	}
}

func TestServeNotFoundIsJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	newServeServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"error":"not found"`) {
		t.Fatalf("body = %q, want structured not found", got)
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

func TestServeFormatOversizedBody(t *testing.T) {
	body := io.MultiReader(
		strings.NewReader(`{"prompt":"`),
		strings.NewReader(strings.Repeat("x", maxServeBodyBytes)),
		strings.NewReader(`"}`),
	)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/format", body)
	rec := httptest.NewRecorder()

	newServeServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
	if got := rec.Body.String(); !strings.Contains(got, `"error":"request body too large"`) {
		t.Fatalf("body = %q, want size error", got)
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

func TestServeRender(t *testing.T) {
	body := bytes.NewBufferString(`{"prompt":"Hello {{.name}}","variables":{"name":"PE"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/render", body)
	rec := httptest.NewRecorder()

	newServeServer().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got renderResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.Prompt != "Hello PE" {
		t.Fatalf("prompt = %q, want rendered prompt", got.Prompt)
	}
}

func TestServeRenderBadRequests(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed json", body: `{`},
		{name: "empty prompt", body: `{"prompt":"   "}`},
		{name: "unknown field", body: `{"prompt":"hello","extra":true}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/render", strings.NewReader(tt.body))
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
