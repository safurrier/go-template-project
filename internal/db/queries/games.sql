-- name: UpsertGame :one
INSERT INTO games (
    team_id,
    game_date,
    opponent,
    points_for,
    points_against,
    possessions,
    result,
    offensive_rating,
    defensive_rating
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (team_id, game_date, opponent)
DO UPDATE SET points_for = EXCLUDED.points_for,
              points_against = EXCLUDED.points_against,
              possessions = EXCLUDED.possessions,
              result = EXCLUDED.result,
              offensive_rating = EXCLUDED.offensive_rating,
              defensive_rating = EXCLUDED.defensive_rating
RETURNING *;

-- name: ListGamesByTeam :many
SELECT * FROM games
WHERE team_id = $1
ORDER BY game_date DESC;

-- name: ListGamesByTeamAsc :many
SELECT * FROM games
WHERE team_id = $1
ORDER BY game_date ASC;
