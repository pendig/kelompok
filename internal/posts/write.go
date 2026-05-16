package posts

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pendig/kelompok/internal/audit"
)

type AdminInput struct {
	OrganizationSlug string          `json:"organization_slug"`
	Slug             string          `json:"slug"`
	Title            string          `json:"title"`
	Summary          string          `json:"summary"`
	Content          string          `json:"content"`
	CategorySlug     string          `json:"category_slug"`
	Status           string          `json:"status"`
	PostData         json.RawMessage `json:"post_data"`
	SEOData          json.RawMessage `json:"seo_data"`
	PublishedAt      *time.Time      `json:"published_at"`
}

func (r *Repository) ListAdminByOrganizationSlug(ctx context.Context, organizationSlug string, limit int) ([]Post, error) {
	rows, err := r.db.Query(ctx, postSelect(`
		WHERE o.slug = $1
		ORDER BY p.updated_at DESC, p.created_at DESC
		LIMIT $2
	`), organizationSlug, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Post, 0, limit)
	for rows.Next() {
		item, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) ListAdmin(ctx context.Context, limit int) ([]Post, error) {
	rows, err := r.db.Query(ctx, postSelect(`
		ORDER BY p.updated_at DESC, p.created_at DESC
		LIMIT $1
	`), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Post, 0, limit)
	for rows.Next() {
		item, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) FindByID(ctx context.Context, id string) (Post, error) {
	row := r.db.QueryRow(ctx, postSelect(`
		WHERE p.id = $1
		LIMIT 1
	`), id)

	item, err := scanPost(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, input AdminInput) (Post, error) {
	organizationID, err := r.lookupOrganizationID(ctx, input.OrganizationSlug)
	if err != nil {
		return Post{}, err
	}

	categoryID, err := r.ensureCategoryID(ctx, input.CategorySlug)
	if err != nil {
		return Post{}, err
	}

	var publishedAt any
	if input.PublishedAt != nil {
		publishedAt = input.PublishedAt
	}

	row := r.db.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO posts (
				organization_id,
				category_id,
				slug,
				title,
				summary,
				content,
				status,
				post_data,
				seo_data,
				published_at
			)
			VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, $8::jsonb, $9::jsonb, $10)
			RETURNING *
		)
		SELECT
			p.id::text,
			p.organization_id::text,
			o.slug,
			o.name,
			COALESCE(c.slug, ''),
			p.slug,
			p.title,
			COALESCE(p.summary, ''),
			COALESCE(p.content, ''),
			p.status,
			COALESCE(p.post_data::text, '{}'),
			COALESCE(p.seo_data::text, '{}'),
			p.published_at,
			p.created_at,
			p.updated_at
		FROM inserted p
		JOIN organizations o ON o.id = p.organization_id
		LEFT JOIN post_categories c ON c.id = p.category_id
	`,
		organizationID,
		categoryID,
		normalizeSlug(input.Slug),
		normalizeText(input.Title),
		normalizeText(input.Summary),
		normalizeText(input.Content),
		normalizedStatus(input.Status),
		jsonOrFallback(input.PostData),
		jsonOrFallback(input.SEOData),
		publishedAt,
	)

	item, err := scanPost(row)
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "post", item.ID, "create", nil, item, map[string]any{"organization_slug": input.OrganizationSlug})
	}
	return item, err
}

func (r *Repository) UpdateByID(ctx context.Context, id string, input AdminInput) (Post, error) {
	organizationID, err := r.lookupOrganizationID(ctx, input.OrganizationSlug)
	if err != nil {
		return Post{}, err
	}

	categoryID, err := r.ensureCategoryID(ctx, input.CategorySlug)
	if err != nil {
		return Post{}, err
	}

	publishedAt := normalizedPublishedAt(input.Status, input.PublishedAt)

	row := r.db.QueryRow(ctx, `
		WITH updated AS (
			UPDATE posts
			SET
				organization_id = $2,
				category_id = $3,
				slug = $4,
				title = $5,
				summary = NULLIF($6, ''),
				content = NULLIF($7, ''),
				status = $8,
				post_data = $9::jsonb,
				seo_data = $10::jsonb,
				published_at = COALESCE($11, published_at),
				updated_at = now()
			WHERE id = $1
			RETURNING *
		)
		SELECT
			p.id::text,
			p.organization_id::text,
			o.slug,
			o.name,
			COALESCE(c.slug, ''),
			p.slug,
			p.title,
			COALESCE(p.summary, ''),
			COALESCE(p.content, ''),
			p.status,
			COALESCE(p.post_data::text, '{}'),
			COALESCE(p.seo_data::text, '{}'),
			p.published_at,
			p.created_at,
			p.updated_at
		FROM updated p
		JOIN organizations o ON o.id = p.organization_id
		LEFT JOIN post_categories c ON c.id = p.category_id
	`,
		id,
		organizationID,
		categoryID,
		normalizeSlug(input.Slug),
		normalizeText(input.Title),
		normalizeText(input.Summary),
		normalizeText(input.Content),
		normalizedStatus(input.Status),
		jsonOrFallback(input.PostData),
		jsonOrFallback(input.SEOData),
		publishedAt,
	)

	item, err := scanPost(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "post", item.ID, "update", nil, item, nil)
	}
	return item, err
}

func (r *Repository) PublishByID(ctx context.Context, id string) (Post, error) {
	return r.setStatusByID(ctx, id, "published", true)
}

func (r *Repository) ArchiveByID(ctx context.Context, id string) (Post, error) {
	return r.setStatusByID(ctx, id, "archived", false)
}

func (r *Repository) setStatusByID(ctx context.Context, id, status string, setPublishedAt bool) (Post, error) {
	publishedAt := any(nil)
	if setPublishedAt {
		now := time.Now().UTC()
		publishedAt = now
	}

	row := r.db.QueryRow(ctx, `
		WITH updated AS (
			UPDATE posts
			SET
				status = $2,
				published_at = COALESCE($3, published_at),
				updated_at = now()
			WHERE id = $1
			RETURNING *
		)
		SELECT
			p.id::text,
			p.organization_id::text,
			o.slug,
			o.name,
			COALESCE(c.slug, ''),
			p.slug,
			p.title,
			COALESCE(p.summary, ''),
			COALESCE(p.content, ''),
			p.status,
			COALESCE(p.post_data::text, '{}'),
			COALESCE(p.seo_data::text, '{}'),
			p.published_at,
			p.created_at,
			p.updated_at
		FROM updated p
		JOIN organizations o ON o.id = p.organization_id
		LEFT JOIN post_categories c ON c.id = p.category_id
	`, id, status, publishedAt)

	item, err := scanPost(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	if err == nil {
		_ = audit.Record(ctx, r.db, nil, "post", item.ID, "update_status", nil, item, map[string]any{"status": status})
	}
	return item, err
}

func (r *Repository) lookupOrganizationID(ctx context.Context, slug string) (string, error) {
	var id string
	if err := r.db.QueryRow(ctx, `SELECT id::text FROM organizations WHERE slug = $1`, slug).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return id, nil
}

func (r *Repository) ensureCategoryID(ctx context.Context, slug string) (any, error) {
	normalized := strings.TrimSpace(slug)
	if normalized == "" {
		return nil, nil
	}

	var id string
	if err := r.db.QueryRow(ctx, `
		INSERT INTO post_categories (slug, name)
		VALUES ($1, initcap(replace($1, '-', ' ')))
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = now()
		RETURNING id::text
	`, normalized).Scan(&id); err != nil {
		return nil, err
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

func normalizeSlug(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return "untitled"
	}
	return strings.NewReplacer(" ", "-", "_", "-", "/", "-").Replace(trimmed)
}

func normalizeText(value string) string {
	return strings.TrimSpace(value)
}

func jsonOrFallback(value json.RawMessage) any {
	if len(value) == 0 {
		return "{}"
	}
	return string(value)
}
