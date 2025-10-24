package tui

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/your-org/go-template-project/internal/telemetry"
)

// Config defines how the telemetry TUI behaves.
type Config struct {
	AppName         string
	Version         string
	Service         telemetry.DataSource
	SessionKey      int
	RefreshInterval time.Duration
	DriverFilter    string
	Debug           bool
}

type state int

const (
	stateLoading state = iota
	stateReady
	stateError
)

type screen int

const (
	screenTower screen = iota
	screenRaceControl
	screenStrategy
	screenTracker
	screenHistory
)

var screenSequence = []screen{screenTower, screenRaceControl, screenStrategy, screenTracker, screenHistory}

// Model implements tea.Model to render the telemetry dashboard.
type Model struct {
	cfg             Config
	spinner         spinner.Model
	state           state
	screen          screen
	session         telemetry.Session
	standings       []telemetry.DriverStanding
	tower           []telemetry.TowerEntry
	lapHistory      map[int][]telemetry.Lap
	selected        int
	activeDriver    int
	raceControl     []telemetry.RaceControlMessage
	stints          map[int][]telemetry.Stint
	err             error
	status          string
	width           int
	height          int
	lastTowerUpdate time.Time
}

// NewModel constructs a bubbletea model for the telemetry dashboard.
func NewModel(cfg Config) *Model {
	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return &Model{
		cfg:        cfg,
		spinner:    sp,
		state:      stateLoading,
		screen:     screenTower,
		stints:     make(map[int][]telemetry.Stint),
		lapHistory: make(map[int][]telemetry.Lap),
	}
}

// Init bootstraps the program.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadInitialDataCmd(m.cfg.Service, m.cfg.SessionKey),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h", "shift+tab":
			cmd := m.cycleScreen(-1)
			return m, cmd
		case "right", "l", "tab":
			cmd := m.cycleScreen(1)
			return m, cmd
		case "1", "2", "3", "4", "5":
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(screenSequence) {
				cmd := m.setScreen(screenSequence[idx])
				return m, cmd
			}
			return m, nil
		case "up", "k":
			if m.state != stateReady || m.screen != screenTower || len(m.standings) == 0 {
				break
			}
			if m.selected > 0 {
				m.selected--
				m.setActiveDriver(m.standings[m.selected].Driver.DriverNumber)
				cmd := m.onActiveDriverChanged()
				m.refreshStatusForScreen()
				return m, cmd
			}
		case "down", "j":
			if m.state != stateReady || m.screen != screenTower || len(m.standings) == 0 {
				break
			}
			if m.selected < len(m.standings)-1 {
				m.selected++
				m.setActiveDriver(m.standings[m.selected].Driver.DriverNumber)
				cmd := m.onActiveDriverChanged()
				m.refreshStatusForScreen()
				return m, cmd
			}
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case snapshotLoadedMsg:
		m.session = msg.session
		m.state = stateReady
		m.standings = msg.standings
		m.tower = msg.tower
		if len(m.tower) > 0 {
			m.standings = standingsFromTower(m.tower)
		}
		m.lastTowerUpdate = msg.retrievedAt
		if len(m.standings) == 0 {
			m.status = "Waiting for live timing data"
			m.activeDriver = 0
			m.selected = 0
			cmds := []tea.Cmd{}
			if m.cfg.RefreshInterval > 0 {
				cmds = append(cmds, scheduleRefresh(m.cfg.RefreshInterval))
			}
			cmds = append(cmds,
				loadRaceControlCmd(m.cfg.Service, m.session),
				loadTyreStintsCmd(m.cfg.Service, m.session),
			)
			m.refreshStatusForScreen()
			return m, tea.Batch(cmds...)
		}
		m.applyDriverFilter()
		cmds := []tea.Cmd{}
		if m.cfg.RefreshInterval > 0 {
			cmds = append(cmds, scheduleRefresh(m.cfg.RefreshInterval))
		}
		cmds = append(cmds,
			loadRaceControlCmd(m.cfg.Service, m.session),
			loadTyreStintsCmd(m.cfg.Service, m.session),
		)
		if cmd := m.onActiveDriverChanged(); cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.refreshStatusForScreen()
		return m, tea.Batch(cmds...)
	case standingsUpdatedMsg:
		previousDriver := m.activeDriver
		m.standings = msg.standings
		m.tower = msg.tower
		if len(m.tower) > 0 {
			m.standings = standingsFromTower(m.tower)
		}
		m.lastTowerUpdate = msg.retrievedAt
		if len(m.standings) == 0 {
			m.status = "Waiting for live timing data"
			m.activeDriver = 0
			m.selected = 0
			m.refreshStatusForScreen()
			return m, nil
		}
		if previousDriver != 0 {
			m.moveSelectionToDriver(previousDriver)
		}
		m.refreshStatusForScreen()
		if m.screen == screenHistory {
			if cmd := loadLapHistoryCmd(m.cfg.Service, m.session.SessionKey, m.activeDriver); cmd != nil {
				return m, cmd
			}
		}
		return m, nil
	case lapHistoryLoadedMsg:
		if msg.err != nil {
			if errors.Is(msg.err, telemetry.ErrNoData) {
				m.lapHistory[msg.driverNumber] = nil
				if msg.driverNumber == m.activeDriver {
					m.status = fmt.Sprintf("No lap history yet for #%d", msg.driverNumber)
				}
				break
			}
			m.status = fmt.Sprintf("Lap history error: %v", msg.err)
			break
		}
		m.lapHistory[msg.driverNumber] = msg.laps
		if msg.driverNumber == m.activeDriver {
			m.refreshStatusForScreen()
		}
	case errMsg:
		if errors.Is(msg.err, telemetry.ErrNoData) {
			m.state = stateReady
			m.status = "Waiting for live timing data"
			if len(m.standings) == 0 {
				m.standings = nil
			}
		} else {
			m.err = msg.err
			m.state = stateError
		}
	case raceControlLoadedMsg:
		if msg.err != nil {
			if errors.Is(msg.err, telemetry.ErrNoData) {
				m.raceControl = nil
				m.refreshStatusForScreen()
			} else {
				m.status = fmt.Sprintf("Race control error: %v", msg.err)
			}
			break
		}
		m.raceControl = msg.messages
		m.refreshStatusForScreen()
	case tyreStintsLoadedMsg:
		if msg.err != nil {
			if errors.Is(msg.err, telemetry.ErrNoData) {
				m.stints = make(map[int][]telemetry.Stint)
				m.refreshStatusForScreen()
			} else {
				m.status = fmt.Sprintf("Tyre strategy error: %v", msg.err)
			}
			break
		}
		m.stints = groupStintsByDriver(msg.stints)
		m.refreshStatusForScreen()
	case refreshTickMsg:
		cmds := []tea.Cmd{
			loadStandingsCmd(m.cfg.Service, m.session),
			loadRaceControlCmd(m.cfg.Service, m.session),
			loadTyreStintsCmd(m.cfg.Service, m.session),
		}
		if m.screen == screenHistory {
			if cmd := loadLapHistoryCmd(m.cfg.Service, m.session.SessionKey, m.activeDriver); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if m.cfg.RefreshInterval > 0 {
			cmds = append(cmds, scheduleRefresh(m.cfg.RefreshInterval))
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m *Model) View() string {
	switch m.state {
	case stateLoading:
		return fmt.Sprintf("%s Loading F1 telemetry...", m.spinner.View())
	case stateError:
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		errorText := errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		return errorText + "\nPress q to exit."
	case stateReady:
		if len(m.standings) == 0 {
			return "No driver data available right now. Press q to exit."
		}
		return m.renderLayout()
	default:
		return ""
	}
}

func (m *Model) renderLayout() string {
	header := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle().Render(fmt.Sprintf("%s v%s", m.cfg.AppName, m.cfg.Version)),
		sessionStyle().Render(m.sessionSummary()),
		m.renderTabs(),
	)

	var body string
	switch m.screen {
	case screenRaceControl:
		body = m.renderRaceControl()
	case screenStrategy:
		body = m.renderStrategy()
	case screenTracker:
		body = m.renderTracker()
	case screenHistory:
		body = m.renderHistory()
	default:
		body = m.renderTimingTower()
	}

	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(m.footerText())

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		body,
		footer,
	)
}

