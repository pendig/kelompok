# Release History

This document summarizes public alpha release boundaries. Detailed build artifacts and tags live in GitHub Releases.

## v1.0-alpha.2

Status: controlled alpha.

Highlights:

- Public registration, login, logout, and account view
- Claim review flow that can assign organization ownership
- Organization relationship graph for hierarchy, autonomous bodies, affiliations, network membership, and related organizations
- Relationship API, CLI commands, admin UI, public profile display, and audit coverage
- Review fixes for relationship date clearing, public active-relationship filtering, scoped slug normalization, and audit visibility for previous and current organization scopes
- README and documentation refresh for the current alpha surface

Known limits:

- Not a non-alpha or production SaaS release
- Manual UI/UX review is still tracked separately
- Email verification, password reset, OAuth/social login, upload storage, practical import plugins, event management, and donor management remain future work

## v1.0-alpha.1

Status: first controlled alpha.

Highlights:

- Go API and CLI foundation
- PostgreSQL migrations and demo seed data
- Public organization profiles
- Public posts/articles and impact report pages
- SvelteKit public web UI with Indonesian/English language toggle
- Alpha `/admin` CRM workspace for organizations, members, claims, posts, and impact reports
- Static admin API key protection with optional organization slug scope
- Plugin architecture documentation and registry skeleton

Known limits:

- Full public auth/account onboarding was not included yet
- Organization relationship graph was not included yet
- Event management, donor management, and practical import plugins remained future phases
