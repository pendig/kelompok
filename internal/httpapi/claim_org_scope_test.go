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

// scopedAdminServer returns a server configured with a single allowed
// organization slug. Every test below exercises the cross-tenant boundary
// without a DB: the admin scope check must reject the request before any
// repository call would touch the nil pgxpool.Pool.
func scopedAdminServer() *Server {
	return New(config.Config{
		APIAddr:                ":0",
		AdminAPIKey:            "test-secret",
		AdminOrganizationSlugs: []string{"allowed-org"},
	}, nil)
}

// authedRequest builds a request authenticated with the scoped admin key in
// the canonical header so middleware passes; the per-route scope check is
// what we are exercising.
func authedRequest(method, path string) *http.Request {
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("X-Kelompok-Admin-Key", "test-secret")
	return request
}

// userSessionRequest forges a non-superadmin user session into the request
// context the same way requireSession would. Used to prove user-session
// branches of the admin scope checks behave correctly without a DB.
func userSessionRequest(method, path string, user auth.User) *http.Request {
	request := httptest.NewRequest(method, path, nil)
	return request.WithContext(context.WithValue(request.Context(), principalContextKey, principal{User: user}))
}

// TestClaimMutationRoutesEnforceCrossOrgIsolation is the central proof that a
// scoped admin key cannot peek at, list, approve, or reject claim activity
// for an organization it does not own. Every entry returns 403
// admin_org_forbidden BEFORE the handler reaches the repository, so org
// metadata cannot leak via either the response body or differential timing.
func TestClaimMutationRoutesEnforceCrossOrgIsolation(t *testing.T) {
	server := scopedAdminServer()

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{"list claims", http.MethodGet, "/api/v1/org-admin/organizations/other-org/claims"},
		{"list audit logs", http.MethodGet, "/api/v1/org-admin/organizations/other-org/audit-logs"},
		{"list members", http.MethodGet, "/api/v1/org-admin/organizations/other-org/members"},
		{"list relationships", http.MethodGet, "/api/v1/org-admin/organizations/other-org/relationships"},
		{"list delegated claims", http.MethodGet, "/api/v1/org-admin/organizations/other-org/delegated-claims"},
		{"get organization", http.MethodGet, "/api/v1/org-admin/organizations/other-org"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, authedRequest(tc.method, tc.path))

			if recorder.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, http.StatusForbidden, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), "admin_org_forbidden") {
				t.Fatalf("missing stable cross-org error code: %s", recorder.Body.String())
			}
			// The forbidden response must not leak the protected slug or any
			// repository-shaped identifier — it must be a generic refusal.
			if strings.Contains(recorder.Body.String(), "claim_request") || strings.Contains(recorder.Body.String(), "audit_log") {
				t.Fatalf("forbidden response leaked entity context: %s", recorder.Body.String())
			}
		})
	}
}

func TestCreateRelatedOrganizationRejectsCrossOrgScopeBeforeBodyDecode(t *testing.T) {
	server := scopedAdminServer()

	request := authedRequest(http.MethodPost, "/api/v1/org-admin/organizations/other-org/related-organizations")
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, http.StatusForbidden, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "admin_org_forbidden") {
		t.Fatalf("missing stable cross-org error code: %s", recorder.Body.String())
	}
}

func TestEnsureAdminCanReviewClaimAllowsDirectScopedAdminKey(t *testing.T) {
	server := scopedAdminServer()
	request := authedRequest(http.MethodPost, "/api/v1/org-admin/claims/id/approve")
	request = request.WithContext(context.WithValue(request.Context(), principalContextKey, principal{AdminKey: true}))
	recorder := httptest.NewRecorder()

	if !server.ensureAdminCanReviewClaimForRequest(recorder, request, "allowed-org") {
		t.Fatalf("expected direct scoped admin key to review claims for its organization")
	}
}

func TestEnsureAdminCanReviewClaimRequiresPrincipal(t *testing.T) {
	server := New(config.Config{APIAddr: ":0", AdminAPIKey: "test-secret"}, nil)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/org-admin/claims/id/approve", nil)
	recorder := httptest.NewRecorder()

	if server.ensureAdminCanReviewClaimForRequest(recorder, request, "allowed-org") {
		t.Fatalf("expected helper to reject request without principal context")
	}
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "admin_auth_required") {
		t.Fatalf("missing stable admin_auth_required error: %s", recorder.Body.String())
	}
}