func (m *Model) renderStandings() string {
	width := maxInt(100, (m.width*2)/3)
	if len(m.standings) == 0 {
		return boxStyle().Width(width).Render("No driver data yet")
	}

	ref := m.towerEntryForDriver(m.activeDriver)
	var b strings.Builder
	header := fmt.Sprintf("%-4s %-4s %-5s %-12s %-9s %-9s %-7s %-7s %-7s %-8s %-8s %-8s",
		"Pos", "Drv", "Car", "Tyre", "Last", "Best", "S1", "S2", "S3", "Int", "Gap", "Rel")
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
	b.WriteByte('\n')

	for i, standing := range m.standings {
		entry := m.towerEntryForDriver(standing.Driver.DriverNumber)
		prefix := "  "
		style := lipgloss.NewStyle()
		if i == m.selected {
			prefix = "> "
			style = selectedStyle()
		}

		driver := standing.Driver
		acronym := driver.NameAcronym
		if strings.TrimSpace(acronym) == "" {
			acronym = truncateString(driver.BroadcastName, 3)
		}
		if strings.TrimSpace(acronym) == "" {
			acronym = fmt.Sprintf("%d", driver.DriverNumber)
		}

		car := fmt.Sprintf("#%d", driver.DriverNumber)
		tyre := "--"
		last := "--"
		best := "--"
		s1, s2, s3 := "--", "--", "--"
		interval := "--"
		gap := "--"
		relative := "--"

		if entry != nil {
			tyre = m.tyreSummary(entry)
			last = formatLapTime(entry.LastLap)
			best = formatLapTime(entry.BestLap)
			s1 = formatSectorTime(entry.LastLap, 1)
			s2 = formatSectorTime(entry.LastLap, 2)
			s3 = formatSectorTime(entry.LastLap, 3)
			interval = formatInterval(entry.IntervalSeconds, entry.IntervalLapsBehind)
			gap = formatInterval(entry.GapToLeaderSeconds, entry.LapsBehind)
			relative = formatRelative(entry, ref)
		}

		line := fmt.Sprintf("%2s   %-4s %-5s %-12s %-9s %-9s %-7s %-7s %-7s %-8s %-8s %-8s",
			positionLabel(standing.Position), acronym, car, tyre, last, best, s1, s2, s3, interval, gap, relative)
		b.WriteString(style.Render(prefix + line))
		b.WriteByte('\n')
	}

	return boxStyle().Width(width).Render(strings.TrimRight(b.String(), "\n"))
}

