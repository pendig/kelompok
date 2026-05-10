# Development

This document covers the first local development path for Kelompok.

## Runtime Requirements

- Go 1.26 or newer
- PostgreSQL 15 or newer
- Node.js and pnpm later for the SvelteKit web app

## Environment

Copy the example environment file:

```sh
cp .env.example .env
```

Set a local database URL:

```sh
KELOMPOK_DATABASE_URL=postgres://kelompok:change-me@100.65.30.81:5432/kelompok_dev?sslmode=disable
```

Do not commit `.env` or real credentials.

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

## CLI

The CLI is intentionally dependency-light for now.

```sh
go run ./cmd/kelompok help
go run ./cmd/kelompok serve
go run ./cmd/kelompok migrate
go run ./cmd/kelompok db ping
go run ./cmd/kelompok db migrate
```

Future commands should preserve automation-friendly output and add `--json` where structured responses are needed.
