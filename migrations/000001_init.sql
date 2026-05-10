CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    email text NOT NULL UNIQUE,
    password_hash text,
    role text NOT NULL DEFAULT 'viewer' CHECK (role IN ('superadmin', 'organization_admin', 'member', 'viewer')),
    email_verified_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS files (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    storage_key text NOT NULL UNIQUE,
    filename text NOT NULL,
    content_type text NOT NULL,
    size_bytes bigint NOT NULL DEFAULT 0,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS source_records (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_type text NOT NULL,
    source_url text,
    source_name text,
    external_id text,
    raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    hash text,
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (source_type, external_id)
);

CREATE TABLE IF NOT EXISTS organizations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text NOT NULL UNIQUE,
    name text NOT NULL,
    legal_name text,
    description text,
    history text,
    country text,
    region text,
    city text,
    website_url text,
    official_email text,
    logo_file_id uuid REFERENCES files(id) ON DELETE SET NULL,
    banner_file_id uuid REFERENCES files(id) ON DELETE SET NULL,
    claim_status text NOT NULL DEFAULT 'unclaimed' CHECK (claim_status IN ('unclaimed', 'pending', 'claimed', 'rejected')),
    claimed_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    claimed_at timestamptz,
    profile_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    source_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    sdgs_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    impact_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS organization_sources (
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source_record_id uuid NOT NULL REFERENCES source_records(id) ON DELETE CASCADE,
    match_status text NOT NULL DEFAULT 'candidate' CHECK (match_status IN ('candidate', 'matched', 'rejected')),
    confidence numeric(5,4),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, source_record_id)
);

CREATE TABLE IF NOT EXISTS claim_requests (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    method text NOT NULL CHECK (method IN ('official_email', 'instagram', 'manual_review')),
    target text NOT NULL,
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    evidence jsonb NOT NULL DEFAULT '{}'::jsonb,
    reviewed_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS members (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name text NOT NULL,
    position text,
    bio text,
    email text,
    phone text,
    social_links jsonb NOT NULL DEFAULT '{}'::jsonb,
    start_date date,
    end_date date,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sdgs_signals (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    sdg_code text NOT NULL,
    confidence numeric(5,4),
    evidence_text text,
    evidence_source_record_id uuid REFERENCES source_records(id) ON DELETE SET NULL,
    detected_by text NOT NULL DEFAULT 'manual',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS post_categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text NOT NULL UNIQUE,
    name text NOT NULL,
    description text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS post_tags (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text NOT NULL UNIQUE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS posts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    author_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    category_id uuid REFERENCES post_categories(id) ON DELETE SET NULL,
    slug text NOT NULL,
    title text NOT NULL,
    summary text,
    content text NOT NULL DEFAULT '',
    cover_file_id uuid REFERENCES files(id) ON DELETE SET NULL,
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    post_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    seo_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, slug)
);

CREATE TABLE IF NOT EXISTS post_tag_mappings (
    post_id uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id uuid NOT NULL REFERENCES post_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

CREATE TABLE IF NOT EXISTS impact_reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    title text NOT NULL,
    summary text,
    report_period_start date,
    report_period_end date,
    sdgs jsonb NOT NULL DEFAULT '[]'::jsonb,
    metrics jsonb NOT NULL DEFAULT '{}'::jsonb,
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    slug text NOT NULL,
    title text NOT NULL,
    description text,
    start_at timestamptz,
    end_at timestamptz,
    timezone text NOT NULL DEFAULT 'UTC',
    location_type text NOT NULL DEFAULT 'offline' CHECK (location_type IN ('offline', 'online', 'hybrid')),
    location_name text,
    location_address text,
    online_url text,
    source_record_id uuid REFERENCES source_records(id) ON DELETE SET NULL,
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived', 'cancelled')),
    event_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, slug)
);

CREATE TABLE IF NOT EXISTS donation_campaigns (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    slug text NOT NULL,
    title text NOT NULL,
    summary text,
    description text,
    goal_amount numeric(14,2),
    currency text NOT NULL DEFAULT 'USD',
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'paused', 'closed', 'archived')),
    campaign_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (organization_id, slug)
);

CREATE TABLE IF NOT EXISTS donation_reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id uuid NOT NULL REFERENCES donation_campaigns(id) ON DELETE CASCADE,
    title text NOT NULL,
    summary text,
    amount_used numeric(14,2),
    report_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    entity_type text NOT NULL,
    entity_id uuid,
    action text NOT NULL,
    before_data jsonb,
    after_data jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_organizations_claim_status ON organizations(claim_status);
CREATE INDEX IF NOT EXISTS idx_organizations_location ON organizations(country, region, city);
CREATE INDEX IF NOT EXISTS idx_members_organization_id ON members(organization_id);
CREATE INDEX IF NOT EXISTS idx_posts_published ON posts(status, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_organization_id ON posts(organization_id);
CREATE INDEX IF NOT EXISTS idx_impact_reports_published ON impact_reports(status, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_impact_reports_organization_id ON impact_reports(organization_id);
CREATE INDEX IF NOT EXISTS idx_events_published ON events(status, start_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_organization_id ON events(organization_id);
CREATE INDEX IF NOT EXISTS idx_donation_campaigns_published ON donation_campaigns(status, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id, created_at DESC);
