# Open Source Governance

Kelompok should feel like a real open-source project from the beginning.

The project is intended to be free and nonprofit-friendly, but the license and governance model should still be chosen carefully.

## License Direction

Kelompok uses Apache-2.0.

Why Apache-2.0 fits this project:

- Friendly to contributors and organizations
- Clear patent grant
- Easy for governments, NGOs, universities, and companies to adopt
- Strong fit for a free public-interest platform that should spread widely
- Low friction for international nonprofit usage

## Governance Principles

- Keep the platform free-first
- Keep self-hosting documented
- Keep public APIs documented
- Keep plugin contracts documented
- Keep claim and data moderation transparent
- Avoid storing secrets in docs, issues, or examples
- Clearly separate source evidence from verified organization data
- Make destructive admin operations auditable

## Suggested Repository Files

Add these files before public launch:

```text
README.md
LICENSE
CONTRIBUTING.md
CODE_OF_CONDUCT.md
SECURITY.md
docs/
.github/ISSUE_TEMPLATE/
.github/PULL_REQUEST_TEMPLATE.md
```

## Contribution Areas

Good first contributor areas:

- Documentation
- Public profile UI
- Organization import formats
- Plugin examples
- SDGS keyword mapping
- OpenAPI examples
- CLI commands
- Test fixtures

Advanced contributor areas:

- Integration and import plugins
- Deduplication and matching
- Claim verification
- Audit log
- Payment integrations
- Internationalization
- Recommendation engine

## Public Data Ethics

Kelompok should treat public data responsibly.

Rules:

- Respect robots.txt where applicable
- Store source URL and timestamp
- Do not publish sensitive personal data without review
- Keep official emails as claim targets, not spam targets
- Provide takedown or correction flow
- Show whether a profile is claimed or unclaimed
- Let organizations correct their own data after verification
