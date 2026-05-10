# Kelompok Documentation

This folder contains the planning and architecture notes for Kelompok.

The docs are intentionally written before implementation so the project can start with a clear product boundary, contributor-friendly module ownership, and a simpler technical direction than the previous multi-repo system.

## Index

- [01 Product Vision](01-product-vision.md)
- [02 Architecture](02-architecture.md)
- [03 Data Model](03-data-model.md)
- [04 API and CLI](04-api-cli.md)
- [05 Roadmap](05-roadmap.md)
- [06 Open Source Governance](06-open-source-governance.md)
- [07 Runtime Decisions](07-runtime-decisions.md)
- [08 Plugin System](08-plugin-system.md)

## Current Planning Summary

Kelompok should be a free, open-source organization management platform with three pillars:

1. Data and public organization profiles
2. Event management and lightweight ticketing
3. Donor management and transparent reporting

The recommended implementation starts with a Go backend, SvelteKit frontend, PostgreSQL, JSONB for dynamic metadata, OpenAPI documentation, a CLI for imports and claim operations, and a plugin system for custom integrations.
