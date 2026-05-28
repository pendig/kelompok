package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/pendig/kelompok/internal/organizations"
)

func TestRunClaimRequiresSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runClaim(context.Background(), nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error when no subcommand is provided")
	}
	if !strings.Contains(err.Error(), "subcommand") {
		t.Fatalf("expected error message to mention 'subcommand', got %q", err.Error())
	}
}

func TestRunClaimUnknownSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runClaim(context.Background(), []string{"weird"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "unknown claim subcommand") {
		t.Fatalf("expected 'unknown claim subcommand', got %q", err.Error())
	}
}

func TestRunClaimUpdateStatusValidationErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"missing id", []string{"update-status", "--decision", "approve"}, "--id is required"},
		{"missing decision", []string{"update-status", "--id", "claim-1"}, "invalid claim decision"},
		{"bad decision", []string{"update-status", "--id", "claim-1", "--decision", "deny"}, "invalid claim decision"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := runClaim(context.Background(), tc.args, &stdout, &stderr)
			if err == nil {
				t.Fatalf("expected validation error, got nil; stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error to contain %q, got %q", tc.want, err.Error())
			}
			if stdout.Len() != 0 {
				t.Fatalf("expected no stdout on validation error, got %q", stdout.String())
			}
		})
	}
}

func TestRunClaimListBadStatusFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runClaim(context.Background(), []string{"list", "--status", "weird"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid claim status") {
		t.Fatalf("expected 'invalid claim status', got %q", err.Error())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout on validation error, got %q", stdout.String())
	}
}

func TestRunClaimListBadLimitFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runClaim(context.Background(), []string{"list", "--limit", "0"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected error for non-positive limit")
	}
	if !strings.Contains(err.Error(), "--limit must be greater than zero") {
		t.Fatalf("expected limit error, got %q", err.Error())
	}
}

func TestPrintClaimListTableEmpty(t *testing.T) {
	var stdout bytes.Buffer
	printClaimListTable(&stdout, nil)
	if stdout.Len() != 0 {
		t.Fatalf("expected empty table output for empty list, got %q", stdout.String())
	}
}

func TestPrintClaimListTableOrdering(t *testing.T) {
	created := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	reviewer := "user-7"
	reviewedAt := created.Add(time.Hour)
	items := []organizations.ClaimRequestWithOrganization{
		{
			ClaimRequest: organizations.ClaimRequest{
				ID:             "claim-1",
				OrganizationID: "org-1",
				UserID:         "user-1",
				Method:         "official_email",
				Target:         "ops@example.org",
				Status:         "pending",
				CreatedAt:      created,
				UpdatedAt:      created,
			},
			OrganizationSlug: "green-foundation",
			OrganizationName: "Green Foundation",
		},
		{
			ClaimRequest: organizations.ClaimRequest{
				ID:             "claim-2",
				OrganizationID: "org-2",
				UserID:         "user-2",
				Method:         "instagram",
				Target:         "@example",
				Status:         "approved",
				ReviewedByUser: &reviewer,
				ReviewedAt:     &reviewedAt,
				CreatedAt:      created,
				UpdatedAt:      created,
			},
			OrganizationSlug: "blue-foundation",
			OrganizationName: "Blue Foundation",
		},
	}

	var stdout bytes.Buffer
	printClaimListTable(&stdout, items)
	got := stdout.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d in %q", len(lines), got)
	}

	wantPrefix0 := "claim-1\tgreen-foundation\tpending\tofficial_email\tops@example.org\t2026-05-28T10:00:00Z\t-\t-"
	if lines[0] != wantPrefix0 {
		t.Fatalf("line 0:\n got: %q\nwant: %q", lines[0], wantPrefix0)
	}
	wantPrefix1 := "claim-2\tblue-foundation\tapproved\tinstagram\t@example\t2026-05-28T10:00:00Z\t2026-05-28T11:00:00Z\tuser-7"
	if lines[1] != wantPrefix1 {
		t.Fatalf("line 1:\n got: %q\nwant: %q", lines[1], wantPrefix1)
	}
}

