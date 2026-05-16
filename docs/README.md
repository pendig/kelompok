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

## Current Alpha Summary

Kelompok is preparing `1.0-alpha.1`. The current implementation has:

- Go API and CLI foundation
- PostgreSQL migrations and demo seed data
- Public SvelteKit pages for organizations, posts, and impact reports
- Alpha `/admin` CRM workspace for controlled deployments
- Static admin API key protection with optional organization slug scope
- Plugin architecture documentation and registry skeleton

The alpha is not a full hosted production release yet. Full user login, claim ownership verification, organization roles, practical import plugins, events, and donor management remain later milestones.