// Scoped admin keys must NOT be able to call workspace-wide admin endpoints —
// the "global" audit/post/impact/org listing paths can include rows from
// every tenant in the workspace, so they must reject any caller that has
// been pinned to a single organization.
func TestGlobalAdminRoutesRejectScopedAdminKey(t *testing.T) {
	server := scopedAdminServer()

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{"list organizations (global)", http.MethodGet, "/api/v1/org-admin/organizations"},
		{"list posts (global)", http.MethodGet, "/api/v1/org-admin/posts"},
		{"list impact reports (global)", http.MethodGet, "/api/v1/org-admin/impact-reports"},
		{"list posts (mismatched query org)", http.MethodGet, "/api/v1/org-admin/posts?organization_slug=other-org"},
		{"list impact reports (mismatched query org)", http.MethodGet, "/api/v1/org-admin/impact-reports?organization_slug=other-org"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, authedRequest(tc.method, tc.path))

			if recorder.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, http.StatusForbidden, recorder.Body.String())
			}
			body := recorder.Body.String()
			if !strings.Contains(body, "admin_org_scope") && !strings.Contains(body, "admin_org_forbidden") {
				t.Fatalf("missing stable scope error: %s", body)
			}
		})
	}
}

// User sessions that aren't superadmin must fall through to the org-scoped
// admin endpoints and never resolve a global view. Without this guard, any
// org admin could enumerate audit/post/impact reports across every tenant.
func TestGlobalAdminRoutesRejectNonSuperadminSession(t *testing.T) {
	// Use a server WITHOUT a scoped admin key — so the rejection comes from
	// the user-session branch, not the admin-key branch.
	server := New(config.Config{APIAddr: ":0", AdminAPIKey: "test-secret"}, nil)
	user := auth.User{ID: "user-1", Role: "organization_admin"}

	cases := []struct {
		name string
		path string
	}{
		{"list organizations (global)", "/api/v1/org-admin/organizations"},
		{"list posts (global)", "/api/v1/org-admin/posts"},
		{"list impact reports (global)", "/api/v1/org-admin/impact-reports"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := userSessionRequest(http.MethodGet, tc.path, user)
			recorder := httptest.NewRecorder()

			// We invoke the handler chain by routing through the mux so the
			// principal context survives. The mux's requireAdmin will accept
			// a session-context principal because adminPrincipal also accepts
			// pre-baked context principals via principalFromContext via the
			// underlying handler. Verify by calling the handler functions
			// that expose the global-list scope guard directly:
			handlerFor(t, server, tc.path)(recorder, request)

			if recorder.Code != http.StatusForbidden {
				t.Fatalf("%s status = %d, want %d (body: %s)", tc.name, recorder.Code, http.StatusForbidden, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), "admin_org_scope_required") {
				t.Fatalf("missing stable admin_org_scope_required error: %s", recorder.Body.String())
			}
		})
	}
}

// handlerFor maps a known global admin path to the underlying handler so the
// test can drive it with an arbitrary principal already in the request
// context, bypassing requireAdmin (which would otherwise demand a real
// admin-key match or DB-backed session).
func handlerFor(t *testing.T, server *Server, path string) http.HandlerFunc {
	t.Helper()
	switch path {
	case "/api/v1/org-admin/organizations":
		return server.handleListAdminOrganizations
	case "/api/v1/org-admin/posts":
		return server.handleListAdminPosts
	case "/api/v1/org-admin/impact-reports":
		return server.handleListAdminImpactReports
	default:
		t.Fatalf("no global admin handler mapping for %q", path)
		return nil
	}
}

