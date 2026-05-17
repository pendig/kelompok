package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pendig/kelompok/internal/auth"
	"github.com/pendig/kelompok/internal/config"
)

func TestOrgAdminRequiresConfiguredAdminKey(t *testing.T) {
	server := New(config.Config{APIAddr: ":0"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/org-admin/organizations", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "admin_auth_not_configured") {
		t.Fatalf("missing stable auth configuration error: %s", recorder.Body.String())
	}
}

func TestOrgAdminRejectsMissingAdminKey(t *testing.T) {
	server := New(config.Config{APIAddr: ":0", AdminAPIKey: "test-secret"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/org-admin/organizations", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "admin_auth_required") {
		t.Fatalf("missing stable auth error: %s", recorder.Body.String())
	}
}

func TestOrgAdminAcceptsBearerAdminKey(t *testing.T) {
	server := New(config.Config{APIAddr: ":0", AdminAPIKey: "test-secret"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/org-admin/organizations", nil)
	request.Header.Set("Authorization", "Bearer test-secret")

	if !server.validAdminKey(request) {
		t.Fatal("bearer admin key was rejected")
	}
}

func TestOrgAdminOrganizationScopeRejectsUnlistedSlug(t *testing.T) {
	server := New(config.Config{
		APIAddr:                ":0",
		AdminAPIKey:            "test-secret",
		AdminOrganizationSlugs: []string{"allowed-org"},
	}, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/org-admin/organizations/other-org", nil)
	request.Header.Set("X-Kelompok-Admin-Key", "test-secret")
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "admin_org_forbidden") {
		t.Fatalf("missing stable organization scope error: %s", recorder.Body.String())
	}
}

func TestOrgAdminOrganizationScopeRejectsGlobalRoutes(t *testing.T) {
	server := New(config.Config{
		APIAddr:                ":0",
		AdminAPIKey:            "test-secret",
		AdminOrganizationSlugs: []string{"allowed-org"},
	}, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/org-admin/posts", nil)
	request.Header.Set("X-Kelompok-Admin-Key", "test-secret")
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

func TestOrgAdminListScopeAllowsSuperadminSessionWithScopedAdminKey(t *testing.T) {
	server := New(config.Config{
		APIAddr:                ":0",
		AdminAPIKey:            "test-secret",
		AdminOrganizationSlugs: []string{"allowed-org"},
	}, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/org-admin/posts", nil)
	request = request.WithContext(context.WithValue(request.Context(), principalContextKey, principal{
		User: auth.User{ID: "user-1", Role: "superadmin"},
	}))
	recorder := httptest.NewRecorder()

	if !server.ensureAdminListScope(recorder, request, "") {
		t.Fatalf("expected superadmin session to bypass scoped admin-key global list restriction: %s", recorder.Body.String())
	}
}

func TestOrgAdminAnyOrganizationScopeAllowsListedRelationshipSide(t *testing.T) {
	server := New(config.Config{
		APIAddr:                ":0",
		AdminAPIKey:            "test-secret",
		AdminOrganizationSlugs: []string{"allowed-org"},
	}, nil)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/org-admin/organization-relationships", nil)
	request = request.WithContext(context.WithValue(request.Context(), principalContextKey, principal{AdminKey: true}))
	recorder := httptest.NewRecorder()

	if !server.ensureAdminAnyOrganizationSlugForRequest(recorder, request, " other-org ", " allowed-org ") {
		t.Fatalf("expected scoped admin key to manage a relationship with one allowed side: %s", recorder.Body.String())
	}
}

func TestOrgAdminAnyOrganizationScopeRejectsUnlistedRelationshipSides(t *testing.T) {
	server := New(config.Config{
		APIAddr:                ":0",
		AdminAPIKey:            "test-secret",
		AdminOrganizationSlugs: []string{"allowed-org"},
	}, nil)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/org-admin/organization-relationships", nil)
	request = request.WithContext(context.WithValue(request.Context(), principalContextKey, principal{AdminKey: true}))
	recorder := httptest.NewRecorder()

	if server.ensureAdminAnyOrganizationSlugForRequest(recorder, request, "other-org", "another-org") {
		t.Fatal("expected scoped admin key to reject relationships without an allowed side")
	}
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "admin_org_forbidden") {
		t.Fatalf("missing stable relationship scope error: %s", recorder.Body.String())
	}
}
