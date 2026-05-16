# Kelompok

The Solutions of Movement.

Kelompok is an open-source, free-first platform for organization management, public impact profiles, lightweight event workflows, and donor reporting.

The project is built for community groups, NGOs, student organizations, foundations, volunteer networks, and nonprofit initiatives that need modern operational software without commercial lock-in.

## Product Scope

Kelompok focuses on three core areas:

1. Organization data and public profiles
2. Posts, articles, and public updates
3. Event management and lightweight ticketing
4. Donor management and transparent impact reporting

The long-term goal is to provide a modular API and CLI foundation that can later be used by AI agents, automation jobs, imports, and partner systems.

## Why This Exists

Many organizations have scattered public data, outdated landing pages, limited reporting tools, and no affordable CRM-like platform. Kelompok aims to make basic organization infrastructure free, inspectable, self-hostable, and extensible.

The platform starts with organizations first, then expands into events and donor workflows.

## Core Features

### Organization CRM and Public Profiles

- Auto-generated public landing pages for organizations
- Organization name, history, leadership, contacts, official email, social links, website, location, and SDGS-related metadata
- Claim flow through official email or Instagram-based verification
- Editable public profile after claim and registration
- Public impact reporting mapped to SDGS
- Organization-owned posts, articles, news, and public updates
- Flexible metadata using JSON fields where variables are highly dynamic
- Structured SQL tables for repeated, query-heavy data

### Posts and Articles

- Organization news and announcements
- Public articles connected to organization profiles
- Impact stories and activity updates
- Draft and published status
- Category and tag support for public discovery

### Event Management

- Manual event creation for registered organizations
- Simple ticketing flow for early releases
- Public event pages connected to organization profiles
- Future path for registration, QR check-in, invoices, and payment integrations

### Donor Management

- Public donation pages for organizations
- Donation campaign creation and reporting
- Fund usage updates and impact report timeline
- Future recommendation engine for related international NGOs and donors

### API and CLI

- API-first backend for web, automation, and future AI workflows
- CLI-first operations for imports, claims, exports, sync jobs, and admin maintenance
- Modular services so contributors can work on data, profiles, events, donors, or integrations independently

### Plugin System

- Pluginable import system for organizations, members, and events
- Pluginable import system for posts and articles
- Custom plugin hooks for source normalization, matching, enrichment, and validation
- Plugin commands exposed through the CLI
- API-safe data ingestion so external tools can integrate without touching core tables directly
- Core CRM modules stay maintainable while custom workflows live outside the core

## Recommended Technical Direction

The recommended backend stack is Go for the core API and CLI.

Go is a strong fit because Kelompok needs:

- A small deployment footprint
- Fast CLI tooling
- Efficient background jobs
- Simple static binaries for self-hosting
- Good concurrency for imports and enrichment jobs
- Clear module boundaries without heavy framework overhead

Recommended baseline:

- Go for backend API, CLI, workers, and service integrations
- PostgreSQL for relational core data
- JSONB for dynamic organization metadata, SDGS signals, source fields, and provider-specific evidence
- SQL migrations with a simple migration tool
- OpenAPI for API documentation
- SvelteKit for the frontend web app
- Docker Compose for local development

See [Architecture](docs/02-architecture.md) for more detail.

## Default Ports

Use uncommon default ports to avoid conflicts with other local projects:

- API: `4621`
- Web app: `4622`
- Worker metrics or internal diagnostics: `4623`
- PostgreSQL in Docker Compose: `54621`

## Alpha Status

Kelompok is currently preparing `1.0-alpha.1`.

The current codebase includes a lean Go API/CLI foundation, PostgreSQL migrations and seed data, public SvelteKit pages, and an alpha `/admin` CRM workspace for organization profiles, members, claims, posts, and impact reports.

The alpha is useful for local development, controlled demos, and early self-hosting tests. It is not yet a full public SaaS release because full user login, claim ownership verification, and organization role management are still intentionally scoped after this alpha.

Included alpha surface:

- Public organization profiles
- Organization posts and articles
- Public impact reports
- Alpha admin screens for editing profile, member, claim, post, and impact records
- Static admin API key protection for controlled deployments

Event management, donor management, practical import plugins, and advanced integrations can follow after the minimal public MVP.

## Documentation

- [Product Vision](docs/01-product-vision.md)
- [Architecture](docs/02-architecture.md)
- [Data Model](docs/03-data-model.md)
- [API and CLI](docs/04-api-cli.md)
- [Roadmap](docs/05-roadmap.md)
- [Open Source Governance](docs/06-open-source-governance.md)
- [Runtime Decisions](docs/07-runtime-decisions.md)
- [Plugin System](docs/08-plugin-system.md)
- [Development](docs/09-development.md)

## Quick Development Start

```sh
cp .env.example .env
# Edit KELOMPOK_DATABASE_URL and KELOMPOK_ADMIN_API_KEY before starting.
go run ./cmd/kelompok db migrate
go run ./cmd/kelompok seed demo
go run ./cmd/kelompok-api
```

```sh
cd web
cp .env.example .env
npm install
npm run dev
```

Common shortcuts:

```sh
make test
make db-migrate
make seed-demo
make api
```

The API defaults to `:4621`. Real database credentials should only live in local environment files or secret managers, never in committed files.

Frontend defaults:

- Web app port: `4622`
- API base: `http://localhost:4621`

Alpha admin access:

- API and web server both need `KELOMPOK_ADMIN_API_KEY`
- The web server forwards the key to the API server-side; the browser never receives it
- `KELOMPOK_ADMIN_ORGANIZATION_SLUGS` can restrict an admin key to specific organization slugs

## Development Principles

- Organization-first, not feature-first
- Public data should be useful before login
- Claims must be safe, auditable, and reversible
- SQL for stable entities, JSONB for dynamic evidence and source metadata
- API and CLI must be first-class interfaces
- Plugins must integrate through stable contracts, not core rewrites
- Keep the early product lean and self-hostable
- Make the default path free and nonprofit-friendly

## License

Kelompok is licensed under the Apache License 2.0.