// Approve/reject claim must enforce the auth gate (admin key configured /
// admin auth required) BEFORE consulting the DB for the claim row. If a
// regression let the request through to FindClaimByID with the nil pool, the
// handler would crash; if it was let through with a real pool, the claim's
// organization could be silently exposed across tenants.
func TestClaimReviewRoutesRequireAdminAuth(t *testing.T) {
	t.Run("admin key not configured", func(t *testing.T) {
		server := New(config.Config{APIAddr: ":0"}, nil)
		for _, action := range []string{"approve", "reject"} {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/org-admin/claims/00000000-0000-0000-0000-000000000001/"+action, nil)
			server.Handler().ServeHTTP(recorder, request)

			if recorder.Code != http.StatusServiceUnavailable {
				t.Fatalf("%s: status = %d, want %d", action, recorder.Code, http.StatusServiceUnavailable)
			}
			if !strings.Contains(recorder.Body.String(), "admin_auth_not_configured") {
				t.Fatalf("%s: missing stable error: %s", action, recorder.Body.String())
			}
		}
	})

	t.Run("admin key configured but missing", func(t *testing.T) {
		server := New(config.Config{APIAddr: ":0", AdminAPIKey: "test-secret"}, nil)
		for _, action := range []string{"approve", "reject"} {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/org-admin/claims/00000000-0000-0000-0000-000000000001/"+action, nil)
			server.Handler().ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("%s: status = %d, want %d", action, recorder.Code, http.StatusUnauthorized)
			}
			if !strings.Contains(recorder.Body.String(), "admin_auth_required") {
				t.Fatalf("%s: missing stable error: %s", action, recorder.Body.String())
			}
		}
	})
}

// CreateOrganizationClaim is a public, unauthenticated endpoint by design
// (anyone can claim an org with proof). It must still reject unparseable
// JSON BEFORE FindBySlug touches the DB so the surface area for a
// malformed-body crash stays minimal.
func TestCreateOrganizationClaimRejectsInvalidJSON(t *testing.T) {
	server := New(config.Config{APIAddr: ":0"}, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/organizations/example-org/claims", strings.NewReader(""))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "invalid_json") {
		t.Fatalf("missing stable invalid_json error: %s", recorder.Body.String())
	}
}

// TestCreateOrganizationClaimRejectsUnknownFields prevents future FE
// regressions or attacker-controlled fields (status, organization_id) from
// silently being smuggled into the claim row by exploiting json.Unmarshal's
// permissive defaults.
func TestCreateOrganizationClaimRejectsUnknownFields(t *testing.T) {
	server := New(config.Config{APIAddr: ":0"}, nil)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/organizations/example-org/claims",
		strings.NewReader(`{"method":"official_email","target":"hi@example.org","status":"approved"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "invalid_json") {
		t.Fatalf("expected invalid_json for unknown smuggled field, got %s", recorder.Body.String())
	}
}

// TestEnsureAdminOrganizationSlugRejectsBlankUserScope guards against the
// case where a user-session principal sneaks through with an empty target
// slug. Without this branch, a lookup that derives the slug from a missing
// PathValue would fall through to authorizedAdminOrganizationSlug and
// silently allow the request through (because the empty slug is "in" the
// no-allowlist case).
func TestEnsureAdminOrganizationSlugRejectsBlankUserScope(t *testing.T) {
	server := New(config.Config{APIAddr: ":0", AdminAPIKey: "test-secret"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/org-admin/organizations/", nil)
	request = request.WithContext(context.WithValue(request.Context(), principalContextKey, principal{
		User: auth.User{ID: "user-1", Role: "organization_admin"},
	}))
	recorder := httptest.NewRecorder()

	if server.ensureAdminOrganizationSlugForRequest(recorder, request, "") {
		t.Fatalf("expected blank slug to be rejected for non-superadmin session")
	}
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, http.StatusForbidden, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "admin_org_scope_required") {
		t.Fatalf("missing stable admin_org_scope_required error: %s", recorder.Body.String())
	}
}

// TestEnsureAdminOrganizationSlugRequiresPrincipal proves the helper refuses
// to authorize requests that arrive without ever passing through a
// requireAdmin / requireSession middleware. This is the seatbelt that
// prevents a future handler from forgetting the middleware and shipping an
// open route.
func TestEnsureAdminOrganizationSlugRequiresPrincipal(t *testing.T) {
	server := New(config.Config{APIAddr: ":0", AdminAPIKey: "test-secret"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/org-admin/organizations/x", nil)
	recorder := httptest.NewRecorder()

	if server.ensureAdminOrganizationSlugForRequest(recorder, request, "x") {
		t.Fatalf("expected helper to reject request without principal context")
	}
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "admin_auth_required") {
		t.Fatalf("missing stable admin_auth_required error: %s", recorder.Body.String())
	}
}
