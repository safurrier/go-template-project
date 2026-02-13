-- name: UpsertDashboardSnapshot :one
INSERT INTO dashboard_snapshots (team_id, window_games, as_of_date, payload, updated_at)
VALUES ($1,$2,$3,$4,NOW())
ON CONFLICT (team_id, window_games)
DO UPDATE SET as_of_date = EXCLUDED.as_of_date,
              payload = EXCLUDED.payload,
              updated_at = NOW()
RETURNING *;

-- name: GetDashboardSnapshotByTeamSlugWindow :one
SELECT ds.*
FROM dashboard_snapshots ds
JOIN teams t ON t.id = ds.team_id
WHERE t.slug = $1
  AND ds.window_games = $2;
