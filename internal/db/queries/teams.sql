-- name: UpsertTeam :one
INSERT INTO teams (slug, display_name, conference)
VALUES ($1, $2, $3)
ON CONFLICT (slug)
DO UPDATE SET display_name = EXCLUDED.display_name,
              conference = EXCLUDED.conference
RETURNING *;

-- name: GetTeamBySlug :one
SELECT * FROM teams
WHERE slug = $1;
