CREATE TABLE secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key_name TEXT NOT NULL,
    environment TEXT NOT NULL DEFAULT 'production',
    encrypted_value TEXT NOT NULL,
    nonce TEXT NOT NULL,
    created_by UUID REFERENCES users(id),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, key_name, environment)
);
