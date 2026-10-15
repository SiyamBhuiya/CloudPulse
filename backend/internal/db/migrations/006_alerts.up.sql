CREATE TABLE alert_rules (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID             NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT             NOT NULL,
    server_id    UUID             REFERENCES servers(id) ON DELETE CASCADE, -- NULL = all servers
    metric       TEXT             NOT NULL CHECK (metric IN ('cpu', 'memory', 'disk')),
    threshold    DOUBLE PRECISION NOT NULL,
    duration_min INT              NOT NULL,
    channel      TEXT             NOT NULL CHECK (channel IN ('email', 'slack', 'webhook')),
    enabled      BOOLEAN          NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE INDEX alert_rules_user_idx ON alert_rules (user_id);

-- One row per incident. Names are copied in, so history survives deleting a rule or server.
CREATE TABLE alert_events (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    rule_id     UUID             REFERENCES alert_rules(id) ON DELETE SET NULL,
    user_id     UUID             NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    server_id   UUID             REFERENCES servers(id) ON DELETE SET NULL,
    rule_name   TEXT             NOT NULL,
    server_name TEXT             NOT NULL,
    metric      TEXT             NOT NULL,
    threshold   DOUBLE PRECISION NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    status      TEXT             NOT NULL CHECK (status IN ('firing', 'resolved')),
    fired_at    TIMESTAMPTZ      NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX alert_events_user_idx ON alert_events (user_id, fired_at DESC);
CREATE INDEX alert_events_server_idx ON alert_events (server_id, fired_at DESC);

-- A rule can have only one open incident per server
CREATE UNIQUE INDEX alert_events_open_idx ON alert_events (rule_id, server_id) WHERE status = 'firing';