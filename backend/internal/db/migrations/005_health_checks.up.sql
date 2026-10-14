CREATE TABLE health_checks (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name             TEXT        NOT NULL,
    url              TEXT        NOT NULL,
    interval_sec     INT         NOT NULL,
    expected_status  INT         NOT NULL DEFAULT 200,
    enabled          BOOLEAN     NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_checked     TIMESTAMPTZ,
    last_status      TEXT,          -- up | slow | down
    last_response_ms INT
);

CREATE INDEX health_checks_user_idx ON health_checks (user_id);
CREATE INDEX health_checks_due_idx  ON health_checks (enabled, last_checked);

-- One row per check run. Uptime and the 30-day strip are calculated from these.
CREATE TABLE check_results (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    check_id    UUID        NOT NULL REFERENCES health_checks(id) ON DELETE CASCADE,
    ts          TIMESTAMPTZ NOT NULL DEFAULT now(),
    ok          BOOLEAN     NOT NULL,
    response_ms INT,
    status_code INT
);

CREATE INDEX check_results_check_ts_idx ON check_results (check_id, ts DESC);