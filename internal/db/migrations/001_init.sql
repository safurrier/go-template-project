CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    conference TEXT NOT NULL DEFAULT 'Big 12',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS games (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    game_date DATE NOT NULL,
    opponent TEXT NOT NULL,
    points_for INT NOT NULL,
    points_against INT NOT NULL,
    possessions NUMERIC(8,2) NOT NULL,
    result TEXT NOT NULL CHECK (result IN ('W', 'L')),
    offensive_rating NUMERIC(8,2) NOT NULL,
    defensive_rating NUMERIC(8,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(team_id, game_date, opponent)
);

CREATE TABLE IF NOT EXISTS dashboard_snapshots (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    window_games INT NOT NULL,
    as_of_date DATE NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(team_id, window_games)
);
