# Architecture

Kelompok should be modular, efficient, and self-hostable.

The recommended backend is Go, with PostgreSQL as the primary database and JSONB for dynamic organization metadata and source evidence.

## Backend Recommendation

Use Go for the backend API, CLI, background jobs, and service workers.

Why Go fits this project:

- Small deployment footprint
- Fast startup and low memory usage
- Excellent CLI ergonomics
- Good concurrency for imports and enrichment jobs
- Simple single-binary deployment
- Strong standard library
- Easier self-hosting for nonprofits and small teams
- Less framework overhead than a large Node or PHP backend

Go is especially suitable because Kelompok is not only a web API. It also needs import tools, claim maintenance, background jobs, and future AI-facing automation.

## Alternatives

### TypeScript and NestJS

Good when the team wants full TypeScript across frontend and backend, strong decorators, and familiar enterprise patterns.

Tradeoff: heavier runtime and more framework complexity for CLI-first tooling.

### Laravel

Good when the team wants fast admin/product CRUD and a mature batteries-included web framework.

Tradeoff: less ideal for lightweight CLI distribution and high-volume background jobs compared to Go.

### Go Recommendation

Choose Go unless the team strongly wants to reuse large parts of the old backend.

The product goal is simpler and open-source-first, so a clean Go foundation is a good reset.

## Frontend Recommendation

Use SvelteKit for the frontend.

Why SvelteKit fits Kelompok:

- Fast development for public pages and CRM screens
- Less UI boilerplate than React-heavy stacks
- Good SSR support for public organization, event, and donation pages
- Vite-based development experience
- Simple deployment with adapter-node for self-hosting
- A smaller conceptual surface for contributors building forms, tables, and public pages

Next.js remains a reasonable alternative if React hiring, existing React components, or ecosystem familiarity becomes more important than simplicity. For this project, SvelteKit is the better default.

Recommended frontend stack:

- SvelteKit
- TypeScript
- Tailwind CSS
- shadcn-svelte or a small internal component kit
- TanStack Table where rich data tables are needed
- OpenAPI-generated API client where practical

## Default Ports

Use uncommon default ports:

```text
API: 4621
Web app: 4622
Worker metrics/internal diagnostics: 4623
PostgreSQL Docker port: 54621
```

## Proposed Modules

```text
cmd/
  kelompok-api/        HTTP API server
  kelompok/            CLI entrypoint

internal/
  auth/                login, sessions, roles, claim verification
  organizations/       organization profiles and claims
  members/             people, leadership, positions
  posts/               organization articles, news, and updates
  imports/             CSV and structured imports
  plugins/             plugin registry, contracts, and execution helpers
  enrichment/          SDGS and profile enrichment jobs
  sdgs/                SDGS detection and impact mapping
  events/              event pages and simple registration basics
  donors/              donation campaigns and reports
  files/               logo, banner, attachments
  publicsite/          public page composition helpers
  audit/               audit log and moderation trail
  config/              runtime config
  database/            migrations, db helpers

pkg/
  client/              generated or handwritten API client
  sdk/                 public Go SDK later if needed

docs/
  product and architecture documentation

```

## Plugin Architecture

Kelompok should be pluginable without making the core hard to maintain.

Core modules own stable business rules:

- Organizations
- Members
- Events
- Donation campaigns
- Posts and articles
- Claims
- Impact reports
- Audit logs

Plugins should integrate through stable contracts:

- Import organizations
- Import members
- Import posts and articles
- Import events
- Normalize source payloads
- Match source records to existing entities
- Enrich public profile data
- Validate provider-specific fields
- Register CLI subcommands
- Register background jobs

Plugins should not mutate core tables directly. They should call internal service interfaces or public API-compatible ingestion commands so audit logs, validation, and permissions stay consistent.

## Deployment Shape

Minimum self-hosted stack:

- `kelompok-api`
- PostgreSQL
- Object storage or local file storage
- Optional worker process
- SvelteKit frontend app

Development stack:

- Docker Compose for PostgreSQL and local services
- Go API running locally
- SvelteKit frontend running locally

Production stack:

- One API service
- One worker service
- One PostgreSQL database
- Object storage for media
- Reverse proxy

## Data Strategy

Use relational tables for stable, repeated data:

- Users
- Organizations
- Members
- Leadership positions
- Contacts
- Events
- Donation campaigns
- Claims
- Audit logs
- Source records

Use JSONB for dynamic and provider-specific data:

- Imported raw payload
- Dynamic organization attributes
- SDGS detection evidence
- Social profile metadata
- Impact report custom fields
- Data enrichment outputs
- External source evidence

This keeps the platform flexible without losing queryability for core workflows.

## API Strategy

The API should be public, documented, and stable.

Recommended:

- REST API first
- OpenAPI specification generated or maintained with code
- Versioned routes under `/api/v1`
- Consistent error format
- Cursor pagination for public listing
- Webhook-compatible design for future integrations

GraphQL can be considered later if public profile composition becomes too expensive through REST alone.

## CLI Strategy

The CLI should be a first-class interface.

Example command groups:

```text
kelompok serve
kelompok migrate
kelompok seed
kelompok org import
kelompok member import
kelompok org claim
kelompok event import
kelompok plugin list
kelompok plugin run
kelompok donor report
kelompok export
kelompok admin user
```

CLI commands should support JSON output so AI agents and automation tools can consume them safely.

## Modularity Rules

- Keep domain modules isolated
- Use interfaces at boundaries where external providers may change
- Keep provider-specific logic outside core domain modules
- Keep plugins behind stable contracts
- Keep public profile rendering separate from organization storage
- Avoid hard-coding one provider for email, Instagram, payment, or storage
- Make every background job idempotent when possible
