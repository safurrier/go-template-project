package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the public OpenF1 API root.
	DefaultBaseURL = "https://api.openf1.org/v1"

	defaultHTTPTimeout = 10 * time.Second
)

// ErrNoData is returned when the API responds successfully but does not have data for the request.
var ErrNoData = errors.New("no telemetry data available")

// DataSource exposes the data requirements that the TUI consumes.
type DataSource interface {
	LatestSession(ctx context.Context) (Session, error)
	Session(ctx context.Context, sessionKey int) (Session, error)
	Drivers(ctx context.Context, sessionKey int) ([]Driver, error)
	Positions(ctx context.Context, meetingKey, sessionKey int) ([]Position, error)
	LatestLap(ctx context.Context, sessionKey, driverNumber int) (*Lap, error)
	DriverStandings(ctx context.Context, session Session) ([]DriverStanding, error)
	RaceControlMessages(ctx context.Context, session Session) ([]RaceControlMessage, error)
	TyreStints(ctx context.Context, session Session) ([]Stint, error)
	TimingTower(ctx context.Context, session Session) ([]TowerEntry, error)
	LapHistory(ctx context.Context, sessionKey, driverNumber int) ([]Lap, error)
}

// Service implements DataSource using the OpenF1 HTTP API.
type Service struct {
	client  *http.Client
	baseURL string
}

// Option configures the Service.
type Option func(*Service)

// WithHTTPClient overrides the HTTP client used for requests.
func WithHTTPClient(client *http.Client) Option {
	return func(s *Service) {
		if client != nil {
			s.client = client
		}
	}
}