func (m *Model) renderDetails() string {
	if len(m.standings) == 0 {
		return ""
	}
	standing := m.standings[m.selected]
	entry := m.towerEntryForDriver(standing.Driver.DriverNumber)

	var rows []string
	rows = append(rows, fmt.Sprintf("Driver: %s", formatName(standing.Driver)))
	rows = append(rows, fmt.Sprintf("Team:   %s", standing.Driver.TeamName))
	rows = append(rows, fmt.Sprintf("Car #:  %d", standing.Driver.DriverNumber))
	if standing.Position > 0 {
		rows = append(rows, fmt.Sprintf("Place:  %d", standing.Position))
	}
	if entry == nil || entry.LastLap == nil {
		rows = append(rows, "Telemetry: no laps completed yet")
	} else {
		lap := entry.LastLap
		rows = append(rows, fmt.Sprintf("Last lap: %d", lap.LapNumber))
		rows = append(rows, fmt.Sprintf("Lap time: %s", formatLapTime(lap)))
		rows = append(rows, fmt.Sprintf("Sector 1: %s", formatSectorTime(lap, 1)))
		rows = append(rows, fmt.Sprintf("Sector 2: %s", formatSectorTime(lap, 2)))
		rows = append(rows, fmt.Sprintf("Sector 3: %s", formatSectorTime(lap, 3)))
		rows = append(rows, fmt.Sprintf("Tyre:   %s", m.tyreSummary(entry)))
		rows = append(rows, fmt.Sprintf("I1 speed: %s", formatSpeed(lap.I1Speed)))
		rows = append(rows, fmt.Sprintf("I2 speed: %s", formatSpeed(lap.I2Speed)))
		rows = append(rows, fmt.Sprintf("Trap speed: %s", formatSpeed(lap.StSpeed)))
		if entry.BestLap != nil {
			rows = append(rows, fmt.Sprintf("Best lap: %d • %s", entry.BestLap.LapNumber, formatLapTime(entry.BestLap)))
		}
		if entry.TotalTimeSeconds != nil {
			rows = append(rows, fmt.Sprintf("Total:   %s", formatDuration(*entry.TotalTimeSeconds)))
		}
		if !m.lastTowerUpdate.IsZero() {
			rows = append(rows, fmt.Sprintf("Updated: %s", m.lastTowerUpdate.Local().Format(time.Kitchen)))
		}
	}

	return boxStyle().Width(maxInt(44, m.width/2)).Render(strings.Join(rows, "\n"))
}

func (m *Model) renderTimingTower() string {
	left := m.renderStandings()
	right := m.renderDetails()
	if right == "" {
		return left
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m *Model) renderRaceControl() string {
	width := maxInt(60, m.width-4)
	if len(m.raceControl) == 0 {
		return boxStyle().Width(width).Render("No race control messages yet")
	}
	limit := minInt(20, len(m.raceControl))
	rows := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		rows = append(rows, m.formatRaceControlLine(m.raceControl[i]))
	}
	return boxStyle().Width(width).Render(strings.Join(rows, "\n"))
}

func (m *Model) renderStrategy() string {
	width := maxInt(60, m.width-4)
	if len(m.standings) == 0 {
		return boxStyle().Width(width).Render("Strategy data available once drivers appear")
	}
	rows := make([]string, 0, len(m.standings))
	for _, standing := range m.standings {
		driver := standing.Driver
		stints := m.stints[driver.DriverNumber]
		if len(stints) == 0 {
			rows = append(rows, fmt.Sprintf("%-3s #%d — no stint data yet", driver.NameAcronym, driver.DriverNumber))
			continue
		}
		stintParts := make([]string, 0, len(stints))
		for _, stint := range stints {
			stintParts = append(stintParts, formatStint(stint))
		}
		stintSummary := strings.Join(stintParts, "   ")
		row := fmt.Sprintf("%-3s #%d  %s", driver.NameAcronym, driver.DriverNumber, stintSummary)
		rows = append(rows, row)
	}
	return boxStyle().Width(width).Render(strings.Join(rows, "\n"))
}

