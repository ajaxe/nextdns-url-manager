# Implementation Plan: Add Application Groups via TUI and Daemon Hot-Reload

## Global Context
This document outlines the implementation plan for enhancing the NextDNS client application. The application is a cross-platform Terminal UI (TUI) written in Go using the Bubble Tea framework.

**Objective:**
1. Enable users to create new application groups directly from the TUI.
2. Provide a seamless workflow: after naming a new group, immediately prompt the user to add URLs to it.
3. Ensure the background daemon automatically detects and applies changes made to the configuration file (`config.yaml`) without requiring a restart (hot-reloading).

**Key Files & Architecture:**
*   `internal/tui/tui.go`: Contains the TUI logic (Model, Update, View) built with Bubble Tea.
*   `internal/config/config.go`: Handles loading and saving the YAML configuration.
*   `internal/timer/background.go`: Contains the background daemon logic for monitoring timers and updating NextDNS.

**Data Structures (Reference):**
```go
// internal/config/config.go
type Config struct {
	Applications []Application `yaml:"applications"`
}

type Application struct {
	Name    string   `yaml:"name"`
	URLs    []string `yaml:"urls"`
	Enabled bool     `yaml:"enabled"`
	Timer   string   `yaml:"timer"`
}
```

---

## Phase 1: Implement App Group Input View in TUI

**Goal:** Modify the TUI state machine to support an "add application" view and handle text input for the new application's name.

**Implementation Steps:**
1.  **Update `internal/tui/tui.go` Constants:**
    *   Add a new view constant: `viewAppInput = "app_input"`.
2.  **Update `Model` Struct:**
    *   Add an `appInput string` field to the `Model` struct to store the text typed by the user.
3.  **Update `handleMainView`:**
    *   Add a case for the `"a"` key to trigger adding a new app.
    *   When `"a"` is pressed: set `m.currentView = viewAppInput` and clear `m.appInput = ""`.
4.  **Create `handleAppInput` Function:**
    *   Create a new method `func (m Model) handleAppInput(msg tea.KeyMsg) (tea.Model, tea.Cmd)`.
    *   Handle `"esc"`: return to `viewMain`.
    *   Handle `"backspace"`: remove the last character from `m.appInput`.
    *   Handle default case (typing): append `msg.String()` to `m.appInput` if it's a single character.
    *   Handle `"enter"`:
        *   If `m.appInput` is not empty, append a new `config.Application{Name: m.appInput, URLs: []string{}, Enabled: false}` to `m.config.Applications`.
        *   Call `config.Save(m.config, m.configPath)` to persist the change immediately.
        *   Set `m.activeApp` to the index of the newly added application (`len(m.config.Applications) - 1`).
        *   Set `m.urlInput = ""` and transition `m.currentView = viewUrlInput` to immediately prompt for the first URL.
5.  **Update `Update` Function:**
    *   In the `switch m.currentView` block, add a case for `viewAppInput` to call `m.handleAppInput(msg)`.

**Verification:**
*   Run the TUI. Pressing `a` on the main view should switch to an empty input (though it won't render yet until Phase 2).
*   Typing a name and pressing Enter should theoretically create the app and transition state.

---

## Phase 2: Render App Group Input View and Update Help

**Goal:** Update the TUI rendering logic to display the new input prompt and update the help text on the main view.

**Implementation Steps:**
1.  **Update `View` Function (`internal/tui/tui.go`):**
    *   In the `switch m.currentView` block, add a case for `viewAppInput`.
    *   Render a prompt: `Enter new application group name:\n`
    *   Render the input box using existing styles: `inputStyle.Render(m.appInput + "_")`.
    *   Render help instructions: `\n\n(Enter to save and add URLs, Esc to cancel, Backspace to clear)`.
2.  **Update Main View Help Text:**
    *   In the `default` case of the `View` function, locate the `helpStyle.Render(...)` call at the bottom.
    *   Update the string to include the new action: `↑/↓: navigate • Space: toggle • t: timer • Enter: edit URLs • a: add app • q: quit`.

**Verification:**
*   Run the TUI.
*   The main menu help text should show `a: add app`.
*   Press `a`. You should see the prompt to enter a new application group name.
*   Type a name (e.g., "Work") and press Enter.
*   The UI should immediately switch to the "Add URL for Work" prompt (the existing `viewUrlInput`).
*   Press Esc to return to the URL list, and Esc again to return to the main menu. The new "Work" group should be listed.
*   Check the `config.yaml` file on disk; the new group should be saved.

---

## Phase 3: Implement Daemon Hot-Reloading

**Goal:** Ensure the background daemon automatically detects changes made to `config.yaml` by the TUI without needing a restart.

**Implementation Steps:**
1.  **Update `StartDaemon` Loop (`internal/timer/background.go`):**
    *   Locate the `StartDaemon` function and its `for` loop that runs on a ticker.
    *   Inside the `case <-ticker.C:` block, *before* calling `RunBackgroundCheck`, add logic to reload the configuration file from disk.
    *   Use `config.Load(configPath)`. If successful, update the local `cfg` pointer.
    *   If `config.Load` fails (e.g., due to a temporary file lock during write), log a warning but proceed with the existing `cfg` to maintain stability.

**Code Snippet (for reference):**
```go
// internal/timer/background.go
// Inside StartDaemon for loop:
case <-ticker.C:
    // Hot-reload configuration to detect TUI changes
    if updatedCfg, err := config.Load(configPath); err == nil {
        cfg = updatedCfg
    } else {
        fmt.Printf("Warning: Failed to hot-reload config: %v\n", err)
    }

    if err := RunBackgroundCheck(apiClient, cfg, configPath); err != nil {
        fmt.Printf("Background check error: %v\n", err)
    }
```

**Verification:**
*   Start the daemon in one terminal window (`go run main.go daemon --api-key=... --profile-id=...`).
*   Open the TUI in another terminal window.
*   Add a new application group via the TUI and add a URL to it.
*   Set a short timer (e.g., `1m`) and enable the group in the TUI.
*   Observe the daemon logs. Within 30 seconds (the ticker interval), the daemon should process the timer based on the hot-reloaded configuration, eventually logging that the timer expired and the application group was blocked.
