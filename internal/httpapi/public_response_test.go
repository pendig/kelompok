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
