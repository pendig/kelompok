# Kelompok Documentation

This folder contains the product, architecture, and development notes for Kelompok.

The docs keep the alpha release grounded: Kelompok should stay a lean organization CRM first, with public profiles, posts, impact reporting, API, CLI, and plugin contracts growing in a maintainable way.

## Index

- [01 Product Vision](01-product-vision.md)
- [02 Architecture](02-architecture.md)
- [03 Data Model](03-data-model.md)
- [04 API and CLI](04-api-cli.md)
- [05 Roadmap](05-roadmap.md)
- [06 Open Source Governance](06-open-source-governance.md)
- [07 Runtime Decisions](07-runtime-decisions.md)
- [08 Plugin System](08-plugin-system.md)
- [09 Development](09-development.md)
- [10 Release History](10-release-history.md)
- [OpenAPI contract (`openapi.yaml`)](openapi.yaml) — see [API and CLI → OpenAPI Contract](04-api-cli.md#openapi-contract) for fetch and publish steps.

## Current Alpha Summary

Kelompok is on the `1.0-alpha.2` alpha line. The current implementation has:

- Go API and CLI foundation
- PostgreSQL migrations and demo seed data
- Basic user registration, login sessions, account view, and logout
- Public SvelteKit pages for organizations, posts, and impact reports
- Organization relationship graph for hierarchy, affiliation, and related organization links
- Alpha `/admin` CRM workspace for controlled deployments
- User-session admin access with static admin API key fallback and optional organization slug scope
- Plugin architecture documentation and registry skeleton

The alpha is not a full hosted production release yet. Email verification, password reset, OAuth/social login, production upload storage, practical import plugins, events, donor management, and a full manual UI/UX polish pass remain later milestones.
