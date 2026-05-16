package organizations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pendig/kelompok/internal/audit"
	"github.com/pendig/kelompok/internal/jsonvalue"
)

var (
	ErrClaimNotFound   = errors.New("claim request not found")
	ErrClaimNotPending = errors.New("claim request is not pending")
)

type AdminInput struct {
	Slug          string          `json:"slug"`
	Name          string          `json:"name"`
	LegalName     string          `json:"legal_name"`
	Description   string          `json:"description"`
	History       string          `json:"history"`
	Country       string          `json:"country"`
	Region        string          `json:"region"`
	City          string          `json:"city"`
	WebsiteURL    string          `json:"website_url"`
	OfficialEmail string          `json:"official_email"`
	ClaimStatus   string          `json:"claim_status"`
	ProfileData   json.RawMessage `json:"profile_data"`
	SDGSData      json.RawMessage `json:"sdgs_data"`
	ImpactData    json.RawMessage `json:"impact_data"`
}

type ClaimInput struct {
	Method         string          `json:"method"`
	Target         string          `json:"target"`
	RequesterEmail string          `json:"requester_email"`
	Evidence       json.RawMessage `json:"evidence"`
}

type ClaimRequest struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organization_id"`
	UserID         string          `json:"user_id"`
	Method         string          `json:"method"`
	Target         string          `json:"target"`
	Status         string          `json:"status"`
	Evidence       json.RawMessage `json:"evidence"`
	ReviewedByUser *string         `json:"reviewed_by_user_id,omitempty"`
	ReviewedAt     *time.Time      `json:"reviewed_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (r *Repository) Create(ctx context.Context, input AdminInput) (Organization, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Organization{}, errors.New("organization name is required")
	}

	slug := normalizeSlug(input.Slug)
	if slug == "" {
		slug = normalizeSlug(name)
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO organizations (
			slug,
			name,
			legal_name,
			description,
			history,
			country,
			region,
			city,
			website_url,
			official_email,
			claim_status,
			profile_data,
			sdgs_data,
			impact_data
		)
		VALUES (
			$1,
			$2,
			NULLIF($3, ''),
			NULLIF($4, ''),
			NULLIF($5, ''),
			NULLIF($6, ''),
			NULLIF($7, ''),
			NULLIF($8, ''),
			NULLIF($9, ''),
			NULLIF($10, ''),
			COALESCE(NULLIF($11, ''), 'unclaimed'),
			$12::jsonb,
			$13::jsonb,
			$14::jsonb
		)
		RETURNING
			id::text,
			slug,
			name,
			COALESCE(legal_name, ''),
			COALESCE(description, ''),
			COALESCE(history, ''),
			COALESCE(country, ''),
			COALESCE(region, ''),
			COALESCE(city, ''),
			COALESCE(website_url, ''),
			COALESCE(official_email, ''),
			claim_status,
			COALESCE(profile_data::text, '{}'),
			COALESCE(sdgs_data::text, '{}'),
			COALESCE(impact_data::text, '{}'),
			created_at,
			updated_at
	`,
		slug,
		name,
		input.LegalName,
		input.Description,
		input.History,
		input.Country,
		input.Region,
		input.City,
		input.WebsiteURL,
		input.OfficialEmail,
		input.ClaimStatus,
		jsonOrFallback(input.ProfileData),
		jsonOrFallback(input.SDGSData),
		jsonOrFallback(input.ImpactData),
	)

	item, err := scanOrganization(row)
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "organization", item.ID, "create", nil, item, nil)
	}
	return item, err
}

