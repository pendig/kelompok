package auth

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestOrganizationClaimMarshalsAccountFields(t *testing.T) {
	reviewed := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)
	item := OrganizationClaim{
		ID:                      "claim-1",
		OrganizationID:          "org-1",
		OrganizationSlug:        "kelompok",
		OrganizationName:        "Kelompok Foundation",
		OrganizationClaimStatus: "claimed",
		Method:                  "official_email",
		Target:                  "team@kelompok.id",
		Status:                  "approved",
		CreatedAt:               time.Date(2026, 5, 27, 8, 0, 0, 0, time.UTC),
		UpdatedAt:               time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC),
		ReviewedAt:              &reviewed,
	}

	encoded, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal organization claim: %v", err)
	}
	body := string(encoded)

	for _, expected := range []string{
		`"id":"claim-1"`,
		`"organization_id":"org-1"`,
		`"organization_slug":"kelompok"`,
		`"organization_name":"Kelompok Foundation"`,
		`"organization_claim_status":"claimed"`,
		`"method":"official_email"`,
		`"target":"team@kelompok.id"`,
		`"status":"approved"`,
		`"created_at":"2026-05-27T08:00:00Z"`,
		`"updated_at":"2026-05-28T09:00:00Z"`,
		`"reviewed_at":"2026-05-28T09:00:00Z"`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("organization claim JSON missing %q in %s", expected, body)
		}
	}

	// The account view must not surface evidence or reviewer identity, both
	// of which are reviewer-only metadata stored on the underlying claim.
	for _, forbidden := range []string{
		"evidence",
		"reviewed_by_user_id",
		"user_id",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("organization claim JSON unexpectedly leaked %q in %s", forbidden, body)
		}
	}
}

func TestOrganizationClaimOmitsReviewedAtWhenPending(t *testing.T) {
	item := OrganizationClaim{
		ID:                      "claim-2",
		OrganizationID:          "org-2",
		OrganizationSlug:        "kelompok-id",
		OrganizationName:        "Kelompok",
		OrganizationClaimStatus: "pending",
		Method:                  "instagram",
		Target:                  "kelompok.id",
		Status:                  "pending",
		CreatedAt:               time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
		UpdatedAt:               time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
	}

	encoded, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal organization claim: %v", err)
	}
	body := string(encoded)

	if strings.Contains(body, "reviewed_at") {
		t.Fatalf("pending claim should omit reviewed_at: %s", body)
	}
}
