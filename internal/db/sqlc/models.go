package sqlc

import (
	"encoding/json"
	"time"
)

type Team struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	DisplayName string    `json:"display_name"`
	Conference  string    `json:"conference"`
	CreatedAt   time.Time `json:"created_at"`
}

type Game struct {
	ID              int64     `json:"id"`
	TeamID          int64     `json:"team_id"`
	GameDate        time.Time `json:"game_date"`
	Opponent        string    `json:"opponent"`
	PointsFor       int32     `json:"points_for"`
	PointsAgainst   int32     `json:"points_against"`
	Possessions     float64   `json:"possessions"`
	Result          string    `json:"result"`
	OffensiveRating float64   `json:"offensive_rating"`
	DefensiveRating float64   `json:"defensive_rating"`
	CreatedAt       time.Time `json:"created_at"`
}

type DashboardSnapshot struct {
	ID          int64           `json:"id"`
	TeamID      int64           `json:"team_id"`
	WindowGames int32           `json:"window_games"`
	AsOfDate    time.Time       `json:"as_of_date"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
