# Data Model

This document describes the first-pass data model for Kelompok.

The model should combine structured SQL tables with JSONB fields for dynamic source data.

## Core Entities

### Organization

Stable fields:

- `id`
- `slug`
- `name`
- `legal_name`
- `description`
- `history`
- `country`
- `region`
- `city`
- `website_url`
- `official_email`
- `logo_file_id`
- `banner_file_id`
- `claim_status`
- `claimed_by_user_id`
- `claimed_at`
- `created_at`
- `updated_at`

Dynamic fields:

- `profile_data jsonb`
- `source_data jsonb`
- `sdgs_data jsonb`
- `impact_data jsonb`

Notes:

- `profile_data` stores dynamic public profile fields.
- `source_data` stores imported or enriched external source results.
- `sdgs_data` stores detected SDGS tags, confidence, and evidence.
- `impact_data` stores custom impact reporting fields.

### Source Record

Tracks where data came from.

Fields:

- `id`
- `source_type`
- `source_url`
- `source_name`
- `external_id`
- `raw_payload jsonb`
- `hash`
- `first_seen_at`
- `last_seen_at`
- `created_at`

Source records allow imported or enriched data to be auditable and refreshed without losing evidence.

### Organization Source

Connects organizations to source records.

Fields:

- `organization_id`
- `source_record_id`
- `match_status`
- `confidence`
- `created_at`

### Claim Request

Tracks organization claim attempts.

Fields:

- `id`
- `organization_id`
- `user_id`
- `method`
- `target`
- `status`
- `evidence jsonb`
- `reviewed_by_user_id`
- `reviewed_at`
- `created_at`
- `updated_at`

Claim methods:

- `official_email`
- `instagram`
- `manual_review`

### User

Fields:

- `id`
- `name`
- `email`
- `password_hash`
- `role`
- `email_verified_at`
- `created_at`
- `updated_at`

Roles:

- `superadmin`
- `organization_admin`
- `member`
- `viewer`

### Member

Represents people related to an organization.

Fields:

- `id`
- `organization_id`
- `name`
- `position`
- `bio`
- `email`
- `phone`
- `social_links jsonb`
- `start_date`
- `end_date`
- `created_at`
- `updated_at`

Use cases:

- Chairperson
- Board member
- Founder
- Public contact
- Volunteer coordinator

### SDGS Signal

Stores SDGS detection results.

Fields:

- `id`
- `organization_id`
- `sdg_code`
- `confidence`
- `evidence_text`
- `evidence_source_record_id`
- `detected_by`
- `created_at`

This makes SDGS data queryable while still allowing full dynamic evidence in JSONB.

### Impact Report

Fields:

- `id`
- `organization_id`
- `title`
- `summary`
- `report_period_start`
- `report_period_end`
- `sdgs jsonb`
- `metrics jsonb`
- `status`
- `published_at`
- `created_at`
- `updated_at`

Impact reports are public when published.

### Post

Represents organization-owned articles, news, announcements, and public updates.

Fields:

- `id`
- `organization_id`
- `author_user_id`
- `slug`
- `title`
- `summary`
- `content`
- `cover_file_id`
- `status`
- `published_at`
- `created_at`
- `updated_at`

Dynamic fields:

- `post_data jsonb`
- `seo_data jsonb`

Post statuses:

- `draft`
- `published`
- `archived`

### Post Category

Fields:

- `id`
- `slug`
- `name`
- `description`
- `created_at`
- `updated_at`

### Post Tag

Fields:

- `id`
- `slug`
- `name`
- `created_at`
- `updated_at`

### Post Tag Mapping

Fields:

- `post_id`
- `tag_id`

### Event

Fields:

- `id`
- `organization_id`
- `slug`
- `title`
- `description`
- `start_at`
- `end_at`
- `timezone`
- `location_type`
- `location_name`
- `location_address`
- `online_url`
- `source_record_id`
- `status`
- `created_at`
- `updated_at`

Dynamic fields:

- `event_data jsonb`
- `source_data jsonb`

### Ticket Type

Fields:

- `id`
- `event_id`
- `name`
- `description`
- `price_amount`
- `currency`
- `quota`
- `sales_start_at`
- `sales_end_at`
- `created_at`
- `updated_at`

The first version can support free and paid tickets in the model, even if payment comes later.

### Event Registration

Fields:

- `id`
- `event_id`
- `ticket_type_id`
- `attendee_name`
- `attendee_email`
- `attendee_phone`
- `status`
- `metadata jsonb`
- `created_at`
- `updated_at`

### Donation Campaign

Fields:

- `id`
- `organization_id`
- `slug`
- `title`
- `description`
- `goal_amount`
- `currency`
- `start_at`
- `end_at`
- `status`
- `created_at`
- `updated_at`

### Donation Report

Fields:

- `id`
- `campaign_id`
- `title`
- `summary`
- `amount_used`
- `currency`
- `report_data jsonb`
- `published_at`
- `created_at`
- `updated_at`

Donation reports should make fund usage visible to the public.

### Audit Log

Fields:

- `id`
- `actor_user_id`
- `entity_type`
- `entity_id`
- `action`
- `before jsonb`
- `after jsonb`
- `created_at`

Important for claims, profile edits, source-data merges, and moderation.

## Data Design Rules

- Stable, repeated fields should be columns
- Dynamic provider-specific fields should be JSONB
- Never overwrite source evidence
- Manual edits should be stored separately from raw source data
- Claims and public data changes should be auditable
- Public pages should be renderable from normalized data without requiring raw source payloads
