package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type testResponder struct {
	sessions  map[string][]Session
	drivers   map[string][]Driver
	positions map[string][]Position
	laps      map[string][]Lap
	raceCtrl  map[string][]RaceControlMessage
	stints    map[string][]Stint
}

func (r *testResponder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := req.URL.Query()
	switch req.URL.Path {
	case "/sessions":
		key := q.Get("session_key")
		writeJSON(w, r.sessions[key])
	case "/drivers":
		key := q.Get("session_key")
		writeJSON(w, r.drivers[key])
	case "/position":
		key := positionKey(q)
		writeJSON(w, r.positions[key])
	case "/laps":
		key := lapsKey(q)
		writeJSON(w, r.laps[key])
	case "/race_control":
		key := q.Get("session_key")
		writeJSON(w, r.raceCtrl[key])
	case "/stints":
		key := q.Get("session_key")
		writeJSON(w, r.stints[key])
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func positionKey(v url.Values) string {
	return v.Get("meeting_key") + ":" + v.Get("session_key")
}

func lapsKey(v url.Values) string {
	return v.Get("session_key") + ":" + v.Get("driver_number")
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if v == nil {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("[]")); err != nil {
			panic(fmt.Sprintf("write empty response: %v", err))
		}
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("marshal response: %v", err))
	}
	if _, err := w.Write(data); err != nil {
		panic(fmt.Sprintf("write response: %v", err))
	}
}

func TestServiceDriverStandings(t *testing.T) {
	responder := &testResponder{
		sessions: map[string][]Session{
			"123": {
				{SessionKey: 123, MeetingKey: 45, SessionName: "Race"},
			},
		},
		drivers: map[string][]Driver{
			"123": {
				{DriverNumber: 44, BroadcastName: "L HAMILTON", NameAcronym: "HAM", TeamName: "Mercedes"},
				{DriverNumber: 1, BroadcastName: "M VERSTAPPEN", NameAcronym: "VER", TeamName: "Red Bull"},
				{DriverNumber: 16, BroadcastName: "C LECLERC", NameAcronym: "LEC", TeamName: "Ferrari"},
			},
		},
		positions: map[string][]Position{
			"45:123": {
				{DriverNumber: 1, Position: 2, Date: time.Date(2024, 3, 16, 12, 0, 0, 0, time.UTC)},
				{DriverNumber: 44, Position: 1, Date: time.Date(2024, 3, 16, 12, 0, 1, 0, time.UTC)},
				{DriverNumber: 16, Position: 3, Date: time.Date(2024, 3, 16, 12, 0, 2, 0, time.UTC)},
			},
		},
	}

	ts := httptest.NewServer(responder)
	t.Cleanup(ts.Close)

	svc := NewService(
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
	)

	session, err := svc.Session(context.Background(), 123)
	if err != nil {
		t.Fatalf("Session returned error: %v", err)
	}

	standings, err := svc.DriverStandings(context.Background(), session)
	if err != nil {
		t.Fatalf("DriverStandings returned error: %v", err)
	}

	if len(standings) != 3 {
		t.Fatalf("expected 3 standings, got %d", len(standings))
	}

	if standings[0].Driver.DriverNumber != 44 || standings[0].Position != 1 {
		t.Errorf("expected driver 44 to lead, got %+v", standings[0])
	}

	if standings[1].Driver.DriverNumber != 1 || standings[1].Position != 2 {
		t.Errorf("expected driver 1 to be second, got %+v", standings[1])
	}

	if standings[2].Driver.DriverNumber != 16 || standings[2].Position != 3 {
		t.Errorf("expected driver 16 to be third, got %+v", standings[2])
	}
}

