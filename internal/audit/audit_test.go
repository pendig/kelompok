package audit

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestOrganizationIDUsesEntityIDForOrganizationEntity proves the audit resolver
// pins the organization_id column directly from entity_id when the entity itself
// IS the organization. Without this, organization audits would land in a NULL
// organization bucket and leak between tenants in scoped audit listings.
func TestOrganizationIDUsesEntityIDForOrganizationEntity(t *testing.T) {
	got := organizationID("organization", "org-123", nil)
	if got != "org-123" {
		t.Fatalf("organization entity_id = %v, want %q", got, "org-123")
	}

	if got := organizationID("organization", "", nil); got != nil {
		t.Fatalf("blank organization entity_id should resolve to nil, got %v", got)
	}
}

// TestOrganizationIDFallsBackToMetadataForChildEntities proves that audit rows
// for entities owned by an organization (claim_request, post, member,
// impact_report, organization_user_role, organization_relationship) still get
// pinned to organization_id via the explicit metadata.organization_id key.
//
// This is the contract claim/role/relationship audit emitters depend on; if it
// regresses, audit listings filtered by organization slug would silently lose
// rows and FE/admin would no longer see mutations against their org.
func TestOrganizationIDFallsBackToMetadataForChildEntities(t *testing.T) {
	cases := []struct {
		entityType string
		metadata   any
		want       any
	}{
		{"claim_request", map[string]any{"organization_id": "org-claim"}, "org-claim"},
		{"organization_user_role", map[string]any{"organization_id": "org-role", "role": "admin"}, "org-role"},
		{"organization_relationship", map[string]any{"organization_id": "org-rel", "relationship_side": "child"}, "org-rel"},
		{"post", map[string]string{"organization_id": "org-post"}, "org-post"},
		{"member", map[string]any{"organization_id": "org-member"}, "org-member"},
		{"impact_report", map[string]any{"organization_id": "org-impact"}, "org-impact"},
	}

	for _, tc := range cases {
		got := organizationID(tc.entityType, "entity-id-irrelevant", tc.metadata)
		if got != tc.want {
			t.Fatalf("organizationID(%q, ...) = %v, want %v", tc.entityType, got, tc.want)
		}
	}
}

// TestOrganizationIDReturnsNilForMissingMetadata makes sure the resolver does
// not silently coerce empty metadata into a bogus organization id (which would
// route audits into the wrong bucket if metadata is ever forgotten).
func TestOrganizationIDReturnsNilForMissingMetadata(t *testing.T) {
	cases := []struct {
		name     string
		metadata any
	}{
		{"nil metadata", nil},
		{"empty map", map[string]any{}},
		{"unrelated key", map[string]any{"role": "admin"}},
		{"blank string value", map[string]any{"organization_id": ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := organizationID("claim_request", "claim-1", tc.metadata)
			if got != nil {
				t.Fatalf("expected nil, got %v", got)
			}
		})
	}
}

// TestMetadataValueHandlesSupportedShapes confirms the metadata bag accepts the
// same map shapes that production audit emitters pass (typed map[string]any
// from claim/role flows; map[string]string in case future emitters send a
// stricter map).
func TestMetadataValueHandlesSupportedShapes(t *testing.T) {
	if got := metadataValue(map[string]any{"organization_id": "org-1"}, "organization_id"); got != "org-1" {
		t.Fatalf("map[string]any lookup = %q", got)
	}
	if got := metadataValue(map[string]string{"organization_id": "org-2"}, "organization_id"); got != "org-2" {
		t.Fatalf("map[string]string lookup = %q", got)
	}
	if got := metadataValue(map[string]any{"organization_id": 42}, "organization_id"); got != "" {
		t.Fatalf("non-string value should be ignored, got %q", got)
	}
	if got := metadataValue(nil, "organization_id"); got != "" {
		t.Fatalf("nil metadata should resolve to empty, got %q", got)
	}
	if got := metadataValue("not-a-map", "organization_id"); got != "" {
		t.Fatalf("unsupported shape should resolve to empty, got %q", got)
	}
}

// TestNormalizeUUIDDropsZeroValues protects audit emitters against accidentally
// inserting empty UUID strings (which would crash the ::uuid cast inside
// audit.Record's INSERT and silently swallow audit rows).
func TestNormalizeUUIDDropsZeroValues(t *testing.T) {
	if got := normalizeUUID(nil); got != nil {
		t.Fatalf("nil → %v", got)
	}
	if got := normalizeUUID(""); got != nil {
		t.Fatalf("empty string → %v", got)
	}
	if got := normalizeUUID("user-1"); got != "user-1" {
		t.Fatalf("string passthrough = %v", got)
	}
}

// TestNormalizeJSONPreservesPayloadIntegrity confirms before/after/metadata
// payloads are faithfully serialised so audit rows do not lose mutation
// context (a regression here would make audit listings useless for review).
func TestNormalizeJSONPreservesPayloadIntegrity(t *testing.T) {
	if got := normalizeJSON(nil); got != nil {
		t.Fatalf("nil → %v", got)
	}
	if got := normalizeJSON(""); got != nil {
		t.Fatalf("empty string → %v", got)
	}
	if got := normalizeJSON([]byte{}); got != nil {
		t.Fatalf("empty bytes → %v", got)
	}
	if got := normalizeJSON(json.RawMessage(``)); got != nil {
		t.Fatalf("empty RawMessage → %v", got)
	}

	if got := normalizeJSON(json.RawMessage(`{"organization_id":"o"}`)); got != `{"organization_id":"o"}` {
		t.Fatalf("RawMessage passthrough = %v", got)
	}

	encoded := normalizeJSON(map[string]any{"organization_id": "o", "role": "admin"})
	str, ok := encoded.(string)
	if !ok {
		t.Fatalf("expected string-encoded JSON, got %T", encoded)
	}
	if !strings.Contains(str, `"organization_id":"o"`) || !strings.Contains(str, `"role":"admin"`) {
		t.Fatalf("encoded payload lost fields: %s", str)
	}
}

// TestRecordSilentlyNoopsOnNilPool guards the call sites that fire-and-forget
// audit.Record (`_ = audit.Record(...)`); a panic here would crash any
// repository call that audits a write when the pool is unconfigured (tests,
// CLI bootstrap, etc.).
func TestRecordSilentlyNoopsOnNilPool(t *testing.T) {
	if err := Record(nil, nil, "actor-id", "claim_request", "claim-1", "approve", nil, nil, map[string]any{
		"organization_id": "org-1",
	}); err != nil {
		t.Fatalf("expected nil-pool noop, got %v", err)
	}
}
