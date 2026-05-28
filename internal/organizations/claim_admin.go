package organizations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/pendig/kelompok/internal/jsonvalue"
)

// ClaimStatusAll is the sentinel filter value that means "any claim status".
const ClaimStatusAll = ""

// ClaimRequestWithOrganization wraps ClaimRequest with the organization slug
// and name so admin/CLI consumers can identify the parent organization without
// performing a second lookup. The embedded ClaimRequest preserves the existing
// JSON field shape, and the two new fields are appended at the end.
type ClaimRequestWithOrganization struct {
	ClaimRequest
	OrganizationSlug string `json:"organization_slug"`
	OrganizationName string `json:"organization_name"`
}

// ClaimListFilter narrows ListClaims results. Empty fields are not applied
// as predicates, so a zero-value filter returns claims across every
// organization regardless of status.
type ClaimListFilter struct {
	Status           string
	OrganizationSlug string
}

// NormalizeClaimStatus accepts the canonical claim status values used by the
// claim_requests table plus the "all" alias for "no status filter". It returns
// the canonical string ("" means no filter) or an error for any other value.
func NormalizeClaimStatus(value string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	switch trimmed {
	case "", "all", "any", "*":
		return ClaimStatusAll, nil
	case "pending", "approved", "rejected":
		return trimmed, nil
	default:
		return "", fmt.Errorf("invalid claim status %q (expected pending, approved, rejected, or all)", value)
	}
}

// ListClaims returns claim requests across organizations, optionally filtered
// by status and/or organization slug. Results are ordered by creation time
// (newest first) for stable, paginatable output.
func (r *Repository) ListClaims(ctx context.Context, filter ClaimListFilter, limit int) ([]ClaimRequestWithOrganization, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}

	status, err := NormalizeClaimStatus(filter.Status)
	if err != nil {
		return nil, err
	}
	organizationSlug := strings.TrimSpace(filter.OrganizationSlug)

	rows, err := r.db.Query(ctx, `
		SELECT
			cr.id::text,
			cr.organization_id::text,
			cr.user_id::text,
			cr.method,
			cr.target,
			cr.status,
			COALESCE(cr.evidence::text, '{}'),
			cr.reviewed_by_user_id::text,
			cr.reviewed_at,
			cr.created_at,
			cr.updated_at,
			o.slug,
			o.name
		FROM claim_requests cr
		JOIN organizations o ON o.id = cr.organization_id
		WHERE ($1 = '' OR cr.status = $1)
			AND ($2 = '' OR o.slug = $2)
		ORDER BY cr.created_at DESC
		LIMIT $3
	`, status, organizationSlug, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ClaimRequestWithOrganization, 0, limit)
	for rows.Next() {
		var item ClaimRequestWithOrganization
		var evidence string
		var reviewedByUser sql.NullString
		var reviewedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.OrganizationID,
			&item.UserID,
			&item.Method,
			&item.Target,
			&item.Status,
			&evidence,
			&reviewedByUser,
			&reviewedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.OrganizationSlug,
			&item.OrganizationName,
		); err != nil {
			return nil, err
		}
		item.Evidence = jsonvalue.Raw(evidence, "{}")
		if reviewedByUser.Valid {
			value := reviewedByUser.String
			item.ReviewedByUser = &value
		}
		if reviewedAt.Valid {
			value := reviewedAt.Time
			item.ReviewedAt = &value
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// FindClaimWithOrganizationByID returns a single claim with the parent
// organization's slug and name attached. It is intended for admin/CLI flows
// that surface organization context alongside the claim before mutating it.
func (r *Repository) FindClaimWithOrganizationByID(ctx context.Context, claimID string) (ClaimRequestWithOrganization, error) {
	claim, organizationSlug, err := r.FindClaimByID(ctx, claimID)
	if err != nil {
		return ClaimRequestWithOrganization{}, err
	}

	var organizationName string
	if err := r.db.QueryRow(ctx, `SELECT name FROM organizations WHERE slug = $1`, organizationSlug).Scan(&organizationName); err != nil {
		return ClaimRequestWithOrganization{}, err
	}

	return ClaimRequestWithOrganization{
		ClaimRequest:     claim,
		OrganizationSlug: organizationSlug,
		OrganizationName: organizationName,
	}, nil
}

// NormalizeClaimDecision accepts approve/reject (and the past-tense forms
// approved/rejected) and returns the canonical action verb used by the
// repository review methods. Anything else is rejected.
func NormalizeClaimDecision(value string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	switch trimmed {
	case "approve", "approved":
		return "approve", nil
	case "reject", "rejected":
		return "reject", nil
	default:
		return "", fmt.Errorf("invalid claim decision %q (expected approve or reject)", value)
	}
}
