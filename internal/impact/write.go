package impact

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pendig/kelompok/internal/audit"
)

var ErrNotFound = errors.New("impact report not found")

type AdminInput struct {
	OrganizationSlug  string          `json:"organization_slug"`
	Title             string          `json:"title"`
	Summary           string          `json:"summary"`
	ReportPeriodStart *time.Time      `json:"report_period_start"`
	ReportPeriodEnd   *time.Time      `json:"report_period_end"`
	SDGS              json.RawMessage `json:"sdgs"`
	Metrics           json.RawMessage `json:"metrics"`
	Status            string          `json:"status"`
	PublishedAt       *time.Time      `json:"published_at"`
}

func (r *Repository) ListAdminByOrganizationSlug(ctx context.Context, organizationSlug string, limit int) ([]Report, error) {
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
			COALESCE(ir.sdgs::text, '[]'),
			COALESCE(ir.metrics::text, '{}'),
			ir.status,
			ir.published_at,
			ir.created_at,
			ir.updated_at
		FROM impact_reports ir
		JOIN organizations o ON o.id = ir.organization_id
		WHERE o.slug = $1
		ORDER BY ir.updated_at DESC, ir.created_at DESC
		LIMIT $2
	`, organizationSlug, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Report, 0, limit)
	for rows.Next() {
		item, err := scanReport(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) ListAdmin(ctx context.Context, limit int) ([]Report, error) {
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
			COALESCE(ir.sdgs::text, '[]'),
			COALESCE(ir.metrics::text, '{}'),
			ir.status,
			ir.published_at,
			ir.created_at,
			ir.updated_at
		FROM impact_reports ir
		JOIN organizations o ON o.id = ir.organization_id
		ORDER BY ir.updated_at DESC, ir.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Report, 0, limit)
	for rows.Next() {
		item, err := scanReport(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) FindByID(ctx context.Context, id string) (Report, error) {
	row := r.db.QueryRow(ctx, `
		SELECT
			ir.id::text,
			ir.organization_id::text,
			o.slug,
			o.name,
			ir.title,
			COALESCE(ir.summary, ''),
			ir.report_period_start,
			ir.report_period_end,
			COALESCE(ir.sdgs::text, '[]'),
			COALESCE(ir.metrics::text, '{}'),
			ir.status,
			ir.published_at,
			ir.created_at,
			ir.updated_at
		FROM impact_reports ir
		JOIN organizations o ON o.id = ir.organization_id
		WHERE ir.id = $1
	`, id)

	item, err := scanReport(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Report{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, input AdminInput) (Report, error) {
	organizationID, err := r.lookupOrganizationID(ctx, input.OrganizationSlug)
	if err != nil {
		return Report{}, err
	}

	var publishedAt any
	if input.PublishedAt != nil {
		publishedAt = input.PublishedAt
	}

	row := r.db.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO impact_reports (
				organization_id,
				title,
				summary,
				report_period_start,
				report_period_end,
				sdgs,
				metrics,
				status,
				published_at
			)
			VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6::jsonb, $7::jsonb, $8, $9)
			RETURNING *
		)
		SELECT
			ir.id::text,
			ir.organization_id::text,
			o.slug,
			o.name,
			ir.title,
			COALESCE(ir.summary, ''),
			ir.report_period_start,
			ir.report_period_end,
			COALESCE(ir.sdgs::text, '[]'),
			COALESCE(ir.metrics::text, '{}'),
			ir.status,
			ir.published_at,
			ir.created_at,
			ir.updated_at
		FROM inserted ir
		JOIN organizations o ON o.id = ir.organization_id
	`,
		organizationID,
		normalizeText(input.Title),
		normalizeText(input.Summary),
		input.ReportPeriodStart,
		input.ReportPeriodEnd,
		jsonOrFallback(input.SDGS, "[]"),
		jsonOrFallback(input.Metrics, "{}"),
		normalizedStatus(input.Status),
		publishedAt,
	)

	item, err := scanReport(row)
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "impact_report", item.ID, "create", nil, item, map[string]any{
			"organization_id":   item.OrganizationID,
			"organization_slug": item.OrganizationSlug,
		})
	}
	return item, err
}