func (m *Model) renderTracker() string {
	width := maxInt(80, m.width-4)
	if len(m.tower) == 0 {
		return boxStyle().Width(width).Render("Driver tracker available once timing data arrives")
	}
	trackWidth := maxInt(40, width-20)
	maxGap := 0.0
	for _, entry := range m.tower {
		if entry.GapToLeaderSeconds != nil && *entry.GapToLeaderSeconds > maxGap {
			maxGap = *entry.GapToLeaderSeconds
		}
		if entry.LapsBehind > 0 {
			if gap := float64(entry.LapsBehind) * 30; gap > maxGap {
				maxGap = gap
			}
		}
	}
	if maxGap <= 0 {
		maxGap = 1
	}
	ref := m.towerEntryForDriver(m.activeDriver)
	rows := make([]string, 0, len(m.tower))
	for i := range m.tower {
		entry := &m.tower[i]
		acronym := entry.Driver.NameAcronym
		if strings.TrimSpace(acronym) == "" {
			acronym = truncateString(entry.Driver.BroadcastName, 3)
		}
		if strings.TrimSpace(acronym) == "" {
			acronym = fmt.Sprintf("%d", entry.Driver.DriverNumber)
		}
		track := []rune(strings.Repeat("·", trackWidth))
		gapValue := 0.0
		if entry.GapToLeaderSeconds != nil {
			gapValue = *entry.GapToLeaderSeconds
		} else if entry.LapsBehind > 0 {
			gapValue = maxGap
		}
		pos := trackWidth - 1 - int((gapValue/maxGap)*float64(trackWidth-1))
		if pos < 0 {
			pos = 0
		}
		if pos >= trackWidth {
			pos = trackWidth - 1
		}
		labelRunes := []rune(strings.ToUpper(acronym))
		if len(labelRunes) == 0 {
			labelRunes = []rune("#")
		}
		marker := labelRunes[0]
		track[pos] = marker
		info := formatInterval(entry.GapToLeaderSeconds, entry.LapsBehind)
		rel := formatRelative(entry, ref)
		row := fmt.Sprintf("%-3s |%s| %8s %8s", acronym, string(track), info, rel)
		if entry.Driver.DriverNumber == m.activeDriver {
			row = selectedStyle().Render(row)
		}
		rows = append(rows, row)
	}
	return boxStyle().Width(width).Render(strings.Join(rows, "\n"))
}

func (m *Model) renderHistory() string {
	width := maxInt(96, m.width-4)
	if m.activeDriver == 0 {
		return boxStyle().Width(width).Render("Select a driver to view lap history")
	}
	laps, ok := m.lapHistory[m.activeDriver]
	if !ok {
		return boxStyle().Width(width).Render("Loading lap history…")
	}
	if len(laps) == 0 {
		return boxStyle().Width(width).Render("No completed laps yet")
	}
	limit := minInt(15, len(laps))
	start := len(laps) - limit
	best, hasBest := bestLapDuration(laps)
	headerFormat := "%-4s %-8s %-8s %-8s %-7s %-7s %-7s %-6s"
	header := lipgloss.NewStyle().Bold(true).
		Render(fmt.Sprintf(headerFormat, "Lap", "Time", "ΔPrev", "ΔBest", "S1", "S2", "S3", "Note"))
	rows := make([]string, 0, limit+1)
	rows = append(rows, header)
	var prevDuration *float64
	for i := start; i < len(laps); i++ {
		lap := laps[i]
		lapCopy := lap
		timeStr := formatLapTime(&lapCopy)
		deltaPrev := "--"
		deltaBest := "--"
		if lap.LapDuration != nil && prevDuration != nil {
			deltaPrev = formatDelta(*lap.LapDuration - *prevDuration)
		}
		if lap.LapDuration != nil && hasBest {
			deltaBest = formatDelta(*lap.LapDuration - best)
		}
		note := ""
		if lap.IsPitOutLap {
			note = "PIT"
		}
		row := fmt.Sprintf("%-4d %-8s %-8s %-8s %-7s %-7s %-7s %-6s",
			lap.LapNumber,
			timeStr,
			deltaPrev,
			deltaBest,
			formatSectorTime(&lapCopy, 1),
			formatSectorTime(&lapCopy, 2),
			formatSectorTime(&lapCopy, 3),
			note,
		)
		rows = append(rows, row)
		if lap.LapDuration != nil && *lap.LapDuration > 0 {
			prevDuration = lap.LapDuration
		}
	}
	return boxStyle().Width(width).Render(strings.Join(rows, "\n"))
}

