package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/pendig/kelompok/internal/config"
)

// loadOpenAPIDocument reads docs/openapi.yaml relative to the repo root.
//
// The test walks up from the package directory until it finds go.mod so that
// running `go test ./...` from anywhere still locates the artifact.
func loadOpenAPIDocument(t *testing.T) (path string, content string) {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for {
		candidate := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(candidate); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate repo root from %s", dir)
		}
		dir = parent
	}

	path = filepath.Join(dir, "docs", "openapi.yaml")
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read openapi document: %v", err)
	}
	return path, string(bytes)
}

func TestOpenAPIDocumentHasRequiredHeader(t *testing.T) {
	_, content := loadOpenAPIDocument(t)

	for _, marker := range []string{
		"openapi: 3.1.0",
		"\ninfo:",
		"\npaths:",
		"\ncomponents:",
	} {
		if !strings.Contains(content, marker) {
			t.Fatalf("openapi document missing required marker %q", marker)
		}
	}
}

// TestOpenAPIDocumentCoversAllRegisteredRoutes verifies that every route the
// API server exposes is documented under `paths:` and references the matching
// HTTP method. RegisteredRoutes is the canonical inventory from server.go, so
// this test fails closed when a new route ships without a matching contract
// entry.
func TestOpenAPIDocumentCoversAllRegisteredRoutes(t *testing.T) {
	path, content := loadOpenAPIDocument(t)

	// `paths:` block runs from `\npaths:\n` to the next top-level key
	// (`components:`). We restrict the search to that block so HTTP verbs
	// elsewhere (for example in `description:` prose) cannot accidentally
	// satisfy the assertion.
	pathsBlock := extractTopLevelBlock(t, content, "paths")

	for _, route := range RegisteredRoutes() {
		operationBlock := extractPathOperationBlock(t, pathsBlock, route.Path)
		method := strings.ToLower(route.Method)
		// OpenAPI operation keys are nested four spaces under the path key
		// (`paths:` two-space, `<path>:` four-space).
		marker := "\n    " + method + ":"
		if !strings.Contains("\n"+operationBlock, marker) {
			t.Fatalf(
				"openapi document %s declares path %q but is missing the %s operation",
				path, route.Path, strings.ToUpper(method),
			)
		}
	}
}

// TestOpenAPIDocumentDoesNotDeclareUnknownRoutes verifies the reverse: every
// path that appears under `paths:` maps to a route the server actually
// registers. This catches drift where a contract entry survives a route
// removal.
func TestOpenAPIDocumentDoesNotDeclareUnknownRoutes(t *testing.T) {
	_, content := loadOpenAPIDocument(t)
	pathsBlock := extractTopLevelBlock(t, content, "paths")

	registered := make(map[string]struct{})
	for _, route := range RegisteredRoutes() {
		registered[route.Path] = struct{}{}
	}

	pathPattern := regexp.MustCompile(`(?m)^  (/[^:\s]*):`)
	for _, match := range pathPattern.FindAllStringSubmatch(pathsBlock, -1) {
		documented := match[1]
		if _, ok := registered[documented]; !ok {
			t.Fatalf("openapi document declares undocumented or removed path: %q", documented)
		}
	}
}

func TestHealthzMatchesDocumentedSuccessEnvelope(t *testing.T) {
	server := New(config.Config{APIAddr: ":0", Env: "test"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("/healthz status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("/healthz content-type = %q, want application/json", got)
	}

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode /healthz body: %v", err)
	}
	if _, ok := body["data"]; !ok {
		t.Fatalf("/healthz body missing documented %q field: %s", "data", recorder.Body.String())
	}
	if message, _ := body["message"].(string); message != "ok" {
		t.Fatalf("/healthz message = %q, want %q", message, "ok")
	}
}

func TestReadyzMatchesDocumentedErrorEnvelope(t *testing.T) {
	server := New(config.Config{APIAddr: ":0"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("/readyz status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}

	envelope := decodeErrorEnvelope(t, recorder.Body.Bytes())
	if envelope.Error.Code != "database_not_ready" {
		t.Fatalf("/readyz error code = %q, want %q", envelope.Error.Code, "database_not_ready")
	}
	if envelope.Error.Message == "" {
		t.Fatalf("/readyz error message must be non-empty")
	}
}

func TestOrgAdminListMatchesDocumentedAuthErrorEnvelope(t *testing.T) {
	server := New(config.Config{APIAddr: ":0", AdminAPIKey: "test-secret"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/org-admin/organizations", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("/org-admin/organizations status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	envelope := decodeErrorEnvelope(t, recorder.Body.Bytes())
	if envelope.Error.Code != "admin_auth_required" {
		t.Fatalf("/org-admin/organizations error code = %q, want %q", envelope.Error.Code, "admin_auth_required")
	}
}

func TestAuthMeMatchesDocumentedSessionRequiredEnvelope(t *testing.T) {
	server := New(config.Config{APIAddr: ":0"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("/auth/me status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	envelope := decodeErrorEnvelope(t, recorder.Body.Bytes())
	if envelope.Error.Code != "session_required" {
		t.Fatalf("/auth/me error code = %q, want %q", envelope.Error.Code, "session_required")
	}
}

type contractError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func decodeErrorEnvelope(t *testing.T, body []byte) contractError {
	t.Helper()
	var envelope contractError
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode error envelope: %v (body=%s)", err, string(body))
	}
	if envelope.Error.Code == "" || envelope.Error.Message == "" {
		t.Fatalf("error envelope missing documented fields: %s", string(body))
	}
	return envelope
}

// extractTopLevelBlock returns the YAML block that begins with `\n<key>:\n`
// and ends just before the next top-level key. It is intentionally a small
// string scan so the smoke test stays dependency-free.
func extractTopLevelBlock(t *testing.T, content string, key string) string {
	t.Helper()
	header := "\n" + key + ":\n"
	start := strings.Index(content, header)
	if start < 0 {
		t.Fatalf("openapi document missing %q section", key)
	}
	start += len(header)

	// Next top-level key starts at column 0 with `<word>:`.
	nextTopLevel := regexp.MustCompile(`(?m)^[A-Za-z_][A-Za-z0-9_-]*:`)
	tail := content[start:]
	if loc := nextTopLevel.FindStringIndex(tail); loc != nil {
		return tail[:loc[0]]
	}
	return tail
}

// extractPathOperationBlock returns the YAML block for a single path entry
// (`  <path>:` ... up to the next sibling path entry). The smoke test only
// inspects the operation method keys nested inside this block, so the
// extraction does not need to be a full YAML parse.
func extractPathOperationBlock(t *testing.T, pathsBlock string, path string) string {
	t.Helper()
	header := "\n  " + path + ":\n"
	start := strings.Index("\n"+pathsBlock, header)
	if start < 0 {
		t.Fatalf("openapi document missing path entry %q", path)
	}
	// Adjust for the synthetic leading newline used to anchor the search.
	start += len(header) - 1
	tail := pathsBlock[start:]
	nextSibling := regexp.MustCompile(`(?m)^  (/[^:\s]*):`)
	if loc := nextSibling.FindStringIndex(tail); loc != nil {
		return tail[:loc[0]]
	}
	return tail
}
