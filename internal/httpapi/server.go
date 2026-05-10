package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pendig/kelompok/internal/config"
	"github.com/pendig/kelompok/internal/database"
)

type Server struct {
	config config.Config
	db     *pgxpool.Pool
	mux    *http.ServeMux
}

func New(config config.Config, db *pgxpool.Pool) *Server {
	server := &Server{
		config: config,
		db:     db,
		mux:    http.NewServeMux(),
	}
	server.routes()
	return server
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:              s.config.APIAddr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /", s.handleRoot)
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /readyz", s.handleReady)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{
		Data: map[string]string{
			"service": "kelompok-api",
			"tagline": "The Solutions of Movement",
		},
		Message: "ok",
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{
		Data: map[string]string{
			"status": "ok",
			"env":    s.config.Env,
		},
		Message: "ok",
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := database.Ping(ctx, s.db); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database_not_ready", "Database is not ready", map[string]string{
			"reason": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: map[string]string{
			"status": "ready",
		},
		Message: "ok",
	})
}

type response struct {
	Data    any    `json:"data"`
	Meta    any    `json:"meta,omitempty"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, message string, details any) {
	writeJSON(w, status, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
