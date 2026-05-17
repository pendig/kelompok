package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pendig/kelompok/internal/config"
	"github.com/pendig/kelompok/internal/organizations"
	"github.com/pendig/kelompok/internal/posts"
)

func TestPublicOrganizationOmitsClaimEmailAndFiltersJSON(t *testing.T) {
	item := organizations.Organization{
		ID:            "org-internal-id",
		Slug:          "safe-org",
		Name:          "Safe Org",
		OfficialEmail: "claim@example.org",
		ClaimStatus:   "claimed",
		ProfileData: json.RawMessage(`{
			"focus":["education"],
			"official_email":"claim@example.org",
			"social_links":{"instagram":"https://instagram.com/safe","private_token":"secret"},
			"source_url":"https://internal.example.org",
			"raw_payload":{"hidden":true}
		}`),
		SDGSData:   json.RawMessage(`{"primary":["SDG 4"],"evidence_text":"internal note"}`),
		ImpactData: json.RawMessage(`{"volunteers":10,"private_note":"do not leak"}`),
		CreatedAt:  time.Date(2026, 5, 10, 1, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 5, 10, 1, 0, 0, 0, time.UTC),
	}

	encoded, err := json.Marshal(publicOrganization(item))
	if err != nil {
		t.Fatalf("marshal public organization: %v", err)
	}
	body := string(encoded)

	for _, forbidden := range []string{
		"claim@example.org",
		"official_email",
		"org-internal-id",
		"private_token",
		"source_url",
		"raw_payload",
		"evidence_text",
		"private_note",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("public organization leaked %q in %s", forbidden, body)
		}
	}
	if !strings.Contains(body, "https://instagram.com/safe") {
		t.Fatalf("public organization removed safe social link: %s", body)
	}
}

func TestPublicOrganizationKeepsReviewedPublicContact(t *testing.T) {
	item := organizations.Organization{
		Slug:        "safe-org",
		Name:        "Safe Org",
		ClaimStatus: "claimed",
		ProfileData: json.RawMessage(`{
			"public_contact":{
				"email":"hello@example.org",
				"phone":"+6200000000",
				"private_token":"hidden"
			}
		}`),
		SDGSData:   json.RawMessage(`{}`),
		ImpactData: json.RawMessage(`{}`),
		CreatedAt:  time.Date(2026, 5, 10, 1, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 5, 10, 1, 0, 0, 0, time.UTC),
	}

	encoded, err := json.Marshal(publicOrganization(item))
	if err != nil {
		t.Fatalf("marshal public organization: %v", err)
	}
	body := string(encoded)

	for _, expected := range []string{"hello@example.org", "+6200000000"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("public organization removed public contact %q in %s", expected, body)
		}
	}
	if strings.Contains(body, "private_token") || strings.Contains(body, "hidden") {
		t.Fatalf("public organization leaked private public_contact field: %s", body)
	}
}

func TestPublicOrganizationRelationshipsOmitInternalIDsAndInactiveRows(t *testing.T) {
	item := organizations.Organization{
		Slug:        "ipm",
		Name:        "IPM",
		ClaimStatus: "claimed",
		ProfileData: json.RawMessage(`{}`),
		SDGSData:    json.RawMessage(`{}`),
		ImpactData:  json.RawMessage(`{}`),
		CreatedAt:   time.Date(2026, 5, 10, 1, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 10, 1, 0, 0, 0, time.UTC),
	}
	relationships := []organizations.Relationship{
		{
			ID:                   "relationship-internal-id",
			ParentOrganizationID: "parent-internal-id",
			Parent:               organizations.OrganizationRef{ID: "parent-internal-id", Slug: "muhammadiyah", Name: "Muhammadiyah"},
			ChildOrganizationID:  "child-internal-id",
			Child:                organizations.OrganizationRef{ID: "child-internal-id", Slug: "ipm", Name: "IPM"},
			RelationshipType:     "autonomous_body",
			Label:                "Autonomous student organization",
			Status:               "active",
		},
		{
			Parent:           organizations.OrganizationRef{Slug: "inactive-parent", Name: "Inactive Parent"},
			Child:            organizations.OrganizationRef{Slug: "ipm", Name: "IPM"},
			RelationshipType: "structural_parent",
			Status:           "inactive",
		},
	}

	encoded, err := json.Marshal(publicOrganizationWithRelationships(item, relationships))
	if err != nil {
		t.Fatalf("marshal public organization relationships: %v", err)
	}
	body := string(encoded)

	for _, forbidden := range []string{"relationship-internal-id", "parent-internal-id", "child-internal-id", "inactive-parent"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("public organization relationship leaked %q in %s", forbidden, body)
		}
	}
	if !strings.Contains(body, `"parents":[{"organization":{"slug":"muhammadiyah","name":"Muhammadiyah"}`) {
		t.Fatalf("public organization missing parent relationship: %s", body)
	}
}

func TestPublicPostOmitsInternalIDsAndFiltersSEOData(t *testing.T) {
	item := posts.Post{
		ID:               "post-internal-id",
		OrganizationID:   "org-internal-id",
		OrganizationSlug: "safe-org",
		OrganizationName: "Safe Org",
		Slug:             "hello",
		Title:            "Hello",
		Status:           "published",
		PostData:         json.RawMessage(`{"kind":"news","featured":true,"source_record_id":"hidden"}`),
		SEOData:          json.RawMessage(`{"title":"Hello","private_note":"hidden","token":"hidden"}`),
		CreatedAt:        time.Date(2026, 5, 10, 1, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2026, 5, 10, 1, 0, 0, 0, time.UTC),
	}

	encoded, err := json.Marshal(publicPost(item))
	if err != nil {
		t.Fatalf("marshal public post: %v", err)
	}
	body := string(encoded)

	for _, forbidden := range []string{
		"post-internal-id",
		"org-internal-id",
		"source_record_id",
		"private_note",
		"token",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("public post leaked %q in %s", forbidden, body)
		}
	}
	if !strings.Contains(body, `"organization":{"slug":"safe-org","name":"Safe Org"}`) {
		t.Fatalf("public post missing organization ref: %s", body)
	}
}

func TestReadyzDoesNotExposeDatabaseError(t *testing.T) {
	server := New(config.Config{APIAddr: ":0"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}

	body := recorder.Body.String()
	if strings.Contains(body, "database pool is not initialized") {
		t.Fatalf("readyz leaked database error detail: %s", body)
	}
	if !strings.Contains(body, "database_not_ready") {
		t.Fatalf("readyz missing stable error code: %s", body)
	}
}
