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

## OpenAPI Contract

The implemented public, auth, and `/api/v1/org-admin` routes are published as
an OpenAPI 3.1 artifact at [`docs/openapi.yaml`](./openapi.yaml). Treat it as
the contract for the alpha line: it reflects only routes that are wired up in
`internal/httpapi/server.go`, not the aspirational endpoint lists below.

### Fetching the artifact

The file ships with the source tree, so any tag, branch, or commit can be
inspected directly:

```sh
# from a local checkout
cat docs/openapi.yaml

# from GitHub (raw contents of a tag or branch)
curl -L https://raw.githubusercontent.com/pendig/kelompok/main/docs/openapi.yaml -o openapi.yaml
```

Drop `openapi.yaml` into Swagger UI, Redocly, Insomnia, Postman, or any other
OpenAPI-aware client to inspect the contract without reading source.

### Publishing a new version

1. Edit `docs/openapi.yaml` together with the matching handler change.
2. Update the `RegisteredRoutes` inventory in `internal/httpapi/server.go`
   when adding or removing routes.
3. Run `go test ./...`. The contract smoke tests in
   `internal/httpapi/openapi_contract_test.go` fail when the YAML and the
   router drift apart, when documented response envelopes change, or when the
   document loses its `openapi`, `info`, `paths`, or `components` headers.
4. Bump `info.version` in `docs/openapi.yaml` for any breaking change.

### What is in scope

The contract documents only routes that are implemented today:

- `GET /healthz` and `GET /readyz`
- The public read API for organizations, posts, and impact reports
- The public organization claim intake (`POST /api/v1/organizations/{slug}/claims`)
- The auth endpoints (`/api/v1/auth/{register,login,logout,me}`)
- The `/api/v1/org-admin` CRM surface (organizations, relationships, claims,
  audit logs, members, posts, impact reports)

Endpoint lists below also describe routes that are planned but not yet
implemented (events, donation campaigns, claim email/instagram verification,
etc.). Those are intentionally omitted from `openapi.yaml` until the matching
handler ships, so consumers who treat the artifact as the contract never see
a route that does not exist in production.

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
PATCH /api/v1/auth/me
```

`PATCH /api/v1/auth/me` accepts `{ "name": "..." }` and updates the signed-in user's display name. The endpoint requires a Bearer session token and returns the same shape as `GET /api/v1/auth/me`.

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

Create organization is an admin-only operation. Public account/onboarding flows
must not call it without an authenticated admin/session context. The create
contract requires `name`; if `slug` is omitted it is normalized from `name`, and
`claim_status` defaults to `unclaimed`. Duplicate slugs return HTTP `409` with
`organization_slug_taken`. Invalid create payloads return HTTP `400` with stable
codes: `organization_name_required`, `organization_slug_required`,
`organization_claim_status_invalid`, `organization_official_email_invalid`, or
`organization_json_invalid`.

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

### Permission matrix

Global user roles are deliberately narrow. `superadmin` can operate across all
organizations and use global admin routes. Non-superadmin users must be checked
against `organization_user_roles` and stay inside organization-scoped routes.

| Actor / role | Scope | Manage org profile/content | Create related org | Review related claims | View audit | Admin override |
| --- | --- | --- | --- | --- | --- | --- |
| `superadmin` | Global | Yes | Yes | Yes | Yes | Yes |
| Org `owner` | Own organization | Yes | Yes | Yes | Yes | No |
| Org `admin` | Own organization | Yes | Yes | Yes | Yes | No |
| Org `member` | Own organization | No | No | No | No | No |
| Org `viewer` | Own organization | No | No | No | No | No |
| Scoped admin key | Configured slugs | Yes | Yes, when one side is allowed | Yes, for allowed org routes | Yes | Operational fallback only |

Backend guards treat org `owner` and `admin` as the canonical manageable roles.
Frontend console navigation must use the same rule, so `member` and `viewer`
roles can see their account association but are not linked into `/admin`.

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

## Claim Maintenance CLI

Implemented in PEN-27. Admins can drive claim review fully from the CLI without
opening the web UI. All commands honor `--json` for stable, automation-friendly
output and exit non-zero on validation, lookup, or database errors.

```text
kelompok claim list [--organization <slug>] [--status pending|approved|rejected|all] [--limit 50] [--json]
kelompok claim pending [--organization <slug>] [--limit 50] [--json]
kelompok claim update-status --id <claim_id> --decision approve|reject [--reviewer-user-id <uuid>] [--dry-run] [--json]
```

### Stable output contracts

`kelompok claim list` and `kelompok claim pending`:

- Human (default): tab-separated columns
  `id\torganization_slug\tstatus\tmethod\ttarget\tcreated_at\treviewed_at\treviewed_by_user_id`
  (`-` is rendered for unreviewed claims). Empty result writes nothing.
- JSON (`--json`): array of claim objects; each item is the existing
  `ClaimRequest` JSON shape with `organization_slug` and `organization_name`
  appended. Order is by `created_at` descending.

`kelompok claim update-status`:

- Human (default): one line per call —
  `claim: <applied|dry-run> decision=<approve|reject> would_become=<approved|rejected> id=<id> organization_slug=<slug> current_status=<state> reviewer_user_id=<uuid|->`
- JSON (`--json`): single object with stable keys
  `{ "dry_run", "decision", "reviewer_user_id", "would_become_status", "claim" }`.
  On `--dry-run` the embedded `claim` reflects the **current** (pre-change) state;
  on a real run it reflects the post-update state. `would_become_status` is
  always the status the claim would land on.

### Recipes

List the pending review queue across every organization:

```bash
kelompok claim pending --json | jq '.[] | {id, organization_slug, target, created_at}'
```

Approve a single claim, but verify first with a dry-run:

```bash
kelompok claim update-status --id $CLAIM_ID --decision approve --dry-run --json
kelompok claim update-status --id $CLAIM_ID --decision approve --reviewer-user-id $ADMIN_USER_ID --json
```

Reject every pending claim for one organization (manual for-loop pattern; the
CLI deliberately keeps mutations one-claim-at-a-time so each is independently
auditable):

```bash
kelompok claim list --status pending --organization green-foundation --json \
  | jq -r '.[].id' \
  | while read -r id; do
      kelompok claim update-status --id "$id" --decision reject --reviewer-user-id "$ADMIN_USER_ID"
    done
```

### Exit codes

- `0` — success
- `1` — any validation, lookup, or database error (including
  `claim_not_found`, `claim_not_pending`, invalid status / decision values, or
  a non-positive `--limit`). Stderr carries the human-readable reason.

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