func TestServiceTimingTowerAggregatesLaps(t *testing.T) {
	responder := &testResponder{
		sessions: map[string][]Session{
			"321": {
				{SessionKey: 321, MeetingKey: 54, SessionName: "Race"},
			},
		},
		drivers: map[string][]Driver{
			"321": {
				{DriverNumber: 44, NameAcronym: "HAM", TeamName: "Mercedes"},
				{DriverNumber: 1, NameAcronym: "VER", TeamName: "Red Bull"},
				{DriverNumber: 16, NameAcronym: "LEC", TeamName: "Ferrari"},
			},
		},
		positions: map[string][]Position{
			"54:321": {
				{DriverNumber: 44, Position: 1, Date: time.Date(2024, 3, 16, 12, 0, 1, 0, time.UTC)},
				{DriverNumber: 1, Position: 2, Date: time.Date(2024, 3, 16, 12, 0, 2, 0, time.UTC)},
				{DriverNumber: 16, Position: 3, Date: time.Date(2024, 3, 16, 12, 0, 3, 0, time.UTC)},
			},
		},
		laps: map[string][]Lap{
			"321:44": {
				{DriverNumber: 44, SessionKey: 321, LapNumber: 1, LapDuration: fptr(90.2)},
				{DriverNumber: 44, SessionKey: 321, LapNumber: 2, LapDuration: fptr(89.7)},
			},
			"321:1": {
				{DriverNumber: 1, SessionKey: 321, LapNumber: 1, LapDuration: fptr(91.1)},
				{DriverNumber: 1, SessionKey: 321, LapNumber: 2, LapDuration: fptr(91.4)},
			},
			"321:16": {
				{DriverNumber: 16, SessionKey: 321, LapNumber: 1, LapDuration: fptr(95.0)},
			},
		},
	}

	ts := httptest.NewServer(responder)
	t.Cleanup(ts.Close)

	svc := NewService(
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
	)

	session, err := svc.Session(context.Background(), 321)
	if err != nil {
		t.Fatalf("Session returned error: %v", err)
	}

	tower, err := svc.TimingTower(context.Background(), session)
	if err != nil {
		t.Fatalf("TimingTower returned error: %v", err)
	}

	if len(tower) != 3 {
		t.Fatalf("expected 3 tower entries, got %d", len(tower))
	}

	leader := tower[0]
	if leader.Driver.DriverNumber != 44 || leader.Position != 1 {
		t.Fatalf("expected driver 44 to lead, got %+v", leader)
	}
	if leader.TotalTimeSeconds == nil || math.Abs(*leader.TotalTimeSeconds-179.9) > 1e-6 {
		t.Fatalf("unexpected leader total time: %+v", leader.TotalTimeSeconds)
	}
	if leader.GapToLeaderSeconds == nil || *leader.GapToLeaderSeconds != 0 {
		t.Fatalf("expected leader gap 0, got %+v", leader.GapToLeaderSeconds)
	}

	second := tower[1]
	if second.Driver.DriverNumber != 1 || second.Position != 2 {
		t.Fatalf("expected driver 1 to be second, got %+v", second)
	}
	if second.TotalTimeSeconds == nil || math.Abs(*second.TotalTimeSeconds-182.5) > 1e-6 {
		t.Fatalf("unexpected second total time: %+v", second.TotalTimeSeconds)
	}
	if second.GapToLeaderSeconds == nil || math.Abs(*second.GapToLeaderSeconds-2.6) > 1e-6 {
		t.Fatalf("expected gap 2.6s, got %+v", second.GapToLeaderSeconds)
	}
	if second.IntervalSeconds == nil || math.Abs(*second.IntervalSeconds-2.6) > 1e-6 {
		t.Fatalf("expected interval 2.6s, got %+v", second.IntervalSeconds)
	}

	third := tower[2]
	if third.Driver.DriverNumber != 16 || third.Position != 3 {
		t.Fatalf("expected driver 16 to be third, got %+v", third)
	}
	if third.LapsBehind != 1 {
		t.Fatalf("expected driver 16 to be 1 lap behind, got %d", third.LapsBehind)
	}
	if third.GapToLeaderSeconds != nil {
		t.Fatalf("expected lapped driver to have nil gap, got %+v", third.GapToLeaderSeconds)
	}
	if third.IntervalLapsBehind != 1 {
		t.Fatalf("expected interval laps behind 1, got %d", third.IntervalLapsBehind)
	}
}

