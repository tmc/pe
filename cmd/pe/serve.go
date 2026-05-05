package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const defaultServeAddr = "127.0.0.1:8080"
const maxServeBodyBytes = 1 << 20

type serveServer struct {
	mux *http.ServeMux
}

type formatRequest struct {
	Prompt string `json:"prompt"`
	Style  string `json:"style,omitempty"`
	Fix    bool   `json:"fix,omitempty"`
}

type formatResponse struct {
	Prompt string `json:"prompt"`
}

type renderRequest struct {
	Prompt    string            `json:"prompt"`
	Variables map[string]string `json:"variables,omitempty"`
}

type renderResponse struct {
	Prompt string `json:"prompt"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func serveCmd() *cobra.Command {
	var addr string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve a local PE HTTP API",
		Long: `Serve a small localhost-first HTTP API.

The server binds to 127.0.0.1:8080 by default. Use --addr explicitly to bind
somewhere else.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateServeAddr(addr); err != nil {
				return err
			}
			srv := newServeServer()
			httpServer := &http.Server{
				Addr:              addr,
				Handler:           srv,
				ReadHeaderTimeout: 5 * time.Second,
				ReadTimeout:       10 * time.Second,
				WriteTimeout:      10 * time.Second,
				IdleTimeout:       30 * time.Second,
			}
			errc := make(chan error, 1)
			go func() {
				<-cmd.Context().Done()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				errc <- httpServer.Shutdown(ctx)
			}()

			fmt.Fprintf(cmd.OutOrStdout(), "serving PE API on http://%s\n", addr)
			err := httpServer.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				err = <-errc
			}
			return err
		},
	}

	cmd.Flags().StringVar(&addr, "addr", defaultServeAddr, "listen address")
	return cmd
}

func validateServeAddr(addr string) error {
	if addr == "" {
		return fmt.Errorf("addr is required")
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid addr %q: %w", addr, err)
	}
	if port == "" {
		return fmt.Errorf("invalid addr %q: missing port", addr)
	}
	if host == "" {
		return fmt.Errorf("invalid addr %q: host is required; use 127.0.0.1:%s for localhost", addr, port)
	}
	return nil
}

func newServeServer() *serveServer {
	s := &serveServer{mux: http.NewServeMux()}
	s.mux.HandleFunc("/healthz", s.handleHealthz)
	s.mux.HandleFunc("/api/v1/format", s.handleFormat)
	s.mux.HandleFunc("/api/v1/render", s.handleRender)
	return s
}

func (s *serveServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/healthz" && r.URL.Path != "/api/v1/format" && r.URL.Path != "/api/v1/render" {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *serveServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *serveServer) handleFormat(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	defer r.Body.Close()

	var req formatRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "prompt is required"})
		return
	}
	style := req.Style
	if style == "" {
		style = "standard"
	}
	switch style {
	case "standard", "anthropic", "openai":
	default:
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid style"})
		return
	}

	writeJSON(w, http.StatusOK, formatResponse{
		Prompt: formatPromptContent(req.Prompt, style, req.Fix),
	})
}

func (s *serveServer) handleRender(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	defer r.Body.Close()

	var req renderRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "prompt is required"})
		return
	}
	writeJSON(w, http.StatusOK, renderResponse{
		Prompt: substituteVariables(req.Prompt, req.Variables),
	})
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	return false
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxServeBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeJSON(w, http.StatusRequestEntityTooLarge, errorResponse{Error: "request body too large"})
		} else {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		}
		return false
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
