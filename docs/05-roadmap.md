# Roadmap

This roadmap keeps Kelompok lean. The goal is to ship a useful open-source CRM foundation before expanding into the full historical feature set.

## Roadmap Rule

API, CLI, and pluginability are not final-phase features.

They are horizontal foundations:

- Every domain feature should expose API endpoints.
- Every operational workflow should have a CLI path where practical.
- Every custom import or provider-specific workflow should go through plugin contracts.
- The end phases are for hardening and stabilization, not for introducing API/CLI/pluginability for the first time.

## Phase 0: Foundation and Contracts

Outcome: contributors can run the project and see the shape of the API, CLI, and plugin system.

Scope:

- README and docs
- Apache-2.0 license
- Contribution guidelines
- Go module initialization
- SvelteKit app initialization
- Database migration setup
- Docker Compose for PostgreSQL
- Default ports: API `4621`, Web `4622`, Worker `4623`, PostgreSQL `54621`
- Health endpoint
- Basic API server
- OpenAPI baseline
- CLI command framework
- Plugin manifest format
- Plugin registry skeleton
- Audit log foundation

Why this is first:

- API and CLI shape should guide the codebase from day one.
- Pluginability must be designed before imports and custom workflows grow.
- It prevents the core from becoming tightly coupled to one provider or one import format.

## Phase 1: Organization CRM Core

Outcome: organizations can be created, listed, viewed, edited, and claimed.

Scope:

- Organization table
- Member table
- Source record table
- User auth
- Roles and permissions
- Public organization list API
- Public organization profile API
- Organization admin API
- Member management API
- Organization and member CLI commands
- Organization profile editing
- Claim request model
- Claim by official email
- Basic audit log records for profile changes and claims
- Public profile skeleton in SvelteKit
- Minimal organization admin screen in SvelteKit
- `OrganizationImporter` contract v0
- `MemberImporter` contract v0

Keep out of Phase 1:

- Complex event ticketing
- Payment integration
- AI features
- Full donor CRM

## Phase 2: Public Profile, Posts, and Impact

Outcome: public profiles become useful communication pages, not just database records.

Scope:

- Logo and banner upload
- Leadership and contact fields
- SDGS public display
- Post/article model
- Post categories and tags
- Draft and published posts
- Public organization post list
- Public post detail page
- Public impact report list
- Public impact report detail page
- Impact report model
- Published impact reports
- Editable public profile sections
- Minimal SvelteKit editor screens for profile, posts, and impact reports
- Post and impact API endpoints
- Post and impact CLI commands
- `PostImporter` contract v0

Why this comes before events and donors:

- Organizations need to communicate before they sell tickets or open campaigns.
- Posts/articles make the public profile feel alive.
- Impact reports support the nonprofit positioning early.

## Phase 3: Pluginable Imports MVP

Outcome: users can bring organization, member, and post data into Kelompok without changing core code.

Scope:

- Import run model
- Import validation result model
- Dry-run mode
- CSV import plugin
- JSON import plugin
- Organization import implementation
- Member import implementation
- Post import implementation
- CLI plugin runner
- Plugin-safe ingestion API
- Audit log integration for plugin writes
- Import status and error reporting

Important distinction:

- Plugin architecture starts in Phase 0.
- Practical import plugins ship here because the core entities now exist.

## Phase 4: Event Management MVP

Outcome: organizations can create simple public event pages and collect basic registrations.

Scope:

- Event model
- Event API endpoints
- Event CLI commands
- Public event page
- Manual event creation
- Event list under organization profile
- Simple registration model
- Optional free ticket model
- `EventImporter` contract v0
- Event import support in CSV/JSON plugins

Keep simple:

- No complex payment system yet
- No advanced QR check-in unless needed for real pilots

## Phase 5: Donor Management MVP

Outcome: organizations can open donation campaigns and publish fund usage reports.

Scope:

- Donation campaign model
- Donation report model
- Donation campaign API endpoints
- Donation report API endpoints
- Donation CLI commands
- Public donation campaign page
- Public fund usage updates
- SDGS linkage
- Basic donor record placeholder

Keep simple:

- Payments can be integrated later.
- The early value is transparency, campaign pages, and reporting.

## Phase 6: MVP Hardening and Release Readiness

Outcome: the project is ready for an open-source MVP release.

Scope:

- OpenAPI completeness
- Stable CLI JSON output
- Plugin contract documentation
- Self-hosting guide
- Backup/export guide
- Seed data
- Test coverage for core services
- Audit viewer basics
- Basic admin operations
- Security review for auth, claims, uploads, and plugin writes

This phase should polish and stabilize API, CLI, and plugin behavior that already exists. It should not introduce them for the first time.

## Phase 7: Post-MVP Ecosystem

Outcome: Kelompok can serve broader international nonprofit and integration use cases.

Scope:

- International address fields
- Multiple languages
- NGO recommendation system
- Partner directory
- Donor and foundation matching
- More import providers
- More integration plugins
- Payment integrations
- QR check-in
- Advanced ticketing
- Advanced analytics

## Minimal Public MVP Cut Line

The minimal public MVP should include Phase 0 through Phase 2.

This MVP is intentionally narrow: a working organization CRM with public profile, posts/articles, and impact reporting in the frontend.

Minimum product surface:

- Go API
- SvelteKit web app
- PostgreSQL
- CLI framework
- Plugin registry skeleton
- Organization CRUD
- Member management
- Public organization profile
- Public organization profile frontend
- Minimal organization admin frontend
- Claim by official email
- Posts/articles
- Public post list and detail frontend
- Manual impact reports
- Public impact report frontend
- OpenAPI baseline
- Local self-hosting guide

Not required for the minimal public MVP:

- Practical CSV/JSON import plugins
- Event management
- Donor management
- Payment integration
- QR check-in
- AI features
- Advanced analytics

## Full MVP Cut Line

The fuller open-source MVP should include Phase 0 through Phase 6.

This is the version that feels more complete for public contributors and self-hosters:

- Pluginable import examples
- Event management MVP
- Donor management MVP
- OpenAPI completeness
- Stable CLI JSON output
- Plugin contract documentation
- Backup/export guide
- Security review

Public alpha can happen after Phase 2. Public MVP should wait until Phase 6.
