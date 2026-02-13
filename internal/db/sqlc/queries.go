package sqlc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Queries struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Queries { return &Queries{db: db} }

type UpsertTeamParams struct{ Slug, DisplayName, Conference string }

func (q *Queries) UpsertTeam(ctx context.Context, arg UpsertTeamParams) (Team, error) {
	const sql = `INSERT INTO teams (slug, display_name, conference)
VALUES ($1, $2, $3)
ON CONFLICT (slug)
DO UPDATE SET display_name = EXCLUDED.display_name,
              conference = EXCLUDED.conference
RETURNING id, slug, display_name, conference, created_at`
	var t Team
	err := q.db.QueryRow(ctx, sql, arg.Slug, arg.DisplayName, arg.Conference).Scan(&t.ID, &t.Slug, &t.DisplayName, &t.Conference, &t.CreatedAt)
	return t, err
}

func (q *Queries) GetTeamBySlug(ctx context.Context, slug string) (Team, error) {
	const sql = `SELECT id, slug, display_name, conference, created_at FROM teams WHERE slug = $1`
	var t Team
	err := q.db.QueryRow(ctx, sql, slug).Scan(&t.ID, &t.Slug, &t.DisplayName, &t.Conference, &t.CreatedAt)
	return t, err
}

type UpsertGameParams struct {
	TeamID                           int64
	GameDate                         time.Time
	Opponent                         string
	PointsFor, PointsAgainst         int32
	Possessions                      float64
	Result                           string
	OffensiveRating, DefensiveRating float64
}

func (q *Queries) UpsertGame(ctx context.Context, arg UpsertGameParams) (Game, error) {
	const sql = `INSERT INTO games (
    team_id, game_date, opponent, points_for, points_against, possessions, result, offensive_rating, defensive_rating)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (team_id, game_date, opponent)
DO UPDATE SET points_for = EXCLUDED.points_for,
              points_against = EXCLUDED.points_against,
              possessions = EXCLUDED.possessions,
              result = EXCLUDED.result,
              offensive_rating = EXCLUDED.offensive_rating,
              defensive_rating = EXCLUDED.defensive_rating
RETURNING id, team_id, game_date, opponent, points_for, points_against, possessions::float8, result, offensive_rating::float8, defensive_rating::float8, created_at`
	var g Game
	err := q.db.QueryRow(ctx, sql, arg.TeamID, arg.GameDate, arg.Opponent, arg.PointsFor, arg.PointsAgainst, arg.Possessions, arg.Result, arg.OffensiveRating, arg.DefensiveRating).
		Scan(&g.ID, &g.TeamID, &g.GameDate, &g.Opponent, &g.PointsFor, &g.PointsAgainst, &g.Possessions, &g.Result, &g.OffensiveRating, &g.DefensiveRating, &g.CreatedAt)
	return g, err
}

func (q *Queries) ListGamesByTeamAsc(ctx context.Context, teamID int64) ([]Game, error) {
	const sql = `SELECT id, team_id, game_date, opponent, points_for, points_against, possessions::float8, result, offensive_rating::float8, defensive_rating::float8, created_at FROM games WHERE team_id = $1 ORDER BY game_date ASC`
	rows, err := q.db.Query(ctx, sql, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var games []Game
	for rows.Next() {
		var g Game
		if err := rows.Scan(&g.ID, &g.TeamID, &g.GameDate, &g.Opponent, &g.PointsFor, &g.PointsAgainst, &g.Possessions, &g.Result, &g.OffensiveRating, &g.DefensiveRating, &g.CreatedAt); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, rows.Err()
}

type UpsertDashboardSnapshotParams struct {
	TeamID      int64
	WindowGames int32
	AsOfDate    time.Time
	Payload     json.RawMessage
}

func (q *Queries) UpsertDashboardSnapshot(ctx context.Context, arg UpsertDashboardSnapshotParams) (DashboardSnapshot, error) {
	const sql = `INSERT INTO dashboard_snapshots (team_id, window_games, as_of_date, payload, updated_at)
VALUES ($1,$2,$3,$4,NOW())
ON CONFLICT (team_id, window_games)
DO UPDATE SET as_of_date = EXCLUDED.as_of_date, payload = EXCLUDED.payload, updated_at = NOW()
RETURNING id, team_id, window_games, as_of_date, payload, created_at, updated_at`
	var ds DashboardSnapshot
	err := q.db.QueryRow(ctx, sql, arg.TeamID, arg.WindowGames, arg.AsOfDate, arg.Payload).Scan(&ds.ID, &ds.TeamID, &ds.WindowGames, &ds.AsOfDate, &ds.Payload, &ds.CreatedAt, &ds.UpdatedAt)
	return ds, err
}

func (q *Queries) GetDashboardSnapshotByTeamSlugWindow(ctx context.Context, slug string, window int32) (DashboardSnapshot, error) {
	const sql = `SELECT ds.id, ds.team_id, ds.window_games, ds.as_of_date, ds.payload, ds.created_at, ds.updated_at
FROM dashboard_snapshots ds JOIN teams t ON t.id = ds.team_id
WHERE t.slug = $1 AND ds.window_games = $2`
	var ds DashboardSnapshot
	err := q.db.QueryRow(ctx, sql, slug, window).Scan(&ds.ID, &ds.TeamID, &ds.WindowGames, &ds.AsOfDate, &ds.Payload, &ds.CreatedAt, &ds.UpdatedAt)
	return ds, err
}
