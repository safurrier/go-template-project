package dashboard

import (
	"encoding/json"
	"time"

	"github.com/your-org/go-template-project/internal/db/sqlc"
)

type TrendPoint struct {
	Date      string  `json:"date"`
	Opponent  string  `json:"opponent"`
	Result    string  `json:"result"`
	PointsFor int32   `json:"points_for"`
	PointsAg  int32   `json:"points_against"`
	NetRating float64 `json:"net_rating"`
}

type Summary struct {
	Wins      int     `json:"wins"`
	Losses    int     `json:"losses"`
	PointsFor float64 `json:"points_for"`
	PointsAg  float64 `json:"points_against"`
	NetRating float64 `json:"net_rating"`
}

type Payload struct {
	TeamSlug string       `json:"team_slug"`
	TeamName string       `json:"team_name"`
	Window   int          `json:"window"`
	AsOfDate string       `json:"as_of_date"`
	Rolling  Summary      `json:"rolling"`
	Season   Summary      `json:"season"`
	Trend    []TrendPoint `json:"trend"`
}

func BuildPayload(team sqlc.Team, games []sqlc.Game, window int) (Payload, error) {
	if len(games) == 0 {
		return Payload{}, nil
	}
	if window > len(games) {
		window = len(games)
	}
	season := summarize(games)
	recent := games[len(games)-window:]
	rolling := summarize(recent)
	trend := make([]TrendPoint, 0, len(recent))
	for _, g := range recent {
		trend = append(trend, TrendPoint{Date: g.GameDate.Format("2006-01-02"), Opponent: g.Opponent, Result: g.Result, PointsFor: g.PointsFor, PointsAg: g.PointsAgainst, NetRating: g.OffensiveRating - g.DefensiveRating})
	}
	return Payload{TeamSlug: team.Slug, TeamName: team.DisplayName, Window: window, AsOfDate: games[len(games)-1].GameDate.Format("2006-01-02"), Rolling: rolling, Season: season, Trend: trend}, nil
}

func summarize(games []sqlc.Game) Summary {
	if len(games) == 0 {
		return Summary{}
	}
	var wins, losses int
	var pf, pa, net float64
	for _, g := range games {
		if g.Result == "W" {
			wins++
		} else {
			losses++
		}
		pf += float64(g.PointsFor)
		pa += float64(g.PointsAgainst)
		net += g.OffensiveRating - g.DefensiveRating
	}
	count := float64(len(games))
	return Summary{Wins: wins, Losses: losses, PointsFor: pf / count, PointsAg: pa / count, NetRating: net / count}
}

func MarshalPayload(p Payload) (json.RawMessage, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func ParseAsOfDate(asOf string) (time.Time, error) {
	return time.Parse("2006-01-02", asOf)
}
