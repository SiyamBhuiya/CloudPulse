CREATE TABLE servers (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name           TEXT        NOT NULL,
    region         TEXT        NOT NULL DEFAULT '',
    agent_key_hash TEXT        NOT NULL UNIQUE,
    last_seen      TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX servers_user_idx ON servers (user_id);