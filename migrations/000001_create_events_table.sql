-- +goose up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    distinct_id TEXT NOT NULL,
    user_id TEXT,
    session_id TEXT,
    ip TEXT,
    user_agent TEXT,
    name TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT events_project_message_unique UNIQUE (project_id, message_id)
);

-- Time-ordered reads per project (dashboards, tail queries).
CREATE INDEX IF NOT EXISTS events_project_occurred_at_idx
    ON events (project_id, occurred_at DESC)
    WHERE deleted_at IS NULL;

-- Per-user / anonymous timelines (distinct_id).
CREATE INDEX IF NOT EXISTS events_project_distinct_occurred_idx
    ON events (project_id, distinct_id, occurred_at DESC)
    WHERE deleted_at IS NULL;

-- Session-scoped queries.
CREATE INDEX IF NOT EXISTS events_project_session_occurred_idx
    ON events (project_id, session_id, occurred_at DESC)
    WHERE deleted_at IS NULL AND session_id IS NOT NULL;

-- Logged-in user slices (nullable column).
CREATE INDEX IF NOT EXISTS events_project_user_occurred_idx
    ON events (project_id, user_id, occurred_at DESC)
    WHERE deleted_at IS NULL AND user_id IS NOT NULL;

-- Ingest / network debugging (nullable column).
CREATE INDEX IF NOT EXISTS events_project_ip_occurred_idx
    ON events (project_id, ip, occurred_at DESC)
    WHERE deleted_at IS NULL AND ip IS NOT NULL;

-- User-Agent equality filters (wide values; use with project_id + time bounds in queries).
CREATE INDEX IF NOT EXISTS events_project_user_agent_occurred_idx
    ON events (project_id, user_agent, occurred_at DESC)
    WHERE deleted_at IS NULL AND user_agent IS NOT NULL;

-- +goose down
DROP TABLE IF EXISTS events;