func TestServiceLapHistorySorted(t *testing.T) {
	responder := &testResponder{
		laps: map[string][]Lap{
			"111:55": {
				{DriverNumber: 55, SessionKey: 111, LapNumber: 3, LapDuration: fptr(95.1)},
				{DriverNumber: 55, SessionKey: 111, LapNumber: 1, LapDuration: fptr(92.4)},
				{DriverNumber: 55, SessionKey: 111, LapNumber: 2, LapDuration: fptr(93.0)},
			},
		},
	}

	ts := httptest.NewServer(responder)
	t.Cleanup(ts.Close)

	svc := NewService(
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
	)

	laps, err := svc.LapHistory(context.Background(), 111, 55)
	if err != nil {
		t.Fatalf("LapHistory returned error: %v", err)
	}

	if len(laps) != 3 {
		t.Fatalf("expected 3 laps, got %d", len(laps))
	}

	for i, lap := range laps {
		expected := i + 1
		if lap.LapNumber != expected {
			t.Fatalf("expected lap %d at index %d, got %d", expected, i, lap.LapNumber)
		}
	}
}

func TestServiceLatestLapReturnsLastEntry(t *testing.T) {
	responder := &testResponder{
		laps: map[string][]Lap{
			"123:44": {
				{DriverNumber: 44, SessionKey: 123, LapNumber: 1},
				{DriverNumber: 44, SessionKey: 123, LapNumber: 2},
			},
		},
	}

	ts := httptest.NewServer(responder)
	t.Cleanup(ts.Close)

	svc := NewService(
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
	)

	lap, err := svc.LatestLap(context.Background(), 123, 44)
	if err != nil {
		t.Fatalf("LatestLap returned error: %v", err)
	}

	if lap == nil || lap.LapNumber != 2 {
		t.Fatalf("expected lap number 2, got %+v", lap)
	}
}

func TestServiceLatestSessionNoData(t *testing.T) {
	responder := &testResponder{}
	ts := httptest.NewServer(responder)
	t.Cleanup(ts.Close)

	svc := NewService(
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
	)

	_, err := svc.LatestSession(context.Background())
	if !errors.Is(err, ErrNoData) {
		t.Fatalf("expected ErrNoData, got %v", err)
	}
}

func TestServiceRaceControlMessagesSorted(t *testing.T) {
	responder := &testResponder{
		raceCtrl: map[string][]RaceControlMessage{
			"123": {
				{SessionKey: 123, Date: time.Date(2024, 3, 16, 13, 0, 0, 0, time.UTC), Message: "Earlier"},
				{SessionKey: 123, Date: time.Date(2024, 3, 16, 13, 5, 0, 0, time.UTC), Message: "Later"},
			},
		},
	}

	ts := httptest.NewServer(responder)
	t.Cleanup(ts.Close)

	svc := NewService(
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
	)

	messages, err := svc.RaceControlMessages(context.Background(), Session{SessionKey: 123})
	if err != nil {
		t.Fatalf("RaceControlMessages returned error: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("expected 2 race control messages, got %d", len(messages))
	}

	if messages[0].Message != "Later" {
		t.Fatalf("expected messages sorted descending by date, got %q first", messages[0].Message)
	}
}

func fptr(v float64) *float64 {
	return &v
}

func TestServiceTyreStintsSorted(t *testing.T) {
	responder := &testResponder{
		stints: map[string][]Stint{
			"123": {
				{DriverNumber: 16, LapStart: 10, StintNumber: 2},
				{DriverNumber: 16, LapStart: 1, StintNumber: 1},
				{DriverNumber: 55, LapStart: 5, StintNumber: 1},
			},
		},
	}

	ts := httptest.NewServer(responder)
	t.Cleanup(ts.Close)

	svc := NewService(
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
	)

	stints, err := svc.TyreStints(context.Background(), Session{SessionKey: 123})
	if err != nil {
		t.Fatalf("TyreStints returned error: %v", err)
	}

	if len(stints) != 3 {
		t.Fatalf("expected 3 stints, got %d", len(stints))
	}

	if stints[0].DriverNumber != 16 || stints[0].LapStart != 1 || stints[1].LapStart != 10 {
		t.Fatalf("expected driver 16 stints ordered by lap start, got %+v %+v", stints[0], stints[1])
	}

	if stints[2].DriverNumber != 55 {
		t.Fatalf("expected driver 55 last due to driver number sort, got %+v", stints[2])
	}
}
