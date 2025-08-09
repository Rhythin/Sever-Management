-- SQL schema for Server Management

CREATE TABLE IF NOT EXISTS servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region VARCHAR(64) NOT NULL,
    type VARCHAR(32) NOT NULL,
    ip_id INTEGER REFERENCES ip_addresses(id),
    state VARCHAR(32) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    stopped_at TIMESTAMP,
    terminated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ip_addresses (
    id SERIAL PRIMARY KEY,
    address VARCHAR(64) UNIQUE NOT NULL,
    allocated BOOLEAN NOT NULL DEFAULT FALSE,
    server_id UUID REFERENCES servers(id)
);

CREATE TABLE IF NOT EXISTS billings (
    id SERIAL PRIMARY KEY,
    server_id UUID UNIQUE REFERENCES servers(id),
    accumulated_seconds BIGINT NOT NULL DEFAULT 0,
    last_billed_at TIMESTAMP,
    total_cost DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS event_logs (
    id SERIAL PRIMARY KEY,
    server_id UUID REFERENCES servers(id),
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    type VARCHAR(32) NOT NULL,
    message TEXT
);

CREATE INDEX IF NOT EXISTS idx_event_logs_server_id ON event_logs(server_id);
