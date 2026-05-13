package tui

import (
	"fmt"
	"strings"

	"nextdns_client/internal/api"
	"nextdns_client/internal/config"
	"nextdns_client/internal/timer"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	viewMain       = "main"
	viewAppInput   = "app_input"
	viewTimerInput = "timer_input"
	viewUrlInput   = "url_input"
	viewUrlList    = "url_list"
)

// Model represents the terminal UI model
type Model struct {
	apiKey     string
	profileID  string
	config     *config.Config
	apiClient  *api.APIClient
	configPath string

	cursor    int
	urlCursor int
	// currentView is the current view
	currentView string
	err         error
	message     string

	// Input states
	timerInput string
	urlInput   string
	appInput   string // for new app name input
	activeApp  int    // Index of app being edited

	// Debug logging
	debugMode bool
	logChan   chan string
	debugLogs []string
}

// NewModel creates a new UI model
func NewModel(apiKey, profileID string, cfg *config.Config, apiClient *api.APIClient, configPath string, debug bool) Model {
	m := Model{
		apiKey:      apiKey,
		profileID:   profileID,
		config:      cfg,
		apiClient:   apiClient,
		configPath:  configPath,
		currentView: viewMain,
		debugMode:   debug,
	}
	if debug {
		m.logChan = make(chan string, 100)
		apiClient.SetLogChannel(m.logChan)
	}
	return m
}

type logMsg string

func waitForLogs(ch chan string) tea.Cmd {
	return func() tea.Msg {
		return logMsg(<-ch)
	}
}

func (m Model) Init() tea.Cmd {
	if m.debugMode {
		return waitForLogs(m.logChan)
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case logMsg:
		m.debugLogs = append(m.debugLogs, string(msg))
		if len(m.debugLogs) > 5 {
			m.debugLogs = m.debugLogs[1:]
		}
		return m, waitForLogs(m.logChan)

	case tea.KeyMsg:
		switch m.currentView {
		case viewAppInput:
			return m.handleAppInput(msg)
		case viewTimerInput:
			return m.handleTimerInput(msg)
		case viewUrlInput:
			return m.handleURLInput(msg)
		case viewUrlList:
			return m.handleURLList(msg)
		default:
			return m.handleMainView(msg)
		}
	}

	return m, nil
}

func (m Model) handleMainView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.config.Applications)-1 {
			m.cursor++
		}

	case " ": // Toggle enabled status
		m.toggleApp(m.cursor)

	case "t": // Set timer
		m.currentView = viewTimerInput
		m.activeApp = m.cursor
		m.timerInput = m.config.Applications[m.cursor].Timer

	case "enter", "e": // Manage URLs
		m.currentView = viewUrlList
		m.activeApp = m.cursor
		m.urlCursor = 0

	case "a": // Add new app group
		m.currentView = viewAppInput
		m.activeApp = m.cursor
		m.appInput = ""
	}
	return m, nil
}

func (m Model) handleURLList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	app := &m.config.Applications[m.activeApp]

	switch msg.String() {
	case "esc", "q":
		m.currentView = viewMain

	case "up", "k":
		if m.urlCursor > 0 {
			m.urlCursor--
		}

	case "down", "j":
		if m.urlCursor < len(app.URLs)-1 {
			m.urlCursor++
		}

	case "a": // Add URL
		m.currentView = viewUrlInput
		m.urlInput = ""

	case "d", "x", "backspace": // Delete URL
		if len(app.URLs) > 0 {
			urlToRemove := app.URLs[m.urlCursor]
			app.URLs = append(app.URLs[:m.urlCursor], app.URLs[m.urlCursor+1:]...)
			if m.urlCursor >= len(app.URLs) && m.urlCursor > 0 {
				m.urlCursor--
			}

			// If app is disabled, remove from NextDNS too (since it was being blocked)
			if !app.Enabled {
				go m.apiClient.RemoveFromDenylist(urlToRemove)
			}
			config.Save(m.config, m.configPath)
		}
	}
	return m, nil
}

func (m *Model) toggleApp(index int) {
	app := &m.config.Applications[index]
	app.Enabled = !app.Enabled

	status := "Blocking"
	if app.Enabled {
		status = "Allowing"
	}
	m.message = fmt.Sprintf("%s NextDNS for %s...", status, app.Name)

	// Create local copies of data needed in goroutine to avoid race conditions
	urls := make([]string, len(app.URLs))
	copy(urls, app.URLs)
	enabled := app.Enabled
	appName := app.Name
	appTimer := app.Timer

	go func() {
		for _, url := range urls {
			if enabled {
				m.apiClient.RemoveFromDenylist(url)
			} else {
				m.apiClient.AddToDenylist(url)
			}
		}

		if enabled && appTimer != "" {
			duration, err := timer.ParseTimer(appTimer)
			if err == nil {
				timer.AddTimer(appName, duration, appName)
			}
		}

		config.Save(m.config, m.configPath)
	}()
}

