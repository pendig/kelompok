package organizations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pendig/kelompok/internal/audit"
	"github.com/pendig/kelompok/internal/jsonvalue"
)

var (
	ErrClaimNotFound   = errors.New("claim request not found")
	ErrClaimNotPending = errors.New("claim request is not pending")
)

var (
	ErrOrganizationNameRequired         = errors.New("organization name is required")
	ErrOrganizationSlugRequired         = errors.New("organization slug is required")
	ErrOrganizationClaimStatusInvalid   = errors.New("organization claim_status is invalid")
	ErrOrganizationOfficialEmailInvalid = errors.New("organization official_email is invalid")
	ErrOrganizationJSONInvalid          = errors.New("organization JSON field is invalid")
	ErrClaimMethodInvalid               = errors.New("claim method is invalid")
	ErrClaimTargetRequired              = errors.New("claim target is required")
	ErrClaimTargetInvalid               = errors.New("claim target is invalid")
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
	SourceData    json.RawMessage `json:"source_data"`
	SDGSData      json.RawMessage `json:"sdgs_data"`
	ImpactData    json.RawMessage `json:"impact_data"`
}

type ClaimInput struct {
	Method         string          `json:"method"`
	Target         string          `json:"target"`
	RequesterEmail string          `json:"requester_email"`
	Evidence       json.RawMessage `json:"evidence"`
}

type OnboardingRequestInput struct {
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
	ProfileData   json.RawMessage `json:"profile_data"`
	SourceData    json.RawMessage `json:"source_data"`
	SDGSData      json.RawMessage `json:"sdgs_data"`
	ImpactData    json.RawMessage `json:"impact_data"`
	Method        string          `json:"method"`
	Target        string          `json:"target"`
	Evidence      json.RawMessage `json:"evidence"`
}

