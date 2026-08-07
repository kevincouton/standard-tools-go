CREATE TABLE IF NOT EXISTS audit_records (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL UNIQUE,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tool_name TEXT NOT NULL,
    input_hash TEXT NOT NULL,
    output_hash TEXT NOT NULL,
    status TEXT NOT NULL,
    error TEXT,
    git_commit_sha TEXT,
    package_version TEXT,
    random_seed BIGINT,
    prev_record_hash TEXT,
    record_hash TEXT NOT NULL,
    input_json TEXT NOT NULL,
    output_json TEXT NOT NULL,
    raw JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_recorded_at ON audit_records(recorded_at);
CREATE INDEX IF NOT EXISTS idx_audit_tool_name ON audit_records(tool_name);
