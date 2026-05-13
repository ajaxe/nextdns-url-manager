package timer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Timer represents a timer configuration
type Timer struct {
	Name      string        `yaml:"name"`
	Duration  time.Duration `yaml:"duration"`
	Start     time.Time     `yaml:"start"`
	Running   bool          `yaml:"running"`
	TargetApp string        `yaml:"target_app"`
}

// ParseTimer parses a timer string into a duration
// Supported formats: "5s", "70s", "1m5s", "1m 5s", "1h5m", "1h 5m"
func ParseTimer(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	// Normalize: remove spaces between numbers and units
	// "1m 5s" -> "1m5s"
	re := regexp.MustCompile(`(\d+)\s+([a-zA-Z]+)`)
	s = re.ReplaceAllString(s, "$1$2")

	// Remove all remaining spaces just in case
	s = strings.ReplaceAll(s, " ", "")

	// time.ParseDuration handles "1h5m", "1m5s", "70s", etc.
	return time.ParseDuration(s)
}

// TimerState represents the state of all active timers
type TimerState struct {
	Timers []Timer `yaml:"timers"`
}

const defaultStateFile = "timer_state.yaml"

var stateDir string

// SetStateDir sets the directory for the timer state file.
func SetStateDir(dir string) {
	stateDir = dir
}

func getStateFilePath() string {
	dir := stateDir
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, defaultStateFile)
}

// SaveState saves the timer state to a file
func SaveState(state *TimerState) error {
	data, err := yaml.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(getStateFilePath(), data, 0644)
}

// LoadState loads the timer state from a file
func LoadState() (*TimerState, error) {
	data, err := os.ReadFile(getStateFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return &TimerState{Timers: []Timer{}}, nil
		}
		return nil, err
	}

	var state TimerState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// AddTimer adds a new timer to the state
func AddTimer(name string, duration time.Duration, targetApp string) error {
	state, err := LoadState()
	if err != nil {
		return err
	}

	t := Timer{
		Name:      name,
		Duration:  duration,
		Start:     time.Now(),
		Running:   true,
		TargetApp: targetApp,
	}

	// Remove existing timer with same name if any
	newTimers := []Timer{}
	for _, t := range state.Timers {
		if t.Name != name {
			newTimers = append(newTimers, t)
		}
	}
	state.Timers = append(newTimers, t)

	return SaveState(state)
}

// GetRemainingTime returns the remaining time for a timer
func (t *Timer) GetRemainingTime() time.Duration {
	if !t.Running {
		return 0
	}
	elapsed := time.Since(t.Start)
	remaining := t.Duration - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}