func (m *Model) renderTabs() string {
	tabs := make([]string, 0, len(screenSequence))
	for i, scr := range screenSequence {
		label := fmt.Sprintf("%d %s", i+1, screenTitle(scr))
		style := tabStyle()
		if scr == m.screen {
			style = tabActiveStyle()
		}
		tabs = append(tabs, style.Render(label))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m *Model) footerText() string {
	controls := []string{fmt.Sprintf("1-%d select view", len(screenSequence)), "←/→ switch view"}
	if (m.screen == screenTower || m.screen == screenHistory || m.screen == screenTracker) && len(m.standings) > 0 {
		controls = append(controls, "↑/↓ change driver")
	}
	controls = append(controls, "q to quit")
	text := strings.Join(controls, " • ")
	if strings.TrimSpace(m.status) != "" {
		text = fmt.Sprintf("%s • %s", text, m.status)
	}
	return text
}

func (m *Model) sessionSummary() string {
	if m.session.SessionKey == 0 {
		return "No session loaded yet"
	}
	parts := []string{summaryName(m.session)}
	if m.session.Location != "" {
		parts = append(parts, m.session.Location)
	}
	if m.session.SessionType != "" {
		parts = append(parts, fmt.Sprintf("(%s)", m.session.SessionType))
	}
	return strings.Join(parts, " • ")
}

func (m *Model) applyDriverFilter() {
	if len(m.standings) == 0 {
		m.activeDriver = 0
		m.selected = 0
		return
	}
	if m.cfg.DriverFilter == "" {
		m.selected = clamp(m.selected, len(m.standings))
		m.setActiveDriver(m.standings[m.selected].Driver.DriverNumber)
		return
	}
	filter := strings.ToLower(strings.TrimSpace(m.cfg.DriverFilter))
	for i, standing := range m.standings {
		driver := standing.Driver
		if strings.ToLower(driver.NameAcronym) == filter ||
			strings.ToLower(driver.BroadcastName) == filter ||
			strings.ToLower(driver.FullName) == filter ||
			strconv.Itoa(driver.DriverNumber) == filter {
			m.selected = i
			m.setActiveDriver(driver.DriverNumber)
			return
		}
	}
	m.selected = 0
	m.setActiveDriver(m.standings[0].Driver.DriverNumber)
}

func (m *Model) moveSelectionToDriver(driverNumber int) {
	for i, standing := range m.standings {
		if standing.Driver.DriverNumber == driverNumber {
			m.selected = i
			m.activeDriver = driverNumber
			return
		}
	}
	m.selected = clamp(0, len(m.standings))
	if len(m.standings) > 0 {
		m.activeDriver = m.standings[m.selected].Driver.DriverNumber
	}
}

func (m *Model) setActiveDriver(driverNumber int) {
	m.activeDriver = driverNumber
}

func (m *Model) onActiveDriverChanged() tea.Cmd {
	if m.session.SessionKey == 0 || m.activeDriver == 0 {
		return nil
	}
	if m.screen == screenHistory {
		return loadLapHistoryCmd(m.cfg.Service, m.session.SessionKey, m.activeDriver)
	}
	return nil
}

func (m *Model) cycleScreen(delta int) tea.Cmd {
	idx := m.screenIndex()
	count := len(screenSequence)
	if count == 0 {
		return nil
	}
	idx = (idx + count + (delta % count)) % count
	return m.setScreen(screenSequence[idx])
}

func (m *Model) setScreen(scr screen) tea.Cmd {
	if scr == m.screen {
		m.refreshStatusForScreen()
		if scr == screenHistory {
			return loadLapHistoryCmd(m.cfg.Service, m.session.SessionKey, m.activeDriver)
		}
		return nil
	}
	m.screen = scr
	cmd := m.onScreenChanged(scr)
	m.refreshStatusForScreen()
	return cmd
}

func (m *Model) onScreenChanged(scr screen) tea.Cmd {
	switch scr {
	case screenHistory:
		return loadLapHistoryCmd(m.cfg.Service, m.session.SessionKey, m.activeDriver)
	default:
		return nil
	}
}

func (m *Model) screenIndex() int {
	for i, scr := range screenSequence {
		if scr == m.screen {
			return i
		}
	}
	return 0
}

func (m *Model) refreshStatusForScreen() {
	switch m.screen {
	case screenRaceControl:
		if len(m.raceControl) == 0 {
			m.status = "Awaiting race control messages"
			return
		}
		latest := m.raceControl[0]
		summary := truncateString(latest.Message, 60)
		timestamp := latest.Date.Local().Format(time.Kitchen)
		m.status = fmt.Sprintf("%s • %s", timestamp, summary)
	case screenStrategy:
		if len(m.stints) == 0 {
			m.status = "Awaiting tyre strategy data"
			return
		}
		m.status = fmt.Sprintf("Strategy info for %d drivers", len(m.stints))
	case screenTracker:
		if len(m.tower) == 0 {
			m.status = "Waiting for timing data"
			return
		}
		if m.activeDriver != 0 {
			if entry := m.towerEntryForDriver(m.activeDriver); entry != nil && entry.LastLap != nil {
				lap := entry.LastLap
				if lap.LapDuration != nil && *lap.LapDuration > 0 {
					m.status = fmt.Sprintf("%s • Lap %d", m.driverLabel(m.activeDriver), lap.LapNumber)
					return
				}
			}
		}
		m.status = fmt.Sprintf("Tracking %d drivers", len(m.tower))
	case screenHistory:
		if m.activeDriver == 0 {
			m.status = "Select a driver to view lap history"
			return
		}
		laps, ok := m.lapHistory[m.activeDriver]
		if !ok {
			m.status = fmt.Sprintf("Loading lap history for #%d", m.activeDriver)
			return
		}
		if len(laps) == 0 {
			m.status = fmt.Sprintf("No lap history yet for #%d", m.activeDriver)
			return
		}
		last := laps[len(laps)-1]
		if last.LapDuration != nil && *last.LapDuration > 0 {
			m.status = fmt.Sprintf("Lap %d • %s", last.LapNumber, formatDuration(*last.LapDuration))
		} else {
			m.status = fmt.Sprintf("Lap %d in progress", last.LapNumber)
		}
	default:
		if len(m.standings) == 0 {
			m.status = "Waiting for live timing data"
			return
		}
		entry := m.towerEntryForDriver(m.activeDriver)
		if entry == nil || entry.LastLap == nil {
			if m.activeDriver != 0 {
				m.status = fmt.Sprintf("Waiting for lap data for #%d", m.activeDriver)
			} else {
				m.status = "Waiting for lap data"
			}
			return
		}
		lap := entry.LastLap
		if lap.LapDuration != nil && *lap.LapDuration > 0 {
			msg := fmt.Sprintf("Lap %d • %s", lap.LapNumber, formatLapTime(lap))
			if !m.lastTowerUpdate.IsZero() {
				msg = fmt.Sprintf("%s @ %s", msg, m.lastTowerUpdate.Local().Format(time.Kitchen))
			}
			m.status = msg
		} else {
			m.status = fmt.Sprintf("Lap %d in progress", lap.LapNumber)
		}
	}
}

func (m *Model) formatRaceControlLine(msg telemetry.RaceControlMessage) string {
	timestamp := msg.Date.Local().Format(time.Kitchen)
	var metadata []string
	if msg.LapNumber != nil && *msg.LapNumber > 0 {
		metadata = append(metadata, fmt.Sprintf("L%d", *msg.LapNumber))
	}
	if msg.DriverNumber != nil {
		metadata = append(metadata, m.driverLabel(*msg.DriverNumber))
	}
	if msg.Flag != nil && strings.TrimSpace(*msg.Flag) != "" {
		metadata = append(metadata, fmt.Sprintf("[%s]", strings.ToUpper(strings.TrimSpace(*msg.Flag))))
	} else if strings.TrimSpace(msg.Category) != "" {
		metadata = append(metadata, msg.Category)
	}
	if msg.Scope != nil && strings.TrimSpace(*msg.Scope) != "" {
		metadata = append(metadata, *msg.Scope)
	}
	meta := strings.Join(metadata, " ")
	if meta != "" {
		meta = " " + meta
	}
	return fmt.Sprintf("%s%s | %s", timestamp, meta, msg.Message)
}

func (m *Model) driverLabel(number int) string {
	for _, standing := range m.standings {
		if standing.Driver.DriverNumber == number {
			acronym := standing.Driver.NameAcronym
			if acronym == "" {
				acronym = standing.Driver.BroadcastName
			}
			return fmt.Sprintf("#%d %s", number, acronym)
		}
	}
	return fmt.Sprintf("#%d", number)
}

func (m *Model) towerEntryForDriver(number int) *telemetry.TowerEntry {
	for i := range m.tower {
		if m.tower[i].Driver.DriverNumber == number {
			return &m.tower[i]
		}
	}
	return nil
}

// loadInitialDataCmd fetches the session, tower, lap history, and driver list
// required for the initial dashboard render in a single call.
func loadInitialDataCmd(service telemetry.DataSource, sessionKey int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()

		var session telemetry.Session
		var err error
		if sessionKey > 0 {
			session, err = service.Session(ctx, sessionKey)
		} else {
			session, err = service.LatestSession(ctx)
		}
		if err != nil {
			return errMsg{err: err}
		}

		standings, err := service.DriverStandings(ctx, session)
		if err != nil && !errors.Is(err, telemetry.ErrNoData) {
			return errMsg{err: err}
		}
		if errors.Is(err, telemetry.ErrNoData) {
			standings = nil
		}

		tower, err := service.TimingTower(ctx, session)
		if err != nil && !errors.Is(err, telemetry.ErrNoData) {
			return errMsg{err: err}
		}
		if errors.Is(err, telemetry.ErrNoData) {
			tower = nil
		}
		if len(tower) > 0 {
			standings = standingsFromTower(tower)
		}

		return snapshotLoadedMsg{session: session, standings: standings, tower: tower, retrievedAt: time.Now()}
	}
}