func (r *Repository) UpdateByID(ctx context.Context, id string, input AdminInput) (Report, error) {
	organizationID, err := r.lookupOrganizationID(ctx, input.OrganizationSlug)
	if err != nil {
		return Report{}, err
	}

	var publishedAt any
	if input.PublishedAt != nil {
		publishedAt = input.PublishedAt
	}

	row := r.db.QueryRow(ctx, `
		WITH updated AS (
			UPDATE impact_reports
			SET
				organization_id = $2,
				title = $3,
				summary = NULLIF($4, ''),
				report_period_start = $5,
				report_period_end = $6,
				sdgs = $7::jsonb,
				metrics = $8::jsonb,
				status = $9,
				published_at = COALESCE($10, published_at),
				updated_at = now()
			WHERE id = $1
			RETURNING *
		)
		SELECT
			ir.id::text,
			ir.organization_id::text,
			o.slug,
			o.name,
			ir.title,
			COALESCE(ir.summary, ''),
			ir.report_period_start,
			ir.report_period_end,
			COALESCE(ir.sdgs::text, '[]'),
			COALESCE(ir.metrics::text, '{}'),
			ir.status,
			ir.published_at,
			ir.created_at,
			ir.updated_at
		FROM updated ir
		JOIN organizations o ON o.id = ir.organization_id
	`,
		id,
		organizationID,
		normalizeText(input.Title),
		normalizeText(input.Summary),
		input.ReportPeriodStart,
		input.ReportPeriodEnd,
		jsonOrFallback(input.SDGS, "[]"),
		jsonOrFallback(input.Metrics, "{}"),
		normalizedStatus(input.Status),
		publishedAt,
	)

	item, err := scanReport(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Report{}, ErrNotFound
	}
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "impact_report", item.ID, "update", nil, item, map[string]any{
			"organization_id":   item.OrganizationID,
			"organization_slug": item.OrganizationSlug,
		})
	}
	return item, err
}

func (r *Repository) PublishByID(ctx context.Context, id string) (Report, error) {
	return r.setStatusByID(ctx, id, "published", true)
}

func (r *Repository) ArchiveByID(ctx context.Context, id string) (Report, error) {
	return r.setStatusByID(ctx, id, "archived", false)
}

func (r *Repository) setStatusByID(ctx context.Context, id, status string, setPublishedAt bool) (Report, error) {
	publishedAt := any(nil)
	if setPublishedAt {
		now := time.Now().UTC()
		publishedAt = now
	}

	row := r.db.QueryRow(ctx, `
		WITH updated AS (
			UPDATE impact_reports
			SET
				status = $2,
				published_at = COALESCE($3, published_at),
				updated_at = now()
			WHERE id = $1
			RETURNING *
		)
		SELECT
			ir.id::text,
			ir.organization_id::text,
			o.slug,
			o.name,
			ir.title,
			COALESCE(ir.summary, ''),
			ir.report_period_start,
			ir.report_period_end,
			COALESCE(ir.sdgs::text, '[]'),
			COALESCE(ir.metrics::text, '{}'),
			ir.status,
			ir.published_at,
			ir.created_at,
			ir.updated_at
		FROM updated ir
		JOIN organizations o ON o.id = ir.organization_id
	`,
		id,
		status,
		publishedAt,
	)

	item, err := scanReport(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Report{}, ErrNotFound
	}
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "impact_report", item.ID, "update_status", nil, item, map[string]any{
			"organization_id":   item.OrganizationID,
			"organization_slug": item.OrganizationSlug,
			"status":            status,
		})
	}
	return item, err
}

func (r *Repository) lookupOrganizationID(ctx context.Context, slug string) (string, error) {
	var id string
	if err := r.db.QueryRow(ctx, `SELECT id::text FROM organizations WHERE slug = $1`, strings.TrimSpace(slug)).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("organization not found")
		}
		return "", err
	}
	return id, nil
}

func normalizedStatus(status string) string {
	trimmed := strings.ToLower(strings.TrimSpace(status))
	switch trimmed {
	case "draft", "published", "archived":
		return trimmed
	default:
		return "draft"
	}
}

func normalizedPublishedAt(status string, provided *time.Time) any {
	if provided != nil {
		return provided
	}
	if normalizedStatus(status) == "published" {
		now := time.Now().UTC()
		return now
	}
	return nil
}

func normalizeText(value string) string {
	return strings.TrimSpace(value)
}

func jsonOrFallback(value json.RawMessage, fallback string) any {
	if len(value) == 0 {
		return fallback
	}
	return string(value)
}
