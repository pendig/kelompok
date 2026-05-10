package seed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DemoResult struct {
	OrganizationSlug string
	Posts            int
	ImpactReports    int
	SDGSSignals      int
}

func Demo(ctx context.Context, pool *pgxpool.Pool) (DemoResult, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return DemoResult{}, fmt.Errorf("begin demo seed: %w", err)
	}
	defer tx.Rollback(ctx)

	var organizationID string
	if err := tx.QueryRow(ctx, `
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
			'gerakan-hijau-nusantara',
			'Gerakan Hijau Nusantara',
			'Gerakan Hijau Nusantara Foundation',
			'A community movement focused on urban climate resilience, youth volunteering, and neighborhood-level environmental education.',
			'Started as a volunteer network in 2021, the movement now coordinates local cleanups, tree stewardship, and public education programs.',
			'Indonesia',
			'DKI Jakarta',
			'Jakarta',
			'https://example.org/gerakan-hijau',
			'hello@example.org',
			'claimed',
			$1::jsonb,
			$2::jsonb,
			$3::jsonb
		)
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			legal_name = EXCLUDED.legal_name,
			description = EXCLUDED.description,
			history = EXCLUDED.history,
			country = EXCLUDED.country,
			region = EXCLUDED.region,
			city = EXCLUDED.city,
			website_url = EXCLUDED.website_url,
			official_email = EXCLUDED.official_email,
			claim_status = EXCLUDED.claim_status,
			profile_data = EXCLUDED.profile_data,
			sdgs_data = EXCLUDED.sdgs_data,
			impact_data = EXCLUDED.impact_data,
			updated_at = now()
		RETURNING id::text
	`,
		`{"focus":["climate education","youth volunteers","urban resilience"],"social_links":{"instagram":"https://instagram.com/example"}}`,
		`{"primary":["SDG 11","SDG 13","SDG 17"],"confidence":0.86}`,
		`{"volunteers":128,"neighborhoods":14,"trees_stewarded":420}`,
	).Scan(&organizationID); err != nil {
		return DemoResult{}, fmt.Errorf("upsert demo organization: %w", err)
	}

	var categoryID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO post_categories (slug, name, description)
		VALUES ('updates', 'Updates', 'Organization news and activity updates')
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			updated_at = now()
		RETURNING id::text
	`).Scan(&categoryID); err != nil {
		return DemoResult{}, fmt.Errorf("upsert demo post category: %w", err)
	}

	if _, err := tx.Exec(ctx, `
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
		VALUES (
			$1,
			$2,
			'first-neighborhood-climate-lab',
			'First Neighborhood Climate Lab',
			'Gerakan Hijau Nusantara launches a neighborhood climate lab with youth volunteers and local mentors.',
			'The first lab brings together residents, students, and local mentors to map heat-risk areas, identify green corridors, and coordinate weekly volunteer actions.',
			'published',
			$3::jsonb,
			$4::jsonb,
			now()
		)
		ON CONFLICT (organization_id, slug) DO UPDATE SET
			category_id = EXCLUDED.category_id,
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			content = EXCLUDED.content,
			status = EXCLUDED.status,
			post_data = EXCLUDED.post_data,
			seo_data = EXCLUDED.seo_data,
			published_at = EXCLUDED.published_at,
			updated_at = now()
	`, organizationID, categoryID, `{"kind":"news","featured":true}`, `{"title":"First Neighborhood Climate Lab"}`); err != nil {
		return DemoResult{}, fmt.Errorf("upsert demo post: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM impact_reports WHERE organization_id = $1 AND title = '2026 Movement Impact Snapshot'`, organizationID); err != nil {
		return DemoResult{}, fmt.Errorf("clear demo impact report: %w", err)
	}

	if _, err := tx.Exec(ctx, `
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
		VALUES (
			$1,
			'2026 Movement Impact Snapshot',
			'A public snapshot of early movement outcomes across volunteers, neighborhoods, and climate education sessions.',
			'2026-01-01',
			'2026-03-31',
			$2::jsonb,
			$3::jsonb,
			'published',
			now()
		)
	`, organizationID, `["SDG 11","SDG 13","SDG 17"]`, `{"volunteers":128,"education_sessions":18,"neighborhoods":14}`); err != nil {
		return DemoResult{}, fmt.Errorf("insert demo impact report: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM sdgs_signals WHERE organization_id = $1 AND detected_by = 'demo_seed'`, organizationID); err != nil {
		return DemoResult{}, fmt.Errorf("clear demo sdgs signals: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO sdgs_signals (organization_id, sdg_code, confidence, evidence_text, detected_by)
		VALUES
			($1, 'SDG 11', 0.88, 'Urban resilience and neighborhood-level action programs', 'demo_seed'),
			($1, 'SDG 13', 0.91, 'Climate education and heat-risk mapping', 'demo_seed'),
			($1, 'SDG 17', 0.72, 'Volunteer, resident, and mentor collaboration', 'demo_seed')
	`, organizationID); err != nil {
		return DemoResult{}, fmt.Errorf("insert demo sdgs signals: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return DemoResult{}, fmt.Errorf("commit demo seed: %w", err)
	}

	return DemoResult{
		OrganizationSlug: "gerakan-hijau-nusantara",
		Posts:            1,
		ImpactReports:    1,
		SDGSSignals:      3,
	}, nil
}
