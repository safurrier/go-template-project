-- Example database initialization script
-- This runs when the PostgreSQL container starts for the first time

-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Example table structure
CREATE TABLE IF NOT EXISTS health_checks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status VARCHAR(50) NOT NULL,
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    details JSONB
);

-- Insert initial health check record
INSERT INTO health_checks (status, details) 
VALUES ('healthy', '{"message": "Database initialized successfully"}');

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_health_checks_checked_at ON health_checks(checked_at);

-- Example user table (commented out - uncomment if needed)
/*
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
*/