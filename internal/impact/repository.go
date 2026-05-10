package impact

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

type Report struct {
	ID                string          `json:"id"`
	OrganizationID    string          `json:"organization_id"`
	OrganizationSlug  string          `json:"organization_slug"`
	OrganizationName  string          `json:"organization_name"`
	Title             string          `json:"title"`
	Summary           string          `json:"summary,omitempty"`
	ReportPeriodStart *time.Time      `json:"report_period_start,omitempty"`
	ReportPeriodEnd   *time.Time      `json:"report_period_end,omitempty"`
	SDGS              json.RawMessage `json:"sdgs"`
	Metrics           json.RawMessage `json:"metrics"`
	Status            string          `json:"status"`
	PublishedAt       *time.Time      `json:"published_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByOrganizationSlug(ctx context.Context, organizationSlug string, limit int) ([]Report, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			ir.id::text,
			ir.organization_id::text,
			o.slug,
			o.name,
			ir.title,
			COALESCE(ir.summary, ''),
			ir.report_period_start,
			ir.report_period_end,
			ir.sdgs::text,
			ir.metrics::text,
			ir.status,
			ir.published_at,
			ir.created_at,
			ir.updated_at
		FROM impact_reports ir
		JOIN organizations o ON o.id = ir.organization_id
		WHERE o.slug = $1
			AND ir.status = 'published'
		ORDER BY ir.published_at DESC NULLS LAST, ir.created_at DESC
		LIMIT $2
	`, organizationSlug, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Report, 0)
	for rows.Next() {
		item, err := scanReport(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

type reportRow interface {
	Scan(dest ...any) error
}

func scanReport(row reportRow) (Report, error) {
	var item Report
	var periodStart sql.NullTime
	var periodEnd sql.NullTime
	var sdgs string
	var metrics string
	var publishedAt sql.NullTime

	err := row.Scan(
		&item.ID,
		&item.OrganizationID,
		&item.OrganizationSlug,
		&item.OrganizationName,
		&item.Title,
		&item.Summary,
		&periodStart,
		&periodEnd,
		&sdgs,
		&metrics,
		&item.Status,
		&publishedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Report{}, err
	}

	if periodStart.Valid {
		item.ReportPeriodStart = &periodStart.Time
	}
	if periodEnd.Valid {
		item.ReportPeriodEnd = &periodEnd.Time
	}
	if publishedAt.Valid {
		item.PublishedAt = &publishedAt.Time
	}
	item.SDGS = rawJSON(sdgs, "[]")
	item.Metrics = rawJSON(metrics, "{}")

	return item, nil
}

func rawJSON(value, fallback string) json.RawMessage {
	if value == "" {
		return json.RawMessage(fallback)
	}
	return json.RawMessage(value)
}
