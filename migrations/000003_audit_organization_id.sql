ALTER TABLE audit_logs
ADD COLUMN IF NOT EXISTS organization_id uuid REFERENCES organizations(id) ON DELETE SET NULL;

UPDATE audit_logs al
SET organization_id = o.id
FROM organizations o
WHERE al.organization_id IS NULL
    AND (
        (al.entity_type = 'organization' AND al.entity_id = o.id)
        OR al.metadata->>'organization_id' = o.id::text
        OR al.metadata->>'organization_slug' = o.slug
    );

UPDATE audit_logs al
SET organization_id = p.organization_id
FROM posts p
WHERE al.organization_id IS NULL
    AND al.entity_type = 'post'
    AND al.entity_id = p.id;

UPDATE audit_logs al
SET organization_id = m.organization_id
FROM members m
WHERE al.organization_id IS NULL
    AND al.entity_type = 'member'
    AND al.entity_id = m.id;

UPDATE audit_logs al
SET organization_id = ir.organization_id
FROM impact_reports ir
WHERE al.organization_id IS NULL
    AND al.entity_type = 'impact_report'
    AND al.entity_id = ir.id;

UPDATE audit_logs al
SET organization_id = cr.organization_id
FROM claim_requests cr
WHERE al.organization_id IS NULL
    AND al.entity_type = 'claim_request'
    AND al.entity_id = cr.id;

CREATE INDEX IF NOT EXISTS idx_audit_logs_organization ON audit_logs(organization_id, created_at DESC);
