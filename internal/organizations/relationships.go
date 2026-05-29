package organizations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pendig/kelompok/internal/audit"
	"github.com/pendig/kelompok/internal/jsonvalue"
)

var (
	ErrRelationshipNotFound             = errors.New("organization relationship not found")
	ErrRelationshipOrganizationNotFound = errors.New("parent or child organization not found")
	ErrRelationshipDuplicate            = errors.New("organization relationship already exists")
	ErrRelationshipSelfLink             = errors.New("organization relationship cannot link an organization to itself")
)

type OrganizationRef struct {
	ID   string `json:"id,omitempty"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Relationship struct {
	ID                   string          `json:"id"`
	ParentOrganizationID string          `json:"parent_organization_id"`
	Parent               OrganizationRef `json:"parent"`
	ChildOrganizationID  string          `json:"child_organization_id"`
	Child                OrganizationRef `json:"child"`
	RelationshipType     string          `json:"relationship_type"`
	Label                string          `json:"label,omitempty"`
	Status               string          `json:"status"`
	StartedAt            *time.Time      `json:"started_at,omitempty"`
	EndedAt              *time.Time      `json:"ended_at,omitempty"`
	Metadata             json.RawMessage `json:"metadata"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

type RelationshipInput struct {
	ParentOrganizationSlug string          `json:"parent_organization_slug"`
	ChildOrganizationSlug  string          `json:"child_organization_slug"`
	RelationshipType       string          `json:"relationship_type"`
	Label                  string          `json:"label"`
	Status                 string          `json:"status"`
	StartedAt              *time.Time      `json:"started_at"`
	EndedAt                *time.Time      `json:"ended_at"`
	Metadata               json.RawMessage `json:"metadata"`
	ClearStartedAt         bool            `json:"-"`
	ClearEndedAt           bool            `json:"-"`
}

type AuditActor struct {
	UserID string
	Type   string
}

type normalizedRelationshipInput struct {
	ParentOrganizationSlug string
	ChildOrganizationSlug  string
	RelationshipType       string
	Label                  string
	Status                 string
	StartedAt              *time.Time
	EndedAt                *time.Time
	Metadata               json.RawMessage
}

var allowedRelationshipTypes = map[string]struct{}{
	"structural_parent": {},
	"autonomous_body":   {},
	"affiliated_with":   {},
	"network_member":    {},
	"related":           {},
}

var allowedRelationshipStatuses = map[string]struct{}{
	"active":   {},
	"pending":  {},
	"inactive": {},
	"archived": {},
}

func (r *Repository) ListRelationshipsByOrganizationSlug(ctx context.Context, organizationSlug string, limit int) ([]Relationship, error) {
	return r.listRelationshipsByOrganizationSlug(ctx, organizationSlug, limit, false)
}

func (r *Repository) ListActiveRelationshipsByOrganizationSlug(ctx context.Context, organizationSlug string, limit int) ([]Relationship, error) {
	return r.listRelationshipsByOrganizationSlug(ctx, organizationSlug, limit, true)
}

func (r *Repository) HasActiveParentRelationship(ctx context.Context, parentSlug string, childSlug string) (bool, error) {
	var allowed bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM organization_relationships rel
			JOIN organizations parent ON parent.id = rel.parent_organization_id
			JOIN organizations child ON child.id = rel.child_organization_id
			WHERE parent.slug = $1
				AND child.slug = $2
				AND rel.status = 'active'
		)
	`, normalizeSlug(parentSlug), normalizeSlug(childSlug)).Scan(&allowed); err != nil {
		return false, err
	}
	return allowed, nil
}

func (r *Repository) listRelationshipsByOrganizationSlug(ctx context.Context, organizationSlug string, limit int, activeOnly bool) ([]Relationship, error) {
	query := `
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
		FROM organization_relationships rel
		JOIN organizations parent ON parent.id = rel.parent_organization_id
		JOIN organizations child ON child.id = rel.child_organization_id
		WHERE (parent.slug = $1 OR child.slug = $1)
	`
	if activeOnly {
		query += ` AND rel.status = 'active'`
	}
	query += `
		ORDER BY
			CASE WHEN child.slug = $1 THEN 0 ELSE 1 END,
			rel.relationship_type ASC,
			parent.name ASC,
			child.name ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, normalizeSlug(organizationSlug), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Relationship, 0, limit)
	for rows.Next() {
		item, err := scanRelationship(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) CreateRelationship(ctx context.Context, input RelationshipInput, actor AuditActor) (Relationship, error) {
	normalized, err := normalizeRelationshipInput(input)
	if err != nil {
		return Relationship{}, err
	}

	row := r.db.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO organization_relationships (
				parent_organization_id,
				child_organization_id,
				relationship_type,
				label,
				status,
				started_at,
				ended_at,
				metadata
			)
			SELECT
				parent.id,
				child.id,
				$3,
				NULLIF($4, ''),
				$5,
				$6,
				$7,
				$8::jsonb
			FROM organizations parent, organizations child
			WHERE parent.slug = $1
				AND child.slug = $2
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
	`,
		normalized.ParentOrganizationSlug,
		normalized.ChildOrganizationSlug,
		normalized.RelationshipType,
		normalized.Label,
		normalized.Status,
		normalized.StartedAt,
		normalized.EndedAt,
		jsonOrFallback(normalized.Metadata),
	)

	item, err := scanRelationship(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Relationship{}, ErrRelationshipOrganizationNotFound
	}
	if err != nil {
		return Relationship{}, relationshipWriteError(err)
	}

	r.recordRelationshipAudit(ctx, actor, "create", nil, item, item)
	return item, nil
}

func (r *Repository) FindRelationshipByID(ctx context.Context, id string) (Relationship, error) {
	row := r.db.QueryRow(ctx, `
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
		FROM organization_relationships rel
		JOIN organizations parent ON parent.id = rel.parent_organization_id
		JOIN organizations child ON child.id = rel.child_organization_id
		WHERE rel.id = $1
	`, id)

	item, err := scanRelationship(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Relationship{}, ErrRelationshipNotFound
	}
	return item, err
}

func (r *Repository) UpdateRelationshipByID(ctx context.Context, id string, input RelationshipInput, actor AuditActor) (Relationship, error) {
	existing, err := r.FindRelationshipByID(ctx, id)
	if err != nil {
		return Relationship{}, err
	}

	if strings.TrimSpace(input.ParentOrganizationSlug) == "" {
		input.ParentOrganizationSlug = existing.Parent.Slug
	}
	if strings.TrimSpace(input.ChildOrganizationSlug) == "" {
		input.ChildOrganizationSlug = existing.Child.Slug
	}
	if strings.TrimSpace(input.RelationshipType) == "" {
		input.RelationshipType = existing.RelationshipType
	}
	if strings.TrimSpace(input.Status) == "" {
		input.Status = existing.Status
	}
	if input.StartedAt == nil && !input.ClearStartedAt {
		input.StartedAt = existing.StartedAt
	}
	if input.EndedAt == nil && !input.ClearEndedAt {
		input.EndedAt = existing.EndedAt
	}
	if len(input.Metadata) == 0 {
		input.Metadata = existing.Metadata
	}

	normalized, err := normalizeRelationshipInput(input)
	if err != nil {
		return Relationship{}, err
	}

	row := r.db.QueryRow(ctx, `
		WITH parent AS (
			SELECT id FROM organizations WHERE slug = $2
		),
		child AS (
			SELECT id FROM organizations WHERE slug = $3
		),
		updated AS (
			UPDATE organization_relationships rel
			SET
				parent_organization_id = parent.id,
				child_organization_id = child.id,
				relationship_type = $4,
				label = NULLIF($5, ''),
				status = $6,
				started_at = $7,
				ended_at = $8,
				metadata = $9::jsonb,
				updated_at = now()
			FROM parent, child
			WHERE rel.id = $1
			RETURNING rel.*
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
		FROM updated rel
		JOIN organizations parent ON parent.id = rel.parent_organization_id
		JOIN organizations child ON child.id = rel.child_organization_id
	`,
		id,
		normalized.ParentOrganizationSlug,
		normalized.ChildOrganizationSlug,
		normalized.RelationshipType,
		normalized.Label,
		normalized.Status,
		normalized.StartedAt,
		normalized.EndedAt,
		jsonOrFallback(normalized.Metadata),
	)

	item, err := scanRelationship(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Relationship{}, ErrRelationshipOrganizationNotFound
	}
	if err != nil {
		return Relationship{}, relationshipWriteError(err)
	}

	r.recordRelationshipUpdateAudit(ctx, actor, existing, item)
	return item, nil
}

func (r *Repository) DeleteRelationshipByID(ctx context.Context, id string, actor AuditActor) (Relationship, error) {
	existing, err := r.FindRelationshipByID(ctx, id)
	if err != nil {
		return Relationship{}, err
	}

	tag, err := r.db.Exec(ctx, `DELETE FROM organization_relationships WHERE id = $1`, id)
	if err != nil {
		return Relationship{}, err
	}
	if tag.RowsAffected() == 0 {
		return Relationship{}, ErrRelationshipNotFound
	}

	r.recordRelationshipAudit(ctx, actor, "delete", existing, nil, existing)
	return existing, nil
}

func (r *Repository) recordRelationshipAudit(ctx context.Context, actor AuditActor, action string, beforeData any, afterData any, item Relationship) {
	parentMetadata := relationshipAuditMetadata(actor, item, item.ParentOrganizationID, "parent")
	_ = audit.Record(ctx, r.db, actor.UserID, "organization_relationship", item.ID, action, beforeData, afterData, parentMetadata)
	if item.ChildOrganizationID == item.ParentOrganizationID {
		return
	}
	childMetadata := relationshipAuditMetadata(actor, item, item.ChildOrganizationID, "child")
	_ = audit.Record(ctx, r.db, actor.UserID, "organization_relationship", item.ID, action, beforeData, afterData, childMetadata)
}

func (r *Repository) recordRelationshipUpdateAudit(ctx context.Context, actor AuditActor, before Relationship, after Relationship) {
	for _, scope := range relationshipAuditScopes(before, after) {
		metadata := relationshipAuditMetadata(actor, after, scope.organizationID, scope.side)
		metadata["previous_parent_organization_id"] = before.ParentOrganizationID
		metadata["previous_parent_organization_slug"] = before.Parent.Slug
		metadata["previous_child_organization_id"] = before.ChildOrganizationID
		metadata["previous_child_organization_slug"] = before.Child.Slug
		_ = audit.Record(ctx, r.db, actor.UserID, "organization_relationship", after.ID, "update", before, after, metadata)
	}
}

type relationshipAuditScope struct {
	organizationID string
	side           string
}

func relationshipAuditScopes(before Relationship, after Relationship) []relationshipAuditScope {
	scopes := make([]relationshipAuditScope, 0, 4)
	seen := map[string]struct{}{}

	add := func(organizationID string, side string) {
		if organizationID == "" {
			return
		}
		if _, ok := seen[organizationID]; ok {
			return
		}
		seen[organizationID] = struct{}{}
		scopes = append(scopes, relationshipAuditScope{organizationID: organizationID, side: side})
	}

	add(after.ParentOrganizationID, "parent")
	add(after.ChildOrganizationID, "child")
	add(before.ParentOrganizationID, "previous_parent")
	add(before.ChildOrganizationID, "previous_child")
	return scopes
}

func relationshipAuditMetadata(actor AuditActor, item Relationship, organizationID string, relationshipSide string) map[string]any {
	metadata := map[string]any{
		"organization_id":           organizationID,
		"relationship_side":         relationshipSide,
		"parent_organization_id":    item.ParentOrganizationID,
		"parent_organization_slug":  item.Parent.Slug,
		"child_organization_id":     item.ChildOrganizationID,
		"child_organization_slug":   item.Child.Slug,
		"relationship_type":         item.RelationshipType,
		"relationship_public_label": item.Label,
	}
	if actor.Type != "" {
		metadata["actor_type"] = actor.Type
	}
	return metadata
}

func normalizeRelationshipInput(input RelationshipInput) (normalizedRelationshipInput, error) {
	parentSlug := normalizeSlug(input.ParentOrganizationSlug)
	childSlug := normalizeSlug(input.ChildOrganizationSlug)
	if parentSlug == "" {
		return normalizedRelationshipInput{}, errors.New("parent organization slug is required")
	}
	if childSlug == "" {
		return normalizedRelationshipInput{}, errors.New("child organization slug is required")
	}
	if parentSlug == childSlug {
		return normalizedRelationshipInput{}, ErrRelationshipSelfLink
	}

	relationshipType := strings.ToLower(strings.TrimSpace(input.RelationshipType))
	if relationshipType == "" {
		relationshipType = "related"
	}
	if _, ok := allowedRelationshipTypes[relationshipType]; !ok {
		return normalizedRelationshipInput{}, errors.New("unsupported relationship type")
	}

	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = "active"
	}
	if _, ok := allowedRelationshipStatuses[status]; !ok {
		return normalizedRelationshipInput{}, errors.New("unsupported relationship status")
	}

	return normalizedRelationshipInput{
		ParentOrganizationSlug: parentSlug,
		ChildOrganizationSlug:  childSlug,
		RelationshipType:       relationshipType,
		Label:                  strings.TrimSpace(input.Label),
		Status:                 status,
		StartedAt:              input.StartedAt,
		EndedAt:                input.EndedAt,
		Metadata:               input.Metadata,
	}, nil
}

type relationshipRow interface {
	Scan(dest ...any) error
}

func scanRelationship(row relationshipRow) (Relationship, error) {
	var item Relationship
	var metadata string
	var startedAt sql.NullTime
	var endedAt sql.NullTime

	err := row.Scan(
		&item.ID,
		&item.ParentOrganizationID,
		&item.Parent.Slug,
		&item.Parent.Name,
		&item.ChildOrganizationID,
		&item.Child.Slug,
		&item.Child.Name,
		&item.RelationshipType,
		&item.Label,
		&item.Status,
		&startedAt,
		&endedAt,
		&metadata,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Relationship{}, err
	}

	item.Parent.ID = item.ParentOrganizationID
	item.Child.ID = item.ChildOrganizationID
	item.Metadata = jsonvalue.Raw(metadata, "{}")
	if startedAt.Valid {
		value := startedAt.Time
		item.StartedAt = &value
	}
	if endedAt.Valid {
		value := endedAt.Time
		item.EndedAt = &value
	}
	return item, nil
}

func relationshipWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.ConstraintName {
	case "organization_relationships_unique_parent_child_type":
		return ErrRelationshipDuplicate
	case "organization_relationships_no_self_link":
		return ErrRelationshipSelfLink
	}

	if pgErr.Code == "23505" {
		return ErrRelationshipDuplicate
	}
	if pgErr.Code == "23514" {
		return ErrRelationshipSelfLink
	}
	return err
}
