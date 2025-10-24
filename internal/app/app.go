package app

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/your-org/go-template-project/internal/telemetry"
	"github.com/your-org/go-template-project/internal/tui"
)

// programRunner is the subset of tea.Program that App interacts with.
type programRunner interface {
	Run() (tea.Model, error)
}

// programFactory creates a new bubbletea program for a model.
type programFactory func(model tea.Model, opts ...tea.ProgramOption) programRunner

// App represents the telemetry CLI application.
type App struct {
	Name    string
	Version string
	Debug   bool

	sessionKey      int
	refreshInterval time.Duration
	driverFilter    string

	telemetryService telemetry.DataSource
	programFactory   programFactory

	httpClient *http.Client
	baseURL    string
}

// Option configures the application.
type Option func(*App)

// WithSessionKey forces the TUI to load a specific session key.
func WithSessionKey(sessionKey int) Option {
	return func(a *App) {
		if sessionKey > 0 {
			a.sessionKey = sessionKey
		}
	}
}

// WithRefreshInterval customises how often the leaderboard refreshes.
func WithRefreshInterval(interval time.Duration) Option {
	return func(a *App) {
		if interval > 0 {
			a.refreshInterval = interval
		}
	}
}

// WithDriverFilter sets a preferred driver to focus on (number, acronym or name).
func WithDriverFilter(filter string) Option {
	return func(a *App) {
		if strings.TrimSpace(filter) != "" {
			a.driverFilter = filter
		}
	}
}

// WithTelemetryService injects a custom telemetry datasource (useful for tests).
func WithTelemetryService(service telemetry.DataSource) Option {
	return func(a *App) {
		if service != nil {
			a.telemetryService = service
		}
	}
}

// WithProgramFactory allows tests to override the bubbletea program creation.
func WithProgramFactory(factory programFactory) Option {
	return func(a *App) {
		if factory != nil {
			a.programFactory = factory
		}
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(a *App) {
		if client != nil {
			a.httpClient = client
		}
	}
}

// WithBaseURL overrides the default OpenF1 API root.
func WithBaseURL(baseURL string) Option {
	return func(a *App) {
		if strings.TrimSpace(baseURL) != "" {
			a.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

// New creates a new application instance.
func New(name, version string, opts ...Option) *App {
	a := &App{
		Name:            name,
		Version:         version,
		Debug:           os.Getenv("DEBUG") == "true",
		refreshInterval: 5 * time.Second,
		baseURL:         telemetry.DefaultBaseURL,
	}

	for _, opt := range opts {
		opt(a)
	}

	if a.programFactory == nil {
		a.programFactory = func(model tea.Model, opts ...tea.ProgramOption) programRunner {
			return tea.NewProgram(model, opts...)
		}
	}

	return a
}

// Run launches the telemetry dashboard.
func (a *App) Run() error {
	service := a.telemetryService
	if service == nil {
		client := a.httpClient
		if client == nil {
			client = &http.Client{Timeout: 10 * time.Second}
		}
		service = telemetry.NewService(
			telemetry.WithHTTPClient(client),
			telemetry.WithBaseURL(a.baseURL),
		)
	}

	model := tui.NewModel(tui.Config{
		AppName:         a.Name,
		Version:         a.Version,
		Service:         service,
		SessionKey:      a.sessionKey,
		RefreshInterval: a.refreshInterval,
		DriverFilter:    a.driverFilter,
		Debug:           a.Debug,
	})

	opts := []tea.ProgramOption{}
	if !a.Debug {
		opts = append(opts, tea.WithAltScreen())
	}
	program := a.programFactory(model, opts...)

	_, err := program.Run()
	if err != nil {
		return err
	}

	return nil
}

// GetInfo returns basic application information for diagnostics.
func (a *App) GetInfo() map[string]string {
	info := map[string]string{
		"name":             a.Name,
		"version":          a.Version,
		"debug":            strconv.FormatBool(a.Debug),
		"refresh_interval": a.refreshInterval.String(),
	}
	if a.sessionKey > 0 {
		info["session_key"] = strconv.Itoa(a.sessionKey)
	}
	if strings.TrimSpace(a.driverFilter) != "" {
		info["driver_filter"] = a.driverFilter
	}
	return info
}
