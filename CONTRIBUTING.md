# Contributing to Kelompok

Thank you for considering a contribution.

Kelompok is intended to be a free, open-source platform for organization management, public impact profiles, event workflows, and donor reporting.

## Current Stage

The project is currently in planning and foundation mode.

The best contributions right now are:

- Product and documentation review
- Architecture feedback
- Data model suggestions
- API and CLI design feedback
- Plugin contract feedback
- Open-source governance feedback
- Early Go backend skeleton work after the foundation is approved

## Contribution Principles

- Keep the platform free-first and nonprofit-friendly
- Prefer simple, maintainable modules over broad abstractions
- Keep public data and source evidence auditable
- Avoid storing secrets in commits, issues, docs, or examples
- Keep API and CLI behavior deterministic
- Keep plugin behavior auditable and documented
- Prioritize organization CRM workflows before advanced event, donor, or provider-specific features

## Development Direction

The recommended backend direction is Go with PostgreSQL. The recommended frontend direction is SvelteKit.

Before implementing a new module, please read:

- [Product Vision](docs/01-product-vision.md)
- [Architecture](docs/02-architecture.md)
- [Data Model](docs/03-data-model.md)
- [API and CLI](docs/04-api-cli.md)
- [Roadmap](docs/05-roadmap.md)

## Pull Request Expectations

For documentation changes:

- Keep language clear and contributor-friendly
- Prefer practical decisions over vague statements
- Update related docs when changing product scope

For code changes, once implementation begins:

- Keep modules focused
- Add tests for meaningful behavior
- Document new CLI commands and API endpoints
- Document new plugin contracts or plugin examples
- Avoid unrelated refactors
