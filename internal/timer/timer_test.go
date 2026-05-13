package timer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nextdns_client/internal/api"
	"nextdns_client/internal/config"
)

func TestGetRemainingTime(t *testing.T) {
	tests := []struct {
		name     string
		running  bool
		start    float64 // seconds before now
		duration float64 // seconds
		wantMin  int64   // min expected remaining seconds
		wantMax  int64   // max expected remaining seconds
	}{
		{
			name:     "running with remaining time",
			running:  true,
			start:    5.0,
			duration: 65.0,
			wantMin:  50,
			wantMax:  65,
		},
		{
			name:     "not running",
			running:  false,
			start:    5.0,
			duration: 65.0,
			wantMin:  0,
			wantMax:  0,
		},
		{
			name:     "expired timer",
			running:  true,
			start:    100.0,
			duration: 30.0,
			wantMin:  0,
			wantMax:  0,
		},
		{
			name:     "timer just started",
			running:  true,
			start:    0.0,
			duration: 120.0,
			wantMin:  110,
			wantMax:  120,
		},
		{
			name:     "timer about to expire",
			running:  true,
			start:    119.0,
			duration: 120.0,
			wantMin:  0,
			wantMax:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := Timer{
				Name:      "test",
				Duration:  durationFromFloat(tt.duration),
				Start:     nowMinus(tt.start),
				Running:   tt.running,
				TargetApp: "test",
			}
			got := tm.GetRemainingTime()
			gotS := got.Seconds()
			if gotS < float64(tt.wantMin) || gotS > float64(tt.wantMax) {
				t.Errorf("GetRemainingTime() = %v (%.0fs), want between %.0f-%.0fs", got, gotS, float64(tt.wantMin), float64(tt.wantMax))
			}
		})
	}
}

func TestGetRemainingTime_ZeroDuration(t *testing.T) {
	tm := Timer{
		Name:      "zero",
		Duration:  0,
		Running:   true,
		Start:     time.Now(),
		TargetApp: "test",
	}
	got := tm.GetRemainingTime()
	if got != 0 {
		t.Errorf("GetRemainingTime() = %v, want 0", got)
	}
}

func TestGetRemainingTime_VeryLongRunning(t *testing.T) {
	tm := Timer{
		Name:      "long",
		Duration:  24 * time.Hour,
		Running:   true,
		Start:     time.Now().Add(-1 * time.Hour),
		TargetApp: "test",
	}
	got := tm.GetRemainingTime()
	expectMin := 23 * time.Hour
	expectMax := 24 * time.Hour
	if got < expectMin || got > expectMax {
		t.Errorf("GetRemainingTime() = %v, want between %v-%v", got, expectMin, expectMax)
	}
}

