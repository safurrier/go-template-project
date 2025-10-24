package telemetry

import "time"

// Session represents a racing session (practice, qualifying, race).
type Session struct {
	MeetingKey       int       `json:"meeting_key"`
	SessionKey       int       `json:"session_key"`
	SessionType      string    `json:"session_type"`
	SessionName      string    `json:"session_name"`
	Location         string    `json:"location"`
	CountryName      string    `json:"country_name"`
	CircuitShortName string    `json:"circuit_short_name"`
	DateStart        time.Time `json:"date_start"`
	DateEnd          time.Time `json:"date_end"`
}

// Driver contains the basic driver information returned by the OpenF1 API.
type Driver struct {
	DriverNumber  int    `json:"driver_number"`
	BroadcastName string `json:"broadcast_name"`
	FullName      string `json:"full_name"`
	NameAcronym   string `json:"name_acronym"`
	TeamName      string `json:"team_name"`
	TeamColour    string `json:"team_colour"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
}

// Position represents the on-track position for a driver at a point in time.
type Position struct {
	Date         time.Time `json:"date"`
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	DriverNumber int       `json:"driver_number"`
	Position     int       `json:"position"`
}

// Lap contains the lap telemetry summary for a specific driver.
type Lap struct {
	MeetingKey      int        `json:"meeting_key"`
	SessionKey      int        `json:"session_key"`
	DriverNumber    int        `json:"driver_number"`
	LapNumber       int        `json:"lap_number"`
	DateStart       *time.Time `json:"date_start"`
	DurationSector1 *float64   `json:"duration_sector_1"`
	DurationSector2 *float64   `json:"duration_sector_2"`
	DurationSector3 *float64   `json:"duration_sector_3"`
	LapDuration     *float64   `json:"lap_duration"`
	I1Speed         *float64   `json:"i1_speed"`
	I2Speed         *float64   `json:"i2_speed"`
	StSpeed         *float64   `json:"st_speed"`
	IsPitOutLap     bool       `json:"is_pit_out_lap"`
}

// DriverStanding combines driver information with their latest classified position.
type DriverStanding struct {
	Driver    Driver
	Position  int
	UpdatedAt time.Time
}

// TowerEntry represents the aggregated timing information for a driver.
type TowerEntry struct {
	Driver             Driver
	Position           int
	PositionUpdatedAt  time.Time
	LastLap            *Lap
	BestLap            *Lap
	LapsCompleted      int
	TotalTimeSeconds   *float64
	GapToLeaderSeconds *float64
	IntervalSeconds    *float64
	LapsBehind         int
	IntervalLapsBehind int
}

// RaceControlMessage represents a message from Race Control during a session.
type RaceControlMessage struct {
	MeetingKey   int       `json:"meeting_key"`
	SessionKey   int       `json:"session_key"`
	Date         time.Time `json:"date"`
	DriverNumber *int      `json:"driver_number"`
	LapNumber    *int      `json:"lap_number"`
	Category     string    `json:"category"`
	Flag         *string   `json:"flag"`
	Scope        *string   `json:"scope"`
	Sector       *int      `json:"sector"`
	Message      string    `json:"message"`
}

// Stint tracks tyre usage across a drivers race.
type Stint struct {
	MeetingKey     int    `json:"meeting_key"`
	SessionKey     int    `json:"session_key"`
	StintNumber    int    `json:"stint_number"`
	DriverNumber   int    `json:"driver_number"`
	LapStart       int    `json:"lap_start"`
	LapEnd         int    `json:"lap_end"`
	Compound       string `json:"compound"`
	TyreAgeAtStart int    `json:"tyre_age_at_start"`
}
