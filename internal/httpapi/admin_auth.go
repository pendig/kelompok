package httpapi

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(s.config.AdminAPIKey) == "" {
			writeError(w, http.StatusServiceUnavailable, "admin_auth_not_configured", "Admin API key is not configured", nil)
			return
		}

		if !s.validAdminKey(r) {
			writeError(w, http.StatusUnauthorized, "admin_auth_required", "Admin authorization is required", nil)
			return
		}

		if !s.authorizedAdminOrganization(r) {
			writeError(w, http.StatusForbidden, "admin_org_forbidden", "Admin key is not authorized for this organization", nil)
			return
		}

		next(w, r)
	}
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

func (s *Server) authorizedAdminOrganization(r *http.Request) bool {
	if len(s.config.AdminOrganizationSlugs) == 0 {
		return true
	}

	slug := strings.ToLower(strings.TrimSpace(r.PathValue("slug")))
	if slug == "" {
		slug = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("organization_slug")))
	}
	if slug == "" {
		return true
	}

	for _, allowed := range s.config.AdminOrganizationSlugs {
		if slug == allowed {
			return true
		}
	}

	return false
}
