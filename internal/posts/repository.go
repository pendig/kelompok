package posts

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pendig/kelompok/internal/jsonvalue"
)

var ErrNotFound = errors.New("post not found")
var ErrAmbiguous = errors.New("post slug is ambiguous")

type Repository struct {
	db *pgxpool.Pool
}

type Post struct {
	ID               string          `json:"id"`
	OrganizationID   string          `json:"organization_id"`
	OrganizationSlug string          `json:"organization_slug"`
	OrganizationName string          `json:"organization_name"`
	CategorySlug     string          `json:"category_slug,omitempty"`
	Slug             string          `json:"slug"`
	Title            string          `json:"title"`
	Summary          string          `json:"summary,omitempty"`
	Content          string          `json:"content,omitempty"`
	Status           string          `json:"status"`
	PostData         json.RawMessage `json:"post_data"`
	SEOData          json.RawMessage `json:"seo_data"`
	PublishedAt      *time.Time      `json:"published_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListPublic(ctx context.Context, limit int) ([]Post, error) {
	rows, err := r.db.Query(ctx, postSelect(`
		WHERE p.status = 'published'
		ORDER BY p.published_at DESC NULLS LAST, p.created_at DESC
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

func (r *Repository) ListByOrganizationSlug(ctx context.Context, organizationSlug string, limit int) ([]Post, error) {
	rows, err := r.db.Query(ctx, postSelect(`
		WHERE o.slug = $1
			AND p.status = 'published'
		ORDER BY p.published_at DESC NULLS LAST, p.created_at DESC
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

func (r *Repository) FindPublishedBySlug(ctx context.Context, slug string) (Post, error) {
	rows, err := r.db.Query(ctx, postSelect(`
		WHERE p.slug = $1
			AND p.status = 'published'
		ORDER BY p.published_at DESC NULLS LAST, p.created_at DESC
		LIMIT 2
	`), slug)
	if err != nil {
		return Post{}, err
	}
	defer rows.Close()

	items := make([]Post, 0, 2)
	for rows.Next() {
		item, err := scanPost(rows)
		if err != nil {
			return Post{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return Post{}, err
	}

	if len(items) == 0 {
		return Post{}, ErrNotFound
	}
	if len(items) > 1 {
		return Post{}, ErrAmbiguous
	}

	return items[0], nil
}

func (r *Repository) FindPublishedByOrganizationAndSlug(ctx context.Context, organizationSlug, postSlug string) (Post, error) {
	row := r.db.QueryRow(ctx, postSelect(`
		WHERE o.slug = $1
			AND p.slug = $2
			AND p.status = 'published'
		LIMIT 1
	`), organizationSlug, postSlug)
	item, err := scanPost(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	return item, err
}

func postSelect(where string) string {
	return `
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
		FROM posts p
		JOIN organizations o ON o.id = p.organization_id
		LEFT JOIN post_categories c ON c.id = p.category_id
	` + where
}

type postRow interface {
	Scan(dest ...any) error
}

func scanPost(row postRow) (Post, error) {
	var item Post
	var postData string
	var seoData string
	var publishedAt sql.NullTime

	err := row.Scan(
		&item.ID,
		&item.OrganizationID,
		&item.OrganizationSlug,
		&item.OrganizationName,
		&item.CategorySlug,
		&item.Slug,
		&item.Title,
		&item.Summary,
		&item.Content,
		&item.Status,
		&postData,
		&seoData,
		&publishedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Post{}, err
	}

	if publishedAt.Valid {
		item.PublishedAt = &publishedAt.Time
	}
	item.PostData = jsonvalue.Raw(postData, "{}")
	item.SEOData = jsonvalue.Raw(seoData, "{}")

	return item, nil
}
