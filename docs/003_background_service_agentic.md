# ADR 003: Cross-Platform Background Service Implementation

## Status
Proposed

## Context
The current `daemon` implementation runs as a foreground console application. To provide a robust user experience, the application needs to run as a native background service on both Windows (Windows Service) and Linux (systemd).

## Objectives
- Implement a true background service that starts on boot.
- Support Windows and Linux using a unified codebase.
- Provide CLI commands to install, uninstall, start, and stop the service.
- Ensure configuration is passed correctly to the background process.

## Architecture: The Cross-Platform Library Approach
We will use the `github.com/kardianos/service` library.

### Phase 1: Service Wrapper (`internal/daemon`)
Create a `Service` struct that implements `service.Interface`.
- **Start()**: Initializes the API client, loads configuration, and starts `timer.StartDaemon` in a separate goroutine.
- **Stop()**: Cancels the context to signal a graceful shutdown.

### Phase 2: CLI Refactoring (`cmd/service.go`)
Introduce a `service` command to replace or extend the current `daemon` command.
- `nextdns-client service install`: Registers the binary with the OS. Must capture and persist `--api-key`, `--profile-id`, and `--config` path.
- `nextdns-client service uninstall`: Removes the service.
- `nextdns-client service start`: Starts the background service.
- `nextdns-client service stop`: Stops the background service.
- `nextdns-client service run`: Internal command used by the OS to execute the service logic.

### Phase 3: Logging & Persistence
- Use `service.Logger` to route logs to the Windows Event Log or Linux Journald.
- Ensure the service runs with the correct permissions (System/Root) to allow configuration file access.

## Detailed Task List

### 1. Dependency Management
- [ ] Run `go get github.com/kardianos/service`

### 2. Implementation of `internal/daemon/service.go`
- [ ] Define `program` struct to hold API client, config, and context.
- [ ] Implement `Start(s service.Service) error`.
- [ ] Implement `Stop(s service.Service) error`.

### 3. Implementation of `cmd/service.go`
- [ ] Create `serviceCmd` with subcommands.
- [ ] Logic for `install` to pass flags as arguments to the service execution.
- [ ] Logic for `run` to initialize the `service` library and call `s.Run()`.

### 4. Verification
- [ ] **Windows**: Verify service appears in `services.msc` and logs to Event Viewer.
- [ ] **Linux**: Verify service is managed by `systemctl` and logs to `journalctl`.
- [ ] **Functional**: Verify timers still expire and block apps when running as a service.
