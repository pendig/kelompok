# API and CLI

Kelompok should be API-first and CLI-first.

The web app, automation jobs, service integrations, and future AI workflows should use the same core API and command interfaces.

## API Principles

- Version all public endpoints under `/api/v1`
- Return consistent response shapes
- Use cursor pagination for large public lists
- Keep public endpoints readable without authentication
- Require authentication for claims, edits, campaigns, and admin actions
- Make claim, import, and enrichment operations auditable
- Publish OpenAPI docs
- Expose plugin-safe ingestion endpoints and commands
- Return public DTOs instead of raw database models
- Keep claim-only contact data, source evidence, raw imports, and private plugin metadata out of public responses

## Response Shape

Recommended success shape:

```json
{
  "data": {},
  "meta": {},
  "message": "ok"
}
```

Recommended error shape:

```json
{
  "error": {
    "code": "organization_not_found",
    "message": "Organization not found",
    "details": {}
  }
}
```

## Public Endpoints

Public endpoints must be treated as an explicit allowlist.

They may include public organization profile fields, public post fields, and public impact metrics. They must not expose internal UUIDs, claim verification emails, raw source records, private evidence, credentials, tokens, or plugin-private metadata.

Dynamic JSON fields are filtered before they leave the API. If a plugin or import pipeline needs to keep raw evidence, it should store that data in internal tables or private JSON fields and expose only reviewed public fields through the stable response DTO.

Organizations:

```text
GET /api/v1/organizations
GET /api/v1/organizations/{slug}
GET /api/v1/organizations/{slug}/events
GET /api/v1/organizations/{slug}/donations
GET /api/v1/organizations/{slug}/impact-reports
GET /api/v1/organizations/{slug}/posts
GET /api/v1/organizations/{slug}/posts/{post_slug}
GET /api/v1/organizations/{slug}/sdgs
```

Implemented in the first public read API slice:

```text
GET /api/v1/organizations
GET /api/v1/organizations/{slug}
POST /api/v1/organizations/{slug}/claims
GET /api/v1/organizations/{slug}/posts
GET /api/v1/organizations/{slug}/posts/{post_slug}
GET /api/v1/organizations/{slug}/impact-reports
```

Posts:

```text
GET /api/v1/posts
GET /api/v1/posts/{slug}
GET /api/v1/post-categories
GET /api/v1/post-tags
```

Implemented in the first public read API slice:

```text
GET /api/v1/posts
GET /api/v1/posts/{slug}
```

Events:

```text
GET /api/v1/events
GET /api/v1/events/{slug}
POST /api/v1/events/{event_id}/registrations
```

Donation campaigns:

```text
GET /api/v1/donation-campaigns
GET /api/v1/donation-campaigns/{slug}
GET /api/v1/donation-campaigns/{slug}/reports
```

## Authenticated Endpoints