type OnboardingRequest struct {
	Organization Organization `json:"organization"`
	Claim        ClaimRequest `json:"claim"`
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

type RelatedOrganizationInput struct {
	AdminInput
	RelationshipType     string          `json:"relationship_type"`
	RelationshipLabel    string          `json:"relationship_label"`
	RelationshipStatus   string          `json:"relationship_status"`
	RelationshipMetadata json.RawMessage `json:"relationship_metadata"`
}

type RelatedOrganizationResult struct {
	Organization Organization `json:"organization"`
	Relationship Relationship `json:"relationship"`
}

func (r *Repository) Create(ctx context.Context, input AdminInput) (Organization, error) {
	normalized, err := NormalizeAdminInput(input)
	if err != nil {
		return Organization{}, err
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
			source_data,
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
			$14::jsonb,
			$15::jsonb
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
			COALESCE(source_data::text, '{}'),
			COALESCE(sdgs_data::text, '{}'),
			COALESCE(impact_data::text, '{}'),
			created_at,
			updated_at
	`,
		normalized.Slug,
		normalized.Name,
		normalized.LegalName,
		normalized.Description,
		normalized.History,
		normalized.Country,
		normalized.Region,
		normalized.City,
		normalized.WebsiteURL,
		normalized.OfficialEmail,
		normalized.ClaimStatus,
		jsonOrFallback(normalized.ProfileData),
		jsonOrFallback(normalized.SourceData),
		jsonOrFallback(normalized.SDGSData),
		jsonOrFallback(normalized.ImpactData),
	)

	item, err := scanOrganization(row)
	if isSlugConflict(err) {
		return Organization{}, ErrSlugTaken
	}
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "organization", item.ID, "create", nil, item, nil)
	}
	return item, err
}

func (r *Repository) CreateRelatedOrganization(ctx context.Context, parentSlug string, input RelatedOrganizationInput, actor AuditActor) (RelatedOrganizationResult, error) {
	organizationInput := input.AdminInput
	if strings.TrimSpace(organizationInput.ClaimStatus) == "" {
		organizationInput.ClaimStatus = "unclaimed"
	}
	normalizedOrganization, err := NormalizeAdminInput(organizationInput)
	if err != nil {
		return RelatedOrganizationResult{}, err
	}

	relationshipInput := RelationshipInput{
		ParentOrganizationSlug: parentSlug,
		ChildOrganizationSlug:  normalizedOrganization.Slug,
		RelationshipType:       input.RelationshipType,
		Label:                  input.RelationshipLabel,
		Status:                 input.RelationshipStatus,
		Metadata:               input.RelationshipMetadata,
	}
	if strings.TrimSpace(relationshipInput.RelationshipType) == "" {
		relationshipInput.RelationshipType = "structural_parent"
	}
	if strings.TrimSpace(relationshipInput.Status) == "" {
		relationshipInput.Status = "active"
	}
	normalizedRelationship, err := normalizeRelationshipInput(relationshipInput)
	if err != nil {
		return RelatedOrganizationResult{}, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return RelatedOrganizationResult{}, err
	}
	defer tx.Rollback(ctx)

	var parentID string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM organizations WHERE slug = $1`, normalizedRelationship.ParentOrganizationSlug).Scan(&parentID); errors.Is(err, pgx.ErrNoRows) {
		return RelatedOrganizationResult{}, ErrNotFound
	} else if err != nil {
		return RelatedOrganizationResult{}, err
	}

	organization, err := scanOrganization(tx.QueryRow(ctx, `
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
			source_data,
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
			$14::jsonb,
			$15::jsonb
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
			COALESCE(source_data::text, '{}'),
			COALESCE(sdgs_data::text, '{}'),
			COALESCE(impact_data::text, '{}'),
			created_at,
			updated_at
	`,
		normalizedOrganization.Slug,
		normalizedOrganization.Name,
		normalizedOrganization.LegalName,
		normalizedOrganization.Description,
		normalizedOrganization.History,
		normalizedOrganization.Country,
		normalizedOrganization.Region,
		normalizedOrganization.City,
		normalizedOrganization.WebsiteURL,
		normalizedOrganization.OfficialEmail,
		normalizedOrganization.ClaimStatus,
		jsonOrFallback(normalizedOrganization.ProfileData),
		jsonOrFallback(normalizedOrganization.SourceData),
		jsonOrFallback(normalizedOrganization.SDGSData),
		jsonOrFallback(normalizedOrganization.ImpactData),
	))
	if err != nil {
		if isSlugConflict(err) {
			return RelatedOrganizationResult{}, ErrSlugTaken
		}
		return RelatedOrganizationResult{}, err
	}

	relationship, err := scanRelationship(tx.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO organization_relationships (
				parent_organization_id,
				child_organization_id,
				relationship_type,
				label,
				status,
				metadata
			)
			VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6::jsonb)
			RETURNING *
		)
		SELECT
			rel.id::text,
			parent.id::text,
			parent.slug,
			parent.name,
			child.id::text,
			child.slug,
			child.name,
			rel.relationship_type,
			COALESCE(rel.label, ''),
			rel.status,
			rel.started_at,
			rel.ended_at,
			COALESCE(rel.metadata::text, '{}'),
			rel.created_at,
			rel.updated_at
		FROM inserted rel
		JOIN organizations parent ON parent.id = rel.parent_organization_id
		JOIN organizations child ON child.id = rel.child_organization_id
	`, parentID, organization.ID, normalizedRelationship.RelationshipType, normalizedRelationship.Label, normalizedRelationship.Status, jsonOrFallback(normalizedRelationship.Metadata)))
	if err != nil {
		return RelatedOrganizationResult{}, relationshipWriteError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return RelatedOrganizationResult{}, err
	}

	_ = audit.Record(ctx, r.db, actor.UserID, "organization", organization.ID, "create_related", nil, organization, map[string]any{
		"organization_id":          organization.ID,
		"parent_organization_id":   parentID,
		"parent_organization_slug": normalizedRelationship.ParentOrganizationSlug,
	})
	r.recordRelationshipAudit(ctx, actor, "create", nil, relationship, relationship)
	return RelatedOrganizationResult{Organization: organization, Relationship: relationship}, nil
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
			source_data = $14::jsonb,
			sdgs_data = $15::jsonb,
			impact_data = $16::jsonb,
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
			COALESCE(source_data::text, '{}'),
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
		jsonOrFallback(input.SourceData),
		jsonOrFallback(input.SDGSData),
		jsonOrFallback(input.ImpactData),
	)

	item, err := scanOrganization(row)
	if isSlugConflict(err) {
		return Organization{}, ErrSlugTaken
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Organization{}, ErrNotFound
	}
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "organization", item.ID, "update", nil, item, nil)
	}
	return item, err
}

func isSlugConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "slug")
}

func (r *Repository) CreateClaim(ctx context.Context, organizationSlug string, input ClaimInput) (ClaimRequest, error) {
	organization, err := r.FindBySlug(ctx, organizationSlug)
	if err != nil {
		return ClaimRequest{}, err
	}

	method, target, err := NormalizeClaimEvidenceInput(input.Method, input.Target)
	if err != nil {
		return ClaimRequest{}, err
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

func (r *Repository) CreateOnboardingRequest(ctx context.Context, userID string, input OnboardingRequestInput) (OnboardingRequest, error) {
	normalizedOrganization, err := NormalizeAdminInput(AdminInput{
		Slug:          input.Slug,
		Name:          input.Name,
		LegalName:     input.LegalName,
		Description:   input.Description,
		History:       input.History,
		Country:       input.Country,
		Region:        input.Region,
		City:          input.City,
		WebsiteURL:    input.WebsiteURL,
		OfficialEmail: input.OfficialEmail,
		ClaimStatus:   "pending",
		ProfileData:   input.ProfileData,
		SourceData:    input.SourceData,
		SDGSData:      input.SDGSData,
		ImpactData:    input.ImpactData,
	})
	if err != nil {
		return OnboardingRequest{}, err
	}

	method, target, err := NormalizeOnboardingClaimEvidenceInput(input.Method, input.Target)
	if err != nil {
		return OnboardingRequest{}, err
	}
	if len(input.Evidence) > 0 && !validJSONObject(input.Evidence) {
		return OnboardingRequest{}, fmt.Errorf("%w: evidence", ErrOrganizationJSONInvalid)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return OnboardingRequest{}, err
	}
	defer tx.Rollback(ctx)

	organization, err := scanOrganization(tx.QueryRow(ctx, `
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
			source_data,
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
			'pending',
			$11::jsonb,
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
			COALESCE(source_data::text, '{}'),
			COALESCE(sdgs_data::text, '{}'),
			COALESCE(impact_data::text, '{}'),
			created_at,
			updated_at
	`,
		normalizedOrganization.Slug,
		normalizedOrganization.Name,
		normalizedOrganization.LegalName,
		normalizedOrganization.Description,
		normalizedOrganization.History,
		normalizedOrganization.Country,
		normalizedOrganization.Region,
		normalizedOrganization.City,
		normalizedOrganization.WebsiteURL,
		normalizedOrganization.OfficialEmail,
		jsonOrFallback(normalizedOrganization.ProfileData),
		jsonOrFallback(normalizedOrganization.SourceData),
		jsonOrFallback(normalizedOrganization.SDGSData),
		jsonOrFallback(normalizedOrganization.ImpactData),
	))
	if isSlugConflict(err) {
		return OnboardingRequest{}, ErrSlugTaken
	}
	if err != nil {
		return OnboardingRequest{}, err
	}

	claim, err := scanClaim(tx.QueryRow(ctx, `
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
	`, organization.ID, userID, method, target, jsonOrFallback(input.Evidence)))
	if err != nil {
		return OnboardingRequest{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return OnboardingRequest{}, err
	}

	_ = audit.Record(ctx, r.db, userID, "organization", organization.ID, "onboarding_request", nil, organization, map[string]any{
		"claim_request_id": claim.ID,
	})
	_ = audit.Record(ctx, r.db, userID, "claim_request", claim.ID, "create", nil, claim, map[string]any{
		"organization_id":   organization.ID,
		"organization_slug": organization.Slug,
	})
	return OnboardingRequest{Organization: organization, Claim: claim}, nil
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

func (r *Repository) ListDelegatedClaimsByReviewerOrganizationSlug(ctx context.Context, reviewerOrganizationSlug string, limit int) ([]ClaimRequestWithOrganization, error) {
	normalizedLimit, err := NormalizeClaimListLimit(limit)
	if err != nil {
		return nil, err
	}

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
			child.slug,
			child.name
		FROM claim_requests cr
		JOIN organizations child ON child.id = cr.organization_id
		JOIN organization_relationships rel ON rel.child_organization_id = child.id
		JOIN organizations parent ON parent.id = rel.parent_organization_id
		WHERE parent.slug = $1
			AND rel.status = 'active'
			AND cr.status = 'pending'
		ORDER BY cr.created_at DESC
		LIMIT $2
	`, normalizeSlug(reviewerOrganizationSlug), normalizedLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ClaimRequestWithOrganization, 0, min(normalizedLimit, claimListPreallocLimit))
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
	return NormalizeSlug(value)
}

