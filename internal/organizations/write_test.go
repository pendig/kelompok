package organizations

import (
	"strings"
	"testing"
)

func TestNormalizeAdminInputRequiresName(t *testing.T) {
	_, err := normalizeAdminInput(AdminInput{Name: "  "}, AdminInput{})
	if err == nil || !strings.Contains(err.Error(), "name") {
		t.Fatalf("expected name error, got %v", err)
	}
}

func TestNormalizeAdminInputDerivesSlugFromNameWhenBlank(t *testing.T) {
	out, err := normalizeAdminInput(AdminInput{Name: "Kelompok Nusantara"}, AdminInput{})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if out.Slug != "kelompok-nusantara" {
		t.Fatalf("slug fallback = %q, want kelompok-nusantara", out.Slug)
	}
	if out.Name != "Kelompok Nusantara" {
		t.Fatalf("name preserved = %q", out.Name)
	}
}

func TestNormalizeAdminInputRespectsExistingSlugDefault(t *testing.T) {
	out, err := normalizeAdminInput(
		AdminInput{Name: "Renamed", Slug: ""},
		AdminInput{Slug: "kept-slug"},
	)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if out.Slug != "kept-slug" {
		t.Fatalf("expected default slug to win when input is blank, got %q", out.Slug)
	}
}

func TestNormalizeAdminInputTrimsTextFields(t *testing.T) {
	out, err := normalizeAdminInput(
		AdminInput{
			Name:        "  Kelompok ",
			Slug:        "  Kelompok Indonesia  ",
			LegalName:   "  Yayasan Kelompok ",
			Description: "  Public ",
			History:     "  Founded ",
			Country:     "  ID ",
			Region:      "  DKI ",
			City:        "  Jakarta ",
		},
		AdminInput{},
	)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if out.Name != "Kelompok" {
		t.Fatalf("name = %q", out.Name)
	}
	if out.Slug != "kelompok-indonesia" {
		t.Fatalf("slug = %q", out.Slug)
	}
	if out.LegalName != "Yayasan Kelompok" || out.Description != "Public" || out.History != "Founded" {
		t.Fatalf("text fields not trimmed: %+v", out)
	}
	if out.Country != "ID" || out.Region != "DKI" || out.City != "Jakarta" {
		t.Fatalf("location fields not trimmed: %+v", out)
	}
}

func TestNormalizeAdminInputRejectsInvalidClaimStatus(t *testing.T) {
	_, err := normalizeAdminInput(AdminInput{Name: "Test", ClaimStatus: "weird"}, AdminInput{})
	if err == nil || !strings.Contains(err.Error(), "claim_status") {
		t.Fatalf("expected claim_status error, got %v", err)
	}
}

func TestNormalizeAdminInputAcceptsKnownClaimStatuses(t *testing.T) {
	for _, status := range []string{"", "unclaimed", "pending", "claimed", "rejected"} {
		out, err := normalizeAdminInput(AdminInput{Name: "Test", ClaimStatus: status}, AdminInput{})
		if err != nil {
			t.Fatalf("status %q rejected: %v", status, err)
		}
		if out.ClaimStatus != status {
			t.Fatalf("status %q normalized to %q", status, out.ClaimStatus)
		}
	}
}

func TestNormalizeAdminInputRejectsBadWebsiteScheme(t *testing.T) {
	_, err := normalizeAdminInput(
		AdminInput{Name: "Test", WebsiteURL: "kelompok.id"},
		AdminInput{},
	)
	if err == nil || !strings.Contains(err.Error(), "website_url") {
		t.Fatalf("expected website_url error, got %v", err)
	}
}

func TestNormalizeAdminInputAcceptsHTTPSchemes(t *testing.T) {
	for _, value := range []string{"http://kelompok.id", "https://kelompok.id"} {
		_, err := normalizeAdminInput(
			AdminInput{Name: "Test", WebsiteURL: value},
			AdminInput{},
		)
		if err != nil {
			t.Fatalf("website %q rejected: %v", value, err)
		}
	}
}

func TestNormalizeAdminInputRejectsBadEmail(t *testing.T) {
	_, err := normalizeAdminInput(
		AdminInput{Name: "Test", OfficialEmail: "not-an-email"},
		AdminInput{},
	)
	if err == nil || !strings.Contains(err.Error(), "official_email") {
		t.Fatalf("expected official_email error, got %v", err)
	}
}

func TestNormalizeAdminInputAcceptsBlankOptionalFields(t *testing.T) {
	out, err := normalizeAdminInput(AdminInput{Name: "Test"}, AdminInput{})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if out.WebsiteURL != "" || out.OfficialEmail != "" || out.LegalName != "" {
		t.Fatalf("expected optional fields to stay empty: %+v", out)
	}
}

func TestNormalizeAdminInputErrorsWhenNameYieldsEmptySlug(t *testing.T) {
	// Name composed only of separators normalizes to an empty slug; the
	// helper should refuse rather than insert a row with an empty slug.
	_, err := normalizeAdminInput(AdminInput{Name: "---"}, AdminInput{})
	if err == nil {
		t.Fatalf("expected slug derivation error")
	}
}