func loadStandingsCmd(service telemetry.DataSource, session telemetry.Session) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		standings, err := service.DriverStandings(ctx, session)
		if err != nil && !errors.Is(err, telemetry.ErrNoData) {
			return errMsg{err: err}
		}
		if errors.Is(err, telemetry.ErrNoData) {
			standings = nil
		}

		tower, err := service.TimingTower(ctx, session)
		if err != nil && !errors.Is(err, telemetry.ErrNoData) {
			return errMsg{err: err}
		}
		if errors.Is(err, telemetry.ErrNoData) {
			tower = nil
		}
		if len(tower) > 0 {
			standings = standingsFromTower(tower)
		}

		return standingsUpdatedMsg{standings: standings, tower: tower, retrievedAt: time.Now()}
	}
}

func loadLapHistoryCmd(service telemetry.DataSource, sessionKey, driverNumber int) tea.Cmd {
	if sessionKey == 0 || driverNumber == 0 {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		laps, err := service.LapHistory(ctx, sessionKey, driverNumber)
		if err != nil {
			return lapHistoryLoadedMsg{driverNumber: driverNumber, err: err}
		}
		return lapHistoryLoadedMsg{driverNumber: driverNumber, laps: laps}
	}
}

func loadRaceControlCmd(service telemetry.DataSource, session telemetry.Session) tea.Cmd {
	if session.SessionKey == 0 {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		messages, err := service.RaceControlMessages(ctx, session)
		if err != nil {
			return raceControlLoadedMsg{err: err}
		}
		return raceControlLoadedMsg{messages: messages}
	}
}

