CREATE TABLE IF NOT EXISTS system_settings (
    key text PRIMARY KEY,
    value text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Seed default maintenance_mode as false
INSERT INTO system_settings (key, value) VALUES ('maintenance_mode', 'false') ON CONFLICT DO NOTHING;
