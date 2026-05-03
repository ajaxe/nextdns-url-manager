# Phase 3: Logging & Persistence Implementation Plan

## Overview
Phase 3 ensures that the background service is reliable, traceable, and correctly handles system-level constraints. This involves bridging the application's logging to system-specific facilities (Windows Event Log, Linux Journald) and ensuring that file paths and permissions are managed correctly for a service context.

## Implementation Details

### 1. Unified Logging with `slog` Bridge
The application uses `log/slog` for structured logging. When running as a service, we must route these logs to the OS-native logging system.

- **New File**: `internal/daemon/logging.go`
- **Component**: `SlogServiceHandler`
    - Implements `slog.Handler`.
    - Wraps `service.Logger`.
    - Maps `slog.LevelDebug` and `slog.LevelInfo` to `service.Logger.Info`.
    - Maps `slog.LevelWarn` to `service.Logger.Warning`.
    - Maps `slog.LevelError` to `service.Logger.Error`.
- **Integration**: In `internal/daemon/service.go`, after the service starts, initialized the `SlogServiceHandler` and call `slog.SetDefault()`.

### 2. Path Persistence & Normalization
Services often run with a different working directory (e.g., `C:\Windows\System32` on Windows). Relative paths like `config.yaml` or `timer_state.yaml` will fail.

- **Config Path**:
    - During `nextdns-client service install`, the CLI will resolve the absolute path of the provided `--config` flag.
    - This absolute path is persisted in the service `Arguments`.
- **Timer State Path**:
    - Refactor `internal/timer/timer.go` to support a configurable state file path.
    - The state file should default to being in the same directory as the config file if not explicitly provided.
    - Logic: `statePath = filepath.Join(filepath.Dir(configPath), "timer_state.yaml")`.

### 3. Permissions & System Context
- **Default User**: The service will run as `SYSTEM` on Windows and `root` on Linux by default. This ensures it has permission to modify its own configuration and state files.
- **Service Config**: In `internal/daemon/service.go`, the `GetConfig` helper will ensure no specific user is set, allowing the OS default for services to take over.

## Task List

- [ ] **Logging**: (priority: high)
    - [ ] Implement `SlogServiceHandler` in `internal/daemon/logging.go`. (priority: high)
    - [ ] Update `internal/daemon/service.go` to initialize the system logger and set it as default. (priority: high)
- [ ] **Persistence**: (priority: high)
    - [ ] Refactor `internal/timer` to accept `statePath` or derive it from `configPath`. (priority: high)
    - [ ] Update `internal/timer/background.go` to pass the correct path to `LoadState` and `SaveState`. (priority: high)
- [ ] **Verification**: (priority: high)
    - [ ] **Windows**: Check "Event Viewer" -> "Windows Logs" -> "Application" for logs from `NextDNS Client`. (priority: high)
    - [ ] **Linux**: Manual verification by developer (see section below). (priority: high)
    - [ ] **State**: Verify `timer_state.yaml` is created in the correct absolute path even when started as a service. (priority: high)

## Verification Plan

1. **Logging Bridge**: Trigger an error (e.g., invalid API key) and verify it appears as an "Error" entry in the system logs.
2. **Absolute Paths**: Install the service from a different directory than the binary and verify it still finds the config file.
3. **Reboot Persistence**: Start a timer, reboot the machine, and verify the background service resumes the timer (by checking the log or observing the block behavior).
4. **Platform Verification**:
    - **Windows**: Automated/Manual verification using Event Viewer and `services.msc`.
    - **Linux**: **Manual verification by the developer** (see section below).

## Manual Linux Verification (Developer Only)
*Note: This section is for the developer to perform manual validation on a Linux environment. It does not require any action by the agentic AI.*

### 1. Setup & Compilation
If developing on Windows, cross-compile the binary for Linux:
```powershell
$env:GOOS='linux'; $env:GOARCH='amd64'; go build -o nextdns-client-linux ./cmd/main.go
```
Transfer the `nextdns-client-linux` binary to your Linux environment (e.g., via SCP or WSL 2).

### 2. Service Installation
Run the following commands on the Linux system:
```bash
# Install the service (requires sudo/root)
sudo ./nextdns-client-linux service install --api-key "YOUR_KEY" --config "/absolute/path/to/config.yaml"

# Start the service
sudo ./nextdns-client-linux service start
```

### 3. Functional Validation
Verify the service status and logs:
```bash
# Check if the service is running
sudo systemctl status nextdns-client

# View logs in the system journal
journalctl -u nextdns-client -f

# Verify state file creation
ls -l /absolute/path/to/timer_state.yaml
```

### 4. Cleanup
```bash
sudo ./nextdns-client-linux service stop
sudo ./nextdns-client-linux service uninstall
```
