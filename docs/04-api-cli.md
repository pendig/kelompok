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

Organizations:

```text
GET /api/v1/organizations
GET /api/v1/organizations/{slug}
GET /api/v1/organizations/{slug}/events
GET /api/v1/organizations/{slug}/donations
GET /api/v1/organizations/{slug}/impact-reports
GET /api/v1/organizations/{slug}/posts
GET /api/v1/organizations/{slug}/sdgs
```

Posts:

```text
GET /api/v1/posts
GET /api/v1/posts/{slug}
GET /api/v1/post-categories
GET /api/v1/post-tags
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
POST /api/v1/organizations/{id}/claims
GET /api/v1/claims
GET /api/v1/claims/{id}
POST /api/v1/claims/{id}/verify-email
POST /api/v1/claims/{id}/verify-instagram
POST /api/v1/admin/claims/{id}/approve
POST /api/v1/admin/claims/{id}/reject
```

Organization management:

```text
PATCH /api/v1/org-admin/organizations/{id}
POST /api/v1/org-admin/organizations/{id}/members
PATCH /api/v1/org-admin/organizations/{id}/members/{member_id}
DELETE /api/v1/org-admin/organizations/{id}/members/{member_id}
POST /api/v1/org-admin/organizations/{id}/impact-reports
PATCH /api/v1/org-admin/impact-reports/{id}
```

Post management:

```text
POST /api/v1/org-admin/posts
PATCH /api/v1/org-admin/posts/{id}
POST /api/v1/org-admin/posts/{id}/publish
POST /api/v1/org-admin/posts/{id}/archive
```

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

Organization data:

```text
kelompok org import --file organizations.csv
kelompok org search "climate foundation"
kelompok org show {slug} --json
kelompok org claim {slug} --email admin@example.org
kelompok org export --format json
```

Members:

```text
kelompok member import --file members.csv --organization {slug}
kelompok member export --organization {slug} --format json
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
