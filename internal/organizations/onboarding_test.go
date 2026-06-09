package organizations

import (
	"errors"
	"testing"
)

func TestNormalizeClaimEvidenceInput(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		target     string
		wantMethod string
		wantTarget string
		wantErr    error
	}{
		{
			name:       "official email",
			method:     " Official_Email ",
			target:     " hello@example.org ",
			wantMethod: "official_email",
			wantTarget: "hello@example.org",
		},
		{
			name:    "invalid official email target",
			method:  "official_email",
			target:  "not-email",
			wantErr: ErrClaimTargetInvalid,
		},
		{
			name:       "instagram",
			method:     "instagram",
			target:     "@example",
			wantMethod: "instagram",
			wantTarget: "@example",
		},
		{
			name:    "manual review is onboarding only",
			method:  "manual_review",
			target:  "uploaded registration document",
			wantErr: ErrClaimMethodInvalid,
		},
		{
			name:    "invalid method",
			method:  "sms",
			target:  "hello",
			wantErr: ErrClaimMethodInvalid,
		},
		{
			name:    "missing target",
			method:  "official_email",
			target:  " ",
			wantErr: ErrClaimTargetRequired,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotMethod, gotTarget, err := NormalizeClaimEvidenceInput(tc.method, tc.target)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotMethod != tc.wantMethod || gotTarget != tc.wantTarget {
				t.Fatalf("NormalizeClaimEvidenceInput() = (%q, %q), want (%q, %q)", gotMethod, gotTarget, tc.wantMethod, tc.wantTarget)
			}
		})
	}
}

func TestNormalizeOnboardingClaimEvidenceInputAllowsManualReview(t *testing.T) {
	method, target, err := NormalizeOnboardingClaimEvidenceInput(" manual_review ", " uploaded registration document ")
	if err != nil {
		t.Fatalf("NormalizeOnboardingClaimEvidenceInput: %v", err)
	}
	if method != "manual_review" || target != "uploaded registration document" {
		t.Fatalf("NormalizeOnboardingClaimEvidenceInput() = (%q, %q)", method, target)
	}
}

func TestNormalizeOnboardingOrganizationForcesPending(t *testing.T) {
	input := OnboardingRequestInput{
		Name:   " Example Org ",
		Method: "manual_review",
		Target: "registration document",
	}

	normalized, err := NormalizeAdminInput(AdminInput{
		Slug:        input.Slug,
		Name:        input.Name,
		ClaimStatus: "pending",
	})
	if err != nil {
		t.Fatalf("NormalizeAdminInput: %v", err)
	}
	if normalized.Slug != "example-org" {
		t.Fatalf("slug = %q, want example-org", normalized.Slug)
	}
	if normalized.ClaimStatus != "pending" {
		t.Fatalf("claim_status = %q, want pending", normalized.ClaimStatus)
	}
}