func loadTyreStintsCmd(service telemetry.DataSource, session telemetry.Session) tea.Cmd {
	if session.SessionKey == 0 {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		stints, err := service.TyreStints(ctx, session)
		if err != nil {
			return tyreStintsLoadedMsg{err: err}
		}
		return tyreStintsLoadedMsg{stints: stints}
	}
}

func scheduleRefresh(interval time.Duration) tea.Cmd {
	if interval <= 0 {
		return nil
	}
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}

func positionLabel(position int) string {
	if position <= 0 {
		return "--"
	}
	return fmt.Sprintf("%d", position)
}

func formatName(driver telemetry.Driver) string {
	if driver.FullName != "" {
		return driver.FullName
	}
	return driver.BroadcastName
}

func formatSpeed(value *float64) string {
	if value == nil {
		return "--"
	}
	return fmt.Sprintf("%.0f km/h", *value)
}

func formatLapTime(lap *telemetry.Lap) string {
	if lap == nil || lap.LapDuration == nil || *lap.LapDuration <= 0 {
		return "--"
	}
	return formatDuration(*lap.LapDuration)
}

func formatSectorTime(lap *telemetry.Lap, sector int) string {
	if lap == nil {
		return "--"
	}
	switch sector {
	case 1:
		return formatDurationPointer(lap.DurationSector1)
	case 2:
		return formatDurationPointer(lap.DurationSector2)
	case 3:
		return formatDurationPointer(lap.DurationSector3)
	default:
		return "--"
	}
}

func formatDurationPointer(value *float64) string {
	if value == nil || *value <= 0 {
		return "--"
	}
	return formatDuration(*value)
}

func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "--"
	}
	minutes := int(seconds) / 60
	remainder := seconds - float64(minutes*60)
	if minutes > 0 {
		return fmt.Sprintf("%d:%06.3f", minutes, remainder)
	}
	return fmt.Sprintf("%.3f", remainder)
}

func formatInterval(seconds *float64, lapsBehind int) string {
	if lapsBehind > 0 {
		return fmt.Sprintf("+%dL", lapsBehind)
	}
	if seconds == nil {
		return "--"
	}
	return fmt.Sprintf("+%.3f", *seconds)
}

func formatRelative(entry, ref *telemetry.TowerEntry) string {
	if entry == nil {
		return "--"
	}
	if ref == nil || entry.Driver.DriverNumber == ref.Driver.DriverNumber {
		return "0.000"
	}
	lapDiff := entry.LapsCompleted - ref.LapsCompleted
	if lapDiff > 0 {
		return fmt.Sprintf("-%dL", lapDiff)
	}
	if lapDiff < 0 {
		return fmt.Sprintf("+%dL", -lapDiff)
	}
	if entry.TotalTimeSeconds == nil || ref.TotalTimeSeconds == nil {
		return "--"
	}
	diff := *entry.TotalTimeSeconds - *ref.TotalTimeSeconds
	sign := "+"
	if diff < 0 {
		sign = "-"
		diff = -diff
	}
	return fmt.Sprintf("%s%.3f", sign, diff)
}

func formatDelta(delta float64) string {
	sign := "+"
	if delta < 0 {
		sign = "-"
		delta = -delta
	}
	return fmt.Sprintf("%s%.3f", sign, delta)
}

func (m *Model) tyreSummary(entry *telemetry.TowerEntry) string {
	if entry == nil {
		return "--"
	}
	stints := m.stints[entry.Driver.DriverNumber]
	if len(stints) == 0 {
		return "--"
	}
	lap := entry.LapsCompleted
	if lap == 0 && entry.LastLap != nil {
		lap = entry.LastLap.LapNumber
	}
	if lap == 0 {
		stint := stints[len(stints)-1]
		return formatCompoundSummary(stint.Compound, stint.TyreAgeAtStart)
	}
	for i := len(stints) - 1; i >= 0; i-- {
		stint := stints[i]
		if stint.LapStart > 0 && lap < stint.LapStart {
			continue
		}
		if stint.LapEnd > 0 && lap > stint.LapEnd {
			continue
		}
		age := stint.TyreAgeAtStart
		if stint.LapStart > 0 && lap >= stint.LapStart {
			age += lap - stint.LapStart + 1
		}
		return formatCompoundSummary(stint.Compound, age)
	}
	stint := stints[len(stints)-1]
	age := stint.TyreAgeAtStart
	if stint.LapStart > 0 && lap >= stint.LapStart {
		age += lap - stint.LapStart + 1
	}
	return formatCompoundSummary(stint.Compound, age)
}