func (r *Repository) UpdateBySlug(ctx context.Context, slug string, input AdminInput) (Organization, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE organizations
		SET
			slug = COALESCE(NULLIF($2, ''), slug),
			name = $3,
			legal_name = NULLIF($4, ''),
			description = NULLIF($5, ''),
			history = NULLIF($6, ''),
			country = NULLIF($7, ''),
			region = NULLIF($8, ''),
			city = NULLIF($9, ''),
			website_url = NULLIF($10, ''),
			official_email = NULLIF($11, ''),
			claim_status = COALESCE(NULLIF($12, ''), claim_status),
			profile_data = $13::jsonb,
			sdgs_data = $14::jsonb,
			impact_data = $15::jsonb,
			updated_at = now()
		WHERE slug = $1
		RETURNING
			id::text,
			slug,
			name,
			COALESCE(legal_name, ''),
			COALESCE(description, ''),
			COALESCE(history, ''),
			COALESCE(country, ''),
			COALESCE(region, ''),
			COALESCE(city, ''),
			COALESCE(website_url, ''),
			COALESCE(official_email, ''),
			claim_status,
			COALESCE(profile_data::text, '{}'),
			COALESCE(sdgs_data::text, '{}'),
			COALESCE(impact_data::text, '{}'),
			created_at,
			updated_at
	`,
		slug,
		normalizeSlug(input.Slug),
		strings.TrimSpace(input.Name),
		input.LegalName,
		input.Description,
		input.History,
		input.Country,
		input.Region,
		input.City,
		input.WebsiteURL,
		input.OfficialEmail,
		input.ClaimStatus,
		jsonOrFallback(input.ProfileData),
		jsonOrFallback(input.SDGSData),
		jsonOrFallback(input.ImpactData),
	)

	item, err := scanOrganization(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Organization{}, ErrNotFound
	}
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "organization", item.ID, "update", nil, item, nil)
	}
	return item, err
}

func (r *Repository) CreateClaim(ctx context.Context, organizationSlug string, input ClaimInput) (ClaimRequest, error) {
	organization, err := r.FindBySlug(ctx, organizationSlug)
	if err != nil {
		return ClaimRequest{}, err
	}

	method := strings.TrimSpace(input.Method)
	if method != "official_email" && method != "instagram" {
		return ClaimRequest{}, errors.New("unsupported claim method")
	}

	target := strings.TrimSpace(input.Target)
	if target == "" {
		return ClaimRequest{}, errors.New("claim target is required")
	}

	requesterEmail := strings.TrimSpace(input.RequesterEmail)
	if requesterEmail == "" {
		requesterEmail = target
	}

	userID, err := r.ensureDemoUser(ctx, requesterEmail)
	if err != nil {
		return ClaimRequest{}, err
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO claim_requests (
			organization_id,
			user_id,
			method,
			target,
			status,
			evidence
		)
		VALUES ($1, $2, $3, $4, 'pending', $5::jsonb)
		RETURNING
			id::text,
			organization_id::text,
			user_id::text,
			method,
			target,
			status,
			COALESCE(evidence::text, '{}'),
			reviewed_by_user_id::text,
			reviewed_at,
			created_at,
			updated_at
	`,
		organization.ID,
		userID,
		method,
		target,
		jsonOrFallback(input.Evidence),
	)

	var item ClaimRequest
	var evidence string
	var reviewedByUser sql.NullString
	var reviewedAt sql.NullTime
	if err := row.Scan(
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
	); err != nil {
		return ClaimRequest{}, err
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

	_ = audit.Record(ctx, r.db, nil, "claim_request", item.ID, "create", nil, item, map[string]any{
		"organization_id":   organization.ID,
		"organization_slug": organizationSlug,
	})
	return item, nil
}