func NormalizeSlug(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	trimmed = slugPattern.ReplaceAllString(trimmed, "-")
	trimmed = strings.Trim(trimmed, "-")
	return trimmed
}

// NormalizeAdminInput validates the create/update payload shape shared by
// admin organization endpoints before it reaches SQL constraints.
func NormalizeAdminInput(input AdminInput) (AdminInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return AdminInput{}, ErrOrganizationNameRequired
	}

	input.Slug = NormalizeSlug(input.Slug)
	if input.Slug == "" {
		input.Slug = NormalizeSlug(input.Name)
	}
	if input.Slug == "" {
		return AdminInput{}, ErrOrganizationSlugRequired
	}

	input.ClaimStatus = strings.ToLower(strings.TrimSpace(input.ClaimStatus))
	if input.ClaimStatus == "" {
		input.ClaimStatus = "unclaimed"
	}
	switch input.ClaimStatus {
	case "unclaimed", "pending", "claimed", "rejected":
	default:
		return AdminInput{}, fmt.Errorf("%w: %s", ErrOrganizationClaimStatusInvalid, input.ClaimStatus)
	}

	input.OfficialEmail = strings.TrimSpace(input.OfficialEmail)
	if input.OfficialEmail != "" {
		if _, err := mail.ParseAddress(input.OfficialEmail); err != nil {
			return AdminInput{}, fmt.Errorf("%w: %s", ErrOrganizationOfficialEmailInvalid, err)
		}
	}

	for name, value := range map[string]json.RawMessage{
		"profile_data": input.ProfileData,
		"source_data":  input.SourceData,
		"sdgs_data":    input.SDGSData,
		"impact_data":  input.ImpactData,
	} {
		if len(value) > 0 && !validJSONObject(value) {
			return AdminInput{}, fmt.Errorf("%w: %s", ErrOrganizationJSONInvalid, name)
		}
	}

	return input, nil
}