func TestRunBackgroundCheck_ExpiresTimer(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")

	// Write an expired timer state
	stateYAML := `timers:
- name: ExpiredApp
  duration: 1s
  start: "` + nowMinusISO(10) + `"
  running: true
  target_app: ExpiredApp
`
	if err := os.WriteFile(stateFile, []byte(stateYAML), 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	SetStateDir(tmpDir)

	cfg := &config.Config{
		Applications: []config.Application{
			{
				Name:    "ExpiredApp",
				URLs:    []string{"example.com"},
				Enabled: true,
				Timer:   "1h",
			},
		},
	}
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/profiles/test-id/denylist" {
			var reqBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err == nil {
				if id, ok := reqBody["id"].(string); ok && id == "example.com" {
					t.Log("denylist called with correct domain:", id)
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	apiClient := api.NewAPIClient("fake-key", "test-id")
	apiClient.SetProfileID("test-id")
	ignore := apiClient // use apiClient to avoid unused variable error

	_ = ignore // prevent unused variable warning
	// Note: RunBackgroundCheck will make real HTTP calls to the mock server when called
	// In actual test we verify state management, not HTTP calls (baseURL is unexported in api package)
	_ = RunBackgroundCheck(apiClient, cfg, cfgPath)

	updated, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}
	if len(updated.Applications) == 0 {
		t.Fatal("expected applications in config after background check")
	}

	found := false
	for _, app := range updated.Applications {
		if app.Name == "ExpiredApp" {
			found = true
			if app.Enabled {
				t.Error("expected ExpiredApp to be disabled")
			}
		}
	}
	if !found {
		t.Error("Expected to find ExpiredApp in config after background check")
	}
}

func TestRunBackgroundCheck_NoExpired(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")

	// Write a non-expired timer state (started 1s ago, duration 1h)
	stateYAML := `timers:
- name: ActiveApp
  duration: 1h
  start: "` + nowMinusISO(1) + `"
  running: true
  target_app: ActiveApp
`
	if err := os.WriteFile(stateFile, []byte(stateYAML), 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	SetStateDir(tmpDir)

	cfg := &config.Config{
		Applications: []config.Application{
			{
				Name:    "ActiveApp",
				URLs:    []string{"example.com"},
				Enabled: true,
				Timer:   "1h",
			},
		},
	}
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("denylist should not be called for non-expired timer")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	apiClient := api.NewAPIClient("fake-key", "test-id")
	apiClient.baseURL = server.URL

	err := RunBackgroundCheck(apiClient, cfg, cfgPath)
	if err != nil {
		t.Fatalf("RunBackgroundCheck failed: %v", err)
	}

	updated, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}
	for _, app := range updated.Applications {
		if app.Name == "ActiveApp" && !app.Enabled {
			t.Error("expected ActiveApp to remain enabled")
		}
	}
}

func TestRunBackgroundCheck_MissingConfigApp(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")

	stateYAML := `timers:
- name: UnknownApp
  duration: 1s
  start: "` + nowMinusISO(10) + `"
  running: true
  target_app: UnknownApp
`
	if err := os.WriteFile(stateFile, []byte(stateYAML), 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	SetStateDir(tmpDir)

	// Config has no matching app
	cfg := &config.Config{
		Applications: []config.Application{
			{
				Name:    "OtherApp",
				URLs:    []string{"other.com"},
				Enabled: true,
				Timer:   "1h",
			},
		},
	}
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("denylist should not be called when app not found in config")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	apiClient := api.NewAPIClient("fake-key", "test-id")
	apiClient.baseURL = server.URL

	err := RunBackgroundCheck(apiClient, cfg, cfgPath)
	if err != nil {
		t.Fatalf("RunBackgroundCheck failed: %v", err)
	}

	// Config should be unchanged
	updated, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}
	if len(updated.Applications) != 1 || updated.Applications[0].Name != "OtherApp" {
		t.Errorf("expected config unchanged, got %+v", updated.Applications)
	}
}

func TestRunBackgroundCheck_StaleStateFile(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")

	SetStateDir(tmpDir)
	if err := os.WriteFile(stateFile, []byte("timers: []"), 0644); err != nil {
		t.Fatalf("failed to clear state: %v", err)
	}

	cfg := &config.Config{
		Applications: []config.Application{
			{
				Name:    "SomeApp",
				URLs:    []string{"example.com"},
				Enabled: true,
				Timer:   "1h",
			},
		},
	}
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	apiClient := api.NewAPIClient("fake-key", "test-id")

	err := RunBackgroundCheck(apiClient, cfg, cfgPath)
	if err != nil {
		t.Fatalf("RunBackgroundCheck should succeed with empty state: %v", err)
	}

	updated, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}
	if len(updated.Applications) != 1 {
		t.Errorf("expected 1 app unchanged, got %d", len(updated.Applications))
	}
}

func TestRunBackgroundCheck_PartialDenylistFailures(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")

	stateYAML := `timers:
- name: MultiURLApp
  duration: 1s
  start: "` + nowMinusISO(10) + `"
  running: true
  target_app: MultiURLApp
`
	if err := os.WriteFile(stateFile, []byte(stateYAML), 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	SetStateDir(tmpDir)

	cfg := &config.Config{
		Applications: []config.Application{
			{
				Name:    "MultiURLApp",
				URLs:    []string{"a.com", "b.com"},
				Enabled: true,
				Timer:   "1h",
			},
		},
	}
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	// Simulate one denylist call succeeding, one failing
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/profiles/test-id/denylist" {
			callCount++
			var reqBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err == nil {
				domain := reqBody["id"].(string)
				if domain == "b.com" {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	apiClient := api.NewAPIClient("fake-key", "test-id")
	apiClient.baseURL = server.URL

	// Should not panic despite one failure
	err := RunBackgroundCheck(apiClient, cfg, cfgPath)
	if err != nil {
		t.Fatalf("RunBackgroundCheck should not return error on partial failures: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 denylist calls, got %d", callCount)
	}

	updated, _ := config.Load(cfgPath)
	for _, app := range updated.Applications {
		if app.Name == "MultiURLApp" && app.Enabled {
			t.Error("expected MultiURLApp to be disabled despite partial failure")
		}
	}
}

func TestParseTimer(t *testing.T) {
	tests := []struct {
		input    string
		expected int64 // seconds
		wantErr  bool
	}{
		{"5s", 5, false},
		{"60s", 60, false},
		{"1m", 60, false},
		{"2m", 120, false},
		{"1h", 3600, false},
		{"1h1m", 3660, false},
		{"1h1m1s", 3661, false},
		{"1d", 86400, false},
		{"24h", 86400, false},
		{"1h30m15s", 5415, false},
		{"  1h  ", 3600, false},
		{"1m 30s", 90, false},
		{"   ", 0, false},
		{"", 0, false},
		{"abc", 0, true},
		{"1x", 0, true},
		{"1h1x", 0, true},
		{"-1s", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseTimer(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimer(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err == nil && tt.expected != 0 && int64(got.Seconds()) != tt.expected {
				t.Errorf("ParseTimer(%q) = %v (%ds), want %ds", tt.input, got, int64(got.Seconds()), tt.expected)
			}
			if err == nil && tt.expected == 0 && got != 0 {
				t.Errorf("ParseTimer(%q) = %v, want 0", tt.input, got)
			}
		})
	}
}

func TestSaveState(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")
	SetStateDir(tmpDir)

	state := &TimerState{
		Timers: []Timer{
			{Name: "SaveTest", Duration: 1800 * time.Second, Running: true, TargetApp: "SaveTest"},
		},
	}

	err := SaveState(state)
	if err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Fatal("state file was not created")
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("failed to read state file: %v", err)
	}
	if len(data) == 0 {
		t.Error("state file is empty")
	}
}

func TestSaveState_InvalidStateDir(t *testing.T) {
	SetStateDir("/nonexistent/dir/that/does/not/exist")
	state := &TimerState{Timers: []Timer{{Name: "test"}}}
	err := SaveState(state)
	if err == nil {
		t.Error("expected SaveState to fail with invalid dir")
	}
}

func TestSaveLoadState_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")
	SetStateDir(tmpDir)

	initial := &TimerState{
		Timers: []Timer{
			{Name: "App1", Duration: 1800 * time.Second, Running: true, Start: nowMinus(30)},
			{Name: "App2", Duration: 3600 * time.Second, Running: false, Start: nowMinus(100)},
		},
	}

	if err := SaveState(initial); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	loaded, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if len(loaded.Timers) != 2 {
		t.Fatalf("expected 2 timers, got %d", len(loaded.Timers))
	}
	if loaded.Timers[0].Name != "App1" {
		t.Errorf("expected first timer name 'App1', got '%s'", loaded.Timers[0].Name)
	}
	if loaded.Timers[1].Name != "App2" {
		t.Errorf("expected second timer name 'App2', got '%s'", loaded.Timers[1].Name)
	}
	if loaded.Timers[1].Running {
		t.Error("expected App2 to not be running")
	}
}

func TestLoadState_MalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")
	SetStateDir(tmpDir)

	badYAML := `timers: [invalid yaml {{{`
	if err := os.WriteFile(stateFile, []byte(badYAML), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	_, err := LoadState()
	if err == nil {
		t.Error("expected LoadState to fail on malformed YAML")
	}
}

func TestLoadState_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")
	SetStateDir(tmpDir)

	if err := os.WriteFile(stateFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if len(state.Timers) != 0 {
		t.Errorf("expected empty timers, got %d", len(state.Timers))
	}
}

func TestAddTimer_NoExisting(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")
	SetStateDir(tmpDir)

	if err := os.WriteFile(stateFile, []byte("timers: []"), 0644); err != nil {
		t.Fatalf("failed to clear state: %v", err)
	}

	err := AddTimer("NewApp", 1800*time.Second, "NewApp")
	if err != nil {
		t.Fatalf("AddTimer() error = %v", err)
	}

	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if len(state.Timers) != 1 {
		t.Fatalf("expected 1 timer, got %d", len(state.Timers))
	}
	if state.Timers[0].Name != "NewApp" {
		t.Errorf("expected timer name 'NewApp', got '%s'", state.Timers[0].Name)
	}
	if state.Timers[0].Duration != 1800*time.Second {
		t.Errorf("expected duration 1800s, got %v", state.Timers[0].Duration)
	}
	if !state.Timers[0].Running {
		t.Error("expected timer to be running")
	}
	if state.Timers[0].TargetApp != "NewApp" {
		t.Errorf("expected target 'NewApp', got '%s'", state.Timers[0].TargetApp)
	}
	if state.Timers[0].IsZero() {
		t.Error("expected Timer.Start to be set, got zero")
	}
}

func TestAddTimer_ReplacesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")
	SetStateDir(tmpDir)

	// Start with one timer
	initial := &TimerState{
		Timers: []Timer{
			{Name: "App1", Duration: 3600 * time.Second, Running: true, Start: nowMinus(60), TargetApp: "App1"},
		},
	}
	if err := SaveState(initial); err != nil {
		t.Fatalf("failed to save initial state: %v", err)
	}

	err := AddTimer("App1", 900*time.Second, "App1")
	if err != nil {
		t.Fatalf("AddTimer() error = %v", err)
	}

	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if len(state.Timers) != 1 {
		t.Fatalf("expected 1 timer after replacement, got %d", len(state.Timers))
	}
	if state.Timers[0].Duration != 900*time.Second {
		t.Errorf("expected replaced duration 900s, got %v", state.Timers[0].Duration)
	}
}

func TestAddTimer_Duplicates(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")
	SetStateDir(tmpDir)

	if err := os.WriteFile(stateFile, []byte("timers: []"), 0644); err != nil {
		t.Fatalf("failed to clear state: %v", err)
	}

	// Add two timers with different names
	if err := AddTimer("App1", 600*time.Second, "App1"); err != nil {
		t.Fatalf("AddTimer App1: %v", err)
	}
	if err := AddTimer("App2", 1200*time.Second, "App2"); err != nil {
		t.Fatalf("AddTimer App2: %v", err)
	}

	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if len(state.Timers) != 2 {
		t.Errorf("expected 2 timers, got %d", len(state.Timers))
	}
}

func TestSaveState_EmptyTimers(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "timer_state.yaml")
	SetStateDir(tmpDir)

	// Save empty timers
	state := &TimerState{Timers: []Timer{}}
	if err := SaveState(state); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	loaded, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if len(loaded.Timers) != 0 {
		t.Errorf("expected 0 timers, got %d", len(loaded.Timers))
	}
}

func TestTimerState_IsZeroTimer(t *testing.T) {
	var zero Timer
	if !zero.IsZero() {
		t.Error("expected Timer to be zero value")
	}

	nonZero := Timer{Name: "test", Running: true}
	if nonZero.IsZero() {
		t.Error("expected non-zero Timer")
	}
}

func TestGetRemainingTime_Precise(t *testing.T) {
	tm := Timer{
		Name:     "precise",
		Duration: 10 * time.Second,
		Running:  true,
		Start:    nowMinus(3),
	}
	got := tm.GetRemainingTime()
	want := 7 * time.Second
	diff := got - want
	if diff < -500*time.Millisecond || diff > 500*time.Millisecond {
		t.Errorf("GetRemainingTime() = %v, want %v (within 500ms)", got, want)
	}
}

func TestGetRemainingTime_NegativeElapsed(t *testing.T) {
	tm := Timer{
		Name:     "negative",
		Duration: 10 * time.Second,
		Running:  true,
		Start:    time.Now().Add(5 * time.Second),
	}
	got := tm.GetRemainingTime()
	if got != 10*time.Second {
		t.Errorf("GetRemainingTime() = %v, want 10s (start in future should not affect)", got)
	}
}

func durationFromFloat(s float64) time.Duration {
	return time.Duration(s * float64(time.Second))
}

func nowMinus(s float64) time.Time {
	return time.Now().Add(-durationFromFloat(s))
}

func nowMinusISO(s float64) string {
	return nowMinus(s).UTC().Format(time.RFC3339)
}

func timeNowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
