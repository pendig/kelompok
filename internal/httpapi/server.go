package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pendig/kelompok/internal/audit"
	"github.com/pendig/kelompok/internal/auth"
	"github.com/pendig/kelompok/internal/config"
	"github.com/pendig/kelompok/internal/database"
	"github.com/pendig/kelompok/internal/impact"
	"github.com/pendig/kelompok/internal/members"
	"github.com/pendig/kelompok/internal/organizations"
	"github.com/pendig/kelompok/internal/posts"
)

type Server struct {
	config        config.Config
	db            *pgxpool.Pool
	mux           *http.ServeMux
	audit         *audit.Repository
	auth          *auth.Repository
	organizations *organizations.Repository
	posts         *posts.Repository
	impact        *impact.Repository
	members       *members.Repository
}

func New(config config.Config, db *pgxpool.Pool) *Server {
	server := &Server{
		config:        config,
		db:            db,
		mux:           http.NewServeMux(),
		audit:         audit.NewRepository(db),
		auth:          auth.NewRepository(db),
		organizations: organizations.NewRepository(db),
		posts:         posts.NewRepository(db),
		impact:        impact.NewRepository(db),
		members:       members.NewRepository(db),
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
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /", s.handleRoot)
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /readyz", s.handleReady)
	s.mux.HandleFunc("GET /api/v1/organizations", s.handleListOrganizations)
	s.mux.HandleFunc("GET /api/v1/organizations/{slug}", s.handleGetOrganization)
	s.mux.HandleFunc("POST /api/v1/organizations/{slug}/claims", s.handleCreateOrganizationClaim)
	s.mux.HandleFunc("GET /api/v1/organizations/{slug}/posts", s.handleListOrganizationPosts)
	s.mux.HandleFunc("GET /api/v1/organizations/{slug}/posts/{post_slug}", s.handleGetOrganizationPost)
	s.mux.HandleFunc("GET /api/v1/organizations/{slug}/impact-reports", s.handleListOrganizationImpactReports)
	s.mux.HandleFunc("GET /api/v1/posts", s.handleListPosts)
	s.mux.HandleFunc("GET /api/v1/posts/{slug}", s.handleGetPost)

	s.mux.HandleFunc("POST /api/v1/auth/register", s.handleRegister)
	s.mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	s.mux.HandleFunc("POST /api/v1/auth/logout", s.requireSession(s.handleLogout))
	s.mux.HandleFunc("GET /api/v1/auth/me", s.requireSession(s.handleMe))
	s.mux.HandleFunc("PATCH /api/v1/auth/me", s.requireSession(s.handleUpdateMe))

	s.mux.HandleFunc("GET /api/v1/org-admin/organizations", s.requireAdmin(s.handleListAdminOrganizations))
	s.mux.HandleFunc("POST /api/v1/org-admin/organizations", s.requireAdmin(s.handleCreateAdminOrganization))
	s.mux.HandleFunc("GET /api/v1/org-admin/organizations/{slug}", s.requireAdmin(s.handleGetAdminOrganization))
	s.mux.HandleFunc("PATCH /api/v1/org-admin/organizations/{slug}", s.requireAdmin(s.handleUpdateAdminOrganization))
	s.mux.HandleFunc("GET /api/v1/org-admin/organizations/{slug}/relationships", s.requireAdmin(s.handleListOrganizationRelationships))
	s.mux.HandleFunc("POST /api/v1/org-admin/organization-relationships", s.requireAdmin(s.handleCreateOrganizationRelationship))
	s.mux.HandleFunc("PATCH /api/v1/org-admin/organization-relationships/{id}", s.requireAdmin(s.handleUpdateOrganizationRelationship))
	s.mux.HandleFunc("DELETE /api/v1/org-admin/organization-relationships/{id}", s.requireAdmin(s.handleDeleteOrganizationRelationship))
	s.mux.HandleFunc("GET /api/v1/org-admin/organizations/{slug}/claims", s.requireAdmin(s.handleListOrganizationClaims))
	s.mux.HandleFunc("POST /api/v1/org-admin/claims/{id}/approve", s.requireAdmin(s.handleApproveOrganizationClaim))
	s.mux.HandleFunc("POST /api/v1/org-admin/claims/{id}/reject", s.requireAdmin(s.handleRejectOrganizationClaim))
	s.mux.HandleFunc("GET /api/v1/org-admin/organizations/{slug}/audit-logs", s.requireAdmin(s.handleListOrganizationAuditLogs))
	s.mux.HandleFunc("GET /api/v1/org-admin/organizations/{slug}/members", s.requireAdmin(s.handleListOrganizationMembers))
	s.mux.HandleFunc("POST /api/v1/org-admin/organizations/{slug}/members", s.requireAdmin(s.handleCreateOrganizationMember))
	s.mux.HandleFunc("PATCH /api/v1/org-admin/members/{id}", s.requireAdmin(s.handleUpdateAdminMember))
	s.mux.HandleFunc("DELETE /api/v1/org-admin/members/{id}", s.requireAdmin(s.handleDeleteAdminMember))
	s.mux.HandleFunc("GET /api/v1/org-admin/posts", s.requireAdmin(s.handleListAdminPosts))
	s.mux.HandleFunc("POST /api/v1/org-admin/posts", s.requireAdmin(s.handleCreateAdminPost))
	s.mux.HandleFunc("PATCH /api/v1/org-admin/posts/{id}", s.requireAdmin(s.handleUpdateAdminPost))
	s.mux.HandleFunc("POST /api/v1/org-admin/posts/{id}/publish", s.requireAdmin(s.handlePublishAdminPost))
	s.mux.HandleFunc("POST /api/v1/org-admin/posts/{id}/archive", s.requireAdmin(s.handleArchiveAdminPost))
	s.mux.HandleFunc("GET /api/v1/org-admin/impact-reports", s.requireAdmin(s.handleListAdminImpactReports))
	s.mux.HandleFunc("POST /api/v1/org-admin/impact-reports", s.requireAdmin(s.handleCreateAdminImpactReport))
	s.mux.HandleFunc("PATCH /api/v1/org-admin/impact-reports/{id}", s.requireAdmin(s.handleUpdateAdminImpactReport))
	s.mux.HandleFunc("POST /api/v1/org-admin/impact-reports/{id}/publish", s.requireAdmin(s.handlePublishAdminImpactReport))
	s.mux.HandleFunc("POST /api/v1/org-admin/impact-reports/{id}/archive", s.requireAdmin(s.handleArchiveAdminImpactReport))
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

func (s *Server) handleListOrganizations(w http.ResponseWriter, r *http.Request) {
	limit := limitFromRequest(r, 50, 100)
	items, err := s.organizations.ListPublic(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "organizations_list_failed", "Failed to list organizations", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: publicOrganizations(items),
		Meta: map[string]any{
			"count": len(items),
			"limit": limit,
		},
		Message: "ok",
	})
}

func (s *Server) handleGetOrganization(w http.ResponseWriter, r *http.Request) {
	item, err := s.organizations.FindBySlug(r.Context(), r.PathValue("slug"))
	if errors.Is(err, organizations.ErrNotFound) {
		writeError(w, http.StatusNotFound, "organization_not_found", "Organization not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "organization_lookup_failed", "Failed to load organization", nil)
		return
	}

	relationships, err := s.organizations.ListActiveRelationshipsByOrganizationSlug(r.Context(), item.Slug, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "organization_relationships_failed", "Failed to load organization relationships", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data:    publicOrganizationWithRelationships(item, relationships),
		Message: "ok",
	})
}

func (s *Server) handleListOrganizationPosts(w http.ResponseWriter, r *http.Request) {
	if !s.ensureOrganization(w, r, r.PathValue("slug")) {
		return
	}

	limit := limitFromRequest(r, 20, 100)
	items, err := s.posts.ListByOrganizationSlug(r.Context(), r.PathValue("slug"), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "organization_posts_failed", "Failed to list organization posts", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: publicPosts(items),
		Meta: map[string]any{
			"count": len(items),
			"limit": limit,
		},
		Message: "ok",
	})
}