// WithBaseURL sets a custom base URL. Useful for tests.
func WithBaseURL(baseURL string) Option {
	return func(s *Service) {
		if strings.TrimSpace(baseURL) != "" {
			s.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

// NewService creates a Service instance.
func NewService(opts ...Option) *Service {
	s := &Service{
		client:  &http.Client{Timeout: defaultHTTPTimeout},
		baseURL: DefaultBaseURL,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// LatestSession fetches metadata about the most recent session with live timing data.
func (s *Service) LatestSession(ctx context.Context) (Session, error) {
	var sessions []Session
	if err := s.get(ctx, "/sessions", url.Values{"session_key": []string{"latest"}}, &sessions); err != nil {
		return Session{}, err
	}
	if len(sessions) == 0 {
		return Session{}, ErrNoData
	}
	return sessions[0], nil
}

// Session fetches information for a specific session key.
func (s *Service) Session(ctx context.Context, sessionKey int) (Session, error) {
	var sessions []Session
	params := url.Values{"session_key": []string{fmt.Sprintf("%d", sessionKey)}}
	if err := s.get(ctx, "/sessions", params, &sessions); err != nil {
		return Session{}, err
	}
	if len(sessions) == 0 {
		return Session{}, ErrNoData
	}
	return sessions[0], nil
}

// Drivers returns the drivers participating in the session.
func (s *Service) Drivers(ctx context.Context, sessionKey int) ([]Driver, error) {
	var drivers []Driver
	params := url.Values{"session_key": []string{fmt.Sprintf("%d", sessionKey)}}
	if err := s.get(ctx, "/drivers", params, &drivers); err != nil {
		return nil, err
	}
	if len(drivers) == 0 {
		return nil, ErrNoData
	}
	return drivers, nil
}

// Positions returns the latest known positions for all drivers in the session.
func (s *Service) Positions(ctx context.Context, meetingKey, sessionKey int) ([]Position, error) {
	var positions []Position
	params := url.Values{
		"meeting_key": []string{fmt.Sprintf("%d", meetingKey)},
		"session_key": []string{fmt.Sprintf("%d", sessionKey)},
	}
	if err := s.get(ctx, "/position", params, &positions); err != nil {
		return nil, err
	}
	if len(positions) == 0 {
		return nil, ErrNoData
	}
	return positions, nil
}

// LatestLap fetches the latest lap telemetry summary for the specified driver.
func (s *Service) LatestLap(ctx context.Context, sessionKey, driverNumber int) (*Lap, error) {
	var laps []Lap
	params := url.Values{
		"session_key":   []string{fmt.Sprintf("%d", sessionKey)},
		"driver_number": []string{fmt.Sprintf("%d", driverNumber)},
	}
	if err := s.get(ctx, "/laps", params, &laps); err != nil {
		return nil, err
	}
	if len(laps) == 0 {
		return nil, ErrNoData
	}
	lap := laps[len(laps)-1]
	return &lap, nil
}

// DriverStandings combines drivers and their positions sorted by place.
func (s *Service) DriverStandings(ctx context.Context, session Session) ([]DriverStanding, error) {
	drivers, err := s.Drivers(ctx, session.SessionKey)
	if err != nil {
		return nil, err
	}

	positions, err := s.Positions(ctx, session.MeetingKey, session.SessionKey)
	if err != nil {
		return nil, err
	}

	latestByDriver := map[int]Position{}
	for _, p := range positions {
		current, ok := latestByDriver[p.DriverNumber]
		if !ok {
			latestByDriver[p.DriverNumber] = p
			continue
		}
		if p.Date.After(current.Date) {
			latestByDriver[p.DriverNumber] = p
		}
	}

	standings := make([]DriverStanding, 0, len(drivers))
	for _, driver := range drivers {
		position := Position{}
		if p, ok := latestByDriver[driver.DriverNumber]; ok {
			position = p
		}
		standings = append(standings, DriverStanding{
			Driver:    driver,
			Position:  position.Position,
			UpdatedAt: position.Date,
		})
	}

	sort.SliceStable(standings, func(i, j int) bool {
		pi := standings[i].Position
		pj := standings[j].Position
		switch {
		case pi == 0 && pj == 0:
			return standings[i].Driver.DriverNumber < standings[j].Driver.DriverNumber
		case pi == 0:
			return false
		case pj == 0:
			return true
		case pi == pj:
			return standings[i].Driver.DriverNumber < standings[j].Driver.DriverNumber
		default:
			return pi < pj
		}
	})

	return standings, nil
}

// TimingTower builds a detailed timing snapshot for every driver in the session.
func (s *Service) TimingTower(ctx context.Context, session Session) ([]TowerEntry, error) {
	drivers, err := s.Drivers(ctx, session.SessionKey)
	if err != nil {
		return nil, err
	}

	positions, err := s.Positions(ctx, session.MeetingKey, session.SessionKey)
	if err != nil {
		return nil, err
	}

	latestByDriver := map[int]Position{}
	for _, p := range positions {
		current, ok := latestByDriver[p.DriverNumber]
		if !ok || p.Date.After(current.Date) {
			latestByDriver[p.DriverNumber] = p
		}
	}

	entries := make([]TowerEntry, 0, len(drivers))
	for _, driver := range drivers {
		position := Position{}
		if p, ok := latestByDriver[driver.DriverNumber]; ok {
			position = p
		}
		entries = append(entries, TowerEntry{
			Driver:            driver,
			Position:          position.Position,
			PositionUpdatedAt: position.Date,
		})
	}

	for i := range entries {
		laps, err := s.fetchLaps(ctx, session.SessionKey, entries[i].Driver.DriverNumber)
		if err != nil {
			if errors.Is(err, ErrNoData) {
				continue
			}
			return nil, err
		}
		summariseLaps(&entries[i], laps)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		pi := entries[i].Position
		pj := entries[j].Position
		switch {
		case pi == 0 && pj == 0:
			return entries[i].Driver.DriverNumber < entries[j].Driver.DriverNumber
		case pi == 0:
			return false
		case pj == 0:
			return true
		case pi == pj:
			return entries[i].Driver.DriverNumber < entries[j].Driver.DriverNumber
		default:
			return pi < pj
		}
	})

	computeGaps(entries)

	return entries, nil
}

// LapHistory returns all completed laps for a driver ordered by lap number.
func (s *Service) LapHistory(ctx context.Context, sessionKey, driverNumber int) ([]Lap, error) {
	laps, err := s.fetchLaps(ctx, sessionKey, driverNumber)
	if err != nil {
		return nil, err
	}
	if len(laps) == 0 {
		return nil, ErrNoData
	}
	return laps, nil
}

func (s *Service) fetchLaps(ctx context.Context, sessionKey, driverNumber int) ([]Lap, error) {
	var laps []Lap
	params := url.Values{
		"session_key":   []string{fmt.Sprintf("%d", sessionKey)},
		"driver_number": []string{fmt.Sprintf("%d", driverNumber)},
	}
	if err := s.get(ctx, "/laps", params, &laps); err != nil {
		return nil, err
	}
	if len(laps) == 0 {
		return nil, ErrNoData
	}
	sortLaps(laps)
	return laps, nil
}

func summariseLaps(entry *TowerEntry, laps []Lap) {
	if len(laps) == 0 {
		entry.LastLap = nil
		entry.BestLap = nil
		entry.TotalTimeSeconds = nil
		entry.LapsCompleted = 0
		entry.GapToLeaderSeconds = nil
		entry.IntervalSeconds = nil
		entry.LapsBehind = 0
		entry.IntervalLapsBehind = 0
		return
	}

	var (
		bestIdx       = -1
		bestDuration  float64
		totalDuration float64
		hasDuration   bool
		lapsCompleted int
	)

	entry.LastLap = &laps[len(laps)-1]

	for i := range laps {
		lap := &laps[i]
		if lap.LapDuration != nil && *lap.LapDuration > 0 {
			totalDuration += *lap.LapDuration
			hasDuration = true
			if lap.LapNumber > lapsCompleted {
				lapsCompleted = lap.LapNumber
			}
			if bestIdx == -1 || *lap.LapDuration < bestDuration {
				bestIdx = i
				bestDuration = *lap.LapDuration
			}
		}
	}

	if bestIdx >= 0 {
		entry.BestLap = &laps[bestIdx]
	} else {
		entry.BestLap = nil
	}

	if hasDuration {
		total := totalDuration
		entry.TotalTimeSeconds = &total
	} else {
		entry.TotalTimeSeconds = nil
	}
	entry.LapsCompleted = lapsCompleted
	entry.GapToLeaderSeconds = nil
	entry.IntervalSeconds = nil
	entry.LapsBehind = 0
	entry.IntervalLapsBehind = 0
}

func computeGaps(entries []TowerEntry) {
	if len(entries) == 0 {
		return
	}

	var leader *TowerEntry
	for i := range entries {
		if entries[i].Position == 1 {
			leader = &entries[i]
			break
		}
	}

	if leader != nil && leader.TotalTimeSeconds != nil {
		leader.GapToLeaderSeconds = floatPtr(0)
	}

	for i := range entries {
		entry := &entries[i]
		if leader == nil || leader.TotalTimeSeconds == nil {
			entry.GapToLeaderSeconds = nil
			entry.LapsBehind = 0
			continue
		}
		if entry.TotalTimeSeconds == nil {
			entry.GapToLeaderSeconds = nil
			entry.LapsBehind = 0
			continue
		}
		lapDiff := leader.LapsCompleted - entry.LapsCompleted
		if lapDiff > 0 {
			entry.LapsBehind = lapDiff
			entry.GapToLeaderSeconds = nil
			continue
		}
		entry.LapsBehind = 0
		diff := *entry.TotalTimeSeconds - *leader.TotalTimeSeconds
		if diff < 0 {
			diff = 0
		}
		entry.GapToLeaderSeconds = floatPtr(diff)
	}

	for i := range entries {
		if i == 0 {
			continue
		}
		curr := &entries[i]
		prev := &entries[i-1]
		if curr.TotalTimeSeconds == nil || prev.TotalTimeSeconds == nil {
			curr.IntervalSeconds = nil
			curr.IntervalLapsBehind = 0
			continue
		}
		lapDiff := prev.LapsCompleted - curr.LapsCompleted
		if lapDiff > 0 {
			curr.IntervalLapsBehind = lapDiff
			curr.IntervalSeconds = nil
			continue
		}
		curr.IntervalLapsBehind = 0
		if prev.GapToLeaderSeconds == nil || curr.GapToLeaderSeconds == nil {
			curr.IntervalSeconds = nil
			continue
		}
		interval := *curr.GapToLeaderSeconds - *prev.GapToLeaderSeconds
		if interval < 0 {
			interval = 0
		}
		curr.IntervalSeconds = floatPtr(interval)
	}
}

func sortLaps(laps []Lap) {
	sort.SliceStable(laps, func(i, j int) bool {
		if laps[i].LapNumber == laps[j].LapNumber {
			if laps[i].DateStart == nil || laps[j].DateStart == nil {
				return laps[i].DateStart != nil
			}
			return laps[i].DateStart.Before(*laps[j].DateStart)
		}
		return laps[i].LapNumber < laps[j].LapNumber
	})
}

func floatPtr(v float64) *float64 {
	return &v
}

// RaceControlMessages fetches the most recent Race Control updates for a session.
func (s *Service) RaceControlMessages(ctx context.Context, session Session) ([]RaceControlMessage, error) {
	var messages []RaceControlMessage
	params := url.Values{
		"session_key": []string{fmt.Sprintf("%d", session.SessionKey)},
	}
	if err := s.get(ctx, "/race_control", params, &messages); err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, ErrNoData
	}

	sort.SliceStable(messages, func(i, j int) bool {
		return messages[i].Date.After(messages[j].Date)
	})

	return messages, nil
}

// TyreStints fetches the tyre stint information for each driver in a session.
func (s *Service) TyreStints(ctx context.Context, session Session) ([]Stint, error) {
	var stints []Stint
	params := url.Values{
		"session_key": []string{fmt.Sprintf("%d", session.SessionKey)},
	}
	if err := s.get(ctx, "/stints", params, &stints); err != nil {
		return nil, err
	}
	if len(stints) == 0 {
		return nil, ErrNoData
	}

	sort.SliceStable(stints, func(i, j int) bool {
		if stints[i].DriverNumber == stints[j].DriverNumber {
			if stints[i].LapStart == stints[j].LapStart {
				return stints[i].StintNumber < stints[j].StintNumber
			}
			return stints[i].LapStart < stints[j].LapStart
		}
		return stints[i].DriverNumber < stints[j].DriverNumber
	})

	return stints, nil
}

func (s *Service) get(ctx context.Context, path string, params url.Values, out any) error {
	endpoint := s.baseURL + path
	if params != nil {
		encoded := params.Encode()
		if encoded != "" {
			endpoint = endpoint + "?" + encoded
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 512))
		if readErr != nil {
			return fmt.Errorf("unexpected status %d: unable to read body: %w", resp.StatusCode, readErr)
		}
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
