package dashboard

import (
	"testing"
	"time"

	"github.com/your-org/go-template-project/internal/db/sqlc"
)

func TestBuildPayload(t *testing.T) {
	team := sqlc.Team{Slug: "arizona-mbb", DisplayName: "Arizona Wildcats"}
	games := []sqlc.Game{
		{GameDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Opponent: "A", Result: "W", PointsFor: 80, PointsAgainst: 70, OffensiveRating: 115, DefensiveRating: 101},
		{GameDate: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), Opponent: "B", Result: "L", PointsFor: 70, PointsAgainst: 75, OffensiveRating: 102, DefensiveRating: 109},
	}
	p, err := BuildPayload(team, games, 1)
	if err != nil {
		t.Fatal(err)
	}
	if p.Rolling.Wins != 0 || p.Rolling.Losses != 1 {
		t.Fatalf("unexpected rolling summary: %+v", p.Rolling)
	}
	if p.Season.Wins != 1 || p.Season.Losses != 1 {
		t.Fatalf("unexpected season summary: %+v", p.Season)
	}
}