func (s *Server) handleGetOrganizationPost(w http.ResponseWriter, r *http.Request) {
	if !s.ensureOrganization(w, r, r.PathValue("slug")) {
		return
	}

	item, err := s.posts.FindPublishedByOrganizationAndSlug(r.Context(), r.PathValue("slug"), r.PathValue("post_slug"))
	if errors.Is(err, posts.ErrNotFound) {
		writeError(w, http.StatusNotFound, "post_not_found", "Post not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "post_lookup_failed", "Failed to load post", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data:    publicPost(item),
		Message: "ok",
	})
}

func (s *Server) handleListOrganizationImpactReports(w http.ResponseWriter, r *http.Request) {
	if !s.ensureOrganization(w, r, r.PathValue("slug")) {
		return
	}

	limit := limitFromRequest(r, 20, 100)
	items, err := s.impact.ListByOrganizationSlug(r.Context(), r.PathValue("slug"), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "organization_impact_reports_failed", "Failed to list organization impact reports", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: publicImpactReports(items),
		Meta: map[string]any{
			"count": len(items),
			"limit": limit,
		},
		Message: "ok",
	})
}

func (s *Server) handleListPosts(w http.ResponseWriter, r *http.Request) {
	limit := limitFromRequest(r, 20, 100)
	items, err := s.posts.ListPublic(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "posts_list_failed", "Failed to list posts", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: publicPosts(items),
		Meta: map[string]any{
			"count": len(items),
			"limit": limit,
		},
		Message: "ok",
	})
}

func (s *Server) handleGetPost(w http.ResponseWriter, r *http.Request) {
	item, err := s.posts.FindPublishedBySlug(r.Context(), r.PathValue("slug"))
	if errors.Is(err, posts.ErrNotFound) {
		writeError(w, http.StatusNotFound, "post_not_found", "Post not found", nil)
		return
	}
	if errors.Is(err, posts.ErrAmbiguous) {
		writeError(w, http.StatusConflict, "post_slug_ambiguous", "Post slug is used by more than one organization; use the organization-scoped post endpoint", map[string]string{
			"endpoint": "/api/v1/organizations/{slug}/posts/{post_slug}",
		})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "post_lookup_failed", "Failed to load post", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data:    publicPost(item),
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
		writeError(w, http.StatusServiceUnavailable, "database_not_ready", "Database is not ready", nil)
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

func limitFromRequest(r *http.Request, fallback, max int) int {
	value := r.URL.Query().Get("limit")
	if value == "" {
		return fallback
	}

	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 {
		return fallback
	}
	if limit > max {
		return max
	}

	return limit
}

func (s *Server) ensureOrganization(w http.ResponseWriter, r *http.Request, slug string) bool {
	if _, err := s.organizations.FindBySlug(r.Context(), slug); errors.Is(err, organizations.ErrNotFound) {
		writeError(w, http.StatusNotFound, "organization_not_found", "Organization not found", nil)
		return false
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "organization_lookup_failed", "Failed to load organization", nil)
		return false
	}

	return true
}
