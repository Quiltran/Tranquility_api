CREATE UNLOGGED TABLE webauthn_cache (
    id serial PRIMARY KEY,
    key text UNIQUE NOT NULL,
    value jsonb,
    inserted_at timestamp
);
CREATE INDEX idx_cache_key ON webauthn_cache(key);

CREATE OR REPLACE PROCEDURE expire_rows (retention_period INTERVAL) as
$$
BEGIN
    DELETE FROM webauthn_cache
    WHERE inserted_at < NOW() - retention_period;

    COMMIT;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE webauthn_credentials (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id BYTEA NOT NULL UNIQUE,
    public_key BYTEA NOT NULL,
    signature_count INT NOT NULL,
    creation_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_date TIMESTAMP WITH TIME ZONE,
    last_updated_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);