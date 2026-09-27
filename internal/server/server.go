package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/example/healthcheckmonitor/internal/logger"
	"github.com/example/healthcheckmonitor/internal/monitor"
	"github.com/example/healthcheckmonitor/internal/types"
)

type Server struct {
	httpServer *http.Server
	monitor    *monitor.Monitor
}

// New creates an HTTP server bound to the given address.
func New(addr string, mon *monitor.Monitor) *Server {
	s := &Server{
		monitor: mon,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleAll)
	mux.HandleFunc("/health/", s.handleOne) // expects /health/{name}
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return s
}

// Start runs the HTTP server in a separate goroutine.
func (s *Server) Start() {
	go func() {
		logger.Log.Infof("starting HTTP server on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalf("HTTP server error: %v", err)
		}
	}()
}

// Shutdown gracefully stops the server with a timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Log.Info("shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

// handleAll returns JSON array of all service statuses.
func (s *Server) handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	statuses := s.monitor.GetAll()
	s.respondJSON(w, statuses, http.StatusOK)
}

// handleOne returns JSON for a single service identified by the path element.
func (s *Server) handleOne(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Expected pattern: /health/{name}
	parts := splitPath(r.URL.Path)
	if len(parts) != 2 {
		http.Error(w, "service name missing", http.StatusBadRequest)
		return
	}
	name := parts[1]
	status, ok := s.monitor.Get(name)
	if !ok {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	s.respondJSON(w, status, http.StatusOK)
}

// splitPath trims leading/trailing slashes and splits.
func splitPath(p string) []string {
	clean := p
	if len(clean) > 0 && clean[0] == '/' {
		clean = clean[1:]
	}
	if len(clean) > 0 && clean[len(clean)-1] == '/' {
		clean = clean[:len(clean)-1]
	}
	if clean == "" {
		return []string{}
	}
	return split(clean, "/")
}

// split is a small helper to avoid importing strings for a single use.
func split(s, sep string) []string {
	var parts []string
	i := 0
	for i < len(s) {
		j := i
		for j < len(s) && s[j:j+len(sep)] != sep {
			j++
			if j+len(sep) > len(s) {
				break
			}
		}
		parts = append(parts, s[i:j])
		i = j + len(sep)
	}
	return parts
}

// respondJSON writes the payload as JSON with proper headers.
func (s *Server) respondJSON(w http.ResponseWriter, payload interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		logger.Log.Errorf("failed to write JSON response: %v", err)
	}
}

// Run starts the server and blocks until an OS interrupt is received.
func (s *Server) Run() {
	s.Start()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		logger.Log.Errorf("server shutdown error: %v", err)
	}
}