func TestEmitClaimUpdateResultJSONShape(t *testing.T) {
	created := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	res := claimUpdateResult{
		DryRun:           true,
		Decision:         "approve",
		ReviewerUserID:   "user-7",
		WouldBecomeState: "approved",
		Claim: organizations.ClaimRequestWithOrganization{
			ClaimRequest: organizations.ClaimRequest{
				ID:             "claim-1",
				OrganizationID: "org-1",
				UserID:         "user-1",
				Method:         "official_email",
				Target:         "ops@example.org",
				Status:         "pending",
				CreatedAt:      created,
				UpdatedAt:      created,
			},
			OrganizationSlug: "green-foundation",
			OrganizationName: "Green Foundation",
		},
	}

	var stdout bytes.Buffer
	if err := emitClaimUpdateResult(&stdout, true, res); err != nil {
		t.Fatalf("emitClaimUpdateResult: %v", err)
	}

	// Must round-trip into the same shape (stable contract).
	var decoded claimUpdateResult
	if err := json.Unmarshal(stdout.Bytes(), &decoded); err != nil {
		t.Fatalf("decode JSON: %v\npayload=%s", err, stdout.String())
	}
	if !decoded.DryRun {
		t.Errorf("expected dry_run=true after roundtrip, got false: %s", stdout.String())
	}
	if decoded.Decision != "approve" {
		t.Errorf("decision = %q, want approve", decoded.Decision)
	}
	if decoded.ReviewerUserID != "user-7" {
		t.Errorf("reviewer_user_id = %q, want user-7", decoded.ReviewerUserID)
	}
	if decoded.WouldBecomeState != "approved" {
		t.Errorf("would_become_status = %q, want approved", decoded.WouldBecomeState)
	}
	if decoded.Claim.OrganizationSlug != "green-foundation" {
		t.Errorf("claim.organization_slug = %q, want green-foundation", decoded.Claim.OrganizationSlug)
	}
	if decoded.Claim.Status != "pending" {
		t.Errorf("claim.status = %q, want pending (current state preserved on dry-run)", decoded.Claim.Status)
	}

	// JSON keys must remain stable for downstream automation.
	for _, key := range []string{`"dry_run"`, `"decision"`, `"reviewer_user_id"`, `"would_become_status"`, `"claim"`, `"organization_slug"`, `"organization_name"`} {
		if !strings.Contains(stdout.String(), key) {
			t.Errorf("expected JSON to contain %s, got %s", key, stdout.String())
		}
	}
}

func TestEmitClaimUpdateResultHumanShape(t *testing.T) {
	res := claimUpdateResult{
		DryRun:           false,
		Decision:         "reject",
		ReviewerUserID:   "",
		WouldBecomeState: "rejected",
		Claim: organizations.ClaimRequestWithOrganization{
			ClaimRequest:     organizations.ClaimRequest{ID: "claim-9", Status: "rejected"},
			OrganizationSlug: "green-foundation",
			OrganizationName: "Green Foundation",
		},
	}

	var stdout bytes.Buffer
	if err := emitClaimUpdateResult(&stdout, false, res); err != nil {
		t.Fatalf("emitClaimUpdateResult: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{
		"claim: applied",
		"decision=reject",
		"would_become=rejected",
		"id=claim-9",
		"organization_slug=green-foundation",
		"current_status=rejected",
		"reviewer_user_id=-",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("human output missing %q in %q", want, got)
		}
	}
}

func TestEmitClaimUpdateResultHumanDryRunMode(t *testing.T) {
	res := claimUpdateResult{
		DryRun:           true,
		Decision:         "approve",
		WouldBecomeState: "approved",
		Claim: organizations.ClaimRequestWithOrganization{
			ClaimRequest:     organizations.ClaimRequest{ID: "claim-9", Status: "pending"},
			OrganizationSlug: "green-foundation",
		},
	}
	var stdout bytes.Buffer
	if err := emitClaimUpdateResult(&stdout, false, res); err != nil {
		t.Fatalf("emitClaimUpdateResult: %v", err)
	}
	if !strings.Contains(stdout.String(), "claim: dry-run") {
		t.Fatalf("expected 'claim: dry-run' in dry-run output, got %q", stdout.String())
	}
}

func TestClaimDecisionToStatus(t *testing.T) {
	if got := claimDecisionToStatus("approve"); got != "approved" {
		t.Errorf("approve -> %q, want approved", got)
	}
	if got := claimDecisionToStatus("reject"); got != "rejected" {
		t.Errorf("reject -> %q, want rejected", got)
	}
	// any non-approve falls through to rejected; this matches the binary
	// API surface (only approve|reject are valid inputs upstream).
	if got := claimDecisionToStatus(""); got != "rejected" {
		t.Errorf("empty -> %q, want rejected (default fallback)", got)
	}
}
