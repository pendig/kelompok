package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pendig/kelompok/internal/organizations"
)

// runClaim dispatches `kelompok claim <subcommand>` invocations.
//
// Supported subcommands:
//
//	list           — list claim requests with optional --status / --organization filters
//	pending        — convenience alias for `list --status pending`
//	update-status  — approve or reject a pending claim, with --dry-run support
//
// All mutating subcommands honor --dry-run; all subcommands honor --json for
// stable, machine-readable output. Validation, lookup, and DB errors are
// returned to the caller so the top-level Run() exits with a non-zero status.
func runClaim(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("claim command requires a subcommand (list, pending, update-status)")
	}

	switch args[0] {
	case "list":
		return runClaimList(ctx, args[1:], stdout, stderr, "")
	case "pending":
		return runClaimList(ctx, args[1:], stdout, stderr, "pending")
	case "update-status":
		return runClaimUpdateStatus(ctx, args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown claim subcommand: %s", args[0])
	}
}

// runClaimList implements `claim list` and `claim pending`. When forcedStatus
// is non-empty (e.g. for `claim pending`), it overrides any --status flag the
// user passes so the convenience alias has a single, predictable contract.
func runClaimList(ctx context.Context, args []string, stdout, stderr io.Writer, forcedStatus string) error {
	flags := flag.NewFlagSet("claim list", flag.ContinueOnError)
	flags.SetOutput(stderr)
	organizationSlug := flags.String("organization", "", "filter by organization slug")
	status := flags.String("status", "", "filter by claim status (pending, approved, rejected, all)")
	limit := flags.Int("limit", 50, fmt.Sprintf("maximum number of claims to return (max %d)", organizations.MaxClaimListLimit))
	jsonOut := flags.Bool("json", false, "print JSON output")
	if err := flags.Parse(args); err != nil {
		return err
	}

	normalizedLimit, err := organizations.NormalizeClaimListLimit(*limit)
	if err != nil {
		return fmt.Errorf("--%s", err)
	}

	effectiveStatus := *status
	if forcedStatus != "" {
		effectiveStatus = forcedStatus
	}
	normalizedStatus, err := organizations.NormalizeClaimStatus(effectiveStatus)
	if err != nil {
		return err
	}

	return withOrganizationRepository(ctx, func(repo *organizations.Repository) error {
		items, err := repo.ListClaims(ctx, organizations.ClaimListFilter{
			Status:           normalizedStatus,
			OrganizationSlug: strings.TrimSpace(*organizationSlug),
		}, normalizedLimit)
		if err != nil {
			return err
		}
		if *jsonOut {
			return writeJSONOutput(stdout, items)
		}
		printClaimListTable(stdout, items)
		return nil
	})
}

