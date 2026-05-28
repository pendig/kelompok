# Development

This document covers the local development path for the Kelompok alpha.

## Runtime Requirements

- Go 1.26 or newer
- PostgreSQL 15 or newer
- Node.js and npm (or pnpm) for the SvelteKit web app

## Environment

Copy the example environment file:

```sh
cp .env.example .env
```

Set a local database URL:

```sh
KELOMPOK_DATABASE_URL=postgres://kelompok:change-me@localhost:54621/kelompok_dev?sslmode=disable
```

Do not commit `.env` or real credentials.

Set an admin key for the alpha admin API and `/admin` route:

```sh
KELOMPOK_ADMIN_API_KEY=change-me-dev-admin-key
```

For controlled demos, the same key must be available to the API process and the SvelteKit server process. The browser never receives this key; SvelteKit forwards it from server-side code only.

Optionally restrict the key to specific organization slugs:

```sh
KELOMPOK_ADMIN_ORGANIZATION_SLUGS=gerakan-hijau-nusantara,another-org
```

Database pool settings can be tuned with:

```text
KELOMPOK_DB_MAX_CONNS=5
KELOMPOK_DB_MIN_CONNS=0
KELOMPOK_DB_MAX_CONN_LIFETIME=30m
KELOMPOK_DB_MAX_CONN_IDLE_TIME=5m
KELOMPOK_DB_HEALTH_CHECK_PERIOD=1m
```

## Database

Apply migrations:

```sh
go run ./cmd/kelompok db migrate
```

Or with the project shortcut:

```sh
make db-migrate
```

Check connectivity:

```sh
go run ./cmd/kelompok db ping
```

Insert demo MVP data:

```sh
go run ./cmd/kelompok seed demo
```

Or with the project shortcut:

```sh
make seed-demo
```

The first migration creates the stable CRM tables for:

- users
- organizations
- organization relationships
- source records
- claim requests
- members
- SDGS signals
- posts, categories, and tags
- impact reports
- events
- donation campaigns and reports
- audit logs

Dynamic and provider-specific data belongs in JSONB columns until it becomes stable enough to promote into structured columns.

Kelompok targets PostgreSQL 15 or newer. The initial schema uses `gen_random_uuid()` for UUID defaults and does not create database extensions from application migrations, so managed environments should provision required database capabilities before running the app user migrations.

## Frontend

The web app uses SvelteKit and is intentionally minimal for the alpha:

```sh
cd web
cp .env.example .env
npm install
npm run dev
```

The frontend uses this API base:

```text
VITE_API_BASE_URL=http://localhost:4621
```

and runs on:

```text
http://localhost:4622
```

If you run API on another host/port, change `VITE_API_BASE_URL` accordingly.

The `/admin` route is the current alpha CRM workspace. It can create and edit organization profiles, create member records, submit/review claim requests, manage organization relationships, create posts, and create impact reports through the local API. Treat it as a controlled development interface until email verification, password reset, production upload storage, and the UI polish pass are complete.
It can use a login session from `/api/v1/auth/login`. For controlled self-hosted operations, it can also read `KELOMPOK_ADMIN_API_KEY` on the server side and send it to the API as `X-Kelompok-Admin-Key`.

## API

Start the API server:

```sh
go run ./cmd/kelompok-api
```

Or:

```sh
make api
```

Default API port:

```text
4621
```

Health endpoints:

```text
GET /healthz
GET /readyz
```

`/healthz` checks that the process is alive. `/readyz` checks database connectivity.

Public MVP endpoints:

```text
GET /api/v1/organizations
GET /api/v1/organizations/{slug}
POST /api/v1/organizations/{slug}/claims
GET /api/v1/organizations/{slug}/posts
GET /api/v1/organizations/{slug}/posts/{post_slug}
GET /api/v1/organizations/{slug}/impact-reports
GET /api/v1/posts
GET /api/v1/posts/{slug}
```

Use the organization-scoped post detail endpoint when a post slug may exist in more than one organization.

Public responses are intentionally smaller than the database rows. The API does not expose internal UUIDs, claim verification emails, raw source evidence, or plugin-private JSON metadata through public endpoints. Dynamic JSON fields are filtered through a public allowlist before response encoding.

Alpha admin endpoints:

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET /api/v1/auth/me
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
GET /api/v1/org-admin/posts
POST /api/v1/org-admin/posts
PATCH /api/v1/org-admin/posts/{id}
POST /api/v1/org-admin/posts/{id}/publish
POST /api/v1/org-admin/posts/{id}/archive
GET /api/v1/org-admin/impact-reports
POST /api/v1/org-admin/impact-reports
PATCH /api/v1/org-admin/impact-reports/{id}
POST /api/v1/org-admin/impact-reports/{id}/publish
POST /api/v1/org-admin/impact-reports/{id}/archive
```

Do not publish the alpha admin API directly to the internet without either user sessions or `KELOMPOK_ADMIN_API_KEY`. For shared environments, prefer user sessions and organization roles. If the fallback key is enabled, also set `KELOMPOK_ADMIN_ORGANIZATION_SLUGS` or place the API behind a stronger auth proxy.

When `KELOMPOK_ADMIN_ORGANIZATION_SLUGS` is set, scoped keys must use organization-scoped routes or include a matching `organization_slug` where supported. Global admin list routes are blocked for scoped keys to avoid cross-organization data leakage. Non-superadmin user sessions are also required to use organization-scoped admin routes.

## Alpha Verification

Before tagging an alpha release, run:

```sh
go test ./...
npm run check --prefix web
npm run build --prefix web
```

For a local smoke test:

In one terminal, prepare the database and start the API:

```sh
go run ./cmd/kelompok db migrate
go run ./cmd/kelompok seed demo
KELOMPOK_ADMIN_API_KEY=change-me-dev-admin-key go run ./cmd/kelompok-api
```

In a second terminal, start the web app:

```sh
KELOMPOK_ADMIN_API_KEY=change-me-dev-admin-key npm run dev --prefix web -- --host localhost --port 4622
```

Then open:

```text
http://localhost:4622/
http://localhost:4622/admin
http://localhost:4622/account
```

Recommended alpha smoke checks:

- Register, log in, open `/account`, then log out.
- Submit and approve an organization claim.
- Create an organization relationship, confirm it appears in `/admin`, and confirm active relationships appear on the public organization profile.
- Confirm organization, post, impact, claim, member, and relationship write actions appear in the organization audit log.

## CLI

The CLI is intentionally dependency-light for now.

```sh
go run ./cmd/kelompok help
go run ./cmd/kelompok serve
go run ./cmd/kelompok migrate
go run ./cmd/kelompok seed demo
go run ./cmd/kelompok db ping
go run ./cmd/kelompok db migrate
go run ./cmd/kelompok org list --json
go run ./cmd/kelompok org create --name "Green Foundation" --slug green-foundation --source-data '{"source":"manual"}'
go run ./cmd/kelompok member list --organization green-foundation --json
go run ./cmd/kelompok member create --organization green-foundation --name "Aisha" --position "Chair"
```

Future commands should preserve automation-friendly output and add `--json` where structured responses are needed.
