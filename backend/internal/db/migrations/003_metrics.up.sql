CREATE TABLE metrics (
    id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    server_id UUID             NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    ts        TIMESTAMPTZ      NOT NULL DEFAULT now(),
    cpu       DOUBLE PRECISION NOT NULL,
    memory    DOUBLE PRECISION NOT NULL,
    disk      DOUBLE PRECISION NOT NULL,
    network   DOUBLE PRECISION NOT NULL  -- MB/s
);

-- "latest row for a server" and "last hour for a server" both use this
CREATE INDEX metrics_server_ts_idx ON metrics (server_id, ts DESC);