// runClaimUpdateStatus implements `claim update-status --id <id> --decision approve|reject`.
//
// Errors that can be surfaced (all of which trigger a non-zero CLI exit):
//   - --id missing or empty
//   - --decision missing or unrecognised
//   - claim_not_found        (organizations.ErrClaimNotFound)
//   - claim_not_pending      (organizations.ErrClaimNotPending)
//   - DB / lookup failures   (raw error from repository)
//
// With --dry-run the command performs all validations and emits the same
// payload it would emit on success, but with `dry_run=true` and no DB write.
func runClaimUpdateStatus(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("claim update-status", flag.ContinueOnError)
	flags.SetOutput(stderr)
	id := flags.String("id", "", "claim request id")
	decision := flags.String("decision", "", "decision: approve or reject")
	reviewerUserID := flags.String("reviewer-user-id", "", "reviewer user id (UUID); optional")
	dryRun := flags.Bool("dry-run", false, "validate without writing to the database")
	jsonOut := flags.Bool("json", false, "print JSON output")
	if err := flags.Parse(args); err != nil {
		return err
	}

	claimID := strings.TrimSpace(*id)
	if claimID == "" {
		return errors.New("--id is required")
	}
	normalizedDecision, err := organizations.NormalizeClaimDecision(*decision)
	if err != nil {
		return err
	}

	return withOrganizationRepository(ctx, func(repo *organizations.Repository) error {
		current, err := repo.FindClaimWithOrganizationByID(ctx, claimID)
		if err != nil {
			return err
		}
		if current.Status != "pending" {
			return fmt.Errorf("%w (current status %q)", organizations.ErrClaimNotPending, current.Status)
		}

		if *dryRun {
			return emitClaimUpdateResult(stdout, *jsonOut, claimUpdateResult{
				DryRun:           true,
				Decision:         normalizedDecision,
				ReviewerUserID:   strings.TrimSpace(*reviewerUserID),
				Claim:            current,
				WouldBecomeState: claimDecisionToStatus(normalizedDecision),
			})
		}

		var updated organizations.ClaimRequest
		if normalizedDecision == "approve" {
			updated, err = repo.ApproveClaim(ctx, claimID, strings.TrimSpace(*reviewerUserID))
		} else {
			updated, err = repo.RejectClaim(ctx, claimID, strings.TrimSpace(*reviewerUserID))
		}
		if err != nil {
			return err
		}

		return emitClaimUpdateResult(stdout, *jsonOut, claimUpdateResult{
			DryRun:           false,
			Decision:         normalizedDecision,
			ReviewerUserID:   strings.TrimSpace(*reviewerUserID),
			Claim:            organizations.ClaimRequestWithOrganization{ClaimRequest: updated, OrganizationSlug: current.OrganizationSlug, OrganizationName: current.OrganizationName},
			WouldBecomeState: updated.Status,
		})
	})
}

// claimUpdateResult is the stable JSON shape emitted by `claim update-status`.
// Field order is fixed to keep automation-friendly diffs stable across runs.
type claimUpdateResult struct {
	DryRun           bool                                       `json:"dry_run"`
	Decision         string                                     `json:"decision"`
	ReviewerUserID   string                                     `json:"reviewer_user_id,omitempty"`
	WouldBecomeState string                                     `json:"would_become_status"`
	Claim            organizations.ClaimRequestWithOrganization `json:"claim"`
}

func claimDecisionToStatus(decision string) string {
	if decision == "approve" {
		return "approved"
	}
	return "rejected"
}

// emitClaimUpdateResult renders update-status output in either JSON envelope
// form (--json) or a single-line human summary. Both shapes carry the same
// fields so automation can rely on either path.
func emitClaimUpdateResult(stdout io.Writer, jsonOut bool, result claimUpdateResult) error {
	if jsonOut {
		return writeJSONOutput(stdout, result)
	}
	mode := "applied"
	if result.DryRun {
		mode = "dry-run"
	}
	reviewer := result.ReviewerUserID
	if reviewer == "" {
		reviewer = "-"
	}
	fmt.Fprintf(stdout,
		"claim: %s decision=%s would_become=%s id=%s organization_slug=%s current_status=%s reviewer_user_id=%s\n",
		mode,
		result.Decision,
		result.WouldBecomeState,
		result.Claim.ID,
		result.Claim.OrganizationSlug,
		result.Claim.Status,
		reviewer,
	)
	return nil
}

// printClaimListTable renders a tab-separated, column-stable view of claims
// suitable for piping through column(1) or awk(1). When the list is empty,
// nothing is written so callers can detect "no rows" by checking output size
// or just relying on exit code 0 with empty stdout.
func printClaimListTable(stdout io.Writer, items []organizations.ClaimRequestWithOrganization) {
	for _, item := range items {
		reviewedAt := "-"
		if item.ReviewedAt != nil {
			reviewedAt = item.ReviewedAt.UTC().Format(time.RFC3339)
		}
		reviewer := "-"
		if item.ReviewedByUser != nil {
			reviewer = *item.ReviewedByUser
		}
		fmt.Fprintf(stdout,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.OrganizationSlug,
			item.Status,
			item.Method,
			item.Target,
			item.CreatedAt.UTC().Format(time.RFC3339),
			reviewedAt,
			reviewer,
		)
	}
}