func formatCompoundSummary(compound string, age int) string {
	label := strings.ToUpper(strings.TrimSpace(compound))
	if label == "" {
		label = "TYR"
	}
	if len(label) > 3 {
		label = label[:3]
	}
	if age <= 0 {
		return label
	}
	return fmt.Sprintf("%s (%d)", label, age)
}

func bestLapDuration(laps []telemetry.Lap) (float64, bool) {
	var best float64
	found := false
	for _, lap := range laps {
		if lap.LapDuration != nil && *lap.LapDuration > 0 {
			if !found || *lap.LapDuration < best {
				best = *lap.LapDuration
				found = true
			}
		}
	}
	return best, found
}

func clamp(selected, length int) int {
	if length == 0 {
		return 0
	}
	if selected < 0 {
		return 0
	}
	if selected >= length {
		return length - 1
	}
	return selected
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func summaryName(session telemetry.Session) string {
	if session.SessionName != "" {
		return session.SessionName
	}
	return session.SessionType
}

func titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).MarginBottom(1)
}

func sessionStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("245")).MarginBottom(1)
}

func boxStyle() lipgloss.Style {
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1).MarginRight(2)
}

func selectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true)
}

func screenTitle(s screen) string {
	switch s {
	case screenTower:
		return "Timing Tower"
	case screenRaceControl:
		return "Race Control"
	case screenStrategy:
		return "Tyre Strategy"
	case screenTracker:
		return "Driver Tracker"
	case screenHistory:
		return "Lap History"
	default:
		return "Timing Tower"
	}
}

func tabStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Padding(0, 2).MarginRight(1)
}

func tabActiveStyle() lipgloss.Style {
	return tabStyle().Foreground(lipgloss.Color("229")).Bold(true).Underline(true)
}

func truncateString(value string, limit int) string {
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	if limit == 1 {
		return string(runes[:limit])
	}
	return string(runes[:limit-1]) + "…"
}

func formatStint(stint telemetry.Stint) string {
	label := compoundLabel(stint.Compound)
	return fmt.Sprintf("%s L%d-%d (age %d)", label, stint.LapStart, stint.LapEnd, stint.TyreAgeAtStart)
}

func compoundLabel(compound string) string {
	c := strings.ToUpper(strings.TrimSpace(compound))
	switch c {
	case "SOFT":
		return "S"
	case "MEDIUM":
		return "M"
	case "HARD":
		return "H"
	case "INTERMEDIATE":
		return "I"
	case "WET":
		return "W"
	default:
		if c == "" {
			return "?"
		}
		if len(c) == 1 {
			return c
		}
		return string([]rune(c)[:1])
	}
}

func groupStintsByDriver(stints []telemetry.Stint) map[int][]telemetry.Stint {
	grouped := make(map[int][]telemetry.Stint)
	for _, stint := range stints {
		grouped[stint.DriverNumber] = append(grouped[stint.DriverNumber], stint)
	}
	for number, entries := range grouped {
		sort.SliceStable(entries, func(i, j int) bool {
			if entries[i].LapStart == entries[j].LapStart {
				return entries[i].StintNumber < entries[j].StintNumber
			}
			return entries[i].LapStart < entries[j].LapStart
		})
		grouped[number] = entries
	}
	return grouped
}

func standingsFromTower(entries []telemetry.TowerEntry) []telemetry.DriverStanding {
	if len(entries) == 0 {
		return nil
	}
	standings := make([]telemetry.DriverStanding, 0, len(entries))
	for _, entry := range entries {
		standings = append(standings, telemetry.DriverStanding{
			Driver:    entry.Driver,
			Position:  entry.Position,
			UpdatedAt: entry.PositionUpdatedAt,
		})
	}
	return standings
}

// Messages used within the update loop.
type (
	snapshotLoadedMsg struct {
		session     telemetry.Session
		standings   []telemetry.DriverStanding
		tower       []telemetry.TowerEntry
		retrievedAt time.Time
	}

	standingsUpdatedMsg struct {
		standings   []telemetry.DriverStanding
		tower       []telemetry.TowerEntry
		retrievedAt time.Time
	}

	lapHistoryLoadedMsg struct {
		driverNumber int
		laps         []telemetry.Lap
		err          error
	}

	raceControlLoadedMsg struct {
		messages []telemetry.RaceControlMessage
		err      error
	}

	tyreStintsLoadedMsg struct {
		stints []telemetry.Stint
		err    error
	}

	errMsg struct {
		err error
	}

	refreshTickMsg struct{}
)
