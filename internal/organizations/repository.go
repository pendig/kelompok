package organizations

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("organization not found")

type Repository struct {
	db *pgxpool.Pool
}

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
			profile_data::text,
			sdgs_data::text,
			impact_data::text,
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

	items := make([]Organization, 0)
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
			profile_data::text,
			sdgs_data::text,
			impact_data::text,
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

	item.ProfileData = rawJSON(profileData, "{}")
	item.SDGSData = rawJSON(sdgsData, "{}")
	item.ImpactData = rawJSON(impactData, "{}")

	return item, nil
}

func rawJSON(value, fallback string) json.RawMessage {
	if value == "" {
		return json.RawMessage(fallback)
	}
	return json.RawMessage(value)
}
