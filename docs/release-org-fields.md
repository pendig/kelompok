# Organization fields — release audit (PEN-25)

This document tracks the full contract of the `organizations` table across
the migration, repository, public API, CLI, and admin UI surfaces. It is the
counterpart to the work completed for [PEN-25] and is updated each time a
deferred field graduates to admin editing.

## Field map

| Field                | Migration                | Repository (`Organization`) | `AdminInput` | CLI `org create` flag    | Admin UI create form | Admin UI edit form | Status              |
| -------------------- | ------------------------ | --------------------------- | ------------ | ------------------------ | -------------------- | ------------------ | ------------------- |
| `slug`               | NOT NULL UNIQUE          | yes                         | yes          | `--slug`                 | yes                  | yes                | shipping            |
| `name`               | NOT NULL                 | yes                         | yes          | `--name`                 | yes                  | yes                | shipping            |
| `legal_name`         | optional                 | yes                         | yes          | `--legal-name`           | yes                  | yes                | shipping            |
| `description`        | optional                 | yes                         | yes          | `--description`          | yes                  | yes                | shipping            |
| `history`            | optional                 | yes                         | yes          | `--history`              | yes                  | yes                | shipping            |
| `country`            | optional                 | yes                         | yes          | `--country`              | yes                  | yes                | shipping            |
| `region`             | optional                 | yes                         | yes          | `--region`               | yes                  | yes                | shipping            |
| `city`               | optional                 | yes                         | yes          | `--city`                 | yes                  | yes                | shipping            |
| `website_url`        | optional                 | yes                         | yes          | `--website-url`          | yes                  | yes                | shipping            |
| `official_email`     | optional                 | yes                         | yes          | `--official-email`       | yes                  | yes                | shipping            |
| `claim_status`       | NOT NULL CHECK enum      | yes                         | yes          | `--claim-status`         | yes                  | yes                | shipping            |
| `profile_data`       | jsonb (focus, programs, languages, public_contact) | yes | yes | n/a (advanced) | yes (form fields) | yes | shipping |
| `sdgs_data`          | jsonb (`primary` list)   | yes                         | yes          | n/a (advanced)           | yes (list field)     | yes (list field)   | shipping            |
| `impact_data`        | jsonb                    | yes                         | yes          | n/a (advanced)           | not in create form   | yes (raw JSON)     | shipping            |
| `logo_file_id`       | optional FK files        | **omitted**                 | omitted      | n/a                      | **deferred**         | **deferred**       | deferred — see note |
| `banner_file_id`     | optional FK files        | omitted                     | omitted      | n/a                      | deferred             | deferred           | deferred — see note |
| `source_data`        | jsonb (ingestion)        | omitted                     | omitted      | n/a                      | not surfaced         | not surfaced       | by design           |
| `claimed_by_user_id` | optional FK users        | omitted                     | omitted      | n/a                      | not surfaced         | not surfaced       | governance flow     |
| `claimed_at`         | optional timestamptz     | omitted                     | omitted      | n/a                      | not surfaced         | not surfaced       | governance flow     |
| `created_at`         | NOT NULL DEFAULT now()   | yes                         | n/a          | n/a                      | n/a                  | n/a                | shipping            |
| `updated_at`         | NOT NULL DEFAULT now()   | yes                         | n/a          | n/a                      | n/a                  | n/a                | shipping            |

Source files:

- `migrations/000001_init.sql` — `organizations` table schema.
- `internal/organizations/repository.go` — `Organization` struct + `ListPublic` /
  `FindBySlug` / `scanOrganization`.
- `internal/organizations/write.go` — `AdminInput`, `Repository.Create`,
  `Repository.UpdateBySlug`, and the `normalizeAdminInput` helper.
- `internal/cli/cli.go` — `kelompok org create` flag set.
- `web/src/routes/admin/+page.server.js` — SvelteKit form action and
  `organizationInput()` form parser.
- `web/src/routes/admin/+page.svelte` — admin create + edit UI.

## Deferred fields and their unblockers

| Field                            | Why deferred for this release                                                                                        | Unblocker                                                                                                                                                  |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `logo_file_id`, `banner_file_id` | The `files` table exists, but there is no admin upload endpoint or attachment widget for organization branding yet. | Ship a file-upload endpoint (PUT to `/api/v1/org-admin/organizations/<slug>/logo` etc.) plus a SvelteKit dropzone component, then expose the file IDs. |
| `source_data`                    | This column is filled by the ingestion pipeline (source records, harvesting). Admin-edit would corrupt provenance.   | Build a separate "source records" admin view if curators ever need to override entries. Until then, admins can use audit logs.                            |
| `claimed_by_user_id`, `claimed_at` | These are written transactionally by `Repository.ApproveClaim` / `RejectClaim`. Direct edit would split state.        | Continue going through the claim approval flow.                                                                                                            |

When any deferred field graduates to admin editing, the unblocker should
update:

1. The `Organization` struct + `scanOrganization` SQL.
2. `AdminInput` + `normalizeAdminInput` (with validation) + the `Create` and
   `UpdateBySlug` SQL.
3. The CLI flag set if appropriate.
4. The SvelteKit `organizationInput()` form parser + the create and edit forms.
5. This audit table.

## Validation contract

`normalizeAdminInput` is the single source of truth for AdminInput validation.
It rejects:

- Empty `name`.
- A name + slug pair that yields an empty derived slug (e.g. `"---"`).
- `claim_status` values outside the migration's CHECK enum.
- `website_url` values that lack an `http://` or `https://` prefix.
- `official_email` values that don't match a `local@host.tld` shape.

The admin web UI mirrors these constraints with HTML5 attributes
(`required`, `pattern`, `type="email"`, `type="url"`) so most invalid input
is caught before submission. The server is still authoritative, and form
submissions surface any rejection inline at the top of the create form.
