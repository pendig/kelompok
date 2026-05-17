package httpapi

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/pendig/kelompok/internal/auth"
)

type authContextKey string

const principalContextKey authContextKey = "admin_principal"

type principal struct {
	User     auth.User
	Token    string
	AdminKey bool
}

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item, ok, err := s.adminPrincipal(r)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "admin_auth_failed", "Failed to validate admin authorization", nil)
			return
		}
		if !ok && strings.TrimSpace(s.config.AdminAPIKey) == "" {
			writeError(w, http.StatusServiceUnavailable, "admin_auth_not_configured", "Admin API key is not configured", nil)
			return
		}
		if !ok {
			writeError(w, http.StatusUnauthorized, "admin_auth_required", "Admin authorization is required", nil)
			return
		}

		next(w, r.WithContext(context.WithValue(r.Context(), principalContextKey, item)))
	}
}

func (s *Server) requireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			writeError(w, http.StatusUnauthorized, "session_required", "A user session is required", nil)
			return
		}

		user, err := s.auth.UserBySessionToken(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "session_invalid", "Session is invalid or expired", nil)
			return
		}

		item := principal{User: user, Token: token}
		next(w, r.WithContext(context.WithValue(r.Context(), principalContextKey, item)))
	}
}

func (s *Server) adminPrincipal(r *http.Request) (principal, bool, error) {
	if s.validAdminKey(r) {
		return principal{AdminKey: true}, true, nil
	}

	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		return principal{}, false, nil
	}
	if s.auth == nil || s.db == nil {
		return principal{}, false, nil
	}

	user, err := s.auth.UserBySessionToken(r.Context(), token)
	if err != nil {
		return principal{}, false, nil
	}
	return principal{User: user, Token: token}, true, nil
}

func (s *Server) validAdminKey(r *http.Request) bool {
	provided := strings.TrimSpace(r.Header.Get("X-Kelompok-Admin-Key"))
	if provided == "" {
		provided = bearerToken(r.Header.Get("Authorization"))
	}
	if provided == "" {
		return false
	}

	expected := strings.TrimSpace(s.config.AdminAPIKey)
	if len(provided) != len(expected) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func bearerToken(header string) string {
	prefix, token, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(prefix, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

func (s *Server) adminScopeConfigured() bool {
	return len(s.config.AdminOrganizationSlugs) > 0
}

func (s *Server) authorizedAdminOrganizationSlug(slug string) bool {
	if len(s.config.AdminOrganizationSlugs) == 0 {
		return true
	}

	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug == "" {
		return false
	}

	for _, allowed := range s.config.AdminOrganizationSlugs {
		if slug == allowed {
			return true
		}
	}

	return false
}

func (s *Server) ensureAdminOrganizationSlug(w http.ResponseWriter, slug string) bool {
	item, _ := principalFromContext(nil)
	return s.ensureAdminOrganizationSlugWithPrincipal(w, nil, item, slug)
}

func (s *Server) ensureAdminOrganizationSlugForRequest(w http.ResponseWriter, r *http.Request, slug string) bool {
	item, ok := principalFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "admin_auth_required", "Admin authorization is required", nil)
		return false
	}
	return s.ensureAdminOrganizationSlugWithPrincipal(w, r, item, slug)
}

func (s *Server) ensureAdminAnyOrganizationSlugForRequest(w http.ResponseWriter, r *http.Request, slugs ...string) bool {
	item, ok := principalFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "admin_auth_required", "Admin authorization is required", nil)
		return false
	}
	if !item.AdminKey && item.User.ID != "" {
		if item.User.Role == "superadmin" {
			return true
		}
		for _, slug := range slugs {
			if strings.TrimSpace(slug) == "" {
				continue
			}
			allowed, err := s.auth.CanManageOrganization(r.Context(), item.User, slug)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "admin_org_scope_failed", "Failed to check organization access", nil)
				return false
			}
			if allowed {
				return true
			}
		}
		writeError(w, http.StatusForbidden, "admin_org_forbidden", "User is not authorized for these organizations", nil)
		return false
	}

	for _, slug := range slugs {
		if s.authorizedAdminOrganizationSlug(slug) {
			return true
		}
	}

	writeError(w, http.StatusForbidden, "admin_org_forbidden", "Admin key is not authorized for these organizations", nil)
	return false
}

func (s *Server) ensureAdminOrganizationSlugWithPrincipal(w http.ResponseWriter, r *http.Request, item principal, slug string) bool {
	if !item.AdminKey && item.User.ID != "" {
		if item.User.Role == "superadmin" {
			return true
		}
		if strings.TrimSpace(slug) == "" {
			writeError(w, http.StatusForbidden, "admin_org_scope_required", "User sessions must be scoped to an organization", nil)
			return false
		}
		allowed, err := s.auth.CanManageOrganization(r.Context(), item.User, slug)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "admin_org_scope_failed", "Failed to check organization access", nil)
			return false
		}
		if allowed {
			return true
		}
		writeError(w, http.StatusForbidden, "admin_org_forbidden", "User is not authorized for this organization", nil)
		return false
	}

	if s.authorizedAdminOrganizationSlug(slug) {
		return true
	}

	writeError(w, http.StatusForbidden, "admin_org_forbidden", "Admin key is not authorized for this organization", nil)
	return false
}

func principalFromContext(r *http.Request) (principal, bool) {
	if r == nil {
		return principal{}, false
	}
	item, ok := r.Context().Value(principalContextKey).(principal)
	return item, ok
}
