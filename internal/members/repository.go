package members

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pendig/kelompok/internal/audit"
	"github.com/pendig/kelompok/internal/jsonvalue"
)

var ErrNotFound = errors.New("member not found")

type Repository struct {
	db *pgxpool.Pool
}

type Member struct {
	ID               string          `json:"id"`
	OrganizationID   string          `json:"organization_id"`
	OrganizationSlug string          `json:"organization_slug"`
	OrganizationName string          `json:"organization_name"`
	Name             string          `json:"name"`
	Position         string          `json:"position,omitempty"`
	Bio              string          `json:"bio,omitempty"`
	Email            string          `json:"email,omitempty"`
	Phone            string          `json:"phone,omitempty"`
	SocialLinks      json.RawMessage `json:"social_links"`
	StartDate        *time.Time      `json:"start_date,omitempty"`
	EndDate          *time.Time      `json:"end_date,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type Input struct {
	Name        string          `json:"name"`
	Position    string          `json:"position"`
	Bio         string          `json:"bio"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	SocialLinks json.RawMessage `json:"social_links"`
	StartDate   *time.Time      `json:"start_date"`
	EndDate     *time.Time      `json:"end_date"`
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByOrganizationSlug(ctx context.Context, organizationSlug string, limit int) ([]Member, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			m.id::text,
			m.organization_id::text,
			o.slug,
			o.name,
			m.name,
			COALESCE(m.position, ''),
			COALESCE(m.bio, ''),
			COALESCE(m.email, ''),
			COALESCE(m.phone, ''),
			COALESCE(m.social_links::text, '{}'),
			m.start_date,
			m.end_date,
			m.created_at,
			m.updated_at
		FROM members m
		JOIN organizations o ON o.id = m.organization_id
		WHERE o.slug = $1
		ORDER BY m.updated_at DESC, m.created_at DESC
		LIMIT $2
	`, organizationSlug, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Member, 0, limit)
	for rows.Next() {
		item, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) ListByOrganizationID(ctx context.Context, organizationID string, limit int) ([]Member, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			m.id::text,
			m.organization_id::text,
			o.slug,
			o.name,
			m.name,
			COALESCE(m.position, ''),
			COALESCE(m.bio, ''),
			COALESCE(m.email, ''),
			COALESCE(m.phone, ''),
			COALESCE(m.social_links::text, '{}'),
			m.start_date,
			m.end_date,
			m.created_at,
			m.updated_at
		FROM members m
		JOIN organizations o ON o.id = m.organization_id
		WHERE m.organization_id = $1
		ORDER BY m.updated_at DESC, m.created_at DESC
		LIMIT $2
	`, organizationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Member, 0, limit)
	for rows.Next() {
		item, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) FindByID(ctx context.Context, id string) (Member, error) {
	row := r.db.QueryRow(ctx, `
		SELECT
			m.id::text,
			m.organization_id::text,
			o.slug,
			o.name,
			m.name,
			COALESCE(m.position, ''),
			COALESCE(m.bio, ''),
			COALESCE(m.email, ''),
			COALESCE(m.phone, ''),
			COALESCE(m.social_links::text, '{}'),
			m.start_date,
			m.end_date,
			m.created_at,
			m.updated_at
		FROM members m
		JOIN organizations o ON o.id = m.organization_id
		WHERE m.id = $1
	`, id)

	item, err := scanMember(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, organizationSlug string, input Input) (Member, error) {
	organizationID, err := r.lookupOrganizationID(ctx, organizationSlug)
	if err != nil {
		return Member{}, err
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO members (
			organization_id,
			name,
			position,
			bio,
			email,
			phone,
			social_links,
			start_date,
			end_date
		)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7::jsonb, $8, $9)
		RETURNING
			id::text,
			organization_id::text,
			(SELECT slug FROM organizations WHERE id = members.organization_id),
			(SELECT name FROM organizations WHERE id = members.organization_id),
			name,
			COALESCE(position, ''),
			COALESCE(bio, ''),
			COALESCE(email, ''),
			COALESCE(phone, ''),
			COALESCE(social_links::text, '{}'),
			start_date,
			end_date,
			created_at,
			updated_at
	`,
		organizationID,
		input.Name,
		input.Position,
		input.Bio,
		input.Email,
		input.Phone,
		jsonOrFallback(input.SocialLinks),
		input.StartDate,
		input.EndDate,
	)

	item, err := scanMember(row)
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "member", item.ID, "create", nil, item, map[string]any{
			"organization_id":   item.OrganizationID,
			"organization_slug": organizationSlug,
		})
	}
	return item, err
}

func (r *Repository) UpdateByID(ctx context.Context, id string, input Input) (Member, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE members
		SET
			name = $2,
			position = NULLIF($3, ''),
			bio = NULLIF($4, ''),
			email = NULLIF($5, ''),
			phone = NULLIF($6, ''),
			social_links = $7::jsonb,
			start_date = $8,
			end_date = $9,
			updated_at = now()
		WHERE id = $1
		RETURNING
			id::text,
			organization_id::text,
			(SELECT slug FROM organizations WHERE id = members.organization_id),
			(SELECT name FROM organizations WHERE id = members.organization_id),
			name,
			COALESCE(position, ''),
			COALESCE(bio, ''),
			COALESCE(email, ''),
			COALESCE(phone, ''),
			COALESCE(social_links::text, '{}'),
			start_date,
			end_date,
			created_at,
			updated_at
	`,
		id,
		input.Name,
		input.Position,
		input.Bio,
		input.Email,
		input.Phone,
		jsonOrFallback(input.SocialLinks),
		input.StartDate,
		input.EndDate,
	)

	item, err := scanMember(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "member", item.ID, "update", nil, item, map[string]any{
			"organization_id":   item.OrganizationID,
			"organization_slug": item.OrganizationSlug,
		})
	}
	return item, err
}

func (r *Repository) DeleteByID(ctx context.Context, id string) error {
	item, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	tag, err := r.db.Exec(ctx, `DELETE FROM members WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	_ = audit.Record(ctx, r.db, nil, "member", item.ID, "delete", item, nil, map[string]any{
		"organization_id":   item.OrganizationID,
		"organization_slug": item.OrganizationSlug,
	})
	return nil
}

func (r *Repository) lookupOrganizationID(ctx context.Context, slug string) (string, error) {
	var id string
	if err := r.db.QueryRow(ctx, `SELECT id::text FROM organizations WHERE slug = $1`, slug).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("organization not found")
		}
		return "", err
	}
	return id, nil
}

type memberRow interface {
	Scan(dest ...any) error
}

func scanMember(row memberRow) (Member, error) {
	var item Member
	var socialLinks string
	var startDate sql.NullTime
	var endDate sql.NullTime

	err := row.Scan(
		&item.ID,
		&item.OrganizationID,
		&item.OrganizationSlug,
		&item.OrganizationName,
		&item.Name,
		&item.Position,
		&item.Bio,
		&item.Email,
		&item.Phone,
		&socialLinks,
		&startDate,
		&endDate,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Member{}, err
	}

	item.SocialLinks = jsonvalue.Raw(socialLinks, "{}")
	if startDate.Valid {
		value := startDate.Time
		item.StartDate = &value
	}
	if endDate.Valid {
		value := endDate.Time
		item.EndDate = &value
	}

	return item, nil
}

func jsonOrFallback(value json.RawMessage) any {
	if len(value) == 0 {
		return "{}"
	}
	return string(value)
}
