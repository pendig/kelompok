package httpapi

import (
	"encoding/json"
	"os"
	"testing"
)

func TestOpenAPIContractCoversReleaseRoutes(t *testing.T) {
	raw, err := os.ReadFile("../../docs/openapi.json")
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}

	var doc struct {
		OpenAPI string `json:"openapi"`
		Info    struct {
			Title   string `json:"title"`
			Version string `json:"version"`
		} `json:"info"`
		Paths map[string]map[string]struct {
			OperationID string                 `json:"operationId"`
			Responses   map[string]any         `json:"responses"`
			RequestBody map[string]interface{} `json:"requestBody"`
		} `json:"paths"`
		Components struct {
			Schemas map[string]any `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("OpenAPI contract must be valid JSON: %v", err)
	}

	if doc.OpenAPI == "" || doc.Info.Title == "" || doc.Info.Version == "" {
		t.Fatalf("OpenAPI contract is missing required root metadata")
	}

	required := map[string][]string{
		"/api/v1/organizations":                                {"get"},
		"/api/v1/organizations/{slug}":                         {"get"},
		"/api/v1/organizations/{slug}/claims":                  {"post"},
		"/api/v1/auth/register":                                {"post"},
		"/api/v1/auth/login":                                   {"post"},
		"/api/v1/auth/me":                                      {"get", "patch"},
		"/api/v1/org-admin/organizations":                      {"get", "post"},
		"/api/v1/org-admin/organizations/{slug}":               {"get", "patch"},
		"/api/v1/org-admin/organization-relationships":         {"post"},
		"/api/v1/org-admin/organization-relationships/{id}":    {"patch", "delete"},
		"/api/v1/org-admin/organizations/{slug}/claims":        {"get"},
		"/api/v1/org-admin/claims/{id}/approve":                {"post"},
		"/api/v1/org-admin/claims/{id}/reject":                 {"post"},
		"/api/v1/org-admin/organizations/{slug}/members":       {"get", "post"},
		"/api/v1/org-admin/members/{id}":                       {"patch", "delete"},
		"/api/v1/org-admin/posts":                              {"get", "post"},
		"/api/v1/org-admin/posts/{id}":                         {"patch"},
		"/api/v1/org-admin/impact-reports":                     {"get", "post"},
		"/api/v1/org-admin/impact-reports/{id}":                {"patch"},
		"/api/v1/org-admin/organizations/{slug}/audit-logs":    {"get"},
		"/api/v1/org-admin/organizations/{slug}/relationships": {"get"},
		"/api/v1/organizations/{slug}/posts":                   {"get"},
		"/api/v1/organizations/{slug}/posts/{post_slug}":       {"get"},
		"/api/v1/organizations/{slug}/impact-reports":          {"get"},
	}

	for path, methods := range required {
		operations, ok := doc.Paths[path]
		if !ok {
			t.Fatalf("OpenAPI contract is missing path %s", path)
		}
		for _, method := range methods {
			operation, ok := operations[method]
			if !ok {
				t.Fatalf("OpenAPI contract is missing %s %s", method, path)
			}
			if operation.OperationID == "" {
				t.Fatalf("OpenAPI operation %s %s is missing operationId", method, path)
			}
			if len(operation.Responses) == 0 {
				t.Fatalf("OpenAPI operation %s %s is missing responses", method, path)
			}
		}
	}

	for _, schema := range []string{"Envelope", "ErrorEnvelope", "Organization", "OrganizationInput", "User"} {
		if _, ok := doc.Components.Schemas[schema]; !ok {
			t.Fatalf("OpenAPI contract is missing schema %s", schema)
		}
	}
}
