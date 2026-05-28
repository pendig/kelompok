package organizations

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pendig/kelompok/internal/jsonvalue"
)

var ErrNotFound = errors.New("organization not found")

type Repository struct {
	db *pgxpool.Pool
}

// Organization is the admin/public-facing representation of an organization
// row. It intentionally omits a few columns that exist on the database table
// but are not editable through the admin surface today:
//
//   - logo_file_id, banner_file_id: still pending a file-upload UX (see
//     docs/release-org-fields.md).
//   - source_data: maintained by the ingestion pipeline, not by admins.
//   - claimed_by_user_id, claimed_at: managed transactionally through the
//     claim approval flow (see Repository.ApproveClaim / RejectClaim).
//
// When any of those fields graduates to admin editing, expand this struct,
// the AdminInput payload, and the Create / UpdateBySlug SQL alongside it.
type Organization struct {
	ID            string          `json:"id"`
	Slug          string          `json:"slug"`
	Name          string          `json:"name"`
	LegalName     string          `json:"legal_name,omitempty"`
	Description   string          `json:"description,omitempty"`
	History       string          `json:"history,omitempty"`
	Country       string          `json:"country,omitempty"`
	Region        string          `json:"region,omitempty"`
	City          string          `json:"city,omitempty"`
	WebsiteURL    string          `json:"website_url,omitempty"`
	OfficialEmail string          `json:"official_email,omitempty"`
	ClaimStatus   string          `json:"claim_status"`
	ProfileData   json.RawMessage `json:"profile_data"`
	SDGSData      json.RawMessage `json:"sdgs_data"`
	ImpactData    json.RawMessage `json:"impact_data"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListPublic(ctx context.Context, limit int) ([]Organization, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
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
		FROM organizations
		ORDER BY updated_at DESC, created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Organization, 0, limit)
	for rows.Next() {
		item, err := scanOrganization(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) FindBySlug(ctx context.Context, slug string) (Organization, error) {
	row := r.db.QueryRow(ctx, `
		SELECT
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
		FROM organizations
		WHERE slug = $1
	`, slug)

	item, err := scanOrganization(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Organization{}, ErrNotFound
	}
	return item, err
}

type organizationRow interface {
	Scan(dest ...any) error
}

func scanOrganization(row organizationRow) (Organization, error) {
	var item Organization
	var profileData string
	var sdgsData string
	var impactData string

	err := row.Scan(
		&item.ID,
		&item.Slug,
		&item.Name,
		&item.LegalName,
		&item.Description,
		&item.History,
		&item.Country,
		&item.Region,
		&item.City,
		&item.WebsiteURL,
		&item.OfficialEmail,
		&item.ClaimStatus,
		&profileData,
		&sdgsData,
		&impactData,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Organization{}, err
	}

	item.ProfileData = jsonvalue.Raw(profileData, "{}")
	item.SDGSData = jsonvalue.Raw(sdgsData, "{}")
	item.ImpactData = jsonvalue.Raw(impactData, "{}")

	return item, nil
}
