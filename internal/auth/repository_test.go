package auth

import (
	"context"
	"encoding/json"
	"errors"
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

// TestNormalizeEmailLowercasesAndTrims protects login/register lookup against
// the most common impersonation foothold: registering as `User@Example.org `
// while another user already owns `user@example.org`. The DB's UNIQUE(email)
// constraint depends on this normaliser staying lowercase + trimmed.
func TestNormalizeEmailLowercasesAndTrims(t *testing.T) {
	cases := map[string]string{
		"User@Example.org":     "user@example.org",
		"  user@example.org  ": "user@example.org",
		"\tuser@EXAMPLE.org\n": "user@example.org",
		"":                     "",
		"\t  ":                 "",
		"MIXEDcase@Domain.COM": "mixedcase@domain.com",
	}
	for input, want := range cases {
		if got := normalizeEmail(input); got != want {
			t.Fatalf("normalizeEmail(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestValidEmailRejectsBareUsernamesAndDisplayForms keeps Register's validation
// strict enough that we never INSERT something the DB UNIQUE index would treat
// as the same address as an attacker's variant.
func TestValidEmailRejectsBareUsernamesAndDisplayForms(t *testing.T) {
	valid := []string{
		"a@b.co",
		"user.name+tag@example.org",
	}
	for _, address := range valid {
		if !validEmail(address) {
			t.Fatalf("expected %q to be valid", address)
		}
	}

	invalid := []string{
		"",
		"plain",
		"missing-at-sign.example",
		"<user@example.org>",
		"User <user@example.org>",
		"user@",
		"@example.org",
	}
	for _, address := range invalid {
		if validEmail(address) {
			t.Fatalf("expected %q to be invalid", address)
		}
	}
}

// TestHashTokenIsDeterministicAndRejectsBlank makes sure session tokens stored
// in user_sessions.token_hash are derived deterministically from the raw token
// (used by both Login -> INSERT and UserBySessionToken -> WHERE clauses), and
// that a blank token never resolves to a non-empty hash (which would otherwise
// match revoked-but-blank rows).
func TestHashTokenIsDeterministicAndRejectsBlank(t *testing.T) {
	first := hashToken("session-token-abc")
	second := hashToken("session-token-abc")
	if first == "" || first != second {
		t.Fatalf("expected deterministic non-empty hash, got %q vs %q", first, second)
	}
	if hashToken(" session-token-abc ") != first {
		t.Fatalf("hashToken should trim whitespace before hashing")
	}
	if hashToken("other-token") == first {
		t.Fatalf("different tokens must hash differently")
	}
	for _, blank := range []string{"", "   ", "\t\n"} {
		if got := hashToken(blank); got != "" {
			t.Fatalf("blank token should resolve to empty hash, got %q for %q", got, blank)
		}
	}
}

// TestNewSessionTokenReturnsMatchingTokenAndHash is the contract Login depends
// on: the plaintext token returned to the client must hash to exactly the
// token_hash stored server-side. A regression here would either lock every
// user out (no session ever validates) or, worse, persist plaintext.
func TestNewSessionTokenReturnsMatchingTokenAndHash(t *testing.T) {
	for i := 0; i < 8; i++ {
		token, tokenHash, err := newSessionToken()
		if err != nil {
			t.Fatalf("newSessionToken: %v", err)
		}
		if token == "" || tokenHash == "" {
			t.Fatalf("token/hash must be non-empty")
		}
		if strings.Contains(tokenHash, token) {
			t.Fatalf("token_hash must not contain the plaintext token")
		}
		if hashToken(token) != tokenHash {
			t.Fatalf("hashToken(token) must match returned hash")
		}
	}
}

// TestNewSessionTokensAreUnique guards the UNIQUE(token_hash) constraint on
// user_sessions: a regression that produced duplicates would either crash
// Login outright or, with weaker entropy, collapse two users to one session.
func TestNewSessionTokensAreUnique(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 32; i++ {
		token, _, err := newSessionToken()
		if err != nil {
			t.Fatalf("newSessionToken: %v", err)
		}
		if _, ok := seen[token]; ok {
			t.Fatalf("duplicate session token after %d iterations: %q", i, token)
		}
		seen[token] = struct{}{}
	}
}

// TestSessionTTLMatchesDocumentedBudget is a contract check on the public
// SessionTTL constant the FE relies on for "remember me" semantics.
// Changing it should be deliberate and visible in code review.
func TestSessionTTLMatchesDocumentedBudget(t *testing.T) {
	want := 30 * 24 * time.Hour
	if SessionTTL != want {
		t.Fatalf("SessionTTL = %v, want %v", SessionTTL, want)
	}
}

func TestUpdateProfileValidationUsesSentinelErrors(t *testing.T) {
	repository := NewRepository(nil)

	if _, err := repository.UpdateProfile(context.Background(), "user-1", UpdateProfileInput{Name: "   "}); !errors.Is(err, ErrProfileNameRequired) {
		t.Fatalf("expected ErrProfileNameRequired, got %v", err)
	}

	longName := strings.Repeat("a", 121)
	if _, err := repository.UpdateProfile(context.Background(), "user-1", UpdateProfileInput{Name: longName}); !errors.Is(err, ErrProfileNameTooLong) {
		t.Fatalf("expected ErrProfileNameTooLong, got %v", err)
	}
}

func TestUpdateProfileNameLimitCountsUTF8Runes(t *testing.T) {
	repository := NewRepository(nil)
	name := strings.Repeat("界", 121)

	if _, err := repository.UpdateProfile(context.Background(), "user-1", UpdateProfileInput{Name: name}); !errors.Is(err, ErrProfileNameTooLong) {
		t.Fatalf("expected ErrProfileNameTooLong for 121 runes, got %v", err)
	}
}

// TestCanManageOrganizationSuperadminBypassesDB ensures the superadmin role
// short-circuits the DB lookup. This is the only branch we can exercise
// without postgres, and it's the security-critical fast path used by every
// admin scope check that flows through ensureAdminOrganizationSlug*.
func TestCanManageOrganizationSuperadminBypassesDB(t *testing.T) {
	repo := &Repository{db: nil} // would panic if the DB path were reached
	allowed, err := repo.CanManageOrganization(context.Background(), User{ID: "u1", Role: "superadmin"}, "any-org")
	if err != nil {
		t.Fatalf("superadmin path returned error: %v", err)
	}
	if !allowed {
		t.Fatalf("superadmin must always be allowed")
	}
}

func TestCanManageOrganizationRoleMatrix(t *testing.T) {
	cases := []struct {
		role string
		want bool
	}{
		{OrganizationRoleOwner, true},
		{OrganizationRoleAdmin, true},
		{OrganizationRoleMember, false},
		{OrganizationRoleViewer, false},
		{" Owner ", true},
		{"unknown", false},
	}

	for _, tc := range cases {
		if got := CanManageOrganizationRole(tc.role); got != tc.want {
			t.Errorf("CanManageOrganizationRole(%q) = %v, want %v", tc.role, got, tc.want)
		}
	}
}

func TestManageOrganizationRolesMatchesMatrix(t *testing.T) {
	for _, role := range ManageOrganizationRoles() {
		if !CanManageOrganizationRole(role) {
			t.Fatalf("ManageOrganizationRoles includes non-manage role %q", role)
		}
	}
}

// TestRoleAssignAuditMetadataPinsOrganization is the regression bait for the
// audit gap that originally let role mutations land in audit_logs with
// organization_id = NULL (because the audit organizationID() resolver only
// pulls entity_id for "organization" entities and otherwise falls back to
// metadata.organization_id). If anyone removes the organization_id key from
// the role-assign metadata bag again, the audit listing endpoint would stop
// surfacing role mutations for the affected org and this test would fail.
func TestRoleAssignAuditMetadataPinsOrganization(t *testing.T) {
	metadata := roleAssignAuditMetadata("org-1", "user-1", "owner")

	if got := metadata["organization_id"]; got != "org-1" {
		t.Fatalf("organization_id missing or wrong: %v", got)
	}
	if got := metadata["user_id"]; got != "user-1" {
		t.Fatalf("user_id = %v, want user-1", got)
	}
	if got := metadata["role"]; got != "owner" {
		t.Fatalf("role = %v, want owner", got)
	}
}