func (r *Repository) ListClaimsByOrganizationSlug(ctx context.Context, organizationSlug string, limit int) ([]ClaimRequest, error) {
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
			cr.updated_at
		FROM claim_requests cr
		JOIN organizations o ON o.id = cr.organization_id
		WHERE o.slug = $1
		ORDER BY cr.created_at DESC
		LIMIT $2
	`, organizationSlug, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ClaimRequest, 0, limit)
	for rows.Next() {
		var item ClaimRequest
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

func (r *Repository) FindClaimByID(ctx context.Context, claimID string) (ClaimRequest, string, error) {
	row := r.db.QueryRow(ctx, `
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
			o.slug
		FROM claim_requests cr
		JOIN organizations o ON o.id = cr.organization_id
		WHERE cr.id = $1
	`, claimID)

	var organizationSlug string
	item, err := scanClaimWithSlug(row, &organizationSlug)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClaimRequest{}, "", ErrClaimNotFound
	}
	return item, organizationSlug, err
}

func (r *Repository) ApproveClaim(ctx context.Context, claimID string, reviewerUserID string) (ClaimRequest, error) {
	return r.reviewClaim(ctx, claimID, reviewerUserID, "approved")
}

func (r *Repository) RejectClaim(ctx context.Context, claimID string, reviewerUserID string) (ClaimRequest, error) {
	return r.reviewClaim(ctx, claimID, reviewerUserID, "rejected")
}

func (r *Repository) reviewClaim(ctx context.Context, claimID string, reviewerUserID string, status string) (ClaimRequest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ClaimRequest{}, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		UPDATE claim_requests
		SET
			status = $2,
			reviewed_by_user_id = NULLIF($3, '')::uuid,
			reviewed_at = now(),
			updated_at = now()
		WHERE id = $1
			AND status = 'pending'
		RETURNING
			id::text,
			organization_id::text,
			user_id::text,
			method,
			target,
			status,
			COALESCE(evidence::text, '{}'),
			reviewed_by_user_id::text,
			reviewed_at,
			created_at,
			updated_at
	`, claimID, status, reviewerUserID)

	item, err := scanClaim(row)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		if lookupErr := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM claim_requests WHERE id = $1)`, claimID).Scan(&exists); lookupErr != nil {
			return ClaimRequest{}, lookupErr
		}
		if exists {
			return ClaimRequest{}, ErrClaimNotPending
		}
		return ClaimRequest{}, ErrClaimNotFound
	}
	if err != nil {
		return ClaimRequest{}, err
	}

	if status == "approved" {
		tag, err := tx.Exec(ctx, `
			UPDATE organizations
			SET
				claim_status = 'claimed',
				claimed_by_user_id = $2,
				claimed_at = now(),
				updated_at = now()
			WHERE id = $1
				AND claim_status <> 'claimed'
		`, item.OrganizationID, item.UserID)
		if err != nil {
			return ClaimRequest{}, err
		}
		if tag.RowsAffected() == 0 {
			return ClaimRequest{}, ErrClaimNotPending
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO organization_user_roles (organization_id, user_id, role)
			VALUES ($1, $2, 'owner')
			ON CONFLICT (organization_id, user_id) DO UPDATE SET
				role = 'owner',
				updated_at = now()
		`, item.OrganizationID, item.UserID); err != nil {
			return ClaimRequest{}, err
		}
	} else if _, err := tx.Exec(ctx, `
		UPDATE organizations
		SET
			claim_status = 'rejected',
			updated_at = now()
		WHERE id = $1
			AND claim_status <> 'claimed'
	`, item.OrganizationID); err != nil {
		return ClaimRequest{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return ClaimRequest{}, err
	}

	_ = audit.Record(ctx, r.db, reviewerUserID, "claim_request", item.ID, status, nil, item, map[string]any{
		"organization_id": item.OrganizationID,
		"user_id":         item.UserID,
	})
	return item, nil
}

func (r *Repository) ensureDemoUser(ctx context.Context, email string) (string, error) {
	var userID string
	if err := r.db.QueryRow(ctx, `
		INSERT INTO users (name, email, role)
		VALUES (split_part($1, '@', 1), $1, 'viewer')
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = now()
		RETURNING id::text
	`, strings.TrimSpace(email)).Scan(&userID); err != nil {
		return "", err
	}

	return userID, nil
}

func scanClaim(row interface{ Scan(dest ...any) error }) (ClaimRequest, error) {
	return scanClaimWithSlug(row, nil)
}

func scanClaimWithSlug(row interface{ Scan(dest ...any) error }, organizationSlug *string) (ClaimRequest, error) {
	var item ClaimRequest
	var evidence string
	var reviewedByUser sql.NullString
	var reviewedAt sql.NullTime
	dest := []any{
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
	}
	if organizationSlug != nil {
		dest = append(dest, organizationSlug)
	}
	if err := row.Scan(dest...); err != nil {
		return ClaimRequest{}, err
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
	return item, nil
}

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func normalizeSlug(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	trimmed = slugPattern.ReplaceAllString(trimmed, "-")
	trimmed = strings.Trim(trimmed, "-")
	return trimmed
}

func jsonOrFallback(value json.RawMessage) any {
	if len(value) == 0 {
		return "{}"
	}
	return string(value)
}
