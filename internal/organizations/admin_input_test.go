package organizations

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNormalizeAdminInputDefaultsAndNormalizes(t *testing.T) {
	got, err := NormalizeAdminInput(AdminInput{
		Name:          "  Taman Baca Pesisir Tarakan  ",
		OfficialEmail: " team@example.org ",
		ProfileData:   json.RawMessage(`{"focus":"education"}`),
	})
	if err != nil {
		t.Fatalf("NormalizeAdminInput returned error: %v", err)
	}

	if got.Name != "Taman Baca Pesisir Tarakan" {
		t.Fatalf("Name = %q, want trimmed name", got.Name)
	}
	if got.Slug != "taman-baca-pesisir-tarakan" {
		t.Fatalf("Slug = %q, want normalized slug from name", got.Slug)
	}
	if got.ClaimStatus != "unclaimed" {
		t.Fatalf("ClaimStatus = %q, want default unclaimed", got.ClaimStatus)
	}
	if got.OfficialEmail != "team@example.org" {
		t.Fatalf("OfficialEmail = %q, want trimmed email", got.OfficialEmail)
	}
}

func TestNormalizeAdminInputRejectsInvalidContractFields(t *testing.T) {
	cases := []struct {
		name  string
		input AdminInput
		want  error
	}{
		{
			name:  "missing name",
			input: AdminInput{},
			want:  ErrOrganizationNameRequired,
		},
		{
			name:  "empty normalized slug",
			input: AdminInput{Name: "---"},
			want:  ErrOrganizationSlugRequired,
		},
		{
			name:  "invalid claim status",
			input: AdminInput{Name: "Org", ClaimStatus: "approved"},
			want:  ErrOrganizationClaimStatusInvalid,
		},
		{
			name:  "invalid official email",
			input: AdminInput{Name: "Org", OfficialEmail: "not an email"},
			want:  ErrOrganizationOfficialEmailInvalid,
		},
		{
			name:  "json field must be object",
			input: AdminInput{Name: "Org", ProfileData: json.RawMessage(`[]`)},
			want:  ErrOrganizationJSONInvalid,
		},
		{
			name:  "json field must be valid",
			input: AdminInput{Name: "Org", SourceData: json.RawMessage(`{`)},
			want:  ErrOrganizationJSONInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NormalizeAdminInput(tc.input)
			if !errors.Is(err, tc.want) {
				t.Fatalf("NormalizeAdminInput error = %v, want %v", err, tc.want)
			}
		})
	}
}
