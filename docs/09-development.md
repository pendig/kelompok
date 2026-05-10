# Development

This document covers the first local development path for Kelompok.

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

The web app uses SvelteKit and is intentionally minimal for MVP:

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
GET /api/v1/organizations/{slug}/posts
GET /api/v1/organizations/{slug}/posts/{post_slug}
GET /api/v1/organizations/{slug}/impact-reports
GET /api/v1/posts
GET /api/v1/posts/{slug}
```

Use the organization-scoped post detail endpoint when a post slug may exist in more than one organization.

Public responses are intentionally smaller than the database rows. The API does not expose internal UUIDs, claim verification emails, raw source evidence, or plugin-private JSON metadata through public endpoints. Dynamic JSON fields are filtered through a public allowlist before response encoding.

## CLI

The CLI is intentionally dependency-light for now.

```sh
go run ./cmd/kelompok help
go run ./cmd/kelompok serve
go run ./cmd/kelompok migrate
go run ./cmd/kelompok seed demo
go run ./cmd/kelompok db ping
go run ./cmd/kelompok db migrate
```

Future commands should preserve automation-friendly output and add `--json` where structured responses are needed.
