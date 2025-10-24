package app

import (
	"context"
	"reflect"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/your-org/go-template-project/internal/telemetry"
	"github.com/your-org/go-template-project/internal/tui"
)

type stubService struct {
	session    telemetry.Session
	standings  []telemetry.DriverStanding
	lap        *telemetry.Lap
	errorToUse error
	raceCtrl   []telemetry.RaceControlMessage
	stints     []telemetry.Stint
	tower      []telemetry.TowerEntry
	history    map[int][]telemetry.Lap
}

func (s *stubService) LatestSession(context.Context) (telemetry.Session, error) {
	if s.session.SessionKey == 0 {
		return telemetry.Session{}, telemetry.ErrNoData
	}
	return s.session, nil
}

func (s *stubService) Session(context.Context, int) (telemetry.Session, error) {
	if s.session.SessionKey == 0 {
		return telemetry.Session{}, telemetry.ErrNoData
	}
	return s.session, nil
}

func (s *stubService) Drivers(context.Context, int) ([]telemetry.Driver, error) {
	return nil, telemetry.ErrNoData
}

func (s *stubService) Positions(context.Context, int, int) ([]telemetry.Position, error) {
	return nil, telemetry.ErrNoData
}

func (s *stubService) LatestLap(context.Context, int, int) (*telemetry.Lap, error) {
	return s.lap, s.errorToUse
}

func (s *stubService) DriverStandings(context.Context, telemetry.Session) ([]telemetry.DriverStanding, error) {
	if s.errorToUse != nil {
		return nil, s.errorToUse
	}
	return s.standings, nil
}

func (s *stubService) RaceControlMessages(context.Context, telemetry.Session) ([]telemetry.RaceControlMessage, error) {
	if s.errorToUse != nil {
		return nil, s.errorToUse
	}
	if len(s.raceCtrl) == 0 {
		return nil, telemetry.ErrNoData
	}
	return s.raceCtrl, nil
}

func (s *stubService) TyreStints(context.Context, telemetry.Session) ([]telemetry.Stint, error) {
	if s.errorToUse != nil {
		return nil, s.errorToUse
	}
	if len(s.stints) == 0 {
		return nil, telemetry.ErrNoData
	}
	return s.stints, nil
}

func (s *stubService) TimingTower(context.Context, telemetry.Session) ([]telemetry.TowerEntry, error) {
	if s.errorToUse != nil {
		return nil, s.errorToUse
	}
	if len(s.tower) == 0 {
		return nil, telemetry.ErrNoData
	}
	return s.tower, nil
}

func (s *stubService) LapHistory(_ context.Context, _ int, driverNumber int) ([]telemetry.Lap, error) {
	if s.errorToUse != nil {
		return nil, s.errorToUse
	}
	if len(s.history) == 0 {
		return nil, telemetry.ErrNoData
	}
	laps, ok := s.history[driverNumber]
	if !ok || len(laps) == 0 {
		return nil, telemetry.ErrNoData
	}
	return laps, nil
}

type fakeProgram struct {
	runs  int
	model tea.Model
	err   error
}

func (p *fakeProgram) Run() (tea.Model, error) {
	p.runs++
	return p.model, p.err
}

func TestNewHonoursDebugEnv(t *testing.T) {
	t.Setenv("DEBUG", "true")
	a := New("test-app", "1.0.0")
	if !a.Debug {
		t.Fatalf("expected debug mode to be enabled when DEBUG env is true")
	}
}

func TestRunUsesInjectedProgramFactory(t *testing.T) {
	service := &stubService{
		session:   telemetry.Session{SessionKey: 123, MeetingKey: 45, SessionName: "Race"},
		standings: []telemetry.DriverStanding{},
	}

	program := &fakeProgram{}
	var capturedOpts []tea.ProgramOption

	a := New(
		"test-app",
		"1.0.0",
		WithTelemetryService(service),
		WithProgramFactory(func(model tea.Model, opts ...tea.ProgramOption) programRunner {
			if _, ok := model.(*tui.Model); !ok {
				t.Fatalf("expected model to be *tui.Model, got %T", model)
			}
			capturedOpts = append([]tea.ProgramOption{}, opts...)
			return program
		}),
	)

	if err := a.Run(); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if program.runs != 1 {
		t.Fatalf("expected program to run once, ran %d times", program.runs)
	}

	if len(capturedOpts) != 1 {
		t.Fatalf("expected alt screen option to be passed when not in debug, got %d options", len(capturedOpts))
	}
}

func TestRunWithDebugDisablesAltScreen(t *testing.T) {
	t.Setenv("DEBUG", "true")
	service := &stubService{
		session: telemetry.Session{SessionKey: 5, MeetingKey: 6, SessionName: "Qualifying"},
	}

	program := &fakeProgram{}
	var capturedOpts []tea.ProgramOption

	a := New(
		"test-app",
		"1.0.0",
		WithTelemetryService(service),
		WithProgramFactory(func(model tea.Model, opts ...tea.ProgramOption) programRunner {
			capturedOpts = append([]tea.ProgramOption{}, opts...)
			return program
		}),
	)

	if err := a.Run(); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(capturedOpts) != 0 {
		t.Fatalf("expected no program options when in debug mode, got %d", len(capturedOpts))
	}
}

func TestGetInfoIncludesConfig(t *testing.T) {
	a := New(
		"telemetry",
		"1.2.3",
		WithSessionKey(999),
		WithRefreshInterval(2*time.Second),
		WithDriverFilter("VER"),
	)

	info := a.GetInfo()
	if info["name"] != "telemetry" || info["version"] != "1.2.3" {
		t.Fatalf("unexpected name/version in info: %#v", info)
	}

	if info["session_key"] != "999" {
		t.Fatalf("expected session_key 999, got %s", info["session_key"])
	}

	if got := info["driver_filter"]; got != "VER" {
		t.Fatalf("expected driver_filter 'VER', got %s", got)
	}

	if got := info["refresh_interval"]; got == "" {
		t.Fatalf("refresh interval missing from info")
	}
}

func TestWithBaseURLOverridesDefault(t *testing.T) {
	a := New("telemetry", "0.0.1", WithBaseURL("https://example.com/api"))
	if a.baseURL != "https://example.com/api" {
		t.Fatalf("expected baseURL to be overridden, got %s", a.baseURL)
	}
}

func TestWithProgramFactoryNil(t *testing.T) {
	// Ensure providing nil program factory leaves the default intact.
	a := New("telemetry", "0.0.1", WithProgramFactory(nil))
	if reflect.ValueOf(a.programFactory).IsNil() {
		t.Fatal("expected default program factory to be set")
	}
}
