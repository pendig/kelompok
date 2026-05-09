# Security Policy

Kelompok handles public organization data, claim verification, contact information, and eventually donor-related records. Security and privacy should be treated as core product requirements.

## Reporting Security Issues

This project is not publicly launched yet.

Until a public reporting channel is created, do not disclose suspected vulnerabilities in public issues. Contact the maintainers privately.

## Security Principles

- Do not commit credentials, tokens, cookies, private keys, or production data
- Treat official organization emails as sensitive claim targets
- Keep claim attempts auditable
- Do not publish sensitive source data without review
- Keep public, claimed, and manually verified data distinguishable
- Log administrative actions that affect public organization data
- Prefer least privilege for API tokens and worker credentials

## Public Data

Public data must be handled responsibly.

The platform should store source URLs, timestamps, and raw evidence where external references are used, but public display must be reviewed carefully when data may include personal or sensitive information.
