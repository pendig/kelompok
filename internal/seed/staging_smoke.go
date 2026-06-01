package seed

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const (
	StagingSmokeUserEmail  = "staging.user@kelompok.id"
	StagingSmokeAdminEmail = "staging.admin@kelompok.id"
	StagingSmokeOrgSlug    = "staging-smoke-org"
)

type StagingSmokeResult struct {
	UserEmail        string
	AdminEmail       string
	OrganizationSlug string
}

func StagingSmoke(ctx context.Context, pool *pgxpool.Pool, password string) (StagingSmokeResult, error) {
	if pool == nil {
		return StagingSmokeResult{}, errors.New("staging smoke seed requires a database pool")
	}

	password = strings.TrimSpace(password)
	if len(password) < 12 {
		return StagingSmokeResult{}, errors.New("staging smoke seed password must be at least 12 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return StagingSmokeResult{}, fmt.Errorf("hash staging smoke password: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return StagingSmokeResult{}, fmt.Errorf("begin staging smoke seed: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO users (name, email, password_hash, role, email_verified_at)
		VALUES
			('Staging Smoke User', $1, $3, 'viewer', now()),
			('Staging Smoke Admin', $2, $3, 'superadmin', now())
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name,
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			email_verified_at = EXCLUDED.email_verified_at,
			updated_at = now()
	`, StagingSmokeUserEmail, StagingSmokeAdminEmail, string(hash)); err != nil {
		return StagingSmokeResult{}, fmt.Errorf("upsert staging smoke users: %w", err)
	}

	if _, err := tx.Exec(ctx, `
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
			claimed_by_user_id,
			claimed_at,
			profile_data,
			source_data,
			sdgs_data,
			impact_data
		)
		VALUES (
			$1,
			'Staging Smoke Organization',
			'Staging Smoke Organization Foundation',
			'Seed organization for staging end-to-end claim and admin moderation smoke tests.',
			'Created by the staging smoke seed and safe to reset between staging validation runs.',
			'Indonesia',
			'DKI Jakarta',
			'Jakarta',
			'https://example.org/staging-smoke-org',
			'staging-smoke-org@kelompok.id',
			'unclaimed',
			NULL,
			NULL,
			$2::jsonb,
			$3::jsonb,
			$4::jsonb,
			$5::jsonb
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
			claimed_by_user_id = EXCLUDED.claimed_by_user_id,
			claimed_at = EXCLUDED.claimed_at,
			profile_data = EXCLUDED.profile_data,
			source_data = EXCLUDED.source_data,
			sdgs_data = EXCLUDED.sdgs_data,
			impact_data = EXCLUDED.impact_data,
			updated_at = now()
	`, StagingSmokeOrgSlug,
		`{"seed":"staging_smoke","focus":["claim flow","profile update","admin moderation"]}`,
		`{"seed":"staging_smoke"}`,
		`{"primary":["SDG 17"],"confidence":0.50}`,
		`{"seed":"staging_smoke"}`,
	); err != nil {
		return StagingSmokeResult{}, fmt.Errorf("upsert staging smoke organization: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return StagingSmokeResult{}, fmt.Errorf("commit staging smoke seed: %w", err)
	}

	return StagingSmokeResult{
		UserEmail:        StagingSmokeUserEmail,
		AdminEmail:       StagingSmokeAdminEmail,
		OrganizationSlug: StagingSmokeOrgSlug,
	}, nil
}
