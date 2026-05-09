# Plugin System

Kelompok should be pluginable while keeping the core CRM easy to maintain.

The core project should provide stable business modules, API contracts, CLI commands, validation, audit logging, and permissions. Custom workflows should live in plugins instead of being patched directly into core modules.

## Goals

- Let users import organization, member, post, and event data from custom sources
- Let maintainers add custom enrichment and normalization workflows
- Keep core tables protected by validation and audit logs
- Make plugins usable from CLI, API, and background jobs
- Keep provider-specific logic outside core modules

## Non-Goals

- Plugins should not bypass authorization
- Plugins should not write directly to database tables
- Plugins should not replace core organization, member, event, claim, or audit services
- Plugins should not require editing core code for every new import format

## Plugin Types

### Import Plugins

Used to bring structured data into Kelompok.

Examples:

- Organization CSV import
- Member spreadsheet import
- Post or article import
- Event JSON import
- Legacy CRM export import

### Enrichment Plugins

Used to add derived metadata after data is imported.

Examples:

- SDGS keyword detection
- Social link normalization
- Location normalization
- Duplicate organization matching

### Integration Plugins

Used to connect external systems.

Examples:

- Email provider
- Object storage provider
- Payment provider
- Analytics export
- Webhook forwarder

## Plugin Contracts

Plugins should communicate with the core through stable contracts.

Recommended contracts:

```text
OrganizationImporter
MemberImporter
PostImporter
EventImporter
SourceNormalizer
SourceMatcher
ProfileEnricher
JobRunner
CLICommandProvider
```

Each contract should return normalized data and validation results. Core services should decide whether data is accepted, rejected, merged, or sent to manual review.

## Example Manifest

```json
{
  "name": "csv-import",
  "version": "0.1.0",
  "description": "Import organizations, members, and events from CSV files.",
  "capabilities": [
    "organization_import",
    "member_import",
    "post_import",
    "event_import"
  ],
  "commands": [
    "import-organizations",
    "import-members",
    "import-posts",
    "import-events"
  ]
}
```

## CLI Shape

```text
kelompok plugin list
kelompok plugin info csv-import
kelompok plugin run csv-import --job import-organizations --file organizations.csv
kelompok plugin run csv-import --job import-members --organization green-foundation --file members.csv
kelompok plugin run csv-import --job import-posts --organization green-foundation --file posts.csv
kelompok plugin run csv-import --job import-events --organization green-foundation --file events.csv
```

## API Shape

Plugin-driven writes should go through ingestion endpoints or internal service contracts.

Possible API endpoints:

```text
POST /api/v1/imports/organizations
POST /api/v1/imports/members
POST /api/v1/imports/posts
POST /api/v1/imports/events
GET /api/v1/imports/{id}
```

The API should return import status, validation errors, and audit references.

## Safety Rules

- Every plugin write must create audit records
- Every imported record should preserve source evidence
- Plugins should support dry-run mode
- Plugins should report validation errors clearly
- Plugins should be idempotent where possible
- Plugins should be disabled by default unless explicitly configured

## First Plugin Examples

Build these early:

- `csv-import`: import organizations, members, posts, and events from CSV
- `json-import`: import organizations, members, posts, and events from JSON
- `sdgs-keyword`: detect SDGS tags from public profile text