func (m Model) handleTimerInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.timerInput == "" {
			m.config.Applications[m.activeApp].Timer = ""
			m.message = fmt.Sprintf("Timer cleared for %s", m.config.Applications[m.activeApp].Name)
		} else {
			duration, err := timer.ParseTimer(m.timerInput)
			if err != nil {
				m.err = err
				m.currentView = "main"
				return m, nil
			}
			m.config.Applications[m.activeApp].Timer = m.timerInput
			if m.config.Applications[m.activeApp].Enabled {
				timer.AddTimer(m.config.Applications[m.activeApp].Name, duration, m.config.Applications[m.activeApp].Name)
			}
			m.message = fmt.Sprintf("Timer set for %s: %s", m.config.Applications[m.activeApp].Name, m.timerInput)
		}
		config.Save(m.config, m.configPath)
		m.currentView = "main"
		m.err = nil

	case "esc":
		m.currentView = "main"

	case "backspace":
		if len(m.timerInput) > 0 {
			m.timerInput = m.timerInput[:len(m.timerInput)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.timerInput += msg.String()
		}
	}
	return m, nil
}

func (m Model) handleAppInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.appInput != "" {
			newApp := config.Application{
				Name:    m.appInput,
				URLs:    []string{},
				Enabled: false,
			}
			m.config.Applications = append(m.config.Applications, newApp)
			config.Save(m.config, m.configPath)
			m.activeApp = len(m.config.Applications) - 1
			m.urlInput = ""
			m.currentView = viewUrlInput
			m.message = fmt.Sprintf("Created app group '%s' — add URLs now", m.appInput)
		}
		m.appInput = ""

	case "esc":
		m.currentView = viewMain

	case "backspace":
		if len(m.appInput) > 0 {
			m.appInput = m.appInput[:len(m.appInput)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.appInput += msg.String()
		}
	}
	return m, nil
}

func (m Model) handleURLInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.urlInput != "" {
			app := &m.config.Applications[m.activeApp]
			app.URLs = append(app.URLs, m.urlInput)

			// If app is disabled, add to NextDNS too (since it is currently blocked)
			if !app.Enabled {
				go m.apiClient.AddToDenylist(m.urlInput)
			}
			config.Save(m.config, m.configPath)
		}
		m.currentView = viewUrlList

	case "esc":
		m.currentView = viewUrlList

	case "backspace":
		if len(m.urlInput) > 0 {
			m.urlInput = m.urlInput[:len(m.urlInput)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.urlInput += msg.String()
		}
	}
	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("NextDNS Client - Manage Application Groups"))
	s.WriteString("\n\n")

	switch m.currentView {
	case viewTimerInput:
		s.WriteString(fmt.Sprintf("Enter timer for %s (e.g., 1h30m, 5m, 70s):\n", m.config.Applications[m.activeApp].Name))
		s.WriteString(inputStyle.Render(m.timerInput + "_"))
		s.WriteString("\n\n(Enter to save, Esc to cancel, Backspace to clear)")

	case viewAppInput:
		s.WriteString("Enter new application group name:\n")
		s.WriteString(inputStyle.Render(m.appInput + "_"))
		s.WriteString("\n\n(Enter to save and add URLs, Esc to cancel, Backspace to clear)")

	case viewUrlInput:
		s.WriteString(fmt.Sprintf("Add URL for %s:\n", m.config.Applications[m.activeApp].Name))
		s.WriteString(inputStyle.Render(m.urlInput + "_"))
		s.WriteString("\n\n(Enter to save, Esc to cancel)")

	case viewUrlList:
		app := m.config.Applications[m.activeApp]
		s.WriteString(fmt.Sprintf("URLs for %s:\n\n", nameStyle.Render(app.Name)))
		if len(app.URLs) == 0 {
			s.WriteString("  No URLs added yet.\n")
		} else {
			for i, url := range app.URLs {
				cursor := " "
				if m.urlCursor == i {
					cursor = ">"
				}
				s.WriteString(fmt.Sprintf("%s %s\n", cursorStyle.Render(cursor), url))
			}
		}
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("↑/↓: navigate • a: add • d: delete • Esc: back"))

	default:
		for i, app := range m.config.Applications {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			status := "Disabled"
			statusStyle := disabledStyle
			if app.Enabled {
				status = "Enabled "
				statusStyle = enabledStyle
			}

			timerStr := ""
			if app.Timer != "" {
				timerStr = fmt.Sprintf(" (Timer: %s)", app.Timer)
			}

			line := fmt.Sprintf("%s %s %s%s",
				cursorStyle.Render(cursor),
				statusStyle.Render(status),
				nameStyle.Render(app.Name),
				timerStr,
			)
			s.WriteString(line + "\n")

			if m.cursor == i {
				s.WriteString(fmt.Sprintf("    URLs: %s\n", strings.Join(app.URLs, ", ")))
			}
		}

		s.WriteString("\n")
		if m.message != "" {
			s.WriteString(messageStyle.Render(m.message) + "\n\n")
		}
		if m.err != nil {
			s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
		}

		s.WriteString(helpStyle.Render("↑/↓: navigate • Space: toggle • t: timer • Enter: edit URLs • a: add app • q: quit"))
	}

	if m.debugMode {
		s.WriteString("\n\n" + debugTitleStyle.Render("--- API Logs ---") + "\n")
		if len(m.debugLogs) == 0 {
			s.WriteString(debugTextStyle.Render("No logs yet..."))
		} else {
			s.WriteString(debugTextStyle.Render(strings.Join(m.debugLogs, "\n")))
		}
	}

	return appStyle.Render(s.String())
}

var (
	appStyle      = lipgloss.NewStyle().Padding(1, 2)
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	enabledStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Reverse(true)
	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	nameStyle     = lipgloss.NewStyle().Bold(true)
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	inputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Border(lipgloss.NormalBorder()).Padding(0, 1)
	messageStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	debugTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	debugTextStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("237")).Faint(true)
)
