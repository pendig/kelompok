CREATE TABLE IF NOT EXISTS organization_relationships (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    child_organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    relationship_type text NOT NULL CHECK (
        relationship_type IN (
            'structural_parent',
            'autonomous_body',
            'affiliated_with',
            'network_member',
            'related'
        )
    ),
    label text,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'pending', 'inactive', 'archived')),
    started_at date,
    ended_at date,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT organization_relationships_no_self_link CHECK (parent_organization_id <> child_organization_id),
    CONSTRAINT organization_relationships_unique_parent_child_type UNIQUE (
        parent_organization_id,
        child_organization_id,
        relationship_type
    )
);

CREATE INDEX IF NOT EXISTS idx_organization_relationships_parent ON organization_relationships(parent_organization_id);
CREATE INDEX IF NOT EXISTS idx_organization_relationships_child ON organization_relationships(child_organization_id);
CREATE INDEX IF NOT EXISTS idx_organization_relationships_type ON organization_relationships(relationship_type);
CREATE INDEX IF NOT EXISTS idx_organization_relationships_status ON organization_relationships(status);