Auth:

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET /api/v1/auth/me
```

Claims:

```text
POST /api/v1/organizations/{slug}/claims
GET /api/v1/claims
GET /api/v1/claims/{id}
POST /api/v1/claims/{id}/verify-email
POST /api/v1/claims/{id}/verify-instagram
POST /api/v1/admin/claims/{id}/approve
POST /api/v1/admin/claims/{id}/reject
```

Organization management:

```text
GET /api/v1/org-admin/organizations
POST /api/v1/org-admin/organizations
GET /api/v1/org-admin/organizations/{slug}
PATCH /api/v1/org-admin/organizations/{slug}
GET /api/v1/org-admin/organizations/{slug}/relationships
POST /api/v1/org-admin/organization-relationships
PATCH /api/v1/org-admin/organization-relationships/{id}
DELETE /api/v1/org-admin/organization-relationships/{id}
GET /api/v1/org-admin/organizations/{slug}/claims
POST /api/v1/org-admin/claims/{id}/approve
POST /api/v1/org-admin/claims/{id}/reject
GET /api/v1/org-admin/organizations/{slug}/audit-logs
GET /api/v1/org-admin/organizations/{slug}/members
POST /api/v1/org-admin/organizations/{slug}/members
PATCH /api/v1/org-admin/members/{id}
DELETE /api/v1/org-admin/members/{id}
GET /api/v1/org-admin/impact-reports
POST /api/v1/org-admin/impact-reports
PATCH /api/v1/org-admin/impact-reports/{id}
POST /api/v1/org-admin/impact-reports/{id}/publish
POST /api/v1/org-admin/impact-reports/{id}/archive
```

For relationship updates, omit `started_at` or `ended_at` to keep the existing date, or send the field as `null` to clear it.

`POST /api/v1/org-admin/organizations` and `PATCH /api/v1/org-admin/organizations/{slug}` accept the release admin organization fields: `slug`, `name`, `legal_name`, `description`, `history`, `country`, `region`, `city`, `website_url`, `official_email`, `claim_status`, `profile_data`, `source_data`, `sdgs_data`, and `impact_data`.

Post management:

```text
GET /api/v1/org-admin/posts
POST /api/v1/org-admin/posts
PATCH /api/v1/org-admin/posts/{id}
POST /api/v1/org-admin/posts/{id}/publish
POST /api/v1/org-admin/posts/{id}/archive
```

The current alpha admin endpoints accept either a real user session from `POST /api/v1/auth/login` or the operations fallback `KELOMPOK_ADMIN_API_KEY`, provided as `X-Kelompok-Admin-Key`.

`KELOMPOK_ADMIN_ORGANIZATION_SLUGS` can restrict fallback key access to specific organizations. Scoped keys must use organization-scoped routes or provide a matching `organization_slug` where supported; global list routes are blocked for scoped keys unless a valid `organization_slug` is provided. User sessions are checked against organization roles, and non-superadmin users must use organization-scoped routes.

The static key gate remains intentionally small and self-hosting friendly. Prefer user login and organization roles for normal admin UI workflows.

Event management:

```text
POST /api/v1/org-admin/events
PATCH /api/v1/org-admin/events/{id}
POST /api/v1/org-admin/events/{id}/ticket-types
GET /api/v1/org-admin/events/{id}/registrations
```

Donor management:

```text
POST /api/v1/org-admin/donation-campaigns
PATCH /api/v1/org-admin/donation-campaigns/{id}
POST /api/v1/org-admin/donation-campaigns/{id}/reports
PATCH /api/v1/org-admin/donation-reports/{id}
```

## CLI Principles

The CLI should be useful for:

- Local development
- Self-hosted maintenance
- Data imports
- Import and enrichment jobs
- Claim operations
- Exporting data
- Future AI agent workflows

CLI commands should support:

- `--json` output
- `--dry-run`
- `--limit`
- `--source`
- `--since`
- clear exit codes

## Proposed CLI Commands

Server and database:

```text
kelompok serve
kelompok migrate up
kelompok migrate down
kelompok seed
kelompok health
```

Implemented early:

```text
kelompok seed demo
```

Organization data:

```text
kelompok org list --json
kelompok org create --name "Green Foundation" --slug green-foundation --official-email hello@example.org --source-data '{"source":"manual"}' --sdgs-data '{"primary":["13"]}'
kelompok org import --file organizations.csv
kelompok org search "climate foundation"
kelompok org show {slug} --json
kelompok org claim {slug} --email admin@example.org
kelompok org export --format json
```

Members:

```text
kelompok member list --organization {slug} --json
kelompok member create --organization {slug} --name "Aisha" --position "Chair"
kelompok member import --file members.csv --organization {slug}
kelompok member export --organization {slug} --format json
```

Organization relationships:

```text
kelompok rel list --organization {slug} --json
kelompok rel create --parent pp-muhammadiyah --child pw-muhammadiyah-dki --type structural_parent
kelompok rel create --parent muhammadiyah --child ipm --type autonomous_body --label "Autonomous student organization"
kelompok rel remove --id {relationship_id}
```

Posts:

```text
kelompok post import --file posts.csv --organization {slug}
kelompok post publish {id}
kelompok post archive {id}
kelompok post export --organization {slug} --format json
```

Imports and source operations:

```text
kelompok source add --type website --url https://example.org
kelompok source normalize --source-record {id}
kelompok source match --dry-run
```

Plugins:

```text
kelompok plugin list
kelompok plugin info {plugin}
kelompok plugin run {plugin} --job import-organizations --file organizations.csv
kelompok plugin run {plugin} --job import-members --organization {slug}
kelompok plugin run {plugin} --job import-posts --organization {slug}
kelompok plugin run {plugin} --job import-events --organization {slug}
```

Events:

```text
kelompok event import --file events.csv
kelompok event publish {id}
```

Donor reports:

```text
kelompok donor campaign create
kelompok donor report publish {id}
kelompok donor export --organization {slug}
```

Admin:

```text
kelompok admin user create
kelompok admin claim approve {claim_id}
kelompok admin claim reject {claim_id}
kelompok admin audit show --entity organization:{id}
```

## AI-Ready CLI Contract

For future AI usage, CLI output should be deterministic and machine-readable.

Example:

```text
kelompok org show green-foundation --json
```

Should return:

```json
{
  "id": "org_123",
  "slug": "green-foundation",
  "name": "Green Foundation",
  "claim_status": "unclaimed",
  "public_url": "https://example.org/o/green-foundation",
  "sdgs": ["13", "15"],
  "sources": [
    {
      "type": "website",
      "url": "https://green.example.org"
    }
  ]
}
```
