package ingest

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/your-org/go-template-project/internal/dashboard"
	"github.com/your-org/go-template-project/internal/db/sqlc"
)

type Service struct{ q *sqlc.Queries }

func NewService(q *sqlc.Queries) *Service { return &Service{q: q} }

func (s *Service) IngestCSV(ctx context.Context, path, slug, name, conf string, windows []int32) error {
	team, err := s.q.UpsertTeam(ctx, sqlc.UpsertTeamParams{Slug: slug, DisplayName: name, Conference: conf})
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)
	_, err = r.Read()
	if err != nil {
		return err
	}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		g, err := parseGame(team.ID, rec)
		if err != nil {
			return err
		}
		if _, err := s.q.UpsertGame(ctx, g); err != nil {
			return err
		}
	}
	games, err := s.q.ListGamesByTeamAsc(ctx, team.ID)
	if err != nil {
		return err
	}
	for _, w := range windows {
		payload, err := dashboard.BuildPayload(team, games, int(w))
		if err != nil {
			return err
		}
		raw, err := dashboard.MarshalPayload(payload)
		if err != nil {
			return err
		}
		asOf, _ := dashboard.ParseAsOfDate(payload.AsOfDate)
		if _, err := s.q.UpsertDashboardSnapshot(ctx, sqlc.UpsertDashboardSnapshotParams{TeamID: team.ID, WindowGames: w, AsOfDate: asOf, Payload: raw}); err != nil {
			return err
		}
	}
	return nil
}

func parseGame(teamID int64, rec []string) (sqlc.UpsertGameParams, error) {
	if len(rec) < 6 {
		return sqlc.UpsertGameParams{}, fmt.Errorf("invalid record length")
	}
	gd, err := time.Parse("2006-01-02", rec[0])
	if err != nil {
		return sqlc.UpsertGameParams{}, err
	}
	pf, _ := strconv.Atoi(rec[3])
	pa, _ := strconv.Atoi(rec[4])
	poss, _ := strconv.ParseFloat(rec[5], 64)
	return sqlc.UpsertGameParams{TeamID: teamID, GameDate: gd, Opponent: rec[1], Result: rec[2], PointsFor: int32(pf), PointsAgainst: int32(pa), Possessions: poss, OffensiveRating: float64(pf) / poss * 100, DefensiveRating: float64(pa) / poss * 100}, nil
}
