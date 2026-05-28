package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pendig/kelompok/internal/config"
)

// newAuthTestServer constructs a Server with no DB pool. Every test in this
// file is intentionally exercised against a Server that would crash if its
// handler reached the database, which proves the handler returned its stable
// error code from a request-shape check (JSON validation, bearer parsing)
// before the auth flow ever consulted the DB.
//
// This matches CI: there is no postgres service container in
// .github/workflows/ci.yml, so no test in this package may rely on a live DB.
func newAuthTestServer() *Server {
	return New(config.Config{APIAddr: ":0"}, nil)
}

func TestRegisterRejectsInvalidJSONBody(t *testing.T) {
	server := newAuthTestServer()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader("not-json"))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), "invalid_json") {
		t.Fatalf("missing stable invalid_json error: %s", recorder.Body.String())
	}
}

// Stable contract: register rejects unknown fields so a future FE typo
// surfaces immediately rather than being silently dropped before INSERT.
func TestRegisterRejectsUnknownFields(t *testing.T) {
	server := newAuthTestServer()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/register",
		strings.NewReader(`{"name":"X","email":"x@example.org","password":"correcthorse","role":"superadmin"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), "invalid_json") {
		t.Fatalf("expected invalid_json for unknown role escalation field, got %s", recorder.Body.String())
	}
}

func TestLoginRejectsInvalidJSONBody(t *testing.T) {
	server := newAuthTestServer()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(""))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), "invalid_json") {
		t.Fatalf("missing stable invalid_json error: %s", recorder.Body.String())
	}
}

// requireSession must reject any request whose Authorization header is missing
// or malformed BEFORE consulting the DB. The auth_handlers wire /logout and
// /me through requireSession; if a regression let an empty/blank/wrong-scheme
// header through, the next call (UserBySessionToken) would either crash
// against the nil pool or silently treat the empty token as valid.
func TestSessionGuardedRoutesRejectMissingOrMalformedBearer(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		header string
	}{
		// /logout
		{"logout no auth", http.MethodPost, "/api/v1/auth/logout", ""},
		{"logout blank bearer", http.MethodPost, "/api/v1/auth/logout", "Bearer "},
		{"logout wrong scheme", http.MethodPost, "/api/v1/auth/logout", "Basic dXNlcjpwYXNz"},
		{"logout token-only", http.MethodPost, "/api/v1/auth/logout", "abcd"},
		// /me
		{"me no auth", http.MethodGet, "/api/v1/auth/me", ""},
		{"me blank bearer", http.MethodGet, "/api/v1/auth/me", "Bearer    "},
		{"me wrong scheme", http.MethodGet, "/api/v1/auth/me", "Token abcd"},
		{"me token-only", http.MethodGet, "/api/v1/auth/me", "raw-token-abcd"},
	}

	server := newAuthTestServer()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.header != "" {
				request.Header.Set("Authorization", tc.header)
			}
			recorder := httptest.NewRecorder()

			server.Handler().ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), "session_required") {
				t.Fatalf("missing stable session_required error: %s", recorder.Body.String())
			}
		})
	}
}

// bearerToken is the parser every requireSession + admin-key path depends on.
// The case-insensitive scheme + whitespace trimming is what lets `Bearer XYZ`
// and `bearer XYZ` resolve to the same token, but it MUST refuse `BearerXYZ`,
// `Token XYZ`, and missing schemes (otherwise an attacker could feed an
// API-key-shaped header through the session path).
func TestBearerTokenAcceptsValidSchemesAndRejectsInvalid(t *testing.T) {
	if got := bearerToken("Bearer abc-token"); got != "abc-token" {
		t.Fatalf("Bearer abc-token → %q", got)
	}
	if got := bearerToken("bearer abc-token"); got != "abc-token" {
		t.Fatalf("case-insensitive scheme failed: %q", got)
	}
	if got := bearerToken(" Bearer  abc-token  "); got != "abc-token" {
		t.Fatalf("whitespace trimming failed: %q", got)
	}

	for _, header := range []string{
		"",
		"Bearer",
		"Bearer ",
		"Token abc",
		"Basic dXNlcjpwYXNz",
		"Bearerabc-token",
	} {
		if got := bearerToken(header); got != "" {
			t.Fatalf("expected empty token for header %q, got %q", header, got)
		}
	}
}