func NormalizeClaimEvidenceInput(method string, target string) (string, string, error) {
	return normalizeClaimEvidenceInput(method, target, false)
}

func NormalizeOnboardingClaimEvidenceInput(method string, target string) (string, string, error) {
	return normalizeClaimEvidenceInput(method, target, true)
}

func normalizeClaimEvidenceInput(method string, target string, allowManualReview bool) (string, string, error) {
	method = strings.ToLower(strings.TrimSpace(method))
	switch method {
	case "official_email", "instagram":
	case "manual_review":
		if !allowManualReview {
			return "", "", fmt.Errorf("%w: %s", ErrClaimMethodInvalid, method)
		}
	default:
		return "", "", fmt.Errorf("%w: %s", ErrClaimMethodInvalid, method)
	}

	target = strings.TrimSpace(target)
	if target == "" {
		return "", "", ErrClaimTargetRequired
	}
	if method == "official_email" {
		if _, err := mail.ParseAddress(target); err != nil {
			return "", "", fmt.Errorf("%w: %s", ErrClaimTargetInvalid, err)
		}
	}

	return method, target, nil
}

func validJSONObject(value json.RawMessage) bool {
	var object map[string]any
	return json.Unmarshal(value, &object) == nil
}

func jsonOrFallback(value json.RawMessage) any {
	if len(value) == 0 {
		return "{}"
	}
	return string(value)
}
